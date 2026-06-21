// Package notify 把 Review Board 事件转成 SeaTalk 通知。
// TODO: 事件到消息的映射随 Review Board webhook 端点一起实现。
package notify

import "github.com/kalulas/review-board-chatbot/internal/seatalk"

// Notifier 负责把 Review Board 事件组装成 SeaTalk 消息并发出去。
type Notifier struct {
	client *seatalk.Client
}

func New(client *seatalk.Client) *Notifier {
	return &Notifier{client: client}
}
