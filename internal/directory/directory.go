package directory

import (
	"sync"

	"github.com/kalulas/review-board-chatbot/internal/seatalk"
)

// Resolver 把 RB username 解析成 SeaTalk employee_code:username -> email(拼 domain)-> code(查 API),带缓存。
type Resolver struct {
	client *seatalk.Client
	domain string // RB username 拼成 email 的域名后缀

	mu    sync.Mutex
	cache map[string]string // username -> employee_code(只缓存查到的)
}

func New(client *seatalk.Client, domain string) *Resolver {
	return &Resolver{client: client, domain: domain, cache: make(map[string]string)}
}

// Codes 返回 username -> employee_code;查不到的 username 不在结果里。命中缓存的不重复查 API。
func (r *Resolver) Codes(usernames []string) (map[string]string, error) {
	result := make(map[string]string)
	var missing []string

	r.mu.Lock()
	for _, u := range usernames {
		if code, ok := r.cache[u]; ok {
			result[u] = code
		} else {
			missing = append(missing, u)
		}
	}
	r.mu.Unlock()

	if len(missing) == 0 {
		return result, nil
	}

	emails := make([]string, len(missing))
	for i, u := range missing {
		emails[i] = u + r.domain
	}

	byEmail, err := r.client.EmployeeCodesByEmail(emails)
	if err != nil {
		return result, err
	}

	r.mu.Lock()
	for _, u := range missing {
		if code := byEmail[u+r.domain]; code != "" {
			r.cache[u] = code // 查不到的不缓存,留待以后(可能后续入职)重试
			result[u] = code
		}
	}
	r.mu.Unlock()

	return result, nil
}
