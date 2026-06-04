package llm

import (
	"context"
	"encoding/json"
	"fmt"
)

type JSONChatClient interface {
	RawChatJSON(ctx context.Context, msg []Message, schema any) (content string, raw string, err error)
}

func ChatJSON[T any](ctx context.Context, client JSONChatClient, msg []Message) (*ChatResponse[T], error) {
	var schemaTarget T
	schema := GenerateSchema(schemaTarget)

	content, raw, err := client.RawChatJSON(ctx, msg, schema)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("解析 LLM 结构化响应失败: %w, content=%s", err, content)
	}

	return &ChatResponse[T]{
		Content: content,
		Raw:     raw,
		Result:  result,
	}, nil
}
