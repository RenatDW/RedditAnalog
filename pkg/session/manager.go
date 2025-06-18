package session

import (
	"errors"
	"github.com/jinzhu/gorm"
	"net/http"
	"sync"
	"time"
)

type SessionsManager struct {
	DB *gorm.DB
	mu *sync.RWMutex
}

func NewSessionsManager() *SessionsManager {
	dsn := "root:@tcp(localhost:3306)/reddit_clone?parseTime=true"
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	db.DB()
	db.DB().Ping()
	return &SessionsManager{
		DB: db,
		//data: make(map[string]*Session, 10),
		mu: &sync.RWMutex{},
	}
}

func (sm *SessionsManager) Check(r *http.Request) (*Session, error) {
	sessionCookie, err := r.Cookie("token")
	if errors.Is(err, http.ErrNoCookie) {
		return nil, ErrNoAuth
	}
	var sess Session
	sm.mu.RLock()
	if result := sm.DB.Where("token = ?", sessionCookie.Value).First(&sess); result.Error != nil {
		return nil, result.Error
	}
	sm.mu.RUnlock()

	return &sess, nil
}

func (sm *SessionsManager) Create(w http.ResponseWriter, token string, userID string, userLogin string) (*Session, error) {
	sess := NewSession(userID, userLogin)
	sess.Token = token
	sm.mu.Lock()
	if result := sm.DB.Create(sess); result.Error != nil {
		return &Session{}, result.Error
	}

	sm.mu.Unlock()

	cookie := &http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: time.Now().Add(90 * 24 * time.Hour),
		Path:    "/",
	}
	http.SetCookie(w, cookie)
	return sess, nil
}

func (sm *SessionsManager) DestroyCurrent(w http.ResponseWriter, r *http.Request) error {
	sess, err := SessionFromContext(r.Context())
	if err != nil {
		return err
	}
	sm.mu.Lock()
	sm.DB.Delete(&sess)
	sm.mu.Unlock()

	cookie := http.Cookie{
		Name:    "session_id",
		Expires: time.Now().AddDate(0, 0, -1),
		Path:    "/",
	}
	http.SetCookie(w, &cookie)
	return nil
}
