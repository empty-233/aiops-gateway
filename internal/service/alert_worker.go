package service

import (
	"context"
	"time"

	"aiops-gateway/internal/model"
)

func (s *AlertService) StartWorkers(ctx context.Context) {
	workerCount := s.aiConfig.WorkerCount
	for i := 0; i < workerCount; i++ {
		workerID := i + 1

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.worker(ctx, workerID)
		}()
	}

	s.logger.Info("AlertService 后台 worker 已启动", "workers", workerCount)
}

func (s *AlertService) Wait() {
	s.wg.Wait()
}

func (s *AlertService) worker(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("AlertService worker 停止", "worker", workerID)
			return

		case payload := <-s.queue:
			s.processAsync(ctx, workerID, &payload)
		}
	}
}

func (s *AlertService) processAsync(parentCtx context.Context, workerID int, payload *model.AlertPayload) {
	timeout := s.aiConfig.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	s.logger.Info(
		"开始后台处理告警",
		"worker", workerID,
		"count", len(payload.Alerts),
		"status", payload.Status,
	)

	if _, err := s.processQueuedPayload(ctx, payload); err != nil {
		s.logger.Error(
			"后台处理告警失败",
			"worker", workerID,
			"error", err,
		)
		return
	}

	s.logger.Info(
		"后台处理告警完成",
		"worker", workerID,
		"count", len(payload.Alerts),
		"status", payload.Status,
	)
}