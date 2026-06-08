package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"aiops-gateway/internal/notify"
)

type BarkConfig struct {
	name      string
	serverURL string
	key       string
	client    *http.Client
}

type barkPushRequest struct {
	DeviceKey string `json:"device_key"`
	Title     string `json:"title,omitempty"`
	Subtitle  string `json:"subtitle,omitempty"`
	Body      string `json:"body"`
}

func NewBark(name, serverURL, key string) *BarkConfig {
	return &BarkConfig{
		name:      name,
		serverURL: serverURL,
		key:       key,
		client:    &http.Client{},
	}
}

func (c *BarkConfig) Name() string {
	return c.name
}

func (c *BarkConfig) Notify(ctx context.Context, msg notify.Message) error {
	if c.serverURL == "" || c.key == "" {
		c.serverURL = "https://api.day.app"
	}

	body := barkPushRequest{
		DeviceKey: c.key,
		Title:     msg.Title,
		Subtitle:  msg.Subtitle,
		Body:      msg.Body,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("序列化 bark 请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverURL+"/push", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("创建 bark 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送 bark 通知失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bark 通知失败，状态码: %d", resp.StatusCode)
	}

	return nil
}
