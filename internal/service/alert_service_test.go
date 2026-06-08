package service_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"aiops-gateway/internal/config"
	"aiops-gateway/internal/model"
	"aiops-gateway/internal/service"

	amtemplate "github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func newPayload(alertName, severity string) *model.AlertPayload {
	alert := amtemplate.Alert{
		Status:      "firing",
		Labels:      map[string]string{"alertname": alertName, "severity": severity},
		Annotations: map[string]string{"summary": "测试告警"},
		Fingerprint: "fp-001",
	}

	return &model.AlertPayload{
		Data: &amtemplate.Data{
			Status: "firing",
			Alerts: amtemplate.Alerts{alert},
		},
	}
}

func newMultiPayload() *model.AlertPayload {
	alerts := amtemplate.Alerts{
		{Labels: map[string]string{"alertname": "A1"}, Annotations: map[string]string{}},
		{Labels: map[string]string{"alertname": "A2"}, Annotations: map[string]string{}},
		{Labels: map[string]string{"alertname": "A3"}, Annotations: map[string]string{}},
	}

	return &model.AlertPayload{
		Data: &amtemplate.Data{
			Status: "firing",
			Alerts: alerts,
		},
	}
}

// TestProcess_Success 正常流程：告警进入队列并返回结果
func TestProcess_Success(t *testing.T) {
	svc := service.NewAlertService(nil, nil, newTestLogger(), nil, config.AIConfig{QueueSize: 1})
	result, err := svc.Process(context.Background(), newPayload("HighCPU", "critical"))

	assert.NoError(t, err)
	assert.Equal(t, 1, result.AlertCount)
	assert.Equal(t, "firing", result.Status)
	assert.Empty(t, result.Analysis)
}

// TestProcess_QueueFull 队列满时返回降级结果
func TestProcess_QueueFull(t *testing.T) {
	svc := service.NewAlertService(nil, nil, newTestLogger(), nil, config.AIConfig{QueueSize: 0})
	result, err := svc.Process(context.Background(), newPayload("HighCPU", "critical"))

	assert.NoError(t, err)
	assert.Equal(t, 1, result.AlertCount)
	assert.Equal(t, "firing", result.Status)
	assert.Equal(t, "analysis queue is full", result.Analysis)
}

// TestProcess_MultipleAlerts 多条告警时返回正确计数
func TestProcess_MultipleAlerts(t *testing.T) {
	svc := service.NewAlertService(nil, nil, newTestLogger(), nil, config.AIConfig{QueueSize: 1})
	result, err := svc.Process(context.Background(), newMultiPayload())

	assert.NoError(t, err)
	assert.Equal(t, 3, result.AlertCount)
	assert.Equal(t, "firing", result.Status)
}

func TestProcess_NilPayload(t *testing.T) {
	svc := service.NewAlertService(nil, nil, newTestLogger(), nil, config.AIConfig{QueueSize: 1})
	_, err := svc.Process(context.Background(), nil)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "告警 payload 为空")
}

func TestProcess_NilData(t *testing.T) {
	svc := service.NewAlertService(nil, nil, newTestLogger(), nil, config.AIConfig{QueueSize: 1})
	_, err := svc.Process(context.Background(), &model.AlertPayload{})

	assert.Error(t, err)
	assert.ErrorContains(t, err, "告警 payload data 为空")
}
