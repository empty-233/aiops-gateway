package service

import (
	"context"
	"strings"
	"testing"

	"aiops-gateway/internal/config"
	"aiops-gateway/internal/model"

	amtemplate "github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
)

func buildTestAlertPayload() *model.AlertPayload {
	return buildTestAlertPayloadWithGeneratorURL("")
}

func buildTestAlertPayloadWithGeneratorURL(generatorURL string) *model.AlertPayload {
	return &model.AlertPayload{
		Data: &amtemplate.Data{
			Status: "firing",
			Alerts: amtemplate.Alerts{
				{
					Status: "firing",
					Labels: map[string]string{
						"alertname": "HighCPU",
						"severity":  "critical",
						"instance":  "node-1",
					},
					Annotations: map[string]string{
						"summary":     "CPU 使用率过高",
						"description": "CPU 持续 5 分钟超过 90%",
					},
					GeneratorURL: generatorURL,
				},
			},
		},
	}
}

// TestAIAnalyzerAnalyzeSuccess 正常流程：仅验证 Analyze 可以返回提示内容
func TestAIAnalyzerAnalyzeSuccess(t *testing.T) {
	analyzer := NewAIAnalyzer(config.AIConfig{
		APIKey:  "test-key",
		BaseURL: "http://example.com",
		Model:   "test-model",
	}, config.QueryConfig{}, config.LogsConfig{}, nil)
	result, err := analyzer.Analyze(context.Background(), buildTestAlertPayload())

	assert.NoError(t, err)
	assert.Contains(t, result, "[待接入 AI] 收到 prompt")
	assert.Contains(t, result, "prompt 预览：")
}

// TestAIAnalyzerBuildPrompt_SkipWithoutGeneratorURL generatorURL 为空时跳过 PromQL 查询
func TestAIAnalyzerBuildPrompt_SkipWithoutGeneratorURL(t *testing.T) {
	analyzer := NewAIAnalyzer(config.AIConfig{}, config.QueryConfig{}, config.LogsConfig{}, nil)
	prompt, err := analyzer.buildPrompt(context.Background(), buildTestAlertPayload())

	assert.NoError(t, err)
	assert.Contains(t, prompt, "历史告警：")
	assert.NotContains(t, prompt, "PromQL：")
	assert.Contains(t, prompt, "请给出：1. 问题判断 2. 可能原因 3. 处理建议")
}

// TestAIAnalyzerBuildPrompt_InvalidPromQL PromQL 为空时返回错误
func TestAIAnalyzerBuildPrompt_InvalidPromQL(t *testing.T) {
	analyzer := NewAIAnalyzer(config.AIConfig{}, config.QueryConfig{}, config.LogsConfig{}, nil)
	payload := buildTestAlertPayloadWithGeneratorURL("http://prometheus:9090/graph?g0.expr=&g0.tab=1")

	_, err := analyzer.buildPrompt(context.Background(), payload)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "提取expr失败")
}

func TestAIAnalyzerCallPreviewMax50Chars(t *testing.T) {
	analyzer := NewAIAnalyzer(config.AIConfig{}, config.QueryConfig{}, config.LogsConfig{}, nil)
	prompt := strings.Repeat("a", 60)

	result, err := analyzer.call(context.Background(), prompt)

	assert.NoError(t, err)
	assert.Contains(t, result, "长度 60 字符")
	assert.Contains(t, result, "prompt 预览："+strings.Repeat("a", 50)+"...")
}

func TestTruncateRunesUsesRuneCount(t *testing.T) {
	assert.Equal(t, "你好", truncateRunes("你好世界", 2))
	assert.Equal(t, "你", truncateRunes("你好", 1))
}
