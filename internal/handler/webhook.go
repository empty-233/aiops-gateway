package handler

import (
	"log/slog"
	"net/http"

	"aiops-gateway/internal/model"
	"aiops-gateway/internal/service"

	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	alertService *service.AlertService
	logger       *slog.Logger
}

func NewWebhookHandler(alertService *service.AlertService, logger *slog.Logger) *WebhookHandler {
	return &WebhookHandler{
		alertService: alertService,
		logger:       logger,
	}
}

func (h *WebhookHandler) HandlePrometheus(ctx *gin.Context) {
	var payload model.AlertPayload

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if payload.Data == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "payload data 不能为空"})
		return
	}

	if len(payload.Alerts) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "alerts 不能为空"})
		return
	}

	result, err := h.alertService.Process(ctx.Request.Context(), &payload)
	if err != nil {
		h.logger.Error("处理告警失败", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *WebhookHandler) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
