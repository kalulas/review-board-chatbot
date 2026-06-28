package reviewboard

import (
	"encoding/json"
	"strconv"
)

// Payload 只挑通知用得到的字段;RB webhook body 很大,其余字段忽略。
// RB 把关联对象序列化成 links.<role>.title,其中 title 就是 username。
type Payload struct {
	Event     string `json:"event"`
	IsNew     bool   `json:"is_new"`
	CloseType string `json:"close_type"`

	ReviewRequest   ReviewRequest `json:"review_request"`
	Review          *Review       `json:"review"`
	Reply           *Reply        `json:"reply"`
	Change          *Change       `json:"change"`
	ClosedBy        *FullUser     `json:"closed_by"`
	DiffComments    []Comment     `json:"diff_comments"`
	GeneralComments []Comment     `json:"general_comments"`
}

type ReviewRequest struct {
	ID           int      `json:"id"`
	Summary      string   `json:"summary"`
	AbsoluteURL  string   `json:"absolute_url"`
	TargetPeople []Titled `json:"target_people"`
	Links        struct {
		Submitter Titled `json:"submitter"`
	} `json:"links"`
}

type Review struct {
	ID          int    `json:"id"`
	ShipIt      bool   `json:"ship_it"`
	BodyTop     string `json:"body_top"`
	AbsoluteURL string `json:"absolute_url"`
	Links       struct {
		User Titled `json:"user"`
	} `json:"links"`
}

type Reply struct {
	ID         int    `json:"id"`
	BodyTop    string `json:"body_top"`    // 对整条 review 的回复,文本落在这里
	BodyBottom string `json:"body_bottom"`
	Links      struct {
		User Titled `json:"user"`
	} `json:"links"`
}

type Comment struct {
	ID    int    `json:"id"`
	Text  string `json:"text"`
	Links struct {
		User    Titled `json:"user"`
		ReplyTo Titled `json:"reply_to"` // title = 被回复的原评论文本(非作者)
	} `json:"links"`
}

type Change struct {
	FieldsChanged struct {
		TargetPeople struct {
			Added []FullUser `json:"added"`
		} `json:"target_people"`
	} `json:"fields_changed"`
}

// Titled 是 links 子字段的形态,title 即 username。
type Titled struct {
	Title string `json:"title"`
}

// FullUser 是 payload 顶层 / change 里的完整 user 对象(含 email)。
type FullUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

func Parse(body []byte) (*Payload, error) {
	var p Payload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Owner 返回 request owner 的 username。
func (p *Payload) Owner() string { return p.ReviewRequest.Links.Submitter.Title }

// Reviewers 返回当前所有 reviewer 的 username。
func (p *Payload) Reviewers() []string {
	out := make([]string, 0, len(p.ReviewRequest.TargetPeople))
	for _, t := range p.ReviewRequest.TargetPeople {
		if t.Title != "" {
			out = append(out, t.Title)
		}
	}
	return out
}

// AddedReviewers 返回本次 republish 新增的 reviewer username(仅 is_new=false 且改了 target_people 时有)。
func (p *Payload) AddedReviewers() []string {
	if p.Change == nil {
		return nil
	}
	out := make([]string, 0, len(p.Change.FieldsChanged.TargetPeople.Added))
	for _, u := range p.Change.FieldsChanged.TargetPeople.Added {
		if u.Username != "" {
			out = append(out, u.Username)
		}
	}
	return out
}

// FirstComment 返回本次 review/reply 的第一条评论文本:diff 优先,其次 general,再次 review.body_top。
func (p *Payload) FirstComment() string {
	if len(p.DiffComments) > 0 && p.DiffComments[0].Text != "" {
		return p.DiffComments[0].Text
	}
	if len(p.GeneralComments) > 0 && p.GeneralComments[0].Text != "" {
		return p.GeneralComments[0].Text
	}
	if p.Review != nil && p.Review.BodyTop != "" {
		return p.Review.BodyTop
	}
	if p.Reply != nil { // 对整条 review 的回复:文本在 body_top / body_bottom
		if p.Reply.BodyTop != "" {
			return p.Reply.BodyTop
		}
		return p.Reply.BodyBottom
	}
	return ""
}

// FirstCommentURL 返回定位到第一条评论的可跳转链接:行内评论用 #comment<id>,通用评论用 #gcomment<id>
// (#review<id> 在 RB 页面里不跳转,故不用)。无评论时回退到 RR 页。
func (p *Payload) FirstCommentURL() string {
	base := p.ReviewRequest.AbsoluteURL
	switch {
	case len(p.DiffComments) > 0:
		return base + "#comment" + strconv.Itoa(p.DiffComments[0].ID)
	case len(p.GeneralComments) > 0:
		return base + "#gcomment" + strconv.Itoa(p.GeneralComments[0].ID)
	default:
		return base
	}
}

// FirstQuote 返回第一条回复所针对的原评论文本(reply_to.title)。
func (p *Payload) FirstQuote() string {
	if len(p.DiffComments) > 0 {
		return p.DiffComments[0].Links.ReplyTo.Title
	}
	if len(p.GeneralComments) > 0 {
		return p.GeneralComments[0].Links.ReplyTo.Title
	}
	return ""
}
