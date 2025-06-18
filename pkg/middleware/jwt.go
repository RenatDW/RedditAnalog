package middleware

import (
	"cmd/redditclone/pkg/user"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

var ExampleTokenSecret = []byte("супер секретный ключ")

func GenerateJWTToken(w http.ResponseWriter, user user.User) ([]byte, string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": map[string]interface{}{
			"username": user.Login,
			"id":       strconv.Itoa(user.ID),
		},
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString(ExampleTokenSecret)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return nil, "", err
	}

	resp, err := json.Marshal(map[string]interface{}{
		"token": tokenString,
	})
	if err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return nil, "", err
	}
	return resp, tokenString, nil
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
