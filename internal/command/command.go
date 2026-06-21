package command

import (
	"log"
	"math/rand"
	"os/exec"
	"strings"
	"sync"
)

// ReplyPool 把用户发来的消息存在内存里,回复时随机挑一条;本机装了 fortune 的话,其输出也作为一个候选参与抽签。
type ReplyPool struct {
	mu          sync.Mutex
	messages    []string
	fortunePath string // fortune 路径,未安装时为空
}

// NewReplyPool 在启动时解析一次 fortune 路径,之后复用,避免每次回复都查 PATH。
func NewReplyPool() *ReplyPool {
	p := &ReplyPool{}
	if path, err := exec.LookPath("fortune"); err == nil {
		p.fortunePath = path
		log.Printf("INFO: fortune found at %s; it will join the reply pool", path)
	} else {
		log.Printf("INFO: fortune not found; replies will use stored messages only")
	}
	return p
}

func (p *ReplyPool) Remember(text string) {
	if text == "" { // 空消息不进池,免得抽到空回复
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = append(p.messages, text)
}

// Pick 从历史消息(外加一条 fortune)里随机返回一条;池子为空时返回 "",兜底交给调用方。
func (p *ReplyPool) Pick() string {
	p.mu.Lock()
	candidates := make([]string, len(p.messages))
	copy(candidates, p.messages)
	p.mu.Unlock()

	if saying, ok := p.fortuneSaying(); ok {
		candidates = append(candidates, saying)
	}
	if len(candidates) == 0 {
		return ""
	}
	return candidates[rand.Intn(len(candidates))]
}

func (p *ReplyPool) fortuneSaying() (string, bool) {
	if p.fortunePath == "" {
		return "", false
	}
	out, err := exec.Command(p.fortunePath).Output()
	if err != nil {
		log.Printf("WARN: fortune execution failed: %v", err)
		return "", false
	}
	saying := strings.TrimSpace(string(out))
	if saying == "" {
		return "", false
	}
	return saying, true
}
