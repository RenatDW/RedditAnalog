package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/posts"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/session"
	"net/http"
)

func (i *ItemsHandler) PostUpVote(w http.ResponseWriter, req *http.Request) {
	postID := mux.Vars(req)["post_id"]
	post, ok := i.ItemsRepo.FindPost(postID)
	if !ok {
		i.Logger.Infof("Пост  не найден %s", postID)
		return
	}

	ss, err := session.SessionFromContext(req.Context())
	if err != nil {
		i.Logger.Error(err)
		return
	}
	userVote := i.UserRepo.GetVote(ss.Login, postID)
	i.Vote(ss, post, 1, userVote)
	err = json.NewEncoder(w).Encode(posts.ConstructPostToFront(post))
	if err != nil {
		i.Logger.Error(err)
		return
	}

}

func (i *ItemsHandler) PostDownVote(w http.ResponseWriter, req *http.Request) {
	postID := mux.Vars(req)["post_id"]
	post, ok := i.ItemsRepo.FindPost(postID)
	if !ok {
		i.Logger.Infof("Пост  не найден %s", postID)
		return
	}

	ss, err := session.SessionFromContext(req.Context())
	if err != nil {
		i.Logger.Error(err)
		return
	}
	userVote := i.UserRepo.GetVote(ss.Login, postID)

	i.Vote(ss, post, -1, userVote)
	err = json.NewEncoder(w).Encode(posts.ConstructPostToFront(post))
	if err != nil {
		i.Logger.Error(err)
		return
	}
}

func (i *ItemsHandler) PostUnVote(w http.ResponseWriter, req *http.Request) {
	postID := mux.Vars(req)["post_id"]

	post, ok := i.ItemsRepo.FindPost(postID)
	if !ok {
		i.Logger.Infof("Пост  не найден %s", postID)
		return
	}

	ss, err := session.SessionFromContext(req.Context())
	if err != nil {
		i.Logger.Error(err)
		return
	}
	userVote := i.UserRepo.GetVote(ss.Login, postID)
	switch userVote {
	case -1:
		post.Score++
	case 1:
		post.Score--
		post.UpVote--
	}

	if post.Score != 0 {
		post.UpvotePercentage = (post.UpVote / post.Score) * 100
	} else {
		post.UpvotePercentage = 0
	}

	err = i.UserRepo.DeleteVote(ss.Login, postID)
	if err != nil {
		i.Logger.Error(err)
		return
	}
	delete(post.Votes, ss.Login)
	err = json.NewEncoder(w).Encode(posts.ConstructPostToFront(post))
	if err != nil {
		i.Logger.Error(err)
	}
}

func (i *ItemsHandler) Vote(ss *session.Session, post *posts.Post, voteValue, userVote int) {
	// Если оценка не изменилась
	if voteValue == userVote {
		return
	}
	// Если оценка поменялась на противоположную
	if voteValue == -userVote {
		post.Score += 2 * voteValue
		post.UpVote += voteValue
		post.UpvotePercentage = (post.UpVote / post.Score) * 100
		post.Votes[ss.Login].Vote = voteValue
		err := i.UserRepo.SetVote(ss.Login, post.ID, voteValue)
		if err != nil {
			i.Logger.Error(err)
			return
		}
		return
	}
	// Если оценки не было
	post.Score += voteValue
	post.UpVote += voteValue
	if post.Score != 0 {
		post.UpvotePercentage = (post.UpVote / post.Score) * 100
	} else {
		post.UpvotePercentage = 0
	}

	vote := posts.Vote{User: ss.UserID, Vote: voteValue}
	post.Votes[ss.Login] = &vote
	err := i.UserRepo.AddVote(ss.Login, post.ID, &vote)
	if err != nil {
		i.Logger.Error(err)
		return
	}

}
