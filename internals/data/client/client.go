package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Udehlee/tweet-stream/models"
)

type Client struct {
	httpClient *http.Client
	apiURL     string
}

func NewClient() *Client {
	cl := &Client{
		httpClient: &http.Client{},
		apiURL:     os.Getenv("API_URL"),
	}

	return cl
}

// RandomTweet returns quote-like tweets
func (cl *Client) RandomTweet(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cl.apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request %w", err)
	}

	req.Header.Set("Accept", "Application/json")
	resp, err := cl.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var quote models.Quote
	if err := json.Unmarshal(body, &quote); err != nil {
		return "", fmt.Errorf("failed to decode JSON: %w", err)
	}

	quoteLikeTweet := quote.Content
	return quoteLikeTweet, nil
}
