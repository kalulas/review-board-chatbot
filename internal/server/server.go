package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kalulas/review-board-chatbot/internal/command"
	"github.com/kalulas/review-board-chatbot/internal/config"
	"github.com/kalulas/review-board-chatbot/internal/notify"
	"github.com/kalulas/review-board-chatbot/internal/seatalk"
)

type Server struct {
	httpServer *http.Server
}

func New(addr string, cfg *config.Config, client *seatalk.Client, pool *command.ReplyPool, notifier *notify.Notifier) *Server {
	r := gin.Default()
	registerRoutes(r, cfg, client, pool, notifier)
	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: r,
		},
	}
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
