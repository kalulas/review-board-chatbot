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
	"github.com/kalulas/review-board-chatbot/internal/employees"
	"github.com/kalulas/review-board-chatbot/internal/message"
	"github.com/kalulas/review-board-chatbot/internal/notify"
	"github.com/kalulas/review-board-chatbot/internal/seatalk"
	"github.com/kalulas/review-board-chatbot/internal/seatalkws"
	"github.com/kalulas/review-board-chatbot/internal/server"
)

func main() {
	port := flag.String("p", "8080", "HTTP listen port")
	configPath := flag.String("config", "config.toml", "path to TOML config file")
	messagesPath := flag.String("messages", "messages.toml", "path to message templates TOML")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config %s: %v", *configPath, err)
	}
	msgs, err := message.Load(*messagesPath)
	if err != nil {
		log.Fatalf("failed to load messages %s: %v", *messagesPath, err)
	}

	client := seatalk.NewClient(cfg.SeaTalk.AppID, cfg.SeaTalk.AppSecret)
	resolver := employees.New(client, cfg.ReviewBoard.EmailDomain)
	notifier := notify.New(client, resolver, msgs)
	pool := command.NewReplyPool()

	addr := ":" + *port
	srv := server.New(addr, cfg, notifier)
	wsConsumer := seatalkws.New(cfg.SeaTalk.AppID, cfg.SeaTalk.AppSecret, client, pool, cfg.SeaTalk.LogPayload)

	// 两条通道长期并存：Review Board 的 webhook -> HTTP server，SeaTalk 事件 -> WebSocket。
	// 共用一个信号 ctx，收到退出信号时二者一起退出。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer stop()

	go func() {
		log.Println("starting web, listening on", addr)
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalln("failed starting web on", addr, err)
		}
	}()

	go wsConsumer.Run(ctx)

	<-ctx.Done()
	log.Println("terminate service")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("shutting down web on", addr)
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalln("failed shutdown server", err)
	}
	log.Println("web gracefully stopped")
}
