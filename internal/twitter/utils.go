package twitter

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"wallet-guesser/internal/domain"
)

var solanaAddressRegex = regexp.MustCompile(`\b[1-9A-HJ-NP-Za-km-z]{32,44}\b`)

// processTwitterAccount converts Apify response to our domain model and extracts addresses
func (c *Client) processTwitterAccount(accountResp ApifyFollowerResponse, progressCallback domain.ProgressCallback) (domain.TwitterUser, error) {
	user := domain.TwitterUser{
		Username:    accountResp.Username,
		DisplayName: accountResp.FullName,
		Bio:         accountResp.Bio,
		Urls:        make([]string, 0),
	}

	// Extract URLs from both profile URL and bio
	distinctUrls := make(map[string]struct{})

	// Add profile URLs
	urlSet := accountResp.Entities.URL.URLSet
	for _, u := range urlSet {
		if _, ok := distinctUrls[u.ExpandedURL]; !ok {
			distinctUrls[u.ExpandedURL] = struct{}{}
			user.Urls = append(user.Urls, u.ExpandedURL)
		}
	}

	// Add bio URLs
	bioUrlSet := accountResp.Entities.Description.URLSet
	for _, u := range bioUrlSet {
		if _, ok := distinctUrls[u.ExpandedURL]; !ok {
			distinctUrls[u.ExpandedURL] = struct{}{}
			user.Urls = append(user.Urls, u.ExpandedURL)
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
	for _, profileUrl := range user.Urls {
		continue
		if !isValidURL(profileUrl) {
			continue
		}
		if progressCallback != nil {
			progressCallback(fmt.Sprintf("Checking @%s's website: %s", accountResp.Username, profileUrl))
		}

		websiteAddresses, err := c.FetchAndExtractAddressesFromWebsite(profileUrl)
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

	return user, nil
}

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
func (c *Client) FetchAndExtractAddressesFromWebsite(websiteURL string) ([]string, error) {
	// Normalize URL
	if !strings.HasPrefix(websiteURL, "http://") && !strings.HasPrefix(websiteURL, "https://") {
		websiteURL = "https://" + websiteURL
	}

	// Check cache first
	c.cacheMutex.RLock()
	cachedContent, found := c.websiteCache[websiteURL]
	c.cacheMutex.RUnlock()

	var bodyText string
	if found {
		// Use cached content
		bodyText = cachedContent
	} else {
		// Fetch from website
		resp, err := c.httpClient.Get(websiteURL)
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

		// Convert to string
		bodyText = string(bodyBytes)

		// Cache the result
		c.cacheMutex.Lock()
		c.websiteCache[websiteURL] = bodyText
		c.cacheMutex.Unlock()
	}

	// Extract addresses
	return ExtractSolanaAddresses(bodyText), nil
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
