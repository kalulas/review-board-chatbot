package message

import (
	"os"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Messages 对应独立的消息模板配置文件(messages.toml),与运行配置 config.toml 分开,方便单独改文案。
type Messages struct {
	Settings struct {
		ImagePlaceholder string `toml:"image_placeholder"`
		CommentMaxLen    int    `toml:"comment_max_len"`
		ReplyQuoteMaxLen int    `toml:"reply_quote_max_len"`
	} `toml:"settings"`
	CloseTypeLabels map[string]string `toml:"close_type_labels"`
	Templates       map[string]string `toml:"templates"`
}

func Load(path string) (*Messages, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Messages
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Render 把模板 key 里的 {字段} 用 vars 替换;模板不存在返回 ""。
func (m *Messages) Render(key string, vars map[string]string) string {
	tmpl, ok := m.Templates[key]
	if !ok {
		return ""
	}
	out := tmpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{"+k+"}", v)
	}
	return strings.TrimSpace(out)
}

var imageMarkdown = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)

// Excerpt 先把图片 markdown 换成占位符,再按 rune 截断到 max,超出补 "..."。max<=0 表示不截断。
func (m *Messages) Excerpt(text string, max int) string {
	s := imageMarkdown.ReplaceAllString(text, m.Settings.ImagePlaceholder)
	s = strings.TrimSpace(s)
	r := []rune(s)
	if max > 0 && len(r) > max {
		return string(r[:max]) + "..."
	}
	return s
}

// CloseTypeLabel 把 close_type 转成配置里的中文标签;无映射则原样返回。
func (m *Messages) CloseTypeLabel(closeType string) string {
	if label, ok := m.CloseTypeLabels[closeType]; ok {
		return label
	}
	return closeType
}
