package simulated

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Udehlee/tweet-stream/models"
	"github.com/Udehlee/tweet-stream/utils"
	"github.com/rs/zerolog"
)

type TweetService struct {
	tweet  map[string]*models.Tweet
	logger *zerolog.Logger
	mu     sync.RWMutex
}

func NewTweetService(logger *zerolog.Logger) *TweetService {
	Ts := &TweetService{
		tweet:  make(map[string]*models.Tweet),
		logger: logger,
	}

	return Ts
}

// PostTweet posts a single tweet
func (ts *TweetService) PostTweet(msg string) (*models.Tweet, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	id := utils.GenerateID()
	name := utils.GenerateUser()
	status := ts.isVerifiedOrNot()

	user := &models.User{
		UserID: utils.GenerateID(),
		Name:   name,
		Status: status,
	}

	tweet := &models.Tweet{
		ID:        id,
		User:      user,
		Message:   msg,
		CreatedAt: time.Now(),
	}

	ts.tweet[id] = tweet
	ts.logger.Info().Msgf("New tweet by %s || %s\n %s", user.Name, user.Status, msg)
	return tweet, nil

}

// UpdateTweet updates an existing tweet
func (ts *TweetService) UpdateTweet(tweetId, msg string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	tweet, found := ts.tweet[tweetId]
	if !found {
		return fmt.Errorf("tweet with the id %s is not found", tweetId)
	}

	tweet.Message = msg
	tweet.UpdatedAt = time.Now()

	ts.logger.Info().Msgf("tweet with the id %s has been updated successfully with the msg %s", tweetId, msg)
	return nil
}

// DeleteTweet deletes a tweet
func (ts *TweetService) DeleteTweet(tweetId string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	_, found := ts.tweet[tweetId]
	if !found {
		return fmt.Errorf("tweet with the id %s is not found", tweetId)
	}

	delete(ts.tweet, tweetId)
	ts.logger.Info().Msgf("tweet with the id %s has been deleted successfully", tweetId)
	return nil
}

// isVerifiedOrNot assigns verified or unverified  to user
func (ts *TweetService) isVerifiedOrNot() string {
	statuses := []string{"verified", "unverified"}
	return statuses[rand.Intn(len(statuses))]
}
