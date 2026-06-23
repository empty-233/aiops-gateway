package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"aiops-gateway/internal/config"
	"aiops-gateway/internal/model"
	"aiops-gateway/internal/notify"
)

type AlertService struct {
	analyzer     Analyzer
	alertRepo    AlertRepository
	logger       *slog.Logger
	notifyClient *notify.Service
	aiConfig     config.AIConfig

	queue chan model.AlertPayload
	wg    sync.WaitGroup
}

func NewAlertService(analyzer Analyzer, alertRepo AlertRepository, logger *slog.Logger, notifyClient *notify.Service, aiConfig config.AIConfig) *AlertService {
	return &AlertService{
		analyzer:     analyzer,
		alertRepo:    alertRepo,
		logger:       logger,
		notifyClient: notifyClient,
		aiConfig:     aiConfig,
		queue:        make(chan model.AlertPayload, aiConfig.QueueSize),
	}
}

func (s *AlertService) Process(ctx context.Context, payload *model.AlertPayload) (*model.AlertResult, error) {
	if payload == nil {
		return nil, fmt.Errorf("告警 payload 为空")
	}
	if payload.Data == nil {
		return nil, fmt.Errorf("告警 payload data 为空")
	}

	s.logger.Info("接收到告警，准备入队", "count", len(payload.Alerts), "status", payload.Status)

	copied, err := cloneAlertPayload(payload)
	if err != nil {
		return nil, fmt.Errorf("复制告警 payload 失败: %w", err)
	}

	select {
	case s.queue <- *copied:
		s.logger.Info("告警已入队，后台处理", "count", len(payload.Alerts), "status", payload.Status)

		return &model.AlertResult{
			AlertCount: len(payload.Alerts),
			Status:     payload.Status,
			Analysis:   "",
			Alerts:     payload.Alerts,
		}, nil

	default:
		s.logger.Error("告警处理队列已满，告警未入队", "count", len(payload.Alerts), "status", payload.Status)

		return &model.AlertResult{
			AlertCount: len(payload.Alerts),
			Status:     payload.Status,
			Analysis:   "analysis queue is full",
			Alerts:     payload.Alerts,
		}, nil
	}
}

func (s *AlertService) processQueuedPayload(ctx context.Context, payload *model.AlertPayload) (*model.AlertResult, error) {
	s.logger.Info("开始后台处理告警", "count", len(payload.Alerts), "status", payload.Status)

	analysis, err := s.analyzer.Analyze(ctx, payload)
	if err != nil {
		s.logger.Error("AI 分析失败", "error", err)
		return nil, fmt.Errorf("AI 分析失败: %w", err)
	}

	var analysisStr string
	if analysis != nil {
		if b, err := json.Marshal(analysis); err != nil {
			s.logger.Error("序列化 analysis 失败", "error", err)
			analysisStr = ""
		} else {
			analysisStr = string(b)
		}
	}

	// fmt.Printf("AI 分析结果: %s\n", analysisStr)

	if analysis.ShouldNotify {
		s.notifyClient.Notify(ctx, notify.Message{
			Title: analysis.NotifyTitle,
			Body:  analysis.NotifyContent,
		})
	}

	for _, alert := range payload.Alerts {
		record := &model.AlertRecord{
			Status:         alert.Status,
			AlertName:      alert.Labels["alertname"],
			Severity:       alert.Labels["severity"],
			Instance:       alert.Labels["instance"],
			Summary:        alert.Annotations["summary"],
			Description:    alert.Annotations["description"],
			Fingerprint:    alert.Fingerprint,
			AnalysisResult: analysisStr,
			StartsAt:       alert.StartsAt,
		}
		if err := s.alertRepo.Save(ctx, record); err != nil {
			s.logger.Error("存储告警记录失败", "error", err)
		}
	}

	s.logger.Info("处理完成", "analysis", analysis)
	return &model.AlertResult{
		AlertCount: len(payload.Alerts),
		Status:     payload.Status,
		Analysis:   analysisStr,
		Alerts:     payload.Alerts,
	}, nil
}

func cloneAlertPayload(payload *model.AlertPayload) (*model.AlertPayload, error) {
	if payload == nil {
		return nil, fmt.Errorf("告警 payload 为空")
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var copied model.AlertPayload
	if err := json.Unmarshal(b, &copied); err != nil {
		return nil, err
	}

	return &copied, nil
}
