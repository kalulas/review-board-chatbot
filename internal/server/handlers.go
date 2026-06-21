package server

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kalulas/review-board-chatbot/internal/command"
	"github.com/kalulas/review-board-chatbot/internal/seatalk"
)

func handleCallback(client *seatalk.Client, pool *command.ReplyPool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req seatalk.EventCallbackReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusInternalServerError, "something wrong")
			return
		}
		log.Printf("INFO: received event with event_type %s", req.EventType)

		switch req.EventType {
		case seatalk.EventVerification:
			c.JSON(http.StatusOK, seatalk.EventVerificationResp{SeatalkChallenge: req.Event.SeatalkChallenge})
		case seatalk.EventMessageFromBotSubscriber:
			text := req.Event.Message.Text.Content
			log.Printf("INFO: message received: %s, with employee_code: %s", text, req.Event.EmployeeCode)
			pool.Remember(text)
			reply := pool.Pick()
			if reply == "" {
				reply = "Hello World"
			}
			log.Printf("INFO: replying with: %s", reply)
			if err := client.SendTextMessage(req.Event.EmployeeCode, reply); err != nil {
				log.Printf("ERROR: failed to send message to user: %v", err)
				c.JSON(http.StatusInternalServerError, "something wrong")
				return
			}
			c.JSON(http.StatusOK, "Success")
		default:
			log.Printf("ERROR: event %s not handled yet!", req.EventType)
			c.JSON(http.StatusOK, "Success")
		}
	}
}
