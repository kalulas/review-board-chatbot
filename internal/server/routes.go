package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kalulas/review-board-chatbot/internal/config"
	"github.com/kalulas/review-board-chatbot/internal/notify"
)

func registerRoutes(r *gin.Engine, cfg *config.Config, notifier *notify.Notifier) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Review Board webhook;验签(HMAC-SHA1)在 handler 内做。
	r.POST("/webhook/reviewboard", handleReviewBoardWebhook(cfg, notifier))
}
