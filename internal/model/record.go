package model

import "gorm.io/gorm"

type AlertRecord struct {
	gorm.Model
	Status      string `gorm:"not null"`       // firing
	AlertName   string `gorm:"index;not null"`
	Severity    string
	Instance    string
	Summary     string
	Description string
	Fingerprint string `gorm:"uniqueIndex"` // 同一条告警唯一标识
	Analysis    string `gorm:"type:text"`   // AI 分析结果，text 类型存长文本
	StartsAt    string
}

type AppConfig struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex;not null"`
	Value string `gorm:"type:text;not null"`
	Desc  string // 配置说明
}
