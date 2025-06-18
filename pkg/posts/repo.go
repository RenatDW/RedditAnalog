package posts

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	_ "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	_ "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"sync"
	"time"
)

type ItemsRepo interface {
	GetAll() []*Post
	AddPost(post *PostToFront)
	AddComment(postID string, comment Comment) *Post
	DeleteComment(postID string, commentID string) *Post
	DeletePost(id string)
	AddVote(postID string, userID string, vote Vote) *Post
	DeleteVote(postID string, userID string) *Post
	FindPost(postID string) (*Post, bool)
}

type ItemMemoryRepository struct {
	lastID uint32
	//Data   map[string]*Post // [PostId]*PostToFront
	DB  *mongo.Collection
	Ctx context.Context
	mu  sync.RWMutex
}

type Post struct {
	Author           Author             `bson:"author" json:"author"`
	Category         string             `bson:"category" json:"category"`
	Comments         map[string]Comment `bson:"comments" json:"comments"`
	Created          time.Time          `bson:"created" json:"created"`
	ID               string             `bson:"_id" json:"id"`
	Score            int                `bson:"score" json:"score"`
	ScoreCount       int                `bson:"scoreCount" json:"upVote"`
	UpvoteCount      int                `bson:"upVoteCount" json:"upVoteCount"`
	Title            string             `bson:"title" json:"title"`
	Type             string             `bson:"type" json:"type"`
	Text             string             `bson:"text" json:"text"`
	URL              string             `bson:"url" json:"url"`
	UpvotePercentage int                `bson:"upvotePercentage" json:"upvotePercentage"`
	Views            int                `bson:"views" json:"views"`
	Votes            map[string]*Vote   `bson:"votes" json:"votes"`
}

func NewMemoryRepo() *ItemMemoryRepository {
	ctx := context.Background()
	sess, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost"))
	if err != nil {
		panic(err)
	}

	collection := sess.Database("reddit_clone").Collection("posts")

	return &ItemMemoryRepository{
		lastID: 0,
		DB:     collection,
		Ctx:    ctx,
		//Data:   make(map[string]*Post),
		mu: sync.RWMutex{},
	}
}

func (i *ItemMemoryRepository) GetAll() []*Post {
	i.mu.RLock()
	defer i.mu.RUnlock()

	var posts []*Post

	c, err := i.DB.Find(i.Ctx, bson.M{})
	if err != nil {
		panic(err)
	}
	err = c.All(i.Ctx, &posts)
	if err != nil {
		panic(err)
	}
	return posts
}

func (i *ItemMemoryRepository) AddComment(postID string, comment Comment) *Post {

	post, ok := i.FindPost(postID)
	if !ok {
		return nil
	}
	i.mu.Lock()
	commentID := primitive.NewObjectID().Hex()
	comment.ID = commentID
	post.Comments[commentID] = comment
	i.DB.UpdateOne(i.Ctx, bson.M{"_id": postID}, bson.M{"$set": post})
	i.mu.Unlock()
	return post
}

func (i *ItemMemoryRepository) DeleteComment(postID string, commentID string) *Post {
	post, ok := i.FindPost(postID)
	if !ok {
		return nil
	}
	i.mu.Lock()
	delete(post.Comments, commentID)
	i.DB.UpdateOne(i.Ctx, bson.M{"_id": postID}, bson.M{"$set": post})
	i.mu.Unlock()
	return post
}

func (i *ItemMemoryRepository) AddPost(post *PostToFront) {
	//todo Точно ли нужно сохранять контекст в структуре или можно всегда использовать context.Background()
	ans := createPost(post)
	i.mu.Lock()
	postID := primitive.NewObjectID().Hex()
	ans.ID = postID
	post.ID = postID
	i.DB.InsertOne(i.Ctx, &ans)
	i.mu.Unlock()
}

func (i *ItemMemoryRepository) DeletePost(id string) {
	i.mu.Lock()
	i.DB.DeleteOne(i.Ctx, bson.M{"_id": id})
	i.mu.Unlock()
}

func (i *ItemMemoryRepository) AddVote(postID string, userID string, vote Vote) *Post {
	post, ok := i.FindPost(postID)
	if !ok {
		return &Post{}
	}
	i.mu.Lock()
	oldVote := post.Votes[userID]
	processVoteValue(oldVote, post, vote)
	post.UpvotePercentage = recalculateUpVotePercentage(post)
	post.Votes[userID] = &vote
	i.DB.UpdateOne(i.Ctx, bson.M{"_id": postID}, bson.M{"$set": post})
	i.mu.Unlock()
	return post

}

func processVoteValue(oldVote *Vote, post *Post, vote Vote) {
	if vote.Vote == 0 || (oldVote != nil && oldVote.Vote == vote.Vote) {
		return
	}
	if oldVote == nil || oldVote.Vote == 0 {
		post.Score += vote.Vote
		post.ScoreCount++
		if vote.Vote == 1 {
			post.UpvoteCount++
		}
		return
	}
	post.Score += 2 * vote.Vote
	post.UpvoteCount += vote.Vote

}

func (i *ItemMemoryRepository) DeleteVote(postID string, userID string) *Post {
	post, ok := i.FindPost(postID)
	if !ok {
		return &Post{}
	}
	i.mu.Lock()
	post.Score -= post.Votes[userID].Vote
	post.ScoreCount--
	if post.Votes[userID].Vote == 1 {
		post.UpvoteCount--
	}

	post.UpvotePercentage = recalculateUpVotePercentage(post)
	delete(post.Votes, userID)
	i.DB.UpdateOne(i.Ctx, bson.M{"_id": postID}, bson.M{"$set": post})
	i.mu.Unlock()
	return post
}

func recalculateUpVotePercentage(post *Post) int {
	percent := int((float32(post.UpvoteCount) / float32(post.ScoreCount)) * 100)
	if percent < 0 {
		percent = 0
	}
	return percent
}

func (i *ItemMemoryRepository) FindPost(id string) (*Post, bool) {
	i.mu.RLock()
	var post *Post
	err := i.DB.FindOne(i.Ctx, bson.M{"_id": id}).Decode(&post)
	if err != nil {
		log.Println(err)
		return nil, false
	}
	i.mu.RUnlock()
	return post, true
}
func createPost(front *PostToFront) *Post {
	answer := Post{
		Author:           front.Author,
		Category:         front.Category,
		Comments:         make(map[string]Comment),
		Created:          front.Created,
		ID:               front.ID,
		Score:            front.Score,
		ScoreCount:       front.UpVote,
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
		UpVote:           post.ScoreCount,
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
