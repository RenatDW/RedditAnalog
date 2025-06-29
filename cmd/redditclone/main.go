package main

import (
	"cmd/redditclone/pkg/handlers"
	"cmd/redditclone/pkg/middleware"
	"cmd/redditclone/pkg/posts"
	"cmd/redditclone/pkg/session"
	"cmd/redditclone/pkg/user"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"html/template"
	"net/http"
)

const (
	pathToStaticDir = "static"
	pathToIndex     = "static/html/index.html"
)

func main() {

	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func(zapLogger *zap.Logger) {
		err = zapLogger.Sync()
		if err != nil {
			panic(err)
		}
	}(zapLogger)
	logger := zapLogger.Sugar()
	items := posts.NewMemoryRepo()

	userRepo := user.NewUserMemoryRepo()
	sm := session.NewSessionsManager()
	userHandler := handlers.UserHandler{
		UserRepo: userRepo,
		Logger:   logger,
		Sessions: sm,
	}

	handlers := &handlers.ItemsHandler{
		Logger:    logger,
		ItemsRepo: items,
		UserRepo:  userRepo,
	}
	r := mux.NewRouter()

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(pathToStaticDir))))
	tmpl := template.Must(template.ParseFiles(pathToIndex))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err = tmpl.Execute(w, nil)
		if err != nil {
			userHandler.Logger.Error(err)
			return
		}
	})

	r.HandleFunc("/api/login", userHandler.LoginPage)
	r.HandleFunc("/api/register", userHandler.RegisterPage)
	// Guest
	r.HandleFunc("/api/posts/", handlers.Posts).Methods(http.MethodGet)
	r.HandleFunc("/api/post/{post_id}", handlers.PostInfo).Methods(http.MethodGet)
	// User
	r.HandleFunc("/api/posts/{category:music|funny|videos|programming|news|fashion}", handlers.PostsWithCategory).Methods(http.MethodGet)
	r.HandleFunc("/api/posts", handlers.AddPosts).Methods(http.MethodPost)
	r.HandleFunc("/api/post/{post_id}", handlers.CommentAdd).Methods(http.MethodPost)
	r.HandleFunc("/api/post/{post_id}/{comment_id}", handlers.CommentDelete).Methods(http.MethodDelete)
	r.HandleFunc("/api/post/{post_id}/upvote", handlers.PostUpVote).Methods(http.MethodGet)
	r.HandleFunc("/api/post/{post_id}/downvote", handlers.PostDownVote).Methods(http.MethodGet)
	r.HandleFunc("/api/post/{post_id}/unvote", handlers.PostUnVote).Methods(http.MethodGet)
	r.HandleFunc("/api/post/{post_id}", handlers.PostDelete).Methods(http.MethodDelete)
	r.HandleFunc("/api/user/{user_login}", handlers.UserPosts).Methods(http.MethodGet)

	mux := middleware.Auth(sm, r)
	mux = middleware.AccessLog(logger, mux)
	mux = middleware.Panic(logger, mux)
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		userHandler.Logger.Error(err)
		return
	}
}
