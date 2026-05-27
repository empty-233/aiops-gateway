package mocks

import (
    "context"

    "github.com/stretchr/testify/mock"
    "aiops-gateway/internal/model"
)

// MockAnalyzer 是 service.Analyzer 接口的 mock 实现
// 嵌入 mock.Mock 获得所有 testify mock 能力
type MockAnalyzer struct {
    mock.Mock
}

func (m *MockAnalyzer) Analyze(ctx context.Context, payload *model.AlertPayload) (string, error) {
    // Called 记录调用，并返回预设的值
    args := m.Called(ctx, payload)
    return args.String(0), args.Error(1)
}