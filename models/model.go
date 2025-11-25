package models

import "time"

// Tweet holds the current tweet data of a single tweet
type Tweet struct {
	ID        string
	User      *User
	Message   string
	HashTag   []string
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

type Quote struct {
	Content string `json:"content"`
}

// WindowMetrics holds all calculated results for a single time window
type WindowMetrics struct {
	WindowStart      time.Time
	WindowEnd        time.Time
	TotalTweets      int
	TrendingHashtags []HashtagCount
	TotalEngagement  int
	VerifiedCount    int
	UnverifiedCount  int

	// Latency Metrics
	AvgLatency time.Duration
	MaxLatency time.Duration
	MinLatency time.Duration

	IsAnomaly bool
}

type HashtagCount struct {
	Tag   string
	Count int
}
