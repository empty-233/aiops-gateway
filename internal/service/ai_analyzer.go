package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"aiops-gateway/internal/config"
	"aiops-gateway/internal/context/logs"
	"aiops-gateway/internal/model"
	"aiops-gateway/internal/prometheus"
)

type AIAnalyzer struct {
	aiConfig    config.AIConfig
	queryConfig config.QueryConfig
	logConfig   config.LogsConfig
	alertRepo   AlertRepository
}

func NewAIAnalyzer(aiConfig config.AIConfig, queryConfig config.QueryConfig, logConfig config.LogsConfig, alertRepo AlertRepository) *AIAnalyzer {
	return &AIAnalyzer{
		aiConfig:    aiConfig,
		queryConfig: queryConfig,
		logConfig:   logConfig,
		alertRepo:   alertRepo,
	}
}

func (a *AIAnalyzer) Analyze(ctx context.Context, payload *model.AlertPayload) (string, error) {
	prompt, err := a.buildPrompt(ctx, payload)
	if err != nil {
		return "", fmt.Errorf("构建 prompt 失败: %w", err)
	}

	response, err := a.call(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("调用 AI 失败: %w", err)
	}

	return response, nil
}

func (a *AIAnalyzer) buildPrompt(ctx context.Context, payload *model.AlertPayload) (string, error) {
	var sb strings.Builder

	sb.WriteString("你是一名 SRE 运维专家，请分析以下 Prometheus 告警并给出处理建议。\n\n")
	sb.WriteString(fmt.Sprintf("告警状态：%s\n", payload.Status))
	sb.WriteString(fmt.Sprintf("告警数量：%d\n\n", len(payload.Alerts)))

	for i, alert := range payload.Alerts {
		logOptions := &logs.Options{
			StartTime: time.Now().Add(-5 * time.Minute),
			EndTime:   time.Now(),
		}

		sb.WriteString(fmt.Sprintf("=== 告警 #%d ===\n", i+1))
		sb.WriteString(fmt.Sprintf("名称：%s\n", alert.Labels["alertname"]))
		sb.WriteString(fmt.Sprintf("级别：%s\n", alert.Labels["severity"]))
		sb.WriteString(fmt.Sprintf("实例：%s\n", alert.Labels["instance"]))
		sb.WriteString(fmt.Sprintf("摘要：%s\n", alert.Annotations["summary"]))
		sb.WriteString(fmt.Sprintf("描述：%s\n", alert.Annotations["description"]))
		sb.WriteString("\n")

		sb.WriteString("历史告警：\n")
		if alert.GeneratorURL == "" {
			// sb.WriteString("- 未提供 generatorURL，跳过 PromQL 查询\n\n")
			continue
		}

		promAddress := extractPrometheusAddress(alert.GeneratorURL)
		if promAddress == "" {
			return "", fmt.Errorf("- 无法从 generatorURL 提取 Prometheus 地址\n\n")
		}

		promql, err := prometheus.BuildPromQL(alert.GeneratorURL, alert.Labels)
		if err != nil {
			return "", err
		}
		sb.WriteString(fmt.Sprintf("PromQL：%s\n", promql))

		prometheusClient, err := prometheus.NewPrometheusClient(promAddress, a.queryConfig)
		if err != nil {
			return "", fmt.Errorf("- %s: 客户端初始化失败: %v\n\n", alert.Labels["alertname"], err)
		}

		end := time.Now()
		step := a.queryConfig.Step
		if step <= 0 {
			step = time.Minute
		}
		rangeTime := a.queryConfig.RangeTime
		if rangeTime <= 0 {
			rangeTime = time.Hour
		}
		start := end.Add(-rangeTime)

		series, err := prometheusClient.QueryRange(ctx, promql, start, end, step)
		if err != nil {
			return "", fmt.Errorf("- %s: 查询失败: %v\n\n", alert.Labels["alertname"], err)
		}
		if len(series) == 0 {
			sb.WriteString("- 查询成功但无数据返回\n\n")
			continue
		}
		sb.WriteString(fmt.Sprintf(
			"query_context: {alert=%q, endpoint=%s, promql=%q, series=%d}\n",
			alert.Labels["alertname"],
			promAddress,
			promql,
			len(series),
		))

		sb.WriteString(prometheus.BuildMetricEvidence(series))

		sb.WriteString("\n")

		matchSources := matchSources(alert, a.logConfig.Source)
		if len(matchSources) == 0 {
			continue
		}

		var stringLog strings.Builder
		for _, source := range matchSources {
			logClient := logs.NewClient(logs.Source{
				Type:     logs.SourceType(source.Type),
				Path:     source.Path,
				MaxLines: a.logConfig.MaxLines,
			})
			logQuery, err := logClient.Query(ctx, logOptions)
			if err != nil {
				return "", fmt.Errorf("- %s: 查询日志失败: %v\n", alert.Labels["alertname"], err)
			}
			stringLog.WriteString("------\n")
			stringLog.WriteString(fmt.Sprintf("日志来源: key-%s|value-%s\n", source.LabelKey, source.LabelValue))
			stringLog.WriteString(logQuery)
			stringLog.WriteString("------\n")
		}

		sb.WriteString("=== 相关日志===\n")
		sb.WriteString(stringLog.String())

		sb.WriteString("\n")
	}

	sb.WriteString("请给出：1. 问题判断 2. 可能原因 3. 处理建议")

	fmt.Print(sb.String())

	return sb.String(), nil
}

func extractPrometheusAddress(generatorURL string) string {
	if generatorURL == "" {
		return ""
	}

	parsedURL, err := url.Parse(generatorURL)
	if err != nil {
		return ""
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return ""
	}

	return parsedURL.Scheme + "://" + parsedURL.Host
}

func (a *AIAnalyzer) call(ctx context.Context, prompt string) (string, error) {
	preview := truncateRunes(prompt, 50)
	return fmt.Sprintf("[待接入 AI] 收到 prompt，长度 %d 字符。\nprompt 预览：%s...", len(prompt), preview), nil
}

func truncateRunes(s string, maxRunes int) string {
	if s == "" || maxRunes <= 0 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}

	return string(runes[:maxRunes])
}

func matchSources(alert model.Alert, sources []config.SourceConfig) []config.SourceConfig {
	var matchedSources []config.SourceConfig

	for _, source := range sources {
		if matchSource(alert, source) {
			matchedSources = append(matchedSources, source)
		}
	}

	return matchedSources
}

func matchSource(alert model.Alert, source config.SourceConfig) bool {
	if source.LabelKey == "" || source.LabelValue == "" {
		return false
	}

	actualValue, ok := alert.Labels[source.LabelKey]
	if !ok {
		return false
	}

	actualValue = strings.TrimSpace(actualValue)
	expectedValue := strings.TrimSpace(source.LabelValue)

	if actualValue == "" || expectedValue == "" {
		return false
	}

	actualValue = strings.ToLower(actualValue)
	expectedValue = strings.ToLower(expectedValue)

	if source.FuzzyMatch {
		return strings.Contains(actualValue, expectedValue)
	}

	return actualValue == expectedValue
}
