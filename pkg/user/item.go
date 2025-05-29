package user

import "gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/posts"

type User struct {
	ID        string
	Login     string
	password  string
	userPosts map[string]bool        // save PostID
	votes     map[string]*posts.Vote // [postID]voteValue
}

type UserRepo interface {
	Authorize(login, pass string) (User, error)
	SignUp(login, pass string) (User, error)
	AddPost(login, postID string) error
	DeletePost(login, postID string) error
	AddVote(login, postID string, vote *posts.Vote) error
	SetVote(login, postID string, voteValue int) error
	GetVote(login, postID string) int
	GetUserPosts(login string) []string
	GetUserVotes(login string) []string
}
