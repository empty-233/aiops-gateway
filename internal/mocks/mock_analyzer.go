package mocks

import (
	"context"

	"aiops-gateway/internal/model"

	"github.com/stretchr/testify/mock"
)

// MockAnalyzer 是 service.Analyzer 接口的 mock 实现
// 嵌入 mock.Mock 获得所有 testify mock 能力
type MockAnalyzer struct {
	mock.Mock
}

func (m *MockAnalyzer) Analyze(ctx context.Context, payload *model.AlertPayload) (*model.AnalysisResult, error) {
	args := m.Called(ctx, payload)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*model.AnalysisResult), args.Error(1)
}
