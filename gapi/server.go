package gapi

import (
	"github.com/Udehlee/tweet-stream/models"
	pb "github.com/Udehlee/tweet-stream/pb"
)

type StreamServer struct {
	pb.UnimplementedTweetServiceServer
	TweetsChan <-chan *models.Tweet
}

func NewStreamServer(tweetCh <-chan *models.Tweet) *StreamServer {
	return &StreamServer{
		TweetsChan: tweetCh,
	}
}

func (s *StreamServer) StreamTweets(req *pb.Empty, stream pb.TweetService_StreamTweetsServer) error {
	c := NewConverter(WithProto())
	for tweet := range s.TweetsChan {
		protoTweet := c.Convert(tweet).(*pb.Tweet)
		if err := stream.Send(protoTweet); err != nil {
			return err
		}
	}
	return nil
}
