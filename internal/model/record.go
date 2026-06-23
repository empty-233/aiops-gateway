package model

import (
	"time"

	"gorm.io/gorm"
)

type AlertRecord struct {
	gorm.Model
	Status         string `gorm:"size:32;not null"`
	AlertName      string `gorm:"size:128;not null;index"`
	Severity       string `gorm:"size:32"`
	Instance       string `gorm:"size:255"`
	Summary        string `gorm:"size:512"`
	Description    string `gorm:"type:text"`
	Fingerprint    string `gorm:"size:128;not null;uniqueIndex;comment:告警唯一标识"`
	AnalysisResult string `gorm:"type:text;comment:告警分析结果"`
	StartsAt       time.Time
}
