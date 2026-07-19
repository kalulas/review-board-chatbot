package config

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Config 对应一份 TOML 配置文件:集中放一处比散在一堆环境变量里改起来直观。
// 端口不在这里——它是部署期才定的,走命令行 -p 参数。
type Config struct {
	SeaTalk     SeaTalk     `toml:"seatalk"`
	ReviewBoard ReviewBoard `toml:"reviewboard"`
}

type SeaTalk struct {
	AppID      string `toml:"app_id"`
	AppSecret  string `toml:"app_secret"`
	LogPayload bool   `toml:"log_seatalk_payload"`
}

type ReviewBoard struct {
	// webhook 验签用;留空则跳过验签。
	WebhookSecret string `toml:"webhook_secret"`
	// 把 RB username 拼成 email 的域名后缀;进内网后可改。
	EmailDomain string `toml:"email_domain"`
	// 是否打印收到的 webhook payload
	LogPayload bool `toml:"log_review_board_payload"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
