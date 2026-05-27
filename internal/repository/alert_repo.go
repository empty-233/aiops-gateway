package repository

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"aiops-gateway/internal/model"
)

type AlertRepository struct {
	db *gorm.DB
}

var orderWhitelist = map[string]string{
	"asc":  "ASC",
	"desc": "DESC",
}

var sortFieldWhitelist = map[string]string{
	"time": "created_at",
	"id":   "id",
}

func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

func (r *AlertRepository) Save(ctx context.Context, record *model.AlertRecord) error {
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "fingerprint"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "analysis", "updated_at"}),
	}).Create(record)
	if result.Error != nil {
		return fmt.Errorf("保存告警记录失败: %w", result.Error)
	}
	return nil
}

func (r *AlertRepository) FindByID(ctx context.Context, id uint) (*model.AlertRecord, error) {
	var record model.AlertRecord
	result := r.db.WithContext(ctx).First(&record, id)
	if result.Error != nil {
		return nil, fmt.Errorf("查询告警失败: %w", result.Error)
	}
	return &record, nil
}

func (r *AlertRepository) List(ctx context.Context, page, pageSize int) ([]model.AlertRecord, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var records []model.AlertRecord
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.AlertRecord{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计总数失败: %w", err)
	}

	result := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("分页查询失败: %w", result.Error)
	}

	return records, total, nil
}

func (r *AlertRepository) FindByAlertName(ctx context.Context, alertName string) ([]model.AlertRecord, error) {
	var records []model.AlertRecord
	result := r.db.WithContext(ctx).
		Where("alert_name = ?", alertName).
		Order("created_at DESC").
		Find(&records)

	if result.Error != nil {
		return nil, fmt.Errorf("按名称查询失败: %w", result.Error)
	}
	return records, nil
}

func (r *AlertRepository) LatestList(ctx context.Context, limit int, sortBy, order string) ([]model.AlertRecord, error) {
	if limit <= 0 {
		return []model.AlertRecord{}, nil
	}

	field, ok := sortFieldWhitelist[strings.ToLower(strings.TrimSpace(sortBy))]
	if !ok {
		return nil, fmt.Errorf("sortBy '%s' 不合法", sortBy)
	}

	direction, ok := orderWhitelist[strings.ToLower(strings.TrimSpace(order))]
	if !ok {
		return nil, fmt.Errorf("order '%s' 不合法", order)
	}

	var records []model.AlertRecord
	orderExpr := fmt.Sprintf("%s %s", field, direction)
	query := r.db.WithContext(ctx).Order(orderExpr)
	if field != "id" {
		query = query.Order(fmt.Sprintf("id %s", direction))
	}

	result := query.
		Limit(limit).
		Find(&records)

	if result.Error != nil {
		return nil, fmt.Errorf("查询最新列表失败: %w", result.Error)
	}

	return records, nil
}
