package seatalkws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	sdk "git.garena.com/seatalk/seatalk-oapi-sdk-go"

	"github.com/kalulas/review-board-chatbot/internal/command"
	"github.com/kalulas/review-board-chatbot/internal/seatalk"
)

// 断线重连的退避区间;连接稳定超过 stableSession 才把退避重置回下限，避免闪断时疯狂重连。
const (
	minBackoff    = time.Second
	maxBackoff    = time.Minute
	stableSession = 30 * time.Second
)

type Consumer struct {
	client     *sdk.Client        // SDK 的 WebSocket 连接，只管单次连接的收发与心跳
	sender     *seatalk.Client    // 回消息走 OpenAPI HTTP，和收事件的通道相互独立
	pool       *command.ReplyPool // 收到的历史消息记录池
	logPayload bool
}

func New(appID, appSecret string, sender *seatalk.Client, pool *command.ReplyPool, logPayload bool) *Consumer {
	c := &Consumer{sender: sender, pool: pool, logPayload: logPayload}
	dispatcher := sdk.NewEventDispatcher().
		OnEnvelope(c.handleEnvelope).
		OnEvent(c.handleEvent).
		OnMessageFromBotSubscriber(c.handleSubscriberMessage).
		OnKick(c.handleKick).
		OnInvalidFrame(c.handleInvalidFrame)
	c.client = sdk.NewClient(appID, appSecret, sdk.WithEventDispatcher(dispatcher))
	return c
}

// 注意: Start 在 Seatalk Platform 侧正常关闭时同样返回 nil，所以只能靠 ctx.Err() 判断是不是我们主动要停。
func (c *Consumer) Run(ctx context.Context) {
	defer c.client.Close()

	backoff := minBackoff
	for {
		if ctx.Err() != nil {
			return
		}

		result, err := c.client.Connect(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("ERROR: seatalk websocket connect: %v", err)
			if !sleep(ctx, backoff) {
				return
			}
			backoff = nextBackoff(backoff)
			continue
		}
		log.Printf("INFO: seatalk websocket registered, sid=%s heartbeat_interval=%s", result.Sid, result.HeartbeatInterval)

		startedAt := time.Now()
		err = c.client.Start(ctx)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Printf("ERROR: seatalk websocket stopped: %v", err)
		} else {
			log.Printf("WARN: seatalk websocket closed by peer")
		}

		if time.Since(startedAt) >= stableSession {
			backoff = minBackoff
		}
		if !sleep(ctx, backoff) {
			return
		}
		backoff = nextBackoff(backoff)
	}
}

// handleEnvelope 覆盖 SDK 的默认实现——后者会把每一帧完整 JSON 用 fmt 打到 stdout，噪音过大。
func (c *Consumer) handleEnvelope(ctx context.Context, envelope *sdk.Envelope) error {
	if !c.logPayload {
		return nil
	}
	pretty, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		log.Printf("WARN: seatalk envelope marshal: %v", err)
		return nil
	}
	log.Printf("DEBUG: seatalk envelope:\n%s", pretty)
	return nil
}

// handleEvent 是所有事件的兜底 handler。SDK 只在事件被某个 handler 处理过时才回 ack，
// 缺了它，我们没写 typed handler 的事件类型会一直得不到确认、可能被平台反复重投。
func (c *Consumer) handleEvent(ctx context.Context, event *sdk.Event) error {
	var meta struct {
		EventID   string `json:"event_id"`
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(event.Data, &meta); err != nil {
		log.Printf("WARN: seatalk event payload unparseable: %v", err)
		return nil
	}
	log.Printf("INFO: seatalk event received, type=%s event_id=%s", meta.EventType, meta.EventID)
	return nil
}

// handleSubscriberMessage 处理私聊消息：记下用户发送的内容，再从已收到的内容里随机回一条。
// handler 一旦返回 error，SDK 会终止整条连接，所以业务失败只记日志，一律返回 nil。
func (c *Consumer) handleSubscriberMessage(ctx context.Context, event *sdk.MessageFromBotSubscriberEvent) error {
	message := event.Event.Message
	if message.Tag != sdk.MessageTagText || message.Text == nil {
		log.Printf("INFO: seatalk message ignored, tag=%s", message.Tag)
		return nil
	}

	text := message.Text.Content
	employeeCode := event.Event.EmployeeCode
	log.Printf("INFO: message received: %s, with employee_code: %s", text, employeeCode)

	c.pool.Remember(text)
	reply := c.pool.Pick()
	if reply == "" {
		return nil
	}
	if err := c.sender.SendTextMessage(employeeCode, reply); err != nil {
		log.Printf("ERROR: failed to send message to %s: %v", employeeCode, err)
	}
	return nil
}

// handleKick 只记录;被踢后连接会断，重连交给 Run。
func (c *Consumer) handleKick(ctx context.Context, envelope *sdk.Envelope) error {
	log.Printf("WARN: seatalk websocket kicked by platform: %s", envelope.Message)
	return nil
}

func (c *Consumer) handleInvalidFrame(ctx context.Context, payload []byte, err error) error {
	log.Printf("WARN: seatalk websocket invalid frame: %v", err)
	return nil
}

func nextBackoff(d time.Duration) time.Duration {
	if d *= 2; d > maxBackoff {
		return maxBackoff
	}
	return d
}

// sleep 等待 d，期间响应 ctx 取消;返回 false 表示应当退出。
func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
