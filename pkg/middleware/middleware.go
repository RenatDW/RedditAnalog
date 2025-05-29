package middleware

import (
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/session"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	noAuthUrls = map[string]struct{}{
		"/":              {},
		"/api/posts/":    {},
		"/manifest.json": {},
		"/api/login":     {},
		"/api/register":  {},
	}
	noSessUrls = map[string]struct{}{
		"/": {},
	}
)

func Auth(sm *session.SessionsManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			log.Println("Не нужна авторизация для static:", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}
		if r.Method == http.MethodGet &&
			(strings.HasPrefix(r.URL.Path, "/api/posts/") || strings.HasPrefix(r.URL.Path, "/api/post/") ||
				strings.HasPrefix(r.URL.Path, "/api/user/")) && !strings.Contains(r.URL.Path, "vote") {
			log.Println("Не нужна авторизация для получения информации о постах", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}
		if _, ok := noAuthUrls[r.URL.Path]; ok {
			log.Println("Не нужна авторизация", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}
		sess, err := sm.Check(r)

		_, canbeWithouthSess := noSessUrls[r.URL.Path]
		if err != nil && !canbeWithouthSess {
			log.Println("no auth", r.URL.Path)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		ctx := session.ContextWithSession(r.Context(), sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AccessLog(logger *zap.SugaredLogger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("access log middleware")
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Infow("New request",
			"method", r.Method,
			"remote_addr", r.RemoteAddr,
			"url", r.URL.Path,
			"time", time.Since(start),
		)
	})
}

func Panic(logger *zap.SugaredLogger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("panicMiddleware", r.URL.Path)
		defer func() {
			if err := recover(); err != nil {
				logger.Error("recovered", err)
				http.Error(w, "Internal server error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
