package database

import (
	"fmt"

	"aiops-gateway/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Open(databasePath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	if err:=db.AutoMigrate(
		&model.AlertRecord{},
        &model.AppConfig{},
    ); err != nil {
        return nil, fmt.Errorf("迁移表结构失败: %w", err)
    }

	return db, nil
}
