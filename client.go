package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

// HTTPClient interface allows for dependency injection and testing.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Logger interface for structured logging.
type Logger interface {
	Debug() *zerolog.Event
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
}

// Client is the main WhatsApp Cloud API client.
type Client struct {
	restyClient   *resty.Client
	logger        Logger
	PhoneNumberID string
	AccessToken   string
	WABAID        string
	APIVersion    string
	rateLimiter   *rate.Limiter
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// NewClient creates a new WhatsApp Cloud API client with the given phone number ID
// and access token. Additional configuration can be provided via Option functions.
func NewClient(phoneNumberID, accessToken string, opts ...Option) (*Client, error) {
	// Validate required parameters
	if strings.TrimSpace(phoneNumberID) == "" {
		return nil, &ErrInvalidPhoneNumberID{}
	}
	if strings.TrimSpace(accessToken) == "" {
		return nil, &ErrInvalidAccessToken{}
	}

	nopLogger := zerolog.Nop()
	c := &Client{
		PhoneNumberID: phoneNumberID,
		AccessToken:   accessToken,
		APIVersion:    "v19.0",    // Default to latest stable version
		logger:        &nopLogger, // Default to no-op logger
	}

	// Apply all options
	for _, opt := range opts {
		opt(c)
	}

	// Initialize Resty client after all options are applied
	c.restyClient = resty.New().
		SetBaseURL(fmt.Sprintf("https://graph.facebook.com/%s", c.APIVersion)).
		SetAuthToken(c.AccessToken).
		SetHeader("Content-Type", "application/json").
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(10 * time.Second)

	// Set up error handling
	c.restyClient.OnAfterResponse(c.handleResponse)

	return c, nil
}

// WithLogger sets a custom logger for the client.
func WithLogger(logger Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithAPIVersion sets a custom API version.
func WithAPIVersion(version string) Option {
	return func(c *Client) {
		c.APIVersion = version
	}
}

// WithWABAID sets the WhatsApp Business Account ID.
func WithWABAID(wabaID string) Option {
	return func(c *Client) {
		c.WABAID = wabaID
	}
}

// WithHTTPClient sets a custom HTTP client.
// Note: This should be called before the client is fully initialized
func WithHTTPClient(httpClient HTTPClient) Option {
	return func(c *Client) {
		// For now, we'll skip custom HTTP client support
		// This can be implemented later using resty.NewWithClient()
	}
}

// WithRateLimiting enables rate limiting with the specified requests per second.
// The default burst size is set to 10 requests.
func WithRateLimiting(requestsPerSecond float64) Option {
	return func(c *Client) {
		c.rateLimiter = rate.NewLimiter(rate.Limit(requestsPerSecond), 10)
	}
}

// WithCustomRateLimiter sets a custom rate limiter.
func WithCustomRateLimiter(limiter *rate.Limiter) Option {
	return func(c *Client) {
		c.rateLimiter = limiter
	}
}

// WithTimeout sets the request timeout for the HTTP client.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if c.restyClient != nil {
			c.restyClient.SetTimeout(timeout)
		}
	}
}

// WithRetryConfig sets custom retry configuration.
func WithRetryConfig(retryCount int, waitTime, maxWaitTime time.Duration) Option {
	return func(c *Client) {
		if c.restyClient != nil {
			c.restyClient.SetRetryCount(retryCount).
				SetRetryWaitTime(waitTime).
				SetRetryMaxWaitTime(maxWaitTime)
		}
	}
}

// handleResponse processes HTTP responses and converts API errors to structured errors.
func (c *Client) handleResponse(client *resty.Client, resp *resty.Response) error {
	if resp.IsError() {
		var apiError APIError
		if err := json.Unmarshal(resp.Body(), &apiError); err != nil {
			// If we can't parse the error response, create a generic error
			return NewAPIError(resp.StatusCode(),
				fmt.Sprintf("HTTP %d: %s", resp.StatusCode(), resp.Status()),
				"HTTPError", "")
		}
		return &apiError
	}
	return nil
}

// waitForRateLimit waits for rate limiting if enabled.
func (c *Client) waitForRateLimit(ctx context.Context) error {
	if c.rateLimiter != nil {
		return c.rateLimiter.Wait(ctx)
	}
	return nil
}

// GetPhoneNumberID returns the configured phone number ID.
func (c *Client) GetPhoneNumberID() string {
	return c.PhoneNumberID
}

// GetWABAID returns the configured WhatsApp Business Account ID.
func (c *Client) GetWABAID() string {
	return c.WABAID
}

// GetAPIVersion returns the configured API version.
func (c *Client) GetAPIVersion() string {
	return c.APIVersion
}

// SetLogger updates the client's logger.
func (c *Client) SetLogger(logger Logger) {
	c.logger = logger
}

// Health checks the health of the WhatsApp Business API.
func (c *Client) Health(ctx context.Context) error {
	if err := c.waitForRateLimit(ctx); err != nil {
		return err
	}

	resp, err := c.restyClient.R().
		SetContext(ctx).
		Get(fmt.Sprintf("/%s", c.PhoneNumberID))

	if err != nil {
		c.logger.Error().Err(err).Msg("Health check failed")
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.IsError() {
		c.logger.Error().
			Int("status_code", resp.StatusCode()).
			Str("response", resp.String()).
			Msg("Health check returned error")
		return fmt.Errorf("health check failed with status %d", resp.StatusCode())
	}

	c.logger.Info().Msg("Health check successful")
	return nil
}

// Close performs any necessary cleanup for the client.
func (c *Client) Close() error {
	// Currently no cleanup needed, but this provides a hook for future use
	return nil
}
