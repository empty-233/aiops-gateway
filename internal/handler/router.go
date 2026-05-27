package handler

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, wh *WebhookHandler) {
	router.GET("/health", wh.HealthCheck)

	webhook := router.Group("/webhook")
	{
		webhook.POST("/prometheus", wh.HandlePrometheus)
	}
}
