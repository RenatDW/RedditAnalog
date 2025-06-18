package posts

import "time"

type Author struct {
	ID       string `bson:"id" json:"id"`
	Username string `bson:"username" json:"username"`
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
	Author  Author    `bson:"author" json:"author"`
	Body    string    `bson:"body" json:"body"`
	Created time.Time `bson:"created" json:"created"`
	ID      string    `bson:"id" json:"id"`
}

type Vote struct {
	User string `bson:"user" json:"user"`
	Vote int    `bson:"vote" json:"vote"`
}
