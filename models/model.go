package models

import "time"

// Tweet holds the current tweet data of a single tweet
type Tweet struct {
	ID        string
	User      *User
	Message   string
	Reactions []Reaction
	Comments  []Comment
	UpdatedAt time.Time
	CreatedAt time.Time
}

type User struct {
	UserID string
	Name   string
	Status string
}

//Reactions holds the reactions to a single tweet
//ReactionType shows it could either be like or retweet
type Reaction struct {
	User         *User
	ReactionType string
	Count        int
}

//Comments holds individual comment to a post
type Comment struct {
	User     *User
	Content  string
	PostedAt time.Time
}
