package grpc

import (
	"github.com/Udehlee/tweet-stream/models"
	"github.com/Udehlee/tweet-stream/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(u *models.User) *pb.User {
	return &pb.User{
		UserId: u.UserID,
		Name:   u.Name,
		Status: u.Status,
	}
}

func convertComment(c models.Comment) *pb.Comment {
	return &pb.Comment{
		User:     convertUser(c.User),
		Content:  c.Content,
		PostedAt: timestamppb.New(c.PostedAt),
	}
}

func convertReaction(r models.Reaction) *pb.Reaction {
	return &pb.Reaction{
		User:         convertUser(r.User),
		ReactionType: r.ReactionType,
		Count:        int32(r.Count),
	}
}

func convertTweet(t *models.Tweet) *pb.Tweet {
	comments := make([]*pb.Comment, len(t.Comments))
	for i, c := range t.Comments {
		comments[i] = convertComment(c)
	}

	reactions := make([]*pb.Reaction, len(t.Reactions))
	for i, r := range t.Reactions {
		reactions[i] = convertReaction(r)
	}

	return &pb.Tweet{
		Id:        t.ID,
		User:      convertUser(t.User),
		Message:   t.Message,
		Hashtags:  t.HashTag,
		Comments:  comments,
		Reactions: reactions,
		CreatedAt: timestamppb.New(t.CreatedAt),
		UpdatedAt: timestamppb.New(t.UpdatedAt),
	}
}
