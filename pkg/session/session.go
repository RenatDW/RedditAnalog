package session

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
)

type Session struct {
	ID     string
	Login  string
	UserID string
}

func NewSession(userID string, userLogin string) *Session {
	// лучше генерировать из заданного алфавита, но так писать меньше и для учебного примера ОК
	randID := make([]byte, 16)
	_, err := rand.Read(randID)
	if err != nil {
		return nil
	}

	return &Session{
		ID:     fmt.Sprintf("%x", randID),
		Login:  userLogin,
		UserID: userID,
	}
}

var (
	ErrNoAuth = errors.New("no session found")
)

type sessKey string

var Key sessKey = "sessionKey"

func SessionFromContext(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value(Key).(*Session)
	if !ok || sess == nil {
		return nil, ErrNoAuth
	}
	return sess, nil
}

func ContextWithSession(ctx context.Context, sess *Session) context.Context {
	return context.WithValue(ctx, Key, sess)
}
