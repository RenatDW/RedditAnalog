package user

import (
	"cmd/redditclone/pkg/posts"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"sync"
)

type UserMemoryRepository struct {
	DB   *gorm.DB
	data map[string]*User
	mu   sync.RWMutex
}

func NewUserMemoryRepo() *UserMemoryRepository {
	dsn := "root:@tcp(localhost:3306)/reddit_clone"
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	db.DB()
	db.DB().Ping()
	return &UserMemoryRepository{
		DB:   db,
		data: map[string]*User{},
		mu:   sync.RWMutex{},
	}
}

func (repo *UserMemoryRepository) Authorize(login, pass string) (User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	var user User
	if result := repo.DB.Where("login = ?", login).First(&user); result.Error != nil {
		return User{}, result.Error
	}

	if !CheckPasswordHash(pass, user.Password) {
		return User{}, fmt.Errorf("неверный логин или пароль")
	}

	return user, nil
}

func (repo *UserMemoryRepository) SignUp(login, pass string) (User, error) {

	repo.mu.Lock()
	defer repo.mu.Unlock()

	hashedPassword, err := HashPassword(pass)
	if err != nil {
		return User{}, err
	}
	user := User{Login: login, Password: hashedPassword}
	if result := repo.DB.Create(&user); result.Error != nil {
		return User{}, result.Error
	}
	return user, nil
}

func (repo *UserMemoryRepository) AddPost(login, postID string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if _, ok := repo.data[login]; !ok {
		return fmt.Errorf("пользователь с таким именем уже существует %s", login)
	}
	repo.data[login].userPosts[postID] = true
	return nil
}

func (repo *UserMemoryRepository) DeletePost(login, postID string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, ok := repo.data[login]; !ok {
		return fmt.Errorf("пользователь с таким именем уже существует %s", login)
	}
	repo.data[login].userPosts[postID] = false
	return nil
}

func (repo *UserMemoryRepository) AddVote(login, postID string, vote *posts.Vote) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if _, ok := repo.data[login]; !ok {
		return fmt.Errorf("пользователь с таким именем уже существует %s", login)
	}
	repo.data[login].votes[postID] = vote
	return nil
}

func (repo *UserMemoryRepository) SetVote(login, postID string, voteValue int) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if _, ok := repo.data[login]; !ok {
		return fmt.Errorf("пользователь с таким именем уже существует %s", login)
	}
	repo.data[login].votes[postID].Vote = voteValue
	return nil
}

func (repo *UserMemoryRepository) DeleteVote(login, postID string) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if _, ok := repo.data[login]; !ok {
		return fmt.Errorf("пользователь с таким именем уже существует %s", login)
	}
	delete(repo.data[login].votes, postID)
	return nil
}

func (repo *UserMemoryRepository) GetUserPosts(login string) []string {
	repo.mu.RLock()
	postsSeq := repo.data[login].userPosts
	repo.mu.RUnlock()

	posts := make([]string, len(postsSeq))
	for post := range postsSeq {
		posts = append(posts, post)
	}
	return posts
}

func (repo *UserMemoryRepository) GetUserVotes(login string) []string {
	repo.mu.RLock()
	votesSeq := repo.data[login].votes
	repo.mu.RUnlock()

	votes := make([]string, len(votesSeq))
	for vote := range votesSeq {
		votes = append(votes, vote)
	}
	return votes
}

func (repo *UserMemoryRepository) GetVote(login, postID string) int {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	vote, ok := repo.data[login].votes[postID]
	if !ok {
		return 0
	}
	return vote.Vote
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
