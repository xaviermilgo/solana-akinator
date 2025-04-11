package twitter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// TwitterUser represents a Twitter user with relevant information
type TwitterUser struct {
	Username              string   `json:"username"`
	DisplayName           string   `json:"displayName,omitempty"`
	Bio                   string   `json:"bio,omitempty"`
	Urls                  []string `json:"urls,omitempty"`
	PossibleMintAddresses []string `json:"possibleAddresses,omitempty"`
}

// Client handles interactions with the Twitter API via Apify
type Client struct {
	apifyToken string
	httpClient *http.Client
}

// ClientOption is a functional option for configuring the Twitter client
type ClientOption func(*Client)

// WithApifyToken sets the Apify token
func WithApifyToken(token string) ClientOption {
	return func(c *Client) {
		c.apifyToken = token
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new Twitter API client
func NewClient(options ...ClientOption) *Client {
	client := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	return client
}

// ApifyInput represents the input for the Apify Twitter follower scraper
type ApifyInput struct {
	UserNames     []string `json:"user_names,omitempty"`
	UserIDs       []string `json:"user_ids,omitempty"`
	MaxFollowers  int      `json:"maxFollowers"`
	MaxFollowings int      `json:"maxFollowings"`
	GetFollowers  bool     `json:"getFollowers"`
	GetFollowing  bool     `json:"getFollowing"`
}

type URLSet []struct {
	ExpandedURL string `json:"expanded_url"`
}

// ApifyFollowerResponse represents the response from Apify
type ApifyFollowerResponse struct {
	Username    string `json:"screen_name"`
	FullName    string `json:"name"`
	Bio         string `json:"description"`
	Website     string `json:"website"`
	ProfileLink string `json:"profileLink"`
	Entities    struct {
		URL struct {
			URLSet `json:"urls"`
		} `json:"url"`
		Description struct {
			URLSet `json:"urls"`
		} `json:"description"`
	} `json:"entities"`
}

// FetchFollowing fetches the accounts a user is following via Apify
func (c *Client) FetchFollowing(username string, limit int, progressCallback func(message string)) ([]TwitterUser, error) {
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

	// Transform to our internal model and extract wallet addresses
	users := make([]TwitterUser, 0, len(apifyResponses))
	for _, accountResp := range apifyResponses {
		user := TwitterUser{
			Username:    accountResp.Username,
			DisplayName: accountResp.FullName,
			Bio:         accountResp.Bio,
			Urls:        make([]string, 0),
		}

		distinctUrls := make(map[string]struct{})
		fullSet := append(accountResp.Entities.URL.URLSet, accountResp.Entities.Description.URLSet...)
		for _, accUrl := range fullSet {
			if _, ok := distinctUrls[accUrl.ExpandedURL]; !ok {
				distinctUrls[accUrl.ExpandedURL] = struct{}{}
				user.Urls = append(user.Urls, accUrl.ExpandedURL)
			}
		}

		// Extract wallet addresses from bio
		bioAddresses := ExtractSolanaAddresses(accountResp.Bio)
		if len(bioAddresses) > 0 {
			user.PossibleMintAddresses = append(user.PossibleMintAddresses, bioAddresses...)
			if progressCallback != nil {
				progressCallback(fmt.Sprintf("Found %d potential wallet address(es) in @%s's bio", len(bioAddresses), accountResp.Username))
			}
		}

		// If a website is provided, fetch and scan it
		for _, accUrl := range user.Urls {
			if !isValidURL(accUrl) {
				continue
			}
			if progressCallback != nil {
				progressCallback(fmt.Sprintf("Checking @%s's website: %s", accountResp.Username, accUrl))
			}

			websiteAddresses, err := FetchAndExtractAddressesFromWebsite(accUrl)
			if err != nil {
				// Just log the error but continue
				if progressCallback != nil {
					progressCallback(fmt.Sprintf("Error scanning website for @%s: %s", accountResp.Username, err.Error()))
				}
			} else if len(websiteAddresses) > 0 {
				user.PossibleMintAddresses = append(user.PossibleMintAddresses, websiteAddresses...)
				if progressCallback != nil {
					progressCallback(fmt.Sprintf("Found %d potential wallet address(es) on @%s's website", len(websiteAddresses), accountResp.Username))
				}
			}
		}

		users = append(users, user)
	}

	if progressCallback != nil {
		progressCallback(fmt.Sprintf("Successfully processed data for %d users followed by @%s", len(users), username))
	}

	return users, nil
}

// isValidURL checks if a string is a valid URL
func isValidURL(urlString string) bool {
	// Add http:// prefix if missing
	if !strings.HasPrefix(urlString, "http://") && !strings.HasPrefix(urlString, "https://") {
		urlString = "https://" + urlString
	}

	_, err := url.ParseRequestURI(urlString)
	return err == nil
}

// solanaAddressRegex is a regular expression for finding Solana addresses
// Solana addresses are base58 encoded and are typically 32-44 characters long
var solanaAddressRegex = regexp.MustCompile(`\b[1-9A-HJ-NP-Za-km-z]{32,44}\b`)

// ExtractSolanaAddresses extracts potential Solana addresses from a string
func ExtractSolanaAddresses(text string) []string {
	if text == "" {
		return nil
	}

	// Find all matches
	matches := solanaAddressRegex.FindAllString(text, -1)

	// Deduplicate the results
	uniqueMatches := make(map[string]struct{})
	for _, match := range matches {
		uniqueMatches[match] = struct{}{}
	}

	// Convert to slice
	result := make([]string, 0, len(uniqueMatches))
	for match := range uniqueMatches {
		result = append(result, match)
	}

	return result
}

// FetchAndExtractAddressesFromWebsite fetches the content of a website and extracts potential Solana addresses
func FetchAndExtractAddressesFromWebsite(websiteURL string) ([]string, error) {
	// Add http:// prefix if missing
	if !strings.HasPrefix(websiteURL, "http://") && !strings.HasPrefix(websiteURL, "https://") {
		websiteURL = "https://" + websiteURL
	}

	// Make a request to the website
	resp, err := http.Get(websiteURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error status: %d", resp.StatusCode)
	}

	// Read the body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Convert to string and extract addresses
	bodyText := string(bodyBytes)
	return ExtractSolanaAddresses(bodyText), nil
}
