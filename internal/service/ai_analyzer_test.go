package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"aiops-gateway/internal/config"
	"aiops-gateway/internal/llm"
	"aiops-gateway/internal/model"

	amtemplate "github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
)

type testLLMClient struct {
	lastMessages []llm.Message
	lastSchema   any
	response     string
	raw          string
	err          error
}

func (c *testLLMClient) RawChatJSON(ctx context.Context, msg []llm.Message, schema any) (string, string, error) {
	c.lastMessages = append([]llm.Message(nil), msg...)
	c.lastSchema = schema

	if c.err != nil {
		return "", "", c.err
	}

	if c.response == "" {
		c.response = `{"should_notify":true,"severity":"critical","confidence":0.93,"summary":"ok","reason":"test","evidence":["e1"],"suggestions":["s1"],"missing_context":[],"tags":["test"]}`
	}
	if c.raw == "" {
		c.raw = c.response
	}

	return c.response, c.raw, nil
}

func newTestAnalyzerClient(result model.AnalysisResult) *testLLMClient {
	content, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	encoded := string(content)
	return &testLLMClient{response: encoded, raw: encoded}
}

func buildTestAlertPayload() *model.AlertPayload {
	return buildTestAlertPayloadWithGeneratorURL("")
}

func buildTestAlertPayloadWithGeneratorURL(generatorURL string) *model.AlertPayload {
	alert := amtemplate.Alert{
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
	}

	return &model.AlertPayload{
		Data: &amtemplate.Data{
			Status: "firing",
			Alerts: amtemplate.Alerts{alert},
		},
	}
}

// TestAIAnalyzerAnalyzeSuccess 正常流程：Analyze 能解析 LLM 返回值并携带提示词
func TestAIAnalyzerAnalyzeSuccess(t *testing.T) {
	client := newTestAnalyzerClient(model.AnalysisResult{
		ShouldNotify: true,
		Severity:     "critical",
		Confidence:   0.93,
		Summary:      "CPU 使用率过高",
		Reason:       "test",
		Evidence:     []string{"e1"},
		Suggestions:  []string{"s1"},
		Tags:         []string{"test"},
	})
	analyzer := NewAIAnalyzer(config.AIConfig{
		APIKey:  "test-key",
		BaseURL: "http://example.com",
		Model:   "test-model",
		Prompt:  "system prompt",
		Timeout: time.Second,
	}, config.QueryConfig{}, config.LogsConfig{}, nil, client)
	result, err := analyzer.Analyze(context.Background(), buildTestAlertPayload())

	assert.NoError(t, err)
	assert.True(t, result.ShouldNotify)
	assert.Equal(t, "critical", result.Severity)
	assert.Equal(t, []string{"e1"}, result.Evidence)
	assert.Len(t, client.lastMessages, 2)
	assert.Equal(t, llm.RoleSystem, client.lastMessages[0].Role)
	assert.Equal(t, "system prompt", client.lastMessages[0].Content)
	assert.Equal(t, llm.RoleUser, client.lastMessages[1].Role)
	assert.Contains(t, client.lastMessages[1].Content, "告警状态：firing")
	assert.Contains(t, client.lastMessages[1].Content, "告警数量：1")
	assert.Contains(t, client.lastMessages[1].Content, "历史告警：")
}

// TestAIAnalyzerBuildPrompt_SkipWithoutGeneratorURL generatorURL 为空时跳过 PromQL 查询
func TestAIAnalyzerBuildPrompt_SkipWithoutGeneratorURL(t *testing.T) {
	analyzer := NewAIAnalyzer(config.AIConfig{}, config.QueryConfig{}, config.LogsConfig{}, nil, newTestAnalyzerClient(model.AnalysisResult{}))
	prompt, err := analyzer.buildPrompt(context.Background(), buildTestAlertPayload())

	assert.NoError(t, err)
	assert.Contains(t, prompt, "历史告警：")
	assert.NotContains(t, prompt, "PromQL：")
	assert.Contains(t, prompt, "请给出：1. 问题判断 2. 可能原因 3. 处理建议")
}

// TestAIAnalyzerBuildPrompt_InvalidPromQL PromQL 为空时返回错误
func TestAIAnalyzerBuildPrompt_InvalidPromQL(t *testing.T) {
	analyzer := NewAIAnalyzer(config.AIConfig{}, config.QueryConfig{}, config.LogsConfig{}, nil, newTestAnalyzerClient(model.AnalysisResult{}))
	payload := buildTestAlertPayloadWithGeneratorURL("http://prometheus:9090/graph?g0.expr=&g0.tab=1")

	_, err := analyzer.buildPrompt(context.Background(), payload)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "提取expr失败")
}

func TestAIAnalyzerCallPreviewMax50Chars(t *testing.T) {
	client := newTestAnalyzerClient(model.AnalysisResult{Summary: "ok"})
	analyzer := NewAIAnalyzer(config.AIConfig{}, config.QueryConfig{}, config.LogsConfig{}, nil, client)
	prompt := strings.Repeat("a", 3100)
	cfg := config.AIConfig{Prompt: "system prompt", Timeout: time.Second}

	result, err := analyzer.call(context.Background(), prompt, cfg)

	assert.NoError(t, err)
	assert.Equal(t, "ok", result.Summary)
	assert.Len(t, client.lastMessages, 2)
	assert.Equal(t, "system prompt", client.lastMessages[0].Content)
	assert.Equal(t, strings.Repeat("a", 3000), client.lastMessages[1].Content)
}

func TestTruncateRunesUsesRuneCount(t *testing.T) {
	assert.Equal(t, "你好", truncateRunes("你好世界", 2))
	assert.Equal(t, "你", truncateRunes("你好", 1))
}
