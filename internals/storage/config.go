package storage

import (
	"log"
	"os"
)

type InfluxConfig struct {
	URL    string
	Token  string
	Org    string
	Bucket string
}

func LoadInfluxConfig() *InfluxConfig {
	url := os.Getenv("INFLUXDB_URL")
	token := os.Getenv("INFLUXDB_TOKEN")
	org := os.Getenv("INFLUXDB_ORG")
	bucket := os.Getenv("INFLUXDB_BUCKET")

	if url == "" || token == "" || org == "" || bucket == "" {
		log.Fatal("InfluxDB env variables are not set")
	}

	cfg := &InfluxConfig{
		URL:    url,
		Token:  token,
		Org:    org,
		Bucket: bucket,
	}

	return cfg
}
