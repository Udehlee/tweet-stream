package aggregator

import (
	"context"
	"log"
	"sort"
	"time"

	"github.com/Udehlee/tweet-stream/internals/storage"
	"github.com/Udehlee/tweet-stream/models"
)

type TweetAggregator struct {
	Batch          []*models.Tweet
	WindowDuration time.Duration
	InChan         <-chan *models.Tweet
	InfluxWriter   *storage.InfluxWriter
}

func NewTweetAggregator(in <-chan *models.Tweet, writer *storage.InfluxWriter, duration time.Duration) *TweetAggregator {
	return &TweetAggregator{
		Batch:          make([]*models.Tweet, 0, 100), // preallocate some space
		WindowDuration: duration,
		InChan:         in,
		InfluxWriter:   writer,
	}
}

// Start begins the aggregation window loop
func (t *TweetAggregator) Start(ctx context.Context) {
	ticker := time.NewTicker(t.WindowDuration)
	defer ticker.Stop()

	windowStart := time.Now()

	for {
		select {
		case <-ctx.Done():
			if len(t.Batch) > 0 {
				t.processBatch(windowStart)
			}
			log.Println("Aggregator received context cancellation. Shutting down gracefully.")
			return

		case tweet, ok := <-t.InChan:
			if !ok {
				if len(t.Batch) > 0 {
					t.processBatch(windowStart)
				}
				return
			}
			t.Batch = append(t.Batch, tweet)

		case windowEnd := <-ticker.C:
			if len(t.Batch) > 0 {
				t.processBatch(windowStart)
			} else {
				log.Println("window passed with zero tweets")
			}
			t.Batch = t.Batch[:0] // reuse slice memory
			windowStart = windowEnd
		}
	}
}

// processBatch calculates metrics
// and writes them to InfluxDB
func (t *TweetAggregator) processBatch(windowStart time.Time) {
	metrics := models.WindowMetrics{
		WindowStart: windowStart,
		WindowEnd:   time.Now(),
		TotalTweets: len(t.Batch),
		MinLatency:  time.Hour,
		MaxLatency:  0,
	}

	if metrics.TotalTweets == 0 {
		return
	}

	hashtagCounts := make(map[string]int)
	var totalLatency time.Duration

	for _, tweet := range t.Batch {
		t.countHashtags(tweet, hashtagCounts)
		t.EngagementStats(tweet, &metrics)
		t.calculateVerifiedStatus(tweet, &metrics)

		latency := t.calculateLatency(tweet, &metrics)
		totalLatency += latency
	}

	metrics.AvgLatency = totalLatency / time.Duration(metrics.TotalTweets)
	metrics.TrendingHashtags = t.findTrendingHashtags(hashtagCounts, 5)
	metrics.IsAnomaly = t.detectAnomaly(&metrics)

	if err := t.InfluxWriter.Insert(metrics); err != nil {
		log.Println("Failed to write metrics to InfluxDB:", err)
	}

	log.Printf("Batch Processed: Tweets=%d, Engagement=%d, Anomaly=%t, AvgLatency=%s, MaxLatency=%s\n",
		metrics.TotalTweets, metrics.TotalEngagement, metrics.IsAnomaly,
		metrics.AvgLatency.Round(time.Millisecond), metrics.MaxLatency.Round(time.Millisecond))
}

// countHashtags counts hashtags for a tweet
func (t *TweetAggregator) countHashtags(tweet *models.Tweet, counts map[string]int) {
	for _, tag := range tweet.HashTag {
		counts[tag]++
	}
}

// EngagementStats sums reactions and comments
func (t *TweetAggregator) EngagementStats(tweet *models.Tweet, metrics *models.WindowMetrics) {
	likes, retweets := 0, 0
	for _, r := range tweet.Reactions {
		switch r.ReactionType {
		case "like":
			likes += r.Count
		case "retweet":
			retweets += r.Count
		}
	}
	metrics.TotalEngagement += likes + retweets + len(tweet.Comments)
}

// calculateVerifiedStatus counts verified and unverified users
func (t *TweetAggregator) calculateVerifiedStatus(tweet *models.Tweet, metrics *models.WindowMetrics) {
	if tweet.User != nil && tweet.User.Status == "verified" {
		metrics.VerifiedCount++
	} else {
		metrics.UnverifiedCount++
	}
}

// calculateLatency updates min/max metrics and returns the latency
func (t *TweetAggregator) calculateLatency(tweet *models.Tweet, metrics *models.WindowMetrics) time.Duration {
	latency := time.Since(tweet.CreatedAt)
	if latency > metrics.MaxLatency {
		metrics.MaxLatency = latency
	}
	if latency < metrics.MinLatency {
		metrics.MinLatency = latency
	}
	return latency
}

// detectAnomaly checks thresholds for anomalies
func (t *TweetAggregator) detectAnomaly(metrics *models.WindowMetrics) bool {
	const (
		TWEET_VOLUME_THRESHOLD     = 50
		LATENCY_SPIKE_MULTIPLIER   = 3.0
		ENGAGEMENT_BURST_THRESHOLD = 1000
	)

	if metrics.TotalTweets > TWEET_VOLUME_THRESHOLD {
		log.Printf("Anomaly: High tweet volume (%d > %d)", metrics.TotalTweets, TWEET_VOLUME_THRESHOLD)
		return true
	}

	if metrics.TotalTweets > 0 {
		avg := float64(metrics.AvgLatency.Milliseconds())
		max := float64(metrics.MaxLatency.Milliseconds())
		if max > avg*LATENCY_SPIKE_MULTIPLIER {
			log.Printf("Anomaly: Latency spike (Max %s > %.1fx Avg %s)", metrics.MaxLatency.Round(time.Millisecond), LATENCY_SPIKE_MULTIPLIER, metrics.AvgLatency.Round(time.Millisecond))
			return true
		}
	}

	if metrics.TotalEngagement > ENGAGEMENT_BURST_THRESHOLD {
		log.Printf("Anomaly: High engagement (%d > %d)", metrics.TotalEngagement, ENGAGEMENT_BURST_THRESHOLD)
		return true
	}

	return false
}

// findTrendingHashtags returns top trending hashtags
func (t *TweetAggregator) findTrendingHashtags(counts map[string]int, n int) []models.HashtagCount {
	tags := make([]models.HashtagCount, 0, len(counts))
	for tag, count := range counts {
		tags = append(tags, models.HashtagCount{Tag: tag, Count: count})
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Count > tags[j].Count
	})

	if len(tags) > n {
		return tags[:n]
	}
	return tags
}
