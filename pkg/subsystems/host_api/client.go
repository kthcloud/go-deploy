package host_api

type Client struct {
	URL string
}

// NewClient creates a new host API client.
func NewClient(url string) *Client {
	return &Client{
		URL: url,
	}
}
