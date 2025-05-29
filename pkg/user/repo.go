package user

import (
	"fmt"
	"gitlab.vk-golang.ru/vk-golang/lectures/05_web_app/99_hw/redditclone/pkg/posts"
	"golang.org/x/crypto/bcrypt"
	"sync"
)

type UserMemoryRepository struct {
	data   map[string]*User
	lastID uint32
	mu     sync.RWMutex
}

func NewUserMemoryRepo() *UserMemoryRepository {
	return &UserMemoryRepository{
		data: map[string]*User{},
		mu:   sync.RWMutex{},
	}
}

func (repo *UserMemoryRepository) Authorize(login, pass string) (User, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	u, ok := repo.data[login]
	if !ok {
		return User{}, fmt.Errorf("пользователь с таким именем не существует")
	}

	if !CheckPasswordHash(pass, u.password) {
		return User{}, fmt.Errorf("неверный логин или пароль")
	}

	return *u, nil
}

func (repo *UserMemoryRepository) SignUp(login, pass string) (User, error) {

	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, ok := repo.data[login]; ok {
		return User{}, fmt.Errorf("пользователь с таким именем уже существует %s", login)
	}

	hashedPassword, err := HashPassword(pass)
	if err != nil {
		return User{}, err
	}

	repo.lastID++
	repo.data[login] = &User{ID: fmt.Sprintf("u%d", repo.lastID), password: hashedPassword, Login: login, userPosts: make(map[string]bool), votes: make(map[string]*posts.Vote)}
	ans := *repo.data[login]
	return ans, nil
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
