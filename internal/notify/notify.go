package notify

import (
	"log"
	"strconv"

	"github.com/kalulas/review-board-chatbot/internal/directory"
	"github.com/kalulas/review-board-chatbot/internal/message"
	"github.com/kalulas/review-board-chatbot/internal/reviewboard"
	"github.com/kalulas/review-board-chatbot/internal/seatalk"
)

// Notifier 把 Review Board 事件转成 SeaTalk 定向通知。
type Notifier struct {
	client   *seatalk.Client
	resolver *directory.Resolver
	msgs     *message.Messages
}

func New(client *seatalk.Client, resolver *directory.Resolver, msgs *message.Messages) *Notifier {
	return &Notifier{client: client, resolver: resolver, msgs: msgs}
}

// Handle 按事件分发。设计为在 webhook 返回 200 之后异步调用。
func (n *Notifier) Handle(event string, p *reviewboard.Payload) {
	switch event {
	case reviewboard.EventReviewRequestPublished:
		n.handlePublished(p)
	case reviewboard.EventReviewRequestClosed:
		n.handleClosed(p)
	case reviewboard.EventReviewPublished:
		n.handleReviewPublished(p)
	case reviewboard.EventReplyPublished:
		n.handleReplyPublished(p)
	default:
		log.Printf("INFO: reviewboard event %s not handled", event)
	}
}

// #1 新建 / 新增 reviewer,#4 republish 更新
func (n *Notifier) handlePublished(p *reviewboard.Payload) {
	owner := p.Owner()
	vars := n.requestVars(p)

	if p.IsNew {
		// #1 新建:通知所有 reviewer
		n.send(exclude(p.Reviewers(), owner), n.msgs.Render("request_published_new", vars))
		return
	}

	// is_new=false:更新。新增的 reviewer 当「新请求」通知,其余当前 reviewer 当「更新内容」通知。
	added := p.AddedReviewers()
	if len(added) > 0 {
		n.send(exclude(added, owner), n.msgs.Render("request_published_new", vars))
	}
	addedSet := toSet(added)
	var existing []string
	for _, u := range p.Reviewers() {
		if !addedSet[u] {
			existing = append(existing, u)
		}
	}
	n.send(exclude(existing, owner), n.msgs.Render("request_published_update", vars))
}

// #5 关闭 request,通知当前 reviewer
func (n *Notifier) handleClosed(p *reviewboard.Payload) {
	actor := ""
	if p.ClosedBy != nil {
		actor = p.ClosedBy.Username
	}
	vars := n.requestVars(p)
	vars["closed_by"] = actor
	vars["close_type"] = n.msgs.CloseTypeLabel(p.CloseType)
	n.send(exclude(p.Reviewers(), actor), n.msgs.Render("request_closed", vars))
}

// #2 评论 / #6 Ship It,通知 owner
func (n *Notifier) handleReviewPublished(p *reviewboard.Payload) {
	if p.Review == nil {
		return
	}
	actor := p.Review.Links.User.Title
	owner := p.Owner()
	if owner == "" || owner == actor { // 自己点评/ship 自己的请求,不通知
		return
	}
	vars := n.requestVars(p)
	vars["user"] = actor

	if p.Review.ShipIt {
		n.send([]string{owner}, n.msgs.Render("ship_it", vars))
		return
	}

	comment := n.msgs.Excerpt(p.FirstComment(), n.msgs.Settings.CommentMaxLen)
	if comment == "" { // 既非 ship_it 又无评论内容,跳过
		return
	}
	vars["comment"] = comment
	vars["review_url"] = p.FirstCommentURL()
	n.send([]string{owner}, n.msgs.Render("review_comment", vars))
}

// #3 回复评论
func (n *Notifier) handleReplyPublished(p *reviewboard.Payload) {
	if p.Reply == nil {
		return
	}
	actor := p.Reply.Links.User.Title
	owner := p.Owner()
	// TODO(Phase 2): 应通知「原评论者」(GET reply_to.href 取其 links.user)。
	// 现部署够不到内网 RB API,退化为通知 owner,且仅当回复者 != owner。
	if owner == "" || owner == actor {
		return
	}
	vars := n.requestVars(p)
	vars["user"] = actor
	vars["comment"] = n.msgs.Excerpt(p.FirstComment(), n.msgs.Settings.CommentMaxLen)
	vars["reply_url"] = p.FirstCommentURL() // 回复挂的也是 comment,#comment<id> 同样可跳转

	quote := n.msgs.Excerpt(p.FirstQuote(), n.msgs.Settings.ReplyQuoteMaxLen)
	if quote == "" {
		// 对整条 review 的回复(reply.body_top),没有被回复的具体评论,用不带「原评论」的模板。
		n.send([]string{owner}, n.msgs.Render("reply_to_review", vars))
		return
	}
	vars["quote"] = quote
	n.send([]string{owner}, n.msgs.Render("reply", vars))
}

// requestVars 准备所有事件公用的 review_request 占位符。
func (n *Notifier) requestVars(p *reviewboard.Payload) map[string]string {
	return map[string]string{
		"owner":      p.Owner(),
		"rr_id":      strconv.Itoa(p.ReviewRequest.ID),
		"rr_summary": p.ReviewRequest.Summary,
		"rr_url":     p.ReviewRequest.AbsoluteURL,
	}
}

// send 解析 username->code 并把同一条消息发给每个收件人。
func (n *Notifier) send(usernames []string, msg string) {
	if msg == "" || len(usernames) == 0 {
		return
	}
	codes, err := n.resolver.Codes(usernames)
	if err != nil {
		log.Printf("ERROR: resolve employee codes for %v: %v", usernames, err)
		// 部分解析成功的仍继续发
	}
	for _, u := range usernames {
		code, ok := codes[u]
		if !ok {
			log.Printf("WARN: no employee code for %s, skip", u)
			continue
		}
		if err := n.client.SendMarkdownMessage(code, msg); err != nil {
			log.Printf("ERROR: send to %s(%s): %v", u, code, err)
		}
	}
}

// exclude 过滤掉动作发起者本人(不给自己发通知)和空 username。
func exclude(users []string, actor string) []string {
	out := make([]string, 0, len(users))
	for _, u := range users {
		if u != "" && u != actor {
			out = append(out, u)
		}
	}
	return out
}

func toSet(users []string) map[string]bool {
	set := make(map[string]bool, len(users))
	for _, u := range users {
		set[u] = true
	}
	return set
}
