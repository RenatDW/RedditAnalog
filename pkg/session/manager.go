package session

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

type SessionsManager struct {
	data map[string]*Session
	mu   *sync.RWMutex
}

func NewSessionsManager() *SessionsManager {
	return &SessionsManager{
		data: make(map[string]*Session, 10),
		mu:   &sync.RWMutex{},
	}
}

func (sm *SessionsManager) Check(r *http.Request) (*Session, error) {
	sessionCookie, err := r.Cookie("session_id")
	if errors.Is(err, http.ErrNoCookie) {
		return nil, ErrNoAuth
	}

	sm.mu.RLock()
	sess, ok := sm.data[sessionCookie.Value]
	sm.mu.RUnlock()

	if !ok {
		return nil, ErrNoAuth
	}

	return sess, nil
}

func (sm *SessionsManager) Create(w http.ResponseWriter, userID string, userLogin string) (*Session, error) {
	sess := NewSession(userID, userLogin)

	sm.mu.Lock()
	_, ok := sm.data[sess.ID]
	for ok {
		sess = NewSession(userID, userLogin)
		_, ok = sm.data[sess.ID]
	}
	sm.data[sess.ID] = sess
	sm.mu.Unlock()

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   sess.ID,
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
	delete(sm.data, sess.ID)
	sm.mu.Unlock()

	cookie := http.Cookie{
		Name:    "session_id",
		Expires: time.Now().AddDate(0, 0, -1),
		Path:    "/",
	}
	http.SetCookie(w, &cookie)
	return nil
}
