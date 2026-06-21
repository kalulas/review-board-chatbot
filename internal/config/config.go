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
	SigningSecret string `toml:"signing_secret"`
	AppID         string `toml:"app_id"`
	AppSecret     string `toml:"app_secret"`
}

type ReviewBoard struct {
	// webhook 验签用;链路验证阶段可留空(跳过验签)。
	WebhookSecret string `toml:"webhook_secret"`
	// 链路验证阶段把收到的事件转发给这个 employee code;留空则只打日志。
	TestEmployeeCode string `toml:"test_employee_code"`
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
