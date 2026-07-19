package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kalulas/review-board-chatbot/internal/config"
	"github.com/kalulas/review-board-chatbot/internal/notify"
	"github.com/kalulas/review-board-chatbot/internal/reviewboard"
)

func handleReviewBoardWebhook(cfg *config.Config, notifier *notify.Notifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
			return
		}

		// 端点挂在公网,配了 secret 必须验签;没配则放行 + 告警。
		if cfg.ReviewBoard.WebhookSecret != "" {
			if !reviewboard.VerifySignature(body, cfg.ReviewBoard.WebhookSecret, c.GetHeader("X-Hub-Signature")) {
				log.Printf("WARN: reviewboard webhook signature mismatch")
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		} else {
			log.Printf("WARN: reviewboard webhook_secret not configured, skipping signature verification")
		}

		event := c.GetHeader("X-ReviewBoard-Event")

		if cfg.ReviewBoard.LogPayload {
			logReviewBoardPayload(event, body)
		}

		payload, err := reviewboard.Parse(body)
		if err != nil {
			// 仍回 200,避免 RB 因非 2xx 重投;解析失败记日志即可。
			log.Printf("WARN: parse reviewboard payload (event=%s): %v", event, err)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
			return
		}
		log.Printf("INFO: reviewboard event=%s rr=#%d", event, payload.ReviewRequest.ID)

		// 先 200 应答,再异步发通知,避免一事件多 reviewer 时阻塞 webhook、触发 RB 重投。
		go notifier.Handle(event, payload)

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func logReviewBoardPayload(event string, body []byte) {
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		log.Printf("DEBUG: reviewboard payload event=%s (raw, unparseable): %s", event, string(body))
		return
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
	log.Printf("DEBUG: reviewboard payload event=%s:\n%s", event, buf.String())
}
