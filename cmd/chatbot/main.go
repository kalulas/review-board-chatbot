package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kalulas/review-board-chatbot/internal/command"
	"github.com/kalulas/review-board-chatbot/internal/config"
	"github.com/kalulas/review-board-chatbot/internal/seatalk"
	"github.com/kalulas/review-board-chatbot/internal/server"
)

func main() {
	port := flag.String("p", "8080", "HTTP listen port")
	configPath := flag.String("config", "config.toml", "path to TOML config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config %s: %v", *configPath, err)
	}

	client := seatalk.NewClient(cfg.SeaTalk.AppID, cfg.SeaTalk.AppSecret)
	pool := command.NewReplyPool()

	addr := ":" + *port
	srv := server.New(addr, cfg, client, pool)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		log.Println("starting web, listening on", addr)
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalln("failed starting web on", addr, err)
		}
	}()

	<-c
	log.Println("terminate service")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("shutting down web on", addr)
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalln("failed shutdown server", err)
	}
	log.Println("web gracefully stopped")
}
