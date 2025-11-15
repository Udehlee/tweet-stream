package generator

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Udehlee/tweet-stream/data/client"
	"github.com/Udehlee/tweet-stream/data/simulated"
	"github.com/Udehlee/tweet-stream/models"
	"github.com/rs/zerolog"
)

type GeneratorService struct {
	tweetSvc   *simulated.TweetService
	client     *client.Client
	logger     *zerolog.Logger
	mu         sync.Mutex
	tweetsChan chan tweetEvent
	done       chan struct{}
}

func NewGeneratorService(tweetSvc *simulated.TweetService, client *client.Client, logger *zerolog.Logger) *GeneratorService {
	gs := &GeneratorService{
		tweetSvc:   tweetSvc,
		client:     client,
		logger:     logger,
		tweetsChan: make(chan tweetEvent, 50),
		done:       make(chan struct{}),
	}

	return gs
}

type tweetEvent struct {
	Type  string
	Tweet *models.Tweet
}

// PostTweet posts random tweets
func (gs *GeneratorService) PostTweet(ctx context.Context) {
	msg, err := gs.client.RandomTweet(ctx)
	if err != nil {
		msg = "use this tweet take flex"
	}

	tweet, err := gs.tweetSvc.CreateTweet(msg)
	if err != nil {
		gs.logger.Info().Msg("failed to post tweet")
		return
	}

	hashTag := gs.tweetSvc.GenerateHashTags(msg)
	tweet.Message = fmt.Sprintf("%s\n %s", tweet.Message, hashTag)

	create := gs.CreateTweetEvent("create", tweet)

	gs.tweetsChan <- create
	gs.logger.Info().Msgf("New tweet posted %s", tweet.Message)
}

// UpdateRandomTweet updates a random existing tweet
func (gs *GeneratorService) UpdateRandomTweet(ctx context.Context) {
	tweet := gs.tweetSvc.GetTweet()
	if tweet == nil {
		return
	}

	msg, err := gs.client.RandomTweet(ctx)
	if err != nil {
		msg = "use this update hold body"
	}

	err = gs.tweetSvc.UpdateTweet(tweet.ID, msg)
	if err != nil {
		fmt.Println("Failed to update tweet:", err)
		return
	}

	update := gs.CreateTweetEvent("update", tweet)

	gs.tweetsChan <- update
	gs.logger.Info().Msgf("Tweet updated %s", msg)
}

// DeleteRandomTweet deletes a random existing tweet
func (gs *GeneratorService) DeleteRandomTweet(ctx context.Context) {
	tweet := gs.tweetSvc.GetTweet()
	if tweet == nil {
		return
	}

	err := gs.tweetSvc.DeleteTweet(tweet.ID)
	if err != nil {
		fmt.Println("Failed to delete tweet:", err)
		return
	}

	delete := gs.CreateTweetEvent("delete", tweet)

	gs.tweetsChan <- delete
	gs.logger.Info().Msgf("Tweet deleted %s", tweet.ID)
}

// CreateTweetEvent creates a tweet Event with the given type and tweet payload.
func (gs *GeneratorService) CreateTweetEvent(eventType string, tweet *models.Tweet) tweetEvent {
	TweetEvt := tweetEvent{
		Type:  eventType,
		Tweet: tweet,
	}

	return TweetEvt
}

// GenerateTweets generates random fake tweet operations at intervals
func (gs *GeneratorService) GenerateTweets(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	opts := []func(context.Context){
		gs.PostTweet,
		gs.UpdateRandomTweet,
		gs.DeleteRandomTweet,
	}

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				gs.mu.Lock()
				var op func(context.Context)
				op = opts[rand.Intn(len(opts))]
				op(ctx)
				gs.mu.Unlock()
			case <-gs.done:
				close(gs.tweetsChan)
				return
			case <-ctx.Done():
				close(gs.tweetsChan)
				return
			}
		}
	}()
}
