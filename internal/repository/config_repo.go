package repository

import (
	"fmt"
	"context"

	"gorm.io/gorm"

	"aiops-gateway/internal/model"
)

type ConfigRepository struct {
    db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
    return &ConfigRepository{db: db}
}

func (r *ConfigRepository) Get(ctx context.Context, key string) (string, error) {
    var cfg model.AppConfig
    result := r.db.WithContext(ctx).Where("key = ?", key).First(&cfg)

    if result.Error == gorm.ErrRecordNotFound {
        return "", nil
    }
    if result.Error != nil {
        return "", fmt.Errorf("读取配置失败: %w", result.Error)
    }

    return cfg.Value, nil
}

func (r *ConfigRepository) Set(ctx context.Context, key, value, desc string) error {
    cfg := model.AppConfig{Key: key, Value: value, Desc: desc}

    result := r.db.WithContext(ctx).
        Where(model.AppConfig{Key: key}).
        Assign(model.AppConfig{Value: value, Desc: desc}).
        FirstOrCreate(&cfg)

    if result.Error != nil {
        return fmt.Errorf("写入配置失败: %w", result.Error)
    }
    return nil
}