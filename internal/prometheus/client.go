package prometheus

import (
	"context"
	"fmt"
	"time"

	"aiops-gateway/internal/config"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
)

type PrometheusClient struct {
	client      v1.API
	queryConfig config.QueryConfig
}

func NewPrometheusClient(address string, queryConfig config.QueryConfig) (*PrometheusClient, error) {
	cfg := api.Config{
		Address: address,
	}
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &PrometheusClient{
		client:      v1.NewAPI(client),
		queryConfig: queryConfig,
	}, nil
}

func (p *PrometheusClient) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]MetricSeries, error) {
	if query == "" {
		return nil, fmt.Errorf("查询条件为空")
	}

	if end.IsZero() {
		end = time.Now()
	}
	if step <= 0 {
		step = p.queryConfig.Step
	}
	if step <= 0 {
		step = time.Minute
	}
	if start.IsZero() {
		rangeTime := p.queryConfig.RangeTime
		if rangeTime <= 0 {
			rangeTime = time.Hour
		}
		start = end.Add(-rangeTime)
	}

	r := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	value, _, err := p.client.QueryRange(ctx, query, r)
	if err != nil {
		return nil, err
	}

	series := make([]MetricSeries, 0)
	switch result := value.(type) {
	case prommodel.Matrix:
		for _, stream := range result {
			metricSeries := MetricSeries{
				Labels: make(map[string]string, len(stream.Metric)),
				Points: make([]MetricPoint, 0, len(stream.Values)),
			}
			for labelName, labelValue := range stream.Metric {
				metricSeries.Labels[string(labelName)] = string(labelValue)
			}
			for _, sample := range stream.Values {
				metricSeries.Points = append(metricSeries.Points, MetricPoint{
					Timestamp: sample.Timestamp.Time(),
					Value:     float64(sample.Value),
				})
			}
			series = append(series, metricSeries)
		}
	case prommodel.Vector:
		for _, sample := range result {
			metricSeries := MetricSeries{
				Labels: make(map[string]string, len(sample.Metric)),
				Points: []MetricPoint{{
					Timestamp: sample.Timestamp.Time(),
					Value:     float64(sample.Value),
				}},
			}
			for labelName, labelValue := range sample.Metric {
				metricSeries.Labels[string(labelName)] = string(labelValue)
			}
			series = append(series, metricSeries)
		}
	default:
		return nil, fmt.Errorf("不支持的 prometheus 结果类型 %T", value)
	}

	return series, nil
}
