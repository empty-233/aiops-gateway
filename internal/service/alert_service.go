package service

import (
	"context"
	"fmt"
	"log/slog"

	"aiops-gateway/internal/model"
)

type AlertService struct {
	analyzer  Analyzer
	alertRepo AlertRepository
	logger    *slog.Logger
}

func NewAlertService(analyzer Analyzer, alertRepo AlertRepository, logger *slog.Logger) *AlertService {
	return &AlertService{
		analyzer:  analyzer,
		alertRepo: alertRepo,
		logger:    logger,
	}
}

func (s *AlertService) Process(ctx context.Context, payload *model.AlertPayload) (*model.AlertResult, error) {
	if payload == nil {
		return nil, fmt.Errorf("告警 payload 为空")
	}
	if payload.Data == nil {
		return nil, fmt.Errorf("告警 payload data 为空")
	}

	s.logger.Info("处理告警", "count", len(payload.Alerts), "status", payload.Status)

	analysis, err := s.analyzer.Analyze(ctx, payload)
	if err != nil {
		s.logger.Error("AI 分析失败", "error", err)
		return nil, fmt.Errorf("AI 分析失败: %w", err)
	}

	for _, alert := range payload.Alerts {
		record := &model.AlertRecord{
			Status:      alert.Status,
			AlertName:   alert.Labels["alertname"],
			Severity:    alert.Labels["severity"],
			Instance:    alert.Labels["instance"],
			Summary:     alert.Annotations["summary"],
			Description: alert.Annotations["description"],
			Fingerprint: alert.Fingerprint,
			Analysis:    analysis,
			StartsAt:    alert.StartsAt.String(),
		}
		if err := s.alertRepo.Save(ctx, record); err != nil {
			s.logger.Error("存储告警记录失败", "error", err)
		}
	}

	s.logger.Info("处理完成", "analysis", analysis)
	return &model.AlertResult{
		AlertCount: len(payload.Alerts),
		Status:     payload.Status,
		Analysis:   analysis,
		Alerts:     payload.Alerts,
	}, nil
}
