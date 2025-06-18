package session

import (
	"context"
	"errors"
	"time"
)

type sqlTime []byte

func (s sqlTime) Time() (time.Time, error) {
	return time.Parse("15:04:05", string(s))
}

type Session struct {
	Token     string `gorm:"primary_key"`
	Login     string
	UserID    string
	IsActive  bool
	CreatedAt time.Time `gorm:"type:timestamp"`
	ExpiresAt time.Time `gorm:"type:timestamp"`
}

func NewSession(userID string, userLogin string) *Session {
	// лучше генерировать из заданного алфавита, но так писать меньше и для учебного примера ОК
	//randID := make([]byte, 16)
	//_, err := rand.Read(randID)
	//if err != nil {
	//	return nil
	//}
	return &Session{
		Login:     userLogin,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
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
