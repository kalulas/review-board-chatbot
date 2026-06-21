package reviewboard

import "encoding/json"

// Payload 只挑监控/通知用得到的字段;RB webhook body 很大,其余字段忽略即可。
// RB 把关联对象序列化成 links.<role>.title,其中 title 就是 username。
type Payload struct {
	Event         string        `json:"event"`
	IsNew         bool          `json:"is_new"`
	CloseType     string        `json:"close_type"`
	ReviewRequest ReviewRequest `json:"review_request"`
	Review        *userLinked   `json:"review"`
	Reply         *userLinked   `json:"reply"`
	ClosedBy      *namedUser    `json:"closed_by"`
	ReopenedBy    *namedUser    `json:"reopened_by"`
}

type ReviewRequest struct {
	AbsoluteURL string `json:"absolute_url"`
	ID          int    `json:"id"`
	Summary     string `json:"summary"`
	Links       struct {
		Submitter titled `json:"submitter"`
	} `json:"links"`
}

type userLinked struct {
	Links struct {
		User titled `json:"user"`
	} `json:"links"`
}

type titled struct {
	Title string `json:"title"`
}

type namedUser struct {
	Username string `json:"username"`
}

func Parse(body []byte) (*Payload, error) {
	var p Payload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// Actor 尽力取本次操作的执行者 username。各事件携带操作者的字段不同,
// 取不到时回退到 request owner(submitter)。
func (p *Payload) Actor() string {
	switch {
	case p.Review != nil && p.Review.Links.User.Title != "":
		return p.Review.Links.User.Title
	case p.Reply != nil && p.Reply.Links.User.Title != "":
		return p.Reply.Links.User.Title
	case p.ClosedBy != nil && p.ClosedBy.Username != "":
		return p.ClosedBy.Username
	case p.ReopenedBy != nil && p.ReopenedBy.Username != "":
		return p.ReopenedBy.Username
	default:
		return p.ReviewRequest.Links.Submitter.Title
	}
}

// URL 返回 review request 页面地址(RB 内网地址,内网点开即可)。
func (p *Payload) URL() string {
	return p.ReviewRequest.AbsoluteURL
}
