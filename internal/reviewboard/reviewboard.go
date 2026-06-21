// Package reviewboard 解析 Review Board webhook 并分类出我们关心的子事件。
// TODO: payload 结构体与分发逻辑随 /webhook/reviewboard 端点一起实现。
package reviewboard

// Review Board 在 Admin UI 里配置的 webhook 事件名。
const (
	EventReviewRequestPublished = "review_request_published"
	EventReviewRequestClosed    = "review_request_closed"
	EventReviewPublished        = "review_published"
	EventReplyPublished         = "reply_published"
)
