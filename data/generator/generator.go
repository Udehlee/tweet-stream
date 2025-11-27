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
	tweetSvc *simulated.TweetService
	client   *client.Client
	logger   *zerolog.Logger
	mu       sync.Mutex
	Publish  chan<- *models.Tweet
	done     chan struct{}
}

func NewGeneratorService(tweetSvc *simulated.TweetService, client *client.Client, logger *zerolog.Logger, publish chan<- *models.Tweet) *GeneratorService {
	return &GeneratorService{
		tweetSvc: tweetSvc,
		client:   client,
		logger:   logger,
		Publish:  publish,
		done:     make(chan struct{}),
	}
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

	gs.publishTweet(tweet)
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
		gs.logger.Info().Msg("Failed to update tweet")
		return
	}

	tweet.Message = msg
	gs.publishTweet(tweet)
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
		gs.logger.Info().Msg("Failed to delete tweet")
		return
	}

	gs.publishTweet(tweet)
	gs.logger.Info().Msgf("Tweet deleted %s", tweet.ID)
}

// PiblishTweet publishes tweet to gRPC
func (gs *GeneratorService) publishTweet(t *models.Tweet) {
	if gs.Publish == nil {
		return
	}
	select {
	case gs.Publish <- t:
	default:
		gs.logger.Info().Msg("publish channel full, dropped tweet")
	}
}

// GenerateTweets generates random fake tweet operations at intervals
func (gs *GeneratorService) GenerateTweets(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	ops := []func(context.Context){
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
				op := ops[rand.Intn(len(ops))]
				op(ctx)
				gs.mu.Unlock()

			case <-gs.done:
				return

			case <-ctx.Done():
				return
			}
		}
	}()
}
