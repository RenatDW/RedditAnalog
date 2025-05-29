package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/posts"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/session"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/user"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type AddComment struct {
	Comment string `json:"comment"`
}

type AddPost struct {
	Category string `json:"category"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	URL      string `json:"url"`
	Text     string `json:"text"`
}
type ItemsHandler struct {
	UserRepo  *user.UserMemoryRepository
	ItemsRepo posts.ItemsRepo
	Logger    *zap.SugaredLogger
}

func (i *ItemsHandler) AddPost(post *posts.PostToFront, ss *session.Session) {
	i.Logger.Info("Adding Post")
	i.ItemsRepo.AddPost(post)
	err := i.UserRepo.AddPost(ss.Login, post.ID)
	if err != nil {
		i.Logger.Error("Failed to add post", err)
		return
	}
}

func (i *ItemsHandler) DeletePost(postID string, ss *session.Session) {
	i.Logger.Info("Deleting Post")
	i.ItemsRepo.DeletePost(postID)
	err := i.UserRepo.DeletePost(ss.Login, postID)
	if err != nil {
		i.Logger.Error("Failed to add post", err)
		return
	}
}

func (i *ItemsHandler) PostsWithCategory(w http.ResponseWriter, req *http.Request) {
	i.Logger.Info("PostsWithCategory start working")
	vars := mux.Vars(req)
	category := vars["category"]

	allPosts := i.ItemsRepo.GetAll()
	postsCopy := make([]*posts.PostToFront, 0, len(allPosts))
	if category != "" {
		for i := 0; i < len(allPosts); i++ {
			if allPosts[i].Category == category {
				postToFront := posts.ConstructPostToFront(allPosts[i])
				postsCopy = append(postsCopy, postToFront)
			}
		}
	} else {
		for i := 0; i < len(allPosts); i++ {
			postToFront := posts.ConstructPostToFront(allPosts[i])
			postsCopy = append(postsCopy, postToFront)
		}
	}
	i.Logger.Infof("Отображены посты с категроией %s", category)
	err := json.NewEncoder(w).Encode(postsCopy)
	if err != nil {
		i.Logger.Error(err)
		return
	}
}

func (i *ItemsHandler) PostInfo(w http.ResponseWriter, req *http.Request) {
	i.Logger.Info("PostInfo start working")
	vars := mux.Vars(req)
	postID := vars["post_id"]
	post, ok := i.ItemsRepo.FindPost(postID)
	if !ok {
		i.Logger.Infof("Пост  не найден %s", postID)
		return
	}
	i.Logger.Infof("Отображен пост с id %s", postID)
	w.Header().Set("Content-Type", "application/json; charset=utf-8\n\n")
	err := json.NewEncoder(w).Encode(posts.ConstructPostToFront(post))
	if err != nil {
		i.Logger.Error(err)
		return
	}

}

func (i *ItemsHandler) Posts(w http.ResponseWriter, req *http.Request) {
	i.Logger.Info("Posts start working")
	allPosts := i.ItemsRepo.GetAll()
	postToFront := make([]*posts.PostToFront, 0, len(allPosts))
	for _, post := range allPosts {
		postToFront = append(postToFront, posts.ConstructPostToFront(post))
	}

	err := json.NewEncoder(w).Encode(postToFront)
	if err != nil {
		i.Logger.Error(err)
		return
	}
}

func (i *ItemsHandler) AddPosts(w http.ResponseWriter, req *http.Request) {
	i.Logger.Info("Add Posts start working")
	var post AddPost
	err := json.NewDecoder(req.Body).Decode(&post)
	if err != nil {
		i.Logger.Error(err)
		return
	}
	ss, err := session.SessionFromContext(req.Context())
	if err != nil {
		i.Logger.Error(err)
		return
	}

	aut := posts.Author{Username: ss.Login, ID: ss.UserID}
	newPost := posts.PostToFront{
		Author:           aut,
		Category:         post.Category,
		Comments:         make([]posts.Comment, 0),
		Created:          time.Now(),
		Score:            0,
		Title:            post.Title,
		Type:             post.Type,
		Text:             post.Text,
		URL:              post.URL,
		UpvotePercentage: 0,
		Views:            0,
		Votes:            []*posts.Vote{},
	}
	i.AddPost(&newPost, ss)

	err = json.NewEncoder(w).Encode(newPost)
	if err != nil {
		i.Logger.Error(err)
		return
	}
}

func (i *ItemsHandler) PostDelete(w http.ResponseWriter, req *http.Request) {
	i.Logger.Info("PostDelete")
	postID := mux.Vars(req)["post_id"]

	ss, err := session.SessionFromContext(req.Context())
	if err != nil {
		i.Logger.Error(err)
		return
	}
	post, ok := i.ItemsRepo.FindPost(postID)
	if !ok {
		i.Logger.Infof("Пост  не найден %s", postID)
		return
	}
	if post.Author.ID == ss.UserID {
		i.DeletePost(post.ID, ss)
		i.Logger.Infof("Пост %s удален", postID)
	} else {
		i.Logger.Infof("Пользователь не имеет права удалить пост %s", postID)
	}
}

func (i *ItemsHandler) UserPosts(w http.ResponseWriter, req *http.Request) {
	i.Logger.Info("UserPosts start working")
	username := mux.Vars(req)["user_login"]

	postIDS := i.UserRepo.GetUserPosts(username)
	userPosts := make([]*posts.PostToFront, 0, len(postIDS))
	for k := 0; k < len(postIDS); k++ {
		item, ok := i.ItemsRepo.FindPost(postIDS[k])
		if !ok {
			i.Logger.Infof("Пост %s не найден", k)
			continue
		}
		postToFront := posts.ConstructPostToFront(item)
		userPosts = append(userPosts, postToFront)

	}

	err := json.NewEncoder(w).Encode(userPosts)
	if err != nil {
		i.Logger.Error(err)
		return
	}
}

func (i *ItemsHandler) CommentAdd(w http.ResponseWriter, req *http.Request) {
	i.Logger.Info("CommentAdd start working")
	comment := AddComment{}
	err := json.NewDecoder(req.Body).Decode(&comment)
	if err != nil {
		i.Logger.Error(err)
		return
	}

	ss, err := session.SessionFromContext(req.Context())
	if err != nil {
		i.Logger.Error(err)
		return
	}

	postID := mux.Vars(req)["post_id"]
	post, ok := i.ItemsRepo.FindPost(postID)
	if !ok {
		i.Logger.Infof("Пост  не найден %s", postID)
		return
	}
	aut := posts.Author{Username: ss.Login, ID: ss.UserID}
	i.ItemsRepo.AddComment(postID, posts.Comment{Author: aut, Body: comment.Comment, Created: time.Now()})
	postToFront := posts.ConstructPostToFront(post)
	err = json.NewEncoder(w).Encode([]posts.PostToFront{*postToFront})
	if err != nil {
		i.Logger.Error(err)
		return
	}

}
func (i *ItemsHandler) CommentDelete(w http.ResponseWriter, req *http.Request) {
	i.Logger.Info("CommentDelete start working")
	vars := mux.Vars(req)
	postID := vars["post_id"]
	commentID := vars["comment_id"]

	post, ok := i.ItemsRepo.FindPost(postID)
	if !ok {
		i.Logger.Infof("Пост  не найден %s", postID)
		return
	}
	i.ItemsRepo.DeleteComment(post.ID, commentID)
	i.Logger.Infof("Комментарий %s удален", commentID)
	err := json.NewEncoder(w).Encode(posts.ConstructPostToFront(post))
	if err != nil {
		i.Logger.Error(err)
		return
	}
}
