package seatalk

// SeaTalk 回调里 event_type 的取值。
const (
	EventVerification             = "event_verification"
	EventMessageFromBotSubscriber = "message_from_bot_subscriber"
)

type EventCallbackReq struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	Timestamp uint64 `json:"timestamp"`
	AppID     string `json:"app_id"`
	Event     Event  `json:"event"`
}

// 配置回调 URL 时 SeaTalk 会发 verification 事件,需把 challenge 原样回传才算验证通过。
type EventVerificationResp struct {
	SeatalkChallenge string `json:"seatalk_challenge"`
}

type Event struct {
	SeatalkChallenge string  `json:"seatalk_challenge"`
	EmployeeCode     string  `json:"employee_code"`
	Message          Message `json:"message"`
}

type Message struct {
	Tag  string      `json:"tag"`
	Text TextMessage `json:"text"`
}

type TextMessage struct {
	Content   string `json:"content"`
	PlainText string `json:"plain_text"`
}
