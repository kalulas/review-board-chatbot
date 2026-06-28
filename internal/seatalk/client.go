package seatalk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	authURL       = "https://openapi.seatalk.io/auth/app_access_token"
	singleChatURL = "https://openapi.seatalk.io/messaging/v2/single_chat"
)

// Client 封装 SeaTalk OpenAPI 调用;token 缓存复用,免得每次发消息都重新鉴权。
type Client struct {
	appID     string
	appSecret string
	http      *http.Client

	mu     sync.Mutex
	token  string
	expire uint64
}

func NewClient(appID, appSecret string) *Client {
	return &Client{
		appID:     appID,
		appSecret: appSecret,
		http:      &http.Client{},
	}
}

type authResp struct {
	Code           int    `json:"code"`
	AppAccessToken string `json:"app_access_token"`
	Expire         uint64 `json:"expire"`
}

func (c *Client) accessToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && c.expire > uint64(time.Now().Unix()) {
		return c.token, nil
	}

	body := []byte(fmt.Sprintf(`{"app_id": "%s", "app_secret": "%s"}`, c.appID, c.appSecret))
	req, err := http.NewRequest(http.MethodPost, authURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth: unexpected status %d", res.StatusCode)
	}

	var resp authResp
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return "", err
	}
	if resp.Code != 0 {
		return "", fmt.Errorf("auth: response code %d (see https://open.seatalk.io/docs/reference_server-api-error-code)", resp.Code)
	}

	c.token = resp.AppAccessToken
	c.expire = resp.Expire
	return c.token, nil
}

type sendTextRequest struct {
	EmployeeCode string      `json:"employee_code"`
	Message      messageBody `json:"message"`
}

type messageBody struct {
	Tag  string   `json:"tag"`
	Text textBody `json:"text"`
}

type textBody struct {
	Format  int8   `json:"format"`
	Content string `json:"content"`
}

type sendResp struct {
	Code int `json:"code"`
}

func (c *Client) SendTextMessage(employeeCode, content string) error {
	return c.sendMessage(employeeCode, content, 2) // 2 = 纯文本
}

// SendMarkdownMessage 以 markdown 格式发送(format 1),用于需要加粗/链接的通知。
func (c *Client) SendMarkdownMessage(employeeCode, content string) error {
	return c.sendMessage(employeeCode, content, 1)
}

func (c *Client) sendMessage(employeeCode, content string, format int8) error {
	token, err := c.accessToken()
	if err != nil {
		return err
	}

	payload, err := json.Marshal(sendTextRequest{
		EmployeeCode: employeeCode,
		Message: messageBody{
			Tag:  "text",
			Text: textBody{Format: format, Content: content},
		},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, singleChatURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var resp sendResp
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("send message: response code %d", resp.Code)
	}
	return nil
}
