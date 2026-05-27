package service

import (
	"context"
	"time"

	"aiops-gateway/internal/model"
	"aiops-gateway/internal/prometheus"
)

type Analyzer interface {
	Analyze(ctx context.Context, payload *model.AlertPayload) (string, error)
}

type AlertRepository interface {
	Save(ctx context.Context, record *model.AlertRecord) error
	FindByID(ctx context.Context, id uint) (*model.AlertRecord, error)
	FindByAlertName(ctx context.Context, alertName string) ([]model.AlertRecord, error)
	LatestList(ctx context.Context, x int, sortBy, order string) ([]model.AlertRecord, error)
	List(ctx context.Context, page, pageSize int) ([]model.AlertRecord, int64, error)
}

type PrometheusQuerier interface {
	QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]prometheus.MetricSeries, error)
}
