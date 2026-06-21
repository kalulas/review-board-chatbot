package server

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kalulas/review-board-chatbot/internal/config"
	"github.com/kalulas/review-board-chatbot/internal/reviewboard"
	"github.com/kalulas/review-board-chatbot/internal/seatalk"
)

// 目前只验证「Review Board -> chatbot」这条链路:打印事件名 + 原始 payload,
// 并可选地把事件转发给一个指定 employee code。
// TODO: 接 notify 包做按事件分发(published/closed/review/reply)与定向通知。
func handleReviewBoardWebhook(cfg *config.Config, client *seatalk.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
			return
		}

		// 端点挂在公网,配了 secret 必须验签;链路验证阶段还没配 secret 时先放行,留一条告警。
		if cfg.ReviewBoard.WebhookSecret != "" {
			if !reviewboard.VerifySignature(body, cfg.ReviewBoard.WebhookSecret, c.GetHeader("X-Hub-Signature")) {
				log.Printf("WARN: reviewboard webhook signature mismatch")
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		} else {
			log.Printf("WARN: REVIEWBOARD_WEBHOOK_SECRET 未配置,暂时跳过验签(仅限链路验证阶段)")
		}

		// RB 把事件名放在 header 里,比解析 body 更稳。
		event := c.GetHeader("X-ReviewBoard-Event")
		log.Printf("INFO: reviewboard event=%s body=%s", event, string(body))

		if code := cfg.ReviewBoard.TestEmployeeCode; code != "" {
			msg := "[Review Board] " + event
			if payload, err := reviewboard.Parse(body); err != nil {
				log.Printf("WARN: failed to parse reviewboard payload: %v", err)
			} else {
				msg = fmt.Sprintf("[Review Board] %s\nuser: %s\n%s", event, payload.Actor(), payload.URL())
			}
			if err := client.SendTextMessage(code, msg); err != nil {
				log.Printf("ERROR: failed to forward reviewboard event to %s: %v", code, err)
			}
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
