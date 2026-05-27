package service_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	amtemplate "github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"aiops-gateway/internal/mocks"
	"aiops-gateway/internal/model"
	"aiops-gateway/internal/service"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func newPayload(alertName, severity string) *model.AlertPayload {
	return &model.AlertPayload{
		Data: &amtemplate.Data{
			Status: "firing",
			Alerts: amtemplate.Alerts{
				{
					Status:      "firing",
					Labels:      map[string]string{"alertname": alertName, "severity": severity},
					Annotations: map[string]string{"summary": "测试告警"},
					Fingerprint: "fp-001",
				},
			},
		},
	}
}

// TestProcess_Success 正常流程：AI 分析成功，存库成功
func TestProcess_Success(t *testing.T) {
	analyzer := new(mocks.MockAnalyzer)
	alertRepo := new(mocks.MockAlertRepository)

	analyzer.On("Analyze", mock.Anything, mock.Anything).Return("建议扩容", nil)
	alertRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewAlertService(analyzer, alertRepo, newTestLogger())
	result, err := svc.Process(context.Background(), newPayload("HighCPU", "critical"))

	assert.NoError(t, err)
	assert.Equal(t, 1, result.AlertCount)
	assert.Equal(t, "建议扩容", result.Analysis)
	assert.Equal(t, "firing", result.Status)

	analyzer.AssertExpectations(t)
	alertRepo.AssertExpectations(t)
}

// TestProcess_AnalyzerError AI 接口报错，应返回错误，不调用 Save
func TestProcess_AnalyzerError(t *testing.T) {
	analyzer := new(mocks.MockAnalyzer)
	alertRepo := new(mocks.MockAlertRepository)

	analyzer.On("Analyze", mock.Anything, mock.Anything).Return("", errors.New("AI 服务不可用"))

	svc := service.NewAlertService(analyzer, alertRepo, newTestLogger())
	_, err := svc.Process(context.Background(), newPayload("HighCPU", "critical"))

	assert.Error(t, err)
	assert.ErrorContains(t, err, "AI 分析失败")

	analyzer.AssertExpectations(t)
	alertRepo.AssertExpectations(t) // Save 没有被调用
}

// TestProcess_SaveError 存库失败，主流程不中断
func TestProcess_SaveError(t *testing.T) {
	analyzer := new(mocks.MockAnalyzer)
	alertRepo := new(mocks.MockAlertRepository)

	analyzer.On("Analyze", mock.Anything, mock.Anything).Return("分析完成", nil)
	alertRepo.On("Save", mock.Anything, mock.Anything).Return(errors.New("数据库连接断开"))

	svc := service.NewAlertService(analyzer, alertRepo, newTestLogger())
	result, err := svc.Process(context.Background(), newPayload("HighDisk", "warning"))

	assert.NoError(t, err)
	assert.Equal(t, "分析完成", result.Analysis)

	analyzer.AssertExpectations(t)
	alertRepo.AssertExpectations(t)
}

// TestProcess_MultipleAlerts 多条告警，Save 调用次数与告警数一致
func TestProcess_MultipleAlerts(t *testing.T) {
	analyzer := new(mocks.MockAnalyzer)
	alertRepo := new(mocks.MockAlertRepository)

	analyzer.On("Analyze", mock.Anything, mock.Anything).Return("批量分析完成", nil)
	alertRepo.On("Save", mock.Anything, mock.Anything).Return(nil).Times(3)

	svc := service.NewAlertService(analyzer, alertRepo, newTestLogger())

	payload := &model.AlertPayload{
		Data: &amtemplate.Data{
			Status: "firing",
			Alerts: amtemplate.Alerts{
				{Labels: map[string]string{"alertname": "A1"}, Annotations: map[string]string{}},
				{Labels: map[string]string{"alertname": "A2"}, Annotations: map[string]string{}},
				{Labels: map[string]string{"alertname": "A3"}, Annotations: map[string]string{}},
			},
		},
	}

	result, err := svc.Process(context.Background(), payload)

	assert.NoError(t, err)
	assert.Equal(t, 3, result.AlertCount)

	analyzer.AssertNumberOfCalls(t, "Analyze", 1)
	alertRepo.AssertNumberOfCalls(t, "Save", 3)
}

// TestProcess_VerifyAnalyzerInput 验证传给 AI 的参数格式正确
func TestProcess_VerifyAnalyzerInput(t *testing.T) {
	analyzer := new(mocks.MockAnalyzer)
	alertRepo := new(mocks.MockAlertRepository)

	analyzer.On("Analyze", mock.Anything,
		mock.MatchedBy(func(p *model.AlertPayload) bool {
			return p.Status == "firing" && len(p.Alerts) == 1
		}),
	).Return("ok", nil)
	alertRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewAlertService(analyzer, alertRepo, newTestLogger())
	_, err := svc.Process(context.Background(), newPayload("HighCPU", "critical"))

	assert.NoError(t, err)
	analyzer.AssertExpectations(t)
}
