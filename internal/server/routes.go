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

	// 签名校验只挂在 /callback 上:/health 不需要,日后 /webhook/reviewboard 会用自己的 HMAC 方案。
	callback := r.Group("/callback")
	callback.Use(seatalkSignature(cfg.SigningSecret))
	callback.POST("", handleCallback(client, pool))
}
