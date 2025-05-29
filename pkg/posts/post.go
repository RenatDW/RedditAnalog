package posts

import "time"

type Author struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}
type PostToFront struct {
	Author           `json:"author"`
	Category         string    `json:"category"`
	Comments         []Comment `json:"comments"`
	Created          time.Time `json:"created"`
	ID               string    `json:"id"`
	Score            int       `json:"score"`
	UpVote           int       `json:"-"`
	Title            string    `json:"title"`
	Type             string    `json:"type"`
	Text             string    `json:"text,omitempty"`
	URL              string    `json:"url,omitempty"`
	UpvotePercentage int       `json:"upvotePercentage"`
	Views            int       `json:"views"`
	Votes            []*Vote   `json:"votes"`
}

type Comment struct {
	Author  Author    `json:"author"`
	Body    string    `json:"body"`
	Created time.Time `json:"created"`
	ID      string    `json:"id"`
}
type Vote struct {
	User string `json:"user"`
	Vote int    `json:"vote"`
}
