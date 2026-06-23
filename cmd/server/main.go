package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"aiops-gateway/internal/app"
	"aiops-gateway/internal/config"
	"aiops-gateway/internal/database"
	"aiops-gateway/internal/handler"
	"aiops-gateway/internal/repository"
	"aiops-gateway/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	logger := buildLogger(cfg.Server.Mode)

	slog.SetDefault(logger)

	db, err := database.Open(cfg.Database)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	alertRepo := repository.NewAlertRepository(db)

	llmClient, err := app.NewLLMClient(cfg.AI)
	if err != nil {
		log.Fatalf("初始化 LLM client 失败: %v", err)
	}
	notifyClient, err := app.NewNotifyClient(cfg.Notify, logger)
	if err != nil {
		log.Fatalf("初始化 Notify client 失败: %v", err)
	}

	analyzer := service.NewAIAnalyzer(
		cfg.AI,
		cfg.Query,
		cfg.Logs,
		alertRepo,
		llmClient,
	)
	alertService := service.NewAlertService(analyzer, alertRepo, logger, notifyClient, cfg.AI)
	alertService.StartWorkers(appCtx)
	webhookHandler := handler.NewWebhookHandler(alertService, logger)

	gin.SetMode(cfg.Server.Mode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(RequestLogger(logger))

	handler.RegisterRoutes(router, webhookHandler)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("服务器启动", "address", addr)
	if err := router.Run(addr); err != nil {
		logger.Error("服务器启动失败", "error", err)
	}
	router.Run(addr)
}

func buildLogger(mode string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	if mode == "release" {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		logger.Info("请求完成",
			"method", ctx.Request.Method,
			"path", ctx.Request.URL.Path,
			"status", ctx.Writer.Status(),
			"ip", ctx.ClientIP(),
		)
	}
}
