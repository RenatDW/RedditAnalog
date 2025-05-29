package handlers

import (
	"encoding/json"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/middleware"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/session"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/user"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type UserHandler struct {
	Logger   *zap.SugaredLogger
	UserRepo user.UserRepo
	Sessions *session.SessionsManager
}

type LoginForm struct {
	Login    string `json:"username"`
	Password string `json:"password"`
}

func (u *UserHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	userData := &LoginForm{}
	err := json.NewDecoder(r.Body).Decode(userData)
	if err != nil {
		middleware.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	us, err := u.UserRepo.Authorize(userData.Login, userData.Password)
	if err != nil {
		middleware.JSONError(w, http.StatusUnauthorized, err.Error())
		return
	}
	err = u.Sessions.DestroyCurrent(w, r)
	if err != nil {
		log.Println(err)
	}
	_, err = u.Sessions.Create(w, us.ID, us.Login)
	if err != nil {
		log.Println(err)
	}
	resp, err := middleware.GenerateJWTToken(w, us)
	if err != nil {
		return
	}
	u.Logger.Infof("Пользователь авторизовался %v", us)
	_, err = w.Write(resp)
	if err != nil {
		log.Println(err)
		return
	}
}

func (u *UserHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	userData := &LoginForm{}
	err := json.NewDecoder(r.Body).Decode(userData)
	if err != nil {
		middleware.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	us, err := u.UserRepo.SignUp(userData.Login, userData.Password)
	if err != nil {
		middleware.JSONError(w, http.StatusUnauthorized, err.Error())
	}

	resp, err := middleware.GenerateJWTToken(w, us)
	if err != nil {
		log.Println(err)
		return
	}
	u.Logger.Infof("Пользователь зарегистрировался %v", us)
	err = u.Sessions.DestroyCurrent(w, r)
	if err != nil {
		log.Println(err)
	}
	_, err = u.Sessions.Create(w, us.ID, us.Login)
	if err != nil {
		log.Println(err)
	}

	_, err = w.Write(resp)
	if err != nil {
		log.Println(err)
	}

}
