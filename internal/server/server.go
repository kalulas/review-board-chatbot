package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kalulas/review-board-chatbot/internal/command"
	"github.com/kalulas/review-board-chatbot/internal/config"
	"github.com/kalulas/review-board-chatbot/internal/seatalk"
)

type Server struct {
	httpServer *http.Server
}

func New(cfg *config.Config, client *seatalk.Client, pool *command.ReplyPool) *Server {
	r := gin.Default()
	registerRoutes(r, cfg, client, pool)
	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + cfg.Port,
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
