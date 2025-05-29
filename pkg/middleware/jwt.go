package middleware

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/user"
	"io"
	"log"
	"net/http"
	"time"
)

var ExampleTokenSecret = []byte("супер секретный ключ")

func GenerateJWTToken(w http.ResponseWriter, user user.User) ([]byte, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": map[string]interface{}{
			"username": user.Login,
			"id":       user.ID,
		},
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString(ExampleTokenSecret)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return nil, err
	}

	resp, err := json.Marshal(map[string]interface{}{
		"token": tokenString,
	})
	if err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return nil, err
	}
	return resp, nil
}

func JSONError(w io.Writer, status int, msg string) {
	resp, err := json.Marshal(map[string]interface{}{
		"status": status,
		"error":  msg,
	})
	if err != nil {
		log.Println(err)
		return
	}
	_, err = w.Write(resp)
	if err != nil {
		log.Println("JSONError:", err)
	}
}
