package config

import "os"

// Config 的字段值来自环境变量,以免把 signing secret / app secret 写进代码库。
type Config struct {
	Port          string
	SigningSecret string
	AppID         string
	AppSecret     string
}

// Load 读取环境变量;PORT 缺省 8080,后续可被命令行 flag 覆盖。
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return &Config{
		Port:          port,
		SigningSecret: os.Getenv("SEATALK_SIGNING_SECRET"),
		AppID:         os.Getenv("SEATALK_APP_ID"),
		AppSecret:     os.Getenv("SEATALK_APP_SECRET"),
	}
}
