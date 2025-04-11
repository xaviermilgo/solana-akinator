package twitter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"wallet-guesser/internal/domain"

	log "github.com/sirupsen/logrus"
)

// Client handles interactions with the Twitter API via Apify
type Client struct {
	apifyToken   string
	httpClient   *http.Client
	timeout      int
	websiteCache map[string]string // URL -> content
	cacheMutex   sync.RWMutex
}

// NewClient creates a new Twitter API client
func NewClient(options ...ClientOption) domain.TwitterService {
	client := &Client{
		timeout:      30, // Default timeout in seconds
		websiteCache: make(map[string]string),
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	client.httpClient = &http.Client{
		Timeout: time.Duration(client.timeout) * time.Second,
	}

	return client
}

// FetchFollowing fetches the accounts a user is following via Apify
func (c *Client) FetchFollowing(username string, limit int, progressCallback domain.ProgressCallback) ([]domain.TwitterUser, error) {
	if c.apifyToken == "" {
		return nil, errors.New("apify token is not set")
	}

	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Fetching users followed by @%s...", username))
	}

	// Prepare the Apify API request
	apifyInput := ApifyInput{
		UserNames:     []string{username},
		MaxFollowers:  200,   // We don't need followers, but api validates this field
		MaxFollowings: limit, // Number of followings to fetch
		GetFollowers:  false, // Don't get followers
		GetFollowing:  true,  // Get following
	}

	inputJSON, err := json.Marshal(apifyInput)
	if err != nil {
		return nil, err
	}

	// Create the request
	apiURL := "https://api.apify.com/v2/acts/kaitoeasyapi~premium-x-follower-scraper-following-data/run-sync-get-dataset-items"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(inputJSON))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Add token as query parameter
	q := req.URL.Query()
	q.Add("token", c.apifyToken)
	req.URL.RawQuery = q.Encode()

	if progressCallback != nil {
		progressCallback("Sending request to Apify...")
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("apify API error (status %d): %s", resp.StatusCode, string(bodyBytes[:min(100, len(bodyBytes))]))
	}

	if progressCallback != nil {
		progressCallback("Processing Apify response...")
	}

	responseContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var apifyResponses []ApifyFollowerResponse
	if err := json.Unmarshal(responseContent, &apifyResponses); err != nil {
		return nil, err
	}

	// Transform to our domain model and extract wallet addresses
	users := make([]domain.TwitterUser, 0, len(apifyResponses))
	for _, accountResp := range apifyResponses {
		// Process account
		user, err := c.processTwitterAccount(accountResp, progressCallback)
		if err != nil {
			log.Warnf("Error processing account @%s: %v", accountResp.Username, err)
			continue
		}
		users = append(users, user)
	}

	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Successfully processed data for %d users followed by @%s", len(users), username))
	}

	return users, nil
}
