package model

type AnalysisResult struct {
	ShouldNotify   bool     `json:"should_notify" jsonschema_description:"是否应该通知用户"`
	Severity       string   `json:"severity" jsonschema:"enum=critical,enum=warning,enum=info,enum=unknown" jsonschema_description:"告警等级"`
	Confidence     float64  `json:"confidence" jsonschema_description:"置信度 0~1"`
	NotifyTitle    string   `json:"notify_title" jsonschema_description:"通知标题"`
	NotifyContent  string   `json:"notify_content" jsonschema_description:"通知内容"`
	Summary        string   `json:"summary" jsonschema_description:"分析摘要"`
	Reason         string   `json:"reason" jsonschema_description:"分析原因"`
	Evidence       []string `json:"evidence" jsonschema_description:"证据"`
	Suggestions    []string `json:"suggestions" jsonschema_description:"建议操作"`
	MissingContext []string `json:"missing_context" jsonschema_description:"缺失信息"`
	Tags           []string `json:"tags" jsonschema_description:"分类标签"`
}