package posts

import (
	"strconv"
	"sync"
	"time"
)

type ItemsRepo interface {
	GetAll() []*Post
	AddPost(post *PostToFront)
	AddComment(postID string, comment Comment)
	DeleteComment(postID string, commentID string)
	DeletePost(id string)
	FindPost(postID string) (*Post, bool)
}

type ItemMemoryRepository struct {
	lastID uint32
	Data   map[string]*Post // [PostId]*PostToFront
	mu     sync.RWMutex
}

type Post struct {
	Author           `json:"author"`
	Category         string             `json:"category"`
	Comments         map[string]Comment `json:"comments"`
	Created          time.Time          `json:"created"`
	ID               string             `json:"id"`
	Score            int                `json:"score"`
	UpVote           int
	Title            string           `json:"title"`
	Type             string           `json:"type"`
	Text             string           `json:"text"`
	URL              string           `json:"url"`
	UpvotePercentage int              `json:"upvotePercentage"`
	Views            int              `json:"views"`
	Votes            map[string]*Vote `json:"votes"`
}

func NewMemoryRepo() *ItemMemoryRepository {
	return &ItemMemoryRepository{
		lastID: 0,
		Data:   make(map[string]*Post),
		mu:     sync.RWMutex{},
	}
}

func (i *ItemMemoryRepository) GetAll() []*Post {
	i.mu.RLock()
	defer i.mu.RUnlock()
	k := 0
	posts := make([]*Post, len(i.Data))
	for _, post := range i.Data {
		posts[k] = post
		k++
	}
	return posts
}

func (i *ItemMemoryRepository) AddComment(postID string, comment Comment) {
	i.mu.Lock()
	i.lastID++

	defer i.mu.Unlock()
	sourcePost, ok := i.Data[postID]
	if !ok {
		return
	}
	comment.ID = "l" + strconv.Itoa(int(i.lastID))
	sourcePost.Comments[comment.ID] = comment
}

func (i *ItemMemoryRepository) DeleteComment(postID string, commentID string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	sourcePost, ok := i.Data[postID]
	if !ok {
		return
	}
	delete(sourcePost.Comments, commentID)
}

func (i *ItemMemoryRepository) AddPost(post *PostToFront) {
	i.mu.Lock()
	i.lastID++
	post.ID = strconv.Itoa(int(i.lastID))
	i.Data[post.ID] = createPost(post)
	i.mu.Unlock()
}

func (i *ItemMemoryRepository) DeletePost(id string) {
	i.mu.Lock()
	delete(i.Data, id)
	i.mu.Unlock()
}

func (i *ItemMemoryRepository) FindPost(id string) (*Post, bool) {
	i.mu.RLock()
	sourcePost, ok := i.Data[id]
	i.mu.RUnlock()
	return sourcePost, ok
}
func createPost(front *PostToFront) *Post {
	answer := Post{
		Author:           front.Author,
		Category:         front.Category,
		Comments:         make(map[string]Comment),
		Created:          front.Created,
		ID:               front.ID,
		Score:            front.Score,
		UpVote:           front.UpVote,
		Title:            front.Title,
		Type:             front.Type,
		Text:             front.Text,
		URL:              front.URL,
		UpvotePercentage: front.UpvotePercentage,
		Views:            front.Views,
		Votes:            make(map[string]*Vote),
	}
	return &answer

}
func ConstructPostToFront(post *Post) *PostToFront {
	votes := make([]*Vote, len(post.Votes))
	i := 0
	for _, vote := range post.Votes {
		votes[i] = vote
		i++
	}

	comments := make([]Comment, len(post.Comments))
	i = 0
	for _, comment := range post.Comments {
		comments[i] = comment
		i++
	}
	constructedAnswer := &PostToFront{
		Author:           post.Author,
		Category:         post.Category,
		Comments:         comments,
		Created:          post.Created,
		ID:               post.ID,
		Score:            post.Score,
		UpVote:           post.UpVote,
		Title:            post.Title,
		Type:             post.Type,
		Text:             post.Text,
		URL:              post.URL,
		UpvotePercentage: post.UpvotePercentage,
		Views:            post.Views,
		Votes:            votes,
	}
	return constructedAnswer
}
