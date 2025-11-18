package gapi

import (
	"github.com/Udehlee/tweet-stream/models"
	"github.com/Udehlee/tweet-stream/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Converter struct {
	ToProto bool
}

type ConvertOption func(*Converter)

func NewConverter(opts ...ConvertOption) *Converter {
	c := &Converter{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithProto() ConvertOption {
	return func(c *Converter) {
		c.ToProto = true
	}
}

// Convert converts between model and proto based on the given supported types
func (c *Converter) Convert(i interface{}) interface{} {
	switch v := i.(type) {
	case *models.User:
		if c.ToProto {
			return &pb.User{
				UserId: v.UserID,
				Name:   v.Name,
				Status: v.Status,
			}
		}
	case *pb.User:
		if !c.ToProto {
			return &models.User{
				UserID: v.UserId,
				Name:   v.Name,
				Status: v.Status,
			}
		}

	case models.Comment:
		if c.ToProto {
			return &pb.Comment{
				User:     c.Convert(&v.User).(*pb.User),
				Content:  v.Content,
				PostedAt: timestamppb.New(v.PostedAt),
			}
		}
	case *pb.Comment:
		if !c.ToProto {
			return models.Comment{
				User:     c.Convert(v.User).(*models.User),
				Content:  v.Content,
				PostedAt: v.PostedAt.AsTime(),
			}
		}

	case models.Reaction:
		if c.ToProto {
			return &pb.Reaction{
				User:         c.Convert(&v.User).(*pb.User),
				ReactionType: v.ReactionType,
				Count:        int32(v.Count),
			}
		}
	case *pb.Reaction:
		if !c.ToProto {
			return models.Reaction{
				User:         c.Convert(v.User).(*models.User),
				ReactionType: v.ReactionType,
				Count:        int(v.Count),
			}
		}

	case *models.Tweet:
		if c.ToProto {
			comments := make([]*pb.Comment, len(v.Comments))
			for i, cm := range v.Comments {
				comments[i] = c.Convert(cm).(*pb.Comment)
			}
			reactions := make([]*pb.Reaction, len(v.Reactions))
			for i, r := range v.Reactions {
				reactions[i] = c.Convert(r).(*pb.Reaction)
			}
			return &pb.Tweet{
				Id:        v.ID,
				User:      c.Convert(v.User).(*pb.User),
				Message:   v.Message,
				Hashtags:  v.HashTag,
				Comments:  comments,
				Reactions: reactions,
				CreatedAt: timestamppb.New(v.CreatedAt),
				UpdatedAt: timestamppb.New(v.UpdatedAt),
			}
		}

	case *pb.Tweet:
		if !c.ToProto {
			comments := make([]models.Comment, len(v.Comments))
			for i, cm := range v.Comments {
				comments[i] = c.Convert(cm).(models.Comment)
			}
			reactions := make([]models.Reaction, len(v.Reactions))
			for i, r := range v.Reactions {
				reactions[i] = c.Convert(r).(models.Reaction)
			}
			return &models.Tweet{
				ID:        v.Id,
				User:      c.Convert(v.User).(*models.User),
				Message:   v.Message,
				HashTag:   v.Hashtags,
				Comments:  comments,
				Reactions: reactions,
				CreatedAt: v.CreatedAt.AsTime(),
				UpdatedAt: v.UpdatedAt.AsTime(),
			}
		}
	}

	return nil
}
