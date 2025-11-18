package processor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand/v2"
	"time"

	"github.com/Udehlee/tweet-stream/gapi"
	"github.com/Udehlee/tweet-stream/models"
	pb "github.com/Udehlee/tweet-stream/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type StreamProcessor struct {
	grpcTarget    string
	aggregateChan chan<- *models.Tweet
	dialOpts      []grpc.DialOption
	Timeout       time.Duration
}

func NewProcessor(grpcTarget string, aggChan chan<- *models.Tweet) *StreamProcessor {
	return &StreamProcessor{
		grpcTarget:    grpcTarget,
		aggregateChan: aggChan,
		dialOpts:      []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
		Timeout:       0,
	}
}

// Run connects to gRPC and starts processing
func (p *StreamProcessor) Run(ctx context.Context) error {
	attempt := 0 //keep track of how many times we've retried

	for {
		attempt++

		conn, err := grpc.NewClient(p.grpcTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logFailure("grpc dial failed", err)
			if err := retry(ctx, attempt); err != nil {
				return err
			}
			continue
		}

		attempt = 0 // reset retries after successful connection
		streamCtx, cancel := context.WithCancel(ctx)
		client := pb.NewTweetServiceClient(conn)

		stream, err := client.StreamTweets(streamCtx, &pb.Empty{})
		if err != nil {
			logFailure("failed to create stream", err)
			cleanup(cancel, conn)
			if err := retry(ctx, 1); err != nil {
				return err
			}
			continue
		}

		if err := p.processStream(stream); err != nil {
			logFailure("process stream", err)
			cleanup(cancel, conn)
			if err := retry(ctx, 1); err != nil {
				return err
			}
			continue
		}

		log.Println("processor: stream processing finished cleanly, restarting connection.")
		cleanup(cancel, conn)
	}
}

// processStream recieves and forwards tweets in the model tweet format
func (p *StreamProcessor) processStream(stream pb.TweetService_StreamTweetsClient) error {
	converter := gapi.NewConverter()

	for {
		msg, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("processor: stream closed by server")
				return nil
			}
			return err
		}

		modelTweet := converter.Convert(msg).(*models.Tweet)
		p.forwardStreams(modelTweet)
	}
}

// forwardStreams forwards model tweets to the aggregator channel
func (p *StreamProcessor) forwardStreams(tweet *models.Tweet) {
	if p.aggregateChan == nil {
		return
	}

	select {
	case p.aggregateChan <- tweet:
	default:
		log.Println("processor: aggregateChan full, dropping tweet")
	}
}

// cleanup cancels the context and closes the gRPC connection
func cleanup(cancel context.CancelFunc, conn *grpc.ClientConn) {
	if cancel != nil {
		cancel()
	}

	if conn != nil {
		if err := conn.Close(); err != nil {
			log.Printf("cleanup: failed to close gRPC connection: %v", err)
		}
	}
}

// retry waits before the next retry attempt using exponential backoff
func retry(ctx context.Context, attempt int) error {
	const (
		MaxRetries = 10
		BaseDelay  = 500 * time.Millisecond
		MaxDelay   = 30 * time.Second
	)

	if attempt > MaxRetries {
		return fmt.Errorf("retry: exceeded max retries (%d)", MaxRetries)
	}

	delay := BaseDelay * time.Duration(math.Pow(2, float64(attempt-1)))
	if delay > MaxDelay {
		delay = MaxDelay
	}

	jitter := time.Duration(rand.Float64() * float64(delay) * 0.2)
	delay += jitter

	// Wait or exit early if context is cancelled
	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func logFailure(failureType string, err error) {
	log.Printf("processor: %s error: %v reconnecting", failureType, err)
}
