package server

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kalulas/review-board-chatbot/internal/seatalk"
)

func seatalkSignature(signingSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		signature := c.GetHeader("signature")
		if signature == "" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
			return
		}

		if !seatalk.VerifySignature(body, signingSecret, signature) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// 读 body 会耗尽 Request.Body,塞回去后续 handler 才解析得到。
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		c.Next()
	}
}
