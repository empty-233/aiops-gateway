package logs

import (
	"context"
	"fmt"
	"strings"
)

type Client struct {
	source Source
}

func NewClient(source Source) *Client {
	return &Client{
		source: source,
	}
}

func (c *Client) Query(ctx context.Context, options *Options) (string, error) {
	switch c.source.Type {
	case SourceFile:
		file, err := NewFileReader().Read(ctx, c.source.Path, options)
		if err != nil {
			return "", fmt.Errorf("读取日志失败: %w", err)
		}
		return LogEntryToString(file), nil

	default:
		return "", fmt.Errorf("不支持的源类型: %s", c.source.Type)
	}
}

func LogEntryToString(entry []LogEntry) string {
	var total int
	for _, e := range entry {
		total += len(e.Message) + 1
	}

	var builder strings.Builder
	builder.Grow(total)
	for _, e := range entry {
		builder.WriteString(e.Message)
		builder.WriteByte('\n')
	}
	return builder.String()
}
