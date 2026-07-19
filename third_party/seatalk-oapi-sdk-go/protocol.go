package seatalkoapisdk

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	DefaultWebSocketURL = "wss://ws-openapi.haiserve.com/ws/bot"
)

type Command string

const (
	CommandRegister Command = "register"
	CommandAck      Command = "ack"
	CommandPing     Command = "ping"
	CommandPong     Command = "pong"
	CommandKick     Command = "kick"
	CommandEvent    Command = "event"
)

const (
	EventTypeUserEnterChatroomWithBot                 = "user_enter_chatroom_with_bot"
	EventTypeMessageFromBotSubscriber                 = "message_from_bot_subscriber"
	EventTypeNewMentionedMessageReceivedFromGroupChat = "new_mentioned_message_received_from_group_chat"
	EventTypeInteractiveMessageClick                  = "interactive_message_click"
	EventTypeNewMessageReceivedFromThread             = "new_message_received_from_thread"
	EventTypeBotAddedToGroupChat                      = "bot_added_to_group_chat"
	EventTypeBotRemovedFromGroupChat                  = "bot_removed_from_group_chat"
	EventTypeGroupChatConvertedToExternalGroup        = "group_chat_converted_to_external_group"
)

const (
	MessageTagText                         = "text"
	MessageTagCombinedForwardedChatHistory = "combined_forwarded_chat_history"
	MessageTagImage                        = "image"
	MessageTagFile                         = "file"
	MessageTagVideo                        = "video"
)

const (
	CodeOK    = 0
	CodeError = 1
)

type Header struct {
	AppID      string `json:"app_id,omitempty"`
	AppSecret  string `json:"app_secret,omitempty"`
	Token      string `json:"token,omitempty"`
	Sid        string `json:"sid,omitempty"`
	CallbackID string `json:"callback_id,omitempty"`
	Rid        string `json:"rid,omitempty"`
}

type Envelope struct {
	Cmd     Command         `json:"cmd"`
	Header  Header          `json:"header,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
	Code    int             `json:"code,omitempty"`
	Message string          `json:"message,omitempty"`
}

type RegisterResult struct {
	AppID             string
	Token             string
	Sid               string
	HeartbeatInterval time.Duration
	HeartbeatTimeout  time.Duration
}

type RegisterSettings struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
	HeartbeatTimeout  int `json:"heartbeat_timeout"`
}

type Event struct {
	AppID      string
	CallbackID string
	Rid        string
	Sid        string
	Data       json.RawMessage
}

type UserEnterChatroomWithBotEvent struct {
	CallbackID string                              `json:"-"`
	EventID    string                              `json:"event_id"`
	EventType  string                              `json:"event_type"`
	Timestamp  uint64                              `json:"timestamp"`
	AppID      string                              `json:"app_id"`
	Event      UserEnterChatroomWithBotEventDetail `json:"event"`
}

type UserEnterChatroomWithBotEventDetail struct {
	EmployeeCode string `json:"employee_code"`
	SeaTalkID    string `json:"seatalk_id"`
	Email        string `json:"email"`
}

type MessageFromBotSubscriberEvent struct {
	CallbackID string                              `json:"-"`
	EventID    string                              `json:"event_id"`
	EventType  string                              `json:"event_type"`
	Timestamp  uint64                              `json:"timestamp"`
	AppID      string                              `json:"app_id"`
	Event      MessageFromBotSubscriberEventDetail `json:"event"`
}

type MessageFromBotSubscriberEventDetail struct {
	SeaTalkID    string               `json:"seatalk_id"`
	EmployeeCode string               `json:"employee_code"`
	Email        string               `json:"email"`
	Message      BotSubscriberMessage `json:"message"`
}

type BotSubscriberMessage struct {
	MessageID       string        `json:"message_id"`
	QuotedMessageID string        `json:"quoted_message_id,omitempty"`
	ThreadID        string        `json:"thread_id,omitempty"`
	Tag             string        `json:"tag"`
	Text            *TextMessage  `json:"text,omitempty"`
	Image           *MediaMessage `json:"image,omitempty"`
	File            *FileMessage  `json:"file,omitempty"`
	Video           *MediaMessage `json:"video,omitempty"`
}

type TextMessage struct {
	Content string `json:"content"`
}

type MediaMessage struct {
	Content string `json:"content"`
}

type FileMessage struct {
	Content  string `json:"content"`
	Filename string `json:"filename"`
}

type NewMentionedMessageReceivedFromGroupChatEvent struct {
	CallbackID string                                            `json:"-"`
	EventID    string                                            `json:"event_id"`
	EventType  string                                            `json:"event_type"`
	Timestamp  uint64                                            `json:"timestamp"`
	AppID      string                                            `json:"app_id"`
	Event      NewMentionedMessageReceivedFromGroupChatEventData `json:"event"`
}

type NewMentionedMessageReceivedFromGroupChatEventData struct {
	GroupID string                    `json:"group_id"`
	Message MentionedGroupChatMessage `json:"message"`
}

type MentionedGroupChatMessage struct {
	MessageID       string                   `json:"message_id"`
	QuotedMessageID string                   `json:"quoted_message_id,omitempty"`
	ThreadID        string                   `json:"thread_id,omitempty"`
	Sender          MentionedGroupChatSender `json:"sender"`
	MessageSentTime uint64                   `json:"message_sent_time"`
	Tag             string                   `json:"tag"`
	Text            *MentionedGroupChatText  `json:"text,omitempty"`
}

type MentionedGroupChatSender struct {
	SeaTalkID    string `json:"seatalk_id"`
	EmployeeCode string `json:"employee_code"`
	Email        string `json:"email"`
	SenderType   int    `json:"sender_type"`
}

type MentionedGroupChatText struct {
	PlainText     string                    `json:"plain_text"`
	MentionedList []MentionedGroupChatActor `json:"mentioned_list"`
}

type MentionedGroupChatActor struct {
	Username     string `json:"username"`
	SeaTalkID    string `json:"seatalk_id"`
	EmployeeCode string `json:"employee_code"`
	Email        string `json:"email"`
}

type InteractiveMessageClickEvent struct {
	CallbackID string                             `json:"-"`
	EventID    string                             `json:"event_id"`
	EventType  string                             `json:"event_type"`
	Timestamp  uint64                             `json:"timestamp"`
	AppID      string                             `json:"app_id"`
	Event      InteractiveMessageClickEventDetail `json:"event"`
}

type InteractiveMessageClickEventDetail struct {
	MessageID    string `json:"message_id"`
	EmployeeCode string `json:"employee_code"`
	Email        string `json:"email"`
	Value        string `json:"value"`
	SeaTalkID    string `json:"seatalk_id"`
	GroupID      string `json:"group_id"`
	ThreadID     string `json:"thread_id"`
}

type NewMessageReceivedFromThreadEvent struct {
	CallbackID string                                `json:"-"`
	EventID    string                                `json:"event_id"`
	EventType  string                                `json:"event_type"`
	Timestamp  uint64                                `json:"timestamp"`
	AppID      string                                `json:"app_id"`
	Event      NewMessageReceivedFromThreadEventData `json:"event"`
}

type NewMessageReceivedFromThreadEventData struct {
	GroupID string        `json:"group_id"`
	Message ThreadMessage `json:"message"`
}

type ThreadMessage struct {
	MessageID       string                 `json:"message_id"`
	QuotedMessageID string                 `json:"quoted_message_id,omitempty"`
	ThreadID        string                 `json:"thread_id"`
	Sender          ThreadMessageSender    `json:"sender"`
	MessageSentTime uint64                 `json:"message_sent_time"`
	Tag             string                 `json:"tag"`
	Text            *ThreadTextMessage     `json:"text,omitempty"`
	Image           *MediaMessage          `json:"image,omitempty"`
	File            *FileMessage           `json:"file,omitempty"`
	Video           *MediaMessage          `json:"video,omitempty"`
	MentionedList   []ThreadMentionedActor `json:"mentioned_list,omitempty"`
}

type ThreadMessageSender struct {
	SeaTalkID    string `json:"seatalk_id"`
	EmployeeCode string `json:"employee_code"`
	Email        string `json:"email"`
}

type ThreadTextMessage struct {
	PlainText string `json:"plain_text"`
}

type ThreadMentionedActor struct {
	Username     string `json:"username"`
	SeaTalkID    string `json:"seatalk_id"`
	EmployeeCode string `json:"employee_code"`
	Email        string `json:"email"`
}

type BotAddedToGroupChatEvent struct {
	CallbackID string                       `json:"-"`
	EventID    string                       `json:"event_id"`
	EventType  string                       `json:"event_type"`
	Timestamp  uint64                       `json:"timestamp"`
	AppID      string                       `json:"app_id"`
	Event      BotAddedToGroupChatEventData `json:"event"`
}

type BotAddedToGroupChatEventData struct {
	Group   BotAddedGroupChat `json:"group"`
	Inviter BotAddedInviter   `json:"inviter"`
}

type BotAddedGroupChat struct {
	GroupID       string                `json:"group_id"`
	GroupName     string                `json:"group_name"`
	GroupSettings BotAddedGroupSettings `json:"group_settings"`
}

type BotAddedGroupSettings struct {
	ChatHistoryForNewMembers string `json:"chat_history_for_new_members"`
	CanNotifyWithAtAll       bool   `json:"can_notify_with_at_all"`
	CanViewMemberList        bool   `json:"can_view_member_list"`
}

type BotAddedInviter struct {
	SeaTalkID    string `json:"seatalk_id"`
	EmployeeCode string `json:"employee_code"`
	Email        string `json:"email"`
}

type BotRemovedFromGroupChatEvent struct {
	CallbackID string                         `json:"-"`
	EventID    string                         `json:"event_id"`
	EventType  string                         `json:"event_type"`
	Timestamp  uint64                         `json:"timestamp"`
	AppID      string                         `json:"app_id"`
	Event      BotRemovedFromGroupChatDetails `json:"event"`
}

type BotRemovedFromGroupChatDetails struct {
	GroupID string            `json:"group_id"`
	Remover BotRemovedRemover `json:"remover"`
}

type BotRemovedRemover struct {
	SeaTalkID    string `json:"seatalk_id"`
	EmployeeCode string `json:"employee_code"`
	Email        string `json:"email"`
}

type GroupChatConvertedToExternalGroupEvent struct {
	CallbackID string                                     `json:"-"`
	EventID    string                                     `json:"event_id"`
	EventType  string                                     `json:"event_type"`
	Timestamp  uint64                                     `json:"timestamp"`
	AppID      string                                     `json:"app_id"`
	Event      GroupChatConvertedToExternalGroupEventData `json:"event"`
}

type GroupChatConvertedToExternalGroupEventData struct {
	GroupID  string                                    `json:"group_id"`
	Operator GroupChatConvertedToExternalGroupOperator `json:"operator"`
}

type GroupChatConvertedToExternalGroupOperator struct {
	SeaTalkID    string `json:"seatalk_id"`
	EmployeeCode string `json:"employee_code"`
	Email        string `json:"email"`
}

type RegisterError struct {
	Code    int
	Message string
}

func (e *RegisterError) Error() string {
	return fmt.Sprintf("register rejected: code=%d msg=%s", e.Code, e.Message)
}

type KickError struct {
	Message  string
	Envelope Envelope
}

func (e *KickError) Error() string {
	if e.Message == "" {
		return "kicked"
	}
	return "kicked: " + e.Message
}
