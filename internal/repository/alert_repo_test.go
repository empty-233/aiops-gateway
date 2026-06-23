package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"aiops-gateway/internal/model"
	"aiops-gateway/internal/repository"
)

// setupTestDB 创建内存数据库，每个测试独立一份，互不干扰
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper() // 标记为辅助函数，出错时报告调用方的行号而不是这里

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	if err := db.AutoMigrate(&model.AlertRecord{}); err != nil {
		t.Fatalf("迁移表结构失败: %v", err)
	}

	return db
}

func TestAlertRepository_Save(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAlertRepository(db)
	ctx := context.Background()

	record := &model.AlertRecord{
		Status:         "firing",
		AlertName:      "HighCPU",
		Severity:       "critical",
		Instance:       "server-01",
		Summary:        "CPU 超过 90%",
		Fingerprint:    "fp-highcpu",
		AnalysisResult: "建议检查进程占用",
	}

	err := repo.Save(ctx, record)
	if err != nil {
		t.Fatalf("Save 失败: %v", err)
	}

	// 保存后应该自动分配 ID
	if record.ID == 0 {
		t.Error("期望 ID > 0，实际为 0")
	}
}

func TestAlertRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAlertRepository(db)
	ctx := context.Background()

	// 先写入一条
	record := &model.AlertRecord{
		Status:      "firing",
		AlertName:   "HighMemory",
		Severity:    "warning",
		Fingerprint: "fp-highmemory",
	}
	if err := repo.Save(ctx, record); err != nil {
		t.Fatalf("准备测试数据失败: %v", err)
	}

	// 再按 ID 查出来
	found, err := repo.FindByID(ctx, record.ID)
	if err != nil {
		t.Fatalf("FindByID 失败: %v", err)
	}
	if found.AlertName != "HighMemory" {
		t.Errorf("期望 AlertName=HighMemory，实际=%s", found.AlertName)
	}
}

func TestAlertRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAlertRepository(db)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, 999) // 不存在的 ID
	if err == nil {
		t.Error("期望返回错误，实际为 nil")
	}
}

func TestAlertRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAlertRepository(db)
	ctx := context.Background()

	// 写入 3 条
	for i := range 3 {
		if err := repo.Save(ctx, &model.AlertRecord{
			AlertName:   fmt.Sprintf("Alert-%d", i),
			Status:      "firing",
			Fingerprint: fmt.Sprintf("fp-list-%d", i),
		}); err != nil {
			t.Fatalf("准备测试数据失败: %v", err)
		}
	}

	records, total, err := repo.List(ctx, 1, 3)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if total != 3 {
		t.Errorf("期望 total=3，实际=%d", total)
	}
	if len(records) != 3 {
		t.Errorf("期望返回 3 条，实际=%d", len(records))
	}
}

func TestAlertRepository_List_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAlertRepository(db)
	ctx := context.Background()

	// 写入 5 条
	for i := range 5 {
		if err := repo.Save(ctx, &model.AlertRecord{
			AlertName:   fmt.Sprintf("Alert-%d", i),
			Status:      "firing",
			Fingerprint: fmt.Sprintf("fp-page-%d", i),
		}); err != nil {
			t.Fatalf("准备测试数据失败: %v", err)
		}
	}

	// 第 1 页，每页 2 条
	records, total, err := repo.List(ctx, 1, 2)
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	if total != 5 {
		t.Errorf("期望 total=5，实际=%d", total)
	}
	if len(records) != 2 {
		t.Errorf("期望返回 2 条，实际=%d", len(records))
	}
}

func TestAlertRepository_LatestList(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAlertRepository(db)
	ctx := context.Background()
	base := time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC)

	for i := range 5 {
		record := &model.AlertRecord{
			AlertName:   fmt.Sprintf("Alert-%d", i),
			Status:      "firing",
			Fingerprint: fmt.Sprintf("fp-lastx-%d", i),
		}
		if err := repo.Save(ctx, record); err != nil {
			t.Fatalf("准备测试数据失败: %v", err)
		}
		if err := db.WithContext(ctx).
			Model(&model.AlertRecord{}).
			Where("id = ?", record.ID).
			Update("created_at", base.Add(time.Duration(i)*time.Minute)).Error; err != nil {
			t.Fatalf("更新 created_at 失败: %v", err)
		}
	}

	records, err := repo.LatestList(ctx, 3, "time", "desc")
	if err != nil {
		t.Fatalf("LatestList 失败: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("期望返回 3 条，实际=%d", len(records))
	}
	if records[0].AlertName != "Alert-4" || records[1].AlertName != "Alert-3" || records[2].AlertName != "Alert-2" {
		t.Fatalf("按 time desc 排序异常: got=%s,%s,%s", records[0].AlertName, records[1].AlertName, records[2].AlertName)
	}

	records, err = repo.LatestList(ctx, 3, "id", "asc")
	if err != nil {
		t.Fatalf("LatestList 失败: %v", err)
	}
	if records[0].AlertName != "Alert-0" || records[1].AlertName != "Alert-1" || records[2].AlertName != "Alert-2" {
		t.Fatalf("按 id asc 排序异常: got=%s,%s,%s", records[0].AlertName, records[1].AlertName, records[2].AlertName)
	}

	records, err = repo.LatestList(ctx, 3, "id", "desc")
	if err != nil {
		t.Fatalf("LatestList 失败: %v", err)
	}
	if records[0].AlertName != "Alert-4" || records[1].AlertName != "Alert-3" || records[2].AlertName != "Alert-2" {
		t.Fatalf("非法 order 未回退到 desc: got=%s,%s,%s", records[0].AlertName, records[1].AlertName, records[2].AlertName)
	}
}

func TestAlertRepository_LatestList_InvalidLimit(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewAlertRepository(db)
	ctx := context.Background()

	records, err := repo.LatestList(ctx, 0, "time", "desc")
	if err != nil {
		t.Fatalf("LatestList 失败: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("期望返回 0 条，实际=%d", len(records))
	}
}
