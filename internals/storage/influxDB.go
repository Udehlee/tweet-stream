package storage

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Udehlee/tweet-stream/models"
	"github.com/Udehlee/tweet-stream/utils"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type InfluxWriter struct {
	Client influxdb2.Client
	Org    string
	Bucket string
}

// ConnectToInfluxDB loads the config from env and connects to InfluxDB
func ConnectToInfluxDB() (*InfluxWriter, error) {
	cfg := LoadInfluxConfig()
	client := influxdb2.NewClient(cfg.URL, cfg.Token)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := client.Health(ctx)
	if err != nil {
		return nil, fmt.Errorf("InfluxDB health check request failed: %v", err)
	}

	if health.Status != "pass" {
		return nil, fmt.Errorf("InfluxDB health check failed: %s", *health.Message)
	}

	fmt.Printf("InfluxDB connection established. Status: %s\n", health.Status)

	influx := &InfluxWriter{
		Client: client,
		Org:    cfg.Org,
		Bucket: cfg.Bucket,
	}

	return influx, nil
}

// Close closes the InfluxDB client
func (iw *InfluxWriter) Close() {
	iw.Client.Close()
	fmt.Println("InfluxDB client closed.")
}

// Insert saves WindowMetrics point to InfluxDB
func (iw *InfluxWriter) Insert(metrics models.WindowMetrics) error {
	writeAPI := iw.Client.WriteAPIBlocking(iw.Org, iw.Bucket)
	hashtags := utils.ExtractHashtags(metrics.TrendingHashtags)

	point := influxdb2.NewPoint(
		"tweet_metrics",
		map[string]string{
			"source": "tweet_stream",
		},
		map[string]interface{}{
			"total_tweets":      metrics.TotalTweets,
			"total_engagement":  metrics.TotalEngagement,
			"min_latency_ms":    metrics.MinLatency.Milliseconds(),
			"max_latency_ms":    metrics.MaxLatency.Milliseconds(),
			"avg_latency_ms":    metrics.AvgLatency.Milliseconds(),
			"is_anomaly":        metrics.IsAnomaly,
			"trending_hashtags": strings.Join(hashtags, ","),
		},
		metrics.WindowEnd,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := writeAPI.WritePoint(ctx, point); err != nil {
		log.Println("error writing to InfluxDB:", err)
		return err
	}

	return nil
}
