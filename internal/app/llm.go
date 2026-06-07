package app

import (
	"fmt"

	"aiops-gateway/internal/config"
	"aiops-gateway/internal/llm"
	compatibleopenai "aiops-gateway/internal/llm/compatible/openai"
)

func NewLLMClient(aiConfig config.AIConfig) (llm.JSONChatClient, error) {
	switch aiConfig.Channel {
	case config.AIChannelCompatibleOpenAI:
		return compatibleopenai.NewClient(aiConfig), nil
	default:
		return nil, fmt.Errorf("不支持的 AI channel: %s", aiConfig.Channel)
	}
}