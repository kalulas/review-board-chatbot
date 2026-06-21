package command

import (
	"math/rand"
	"sync"
)

// ReplyPool 把用户发来的消息存在内存里,回复时随机挑一条回过去。
type ReplyPool struct {
	mu       sync.Mutex
	messages []string
}

func NewReplyPool() *ReplyPool {
	return &ReplyPool{}
}

func (p *ReplyPool) Remember(text string) {
	if text == "" { // 空消息不进池,免得抽到空回复
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = append(p.messages, text)
}

// Pick 从历史消息里随机返回一条;池子为空时返回 ""。
func (p *ReplyPool) Pick() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.messages) == 0 {
		return ""
	}
	return p.messages[rand.Intn(len(p.messages))]
}
