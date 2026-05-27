package mocks

import (
	"context"

	"aiops-gateway/internal/model"

	"github.com/stretchr/testify/mock"
)

// MockAlertRepository 是 service.AlertRepository 接口的 mock 实现
type MockAlertRepository struct {
	mock.Mock
}

func (m *MockAlertRepository) Save(ctx context.Context, record *model.AlertRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockAlertRepository) FindByID(ctx context.Context, id uint) (*model.AlertRecord, error) {
	args := m.Called(ctx, id)
	// get(0) 返回 interface{}，需要类型断言
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AlertRecord), args.Error(1)
}

func (m *MockAlertRepository) FindByAlertName(ctx context.Context, alertName string) ([]model.AlertRecord, error) {
	args := m.Called(ctx, alertName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.AlertRecord), args.Error(1)
}

func (m *MockAlertRepository) LatestList(ctx context.Context, x int, sortBy, order string) ([]model.AlertRecord, error) {
	args := m.Called(ctx, x, sortBy, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.AlertRecord), args.Error(1)
}

func (m *MockAlertRepository) List(ctx context.Context, page, pageSize int) ([]model.AlertRecord, int64, error) {
	args := m.Called(ctx, page, pageSize)
	var records []model.AlertRecord
	var total int64

	if args.Get(0) != nil {
		records = args.Get(0).([]model.AlertRecord)
	}
	if args.Get(1) != nil {
		total = args.Get(1).(int64)
	}

	return records, total, args.Error(2)
}
