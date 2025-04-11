package twitter

// ApifyInput represents the input for the Apify Twitter follower scraper
type ApifyInput struct {
	UserNames     []string `json:"user_names,omitempty"`
	UserIDs       []string `json:"user_ids,omitempty"`
	MaxFollowers  int      `json:"maxFollowers"`
	MaxFollowings int      `json:"maxFollowings"`
	GetFollowers  bool     `json:"getFollowers"`
	GetFollowing  bool     `json:"getFollowing"`
}

// URLSet is a set of URLs from Twitter entities
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

// ClientOption is a functional option for configuring the Twitter client
type ClientOption func(*Client)

// WithApifyToken sets the Apify token
func WithApifyToken(token string) ClientOption {
	return func(c *Client) {
		c.apifyToken = token
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout int) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}
