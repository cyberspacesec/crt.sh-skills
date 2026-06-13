// options.go — ClientOption pattern for configuring the crt.sh SDK client
package crtsh

import "time"

// ClientOption configures a Client during construction.
type ClientOption func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.HTTPClient.Timeout = d
	}
}

// WithRetryCount sets the number of retries for failed requests.
func WithRetryCount(n int) ClientOption {
	return func(c *Client) {
		c.RetryCount = n
	}
}

// WithDebug enables debug logging of HTTP requests and responses.
func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		c.Debug = debug
	}
}

// WithUserAgent sets the User-Agent header for HTTP requests.
func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.UserAgent = ua
	}
}

// WithBaseURL sets the base URL for the crt.sh API.
// Useful for testing or routing through a proxy.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.BaseURL = url
	}
}
