package database

import (
	"fmt"
	"os"
	"path/filepath"

	"aiops-gateway/internal/config"
	"aiops-gateway/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Open(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)

	switch cfg.Driver {
	case "sqlite":
		db,err = OpenSQLite(cfg.SQLite)
	case "mysql":
		db,err = OpenMySQL(cfg.MySQL)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", cfg.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	if err := db.AutoMigrate(
		&model.AlertRecord{},
	); err != nil {
		return nil, fmt.Errorf("迁移表结构失败: %w", err)
	}

	return db, nil
}

func OpenSQLite(cfg config.SQLiteConfig) (*gorm.DB, error) {
	dir := filepath.Dir(cfg.Path)
	_, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("访问目录失败: %w", err)
	}
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("目录不存在: %s", dir)
	}

	db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{})
	return db, err
}

func OpenMySQL(cfg config.MySQLConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
		cfg.Loc,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return db, err
}
