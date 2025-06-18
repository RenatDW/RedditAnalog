package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gitlab.vk-golang.ru/vk-golang/lectures/06_databases/99_hw/db/redditclone/pkg/posts"
	"gitlab.vk-golang.ru/vk-golang/lectures/06_databases/99_hw/db/redditclone/pkg/session"
	"net/http"
)

func (i *ItemsHandler) PostUpVote(w http.ResponseWriter, req *http.Request) {
	ChangeVote(w, req, i, 1)
}

func (i *ItemsHandler) PostDownVote(w http.ResponseWriter, req *http.Request) {
	ChangeVote(w, req, i, -1)
}

func ChangeVote(w http.ResponseWriter, req *http.Request, i *ItemsHandler, voteValue int) {
	postID := mux.Vars(req)["post_id"]

	ss, err := session.SessionFromContext(req.Context())
	if err != nil {
		i.Logger.Error(err)
		return
	}

	newVote := posts.Vote{
		User: ss.UserID,
		Vote: voteValue,
	}

	post := i.ItemsRepo.AddVote(postID, ss.UserID, newVote)

	err = json.NewEncoder(w).Encode(posts.ConstructPostToFront(post))
	if err != nil {
		i.Logger.Error(err)
		return
	}
}

func (i *ItemsHandler) PostUnVote(w http.ResponseWriter, req *http.Request) {
	postID := mux.Vars(req)["post_id"]

	ss, err := session.SessionFromContext(req.Context())
	if err != nil {
		i.Logger.Error(err)
		return
	}

	post := i.ItemsRepo.DeleteVote(postID, ss.UserID)

	err = json.NewEncoder(w).Encode(posts.ConstructPostToFront(post))
	if err != nil {
		i.Logger.Error(err)
		return
	}
}
