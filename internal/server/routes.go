package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kalulas/review-board-chatbot/internal/command"
	"github.com/kalulas/review-board-chatbot/internal/config"
	"github.com/kalulas/review-board-chatbot/internal/seatalk"
)

func registerRoutes(r *gin.Engine, cfg *config.Config, client *seatalk.Client, pool *command.ReplyPool) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 签名校验只挂在 /callback 上:/health 不需要,/webhook/reviewboard 用自己的 HMAC 方案。
	callback := r.Group("/callback")
	callback.Use(seatalkSignature(cfg.SeaTalk.SigningSecret))
	callback.POST("", handleCallback(client, pool))

	// Review Board webhook;验签(HMAC-SHA1)在 handler 内做,方案和 SeaTalk 不同。
	r.POST("/webhook/reviewboard", handleReviewBoardWebhook(cfg, client))
}
