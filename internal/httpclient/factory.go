package httpclient

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"

	"github.com/wongpinter/go-whatsapp/internal/config"
)

// ClientType represents different types of WhatsApp API clients
type ClientType string

const (
	CloudAPIClient    ClientType = "cloudapi"
	BusinessAPIClient ClientType = "business"
	FlowsAPIClient    ClientType = "flows"
)

// ClientConfig holds configuration for HTTP client creation
type ClientConfig struct {
	AccessToken   string
	APIVersion    string
	BaseURL       string
	Timeout       time.Duration
	RetryCount    int
	RetryWaitTime time.Duration
	RetryMaxWait  time.Duration
	RateLimiter   *rate.Limiter
	Logger        *zerolog.Logger
	UserAgent     string
	CustomHeaders map[string]string
}

// Factory creates and manages HTTP clients for WhatsApp API
type Factory struct {
	defaultConfig *config.Config
	logger        *zerolog.Logger
}

// NewFactory creates a new HTTP client factory
func NewFactory(cfg *config.Config, logger *zerolog.Logger) *Factory {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	if logger == nil {
		nopLogger := zerolog.Nop()
		logger = &nopLogger
	}

	return &Factory{
		defaultConfig: cfg,
		logger:        logger,
	}
}

// CreateClient creates a new Resty client with the specified configuration
func (f *Factory) CreateClient(clientType ClientType, cfg *ClientConfig) (*resty.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("client configuration is required")
	}

	// Merge with default configuration
	mergedConfig := f.mergeWithDefaults(cfg)

	// Create base HTTP client with connection pooling
	httpClient := &http.Client{
		Timeout: mergedConfig.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
		},
	}

	// Create Resty client
	client := resty.NewWithClient(httpClient)

	// Configure base settings
	client.SetBaseURL(f.getBaseURL(clientType, mergedConfig.APIVersion)).
		SetAuthToken(mergedConfig.AccessToken).
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", mergedConfig.UserAgent).
		SetTimeout(mergedConfig.Timeout).
		SetRetryCount(mergedConfig.RetryCount).
		SetRetryWaitTime(mergedConfig.RetryWaitTime).
		SetRetryMaxWaitTime(mergedConfig.RetryMaxWait)

	// Add custom headers
	for key, value := range mergedConfig.CustomHeaders {
		client.SetHeader(key, value)
	}

	// Set up retry conditions
	client.AddRetryCondition(func(r *resty.Response, err error) bool {
		// Retry on network errors
		if err != nil {
			return true
		}
		// Retry on specific HTTP status codes
		return r.StatusCode() == 429 || // Rate limited
			r.StatusCode() >= 500 // Server errors
	})

	// Set up response middleware for logging and error handling
	client.OnAfterResponse(f.createResponseHandler(mergedConfig.Logger))

	// Set up request middleware for rate limiting
	if mergedConfig.RateLimiter != nil {
		client.OnBeforeRequest(f.createRateLimitHandler(mergedConfig.RateLimiter))
	}

	f.logger.Info().
		Str("client_type", string(clientType)).
		Str("base_url", client.BaseURL).
		Str("api_version", mergedConfig.APIVersion).
		Msg("HTTP client created successfully")

	return client, nil
}

// mergeWithDefaults merges client config with factory defaults
func (f *Factory) mergeWithDefaults(cfg *ClientConfig) *ClientConfig {
	merged := &ClientConfig{
		AccessToken:   cfg.AccessToken,
		APIVersion:    cfg.APIVersion,
		BaseURL:       cfg.BaseURL,
		Timeout:       cfg.Timeout,
		RetryCount:    cfg.RetryCount,
		RetryWaitTime: cfg.RetryWaitTime,
		RetryMaxWait:  cfg.RetryMaxWait,
		RateLimiter:   cfg.RateLimiter,
		Logger:        cfg.Logger,
		UserAgent:     cfg.UserAgent,
		CustomHeaders: cfg.CustomHeaders,
	}

	// Apply defaults where values are not set
	if merged.APIVersion == "" {
		merged.APIVersion = f.defaultConfig.APIVersion
	}
	if merged.BaseURL == "" {
		merged.BaseURL = f.defaultConfig.BaseURL
	}
	if merged.Timeout == 0 {
		merged.Timeout = f.defaultConfig.RequestTimeout
	}
	if merged.RetryCount == 0 {
		merged.RetryCount = f.defaultConfig.RetryCount
	}
	if merged.RetryWaitTime == 0 {
		merged.RetryWaitTime = f.defaultConfig.RetryWaitTime
	}
	if merged.RetryMaxWait == 0 {
		merged.RetryMaxWait = f.defaultConfig.RetryMaxWait
	}
	if merged.UserAgent == "" {
		merged.UserAgent = f.defaultConfig.UserAgent
	}
	if merged.Logger == nil {
		merged.Logger = f.logger
	}

	return merged
}

// getBaseURL returns the appropriate base URL for the client type
func (f *Factory) getBaseURL(clientType ClientType, apiVersion string) string {
	switch clientType {
	case CloudAPIClient, BusinessAPIClient:
		return fmt.Sprintf("%s/%s", f.defaultConfig.BaseURL, apiVersion)
	case FlowsAPIClient:
		// Flows might use a different API version
		if apiVersion == "" {
			apiVersion = "v18.0" // Default for Flows
		}
		return fmt.Sprintf("%s/%s", f.defaultConfig.BaseURL, apiVersion)
	default:
		return fmt.Sprintf("%s/%s", f.defaultConfig.BaseURL, apiVersion)
	}
}

// createResponseHandler creates a response middleware for logging and error handling
func (f *Factory) createResponseHandler(logger *zerolog.Logger) func(*resty.Client, *resty.Response) error {
	return func(client *resty.Client, resp *resty.Response) error {
		// Log response details
		logger.Debug().
			Str("method", resp.Request.Method).
			Str("url", resp.Request.URL).
			Int("status", resp.StatusCode()).
			Dur("duration", resp.Time()).
			Msg("HTTP response received")

		// Handle errors (this can be customized per client type)
		if resp.IsError() {
			logger.Warn().
				Str("method", resp.Request.Method).
				Str("url", resp.Request.URL).
				Int("status", resp.StatusCode()).
				Str("body", string(resp.Body())).
				Msg("HTTP error response")
		}

		return nil
	}
}

// createRateLimitHandler creates a request middleware for rate limiting
func (f *Factory) createRateLimitHandler(limiter *rate.Limiter) func(*resty.Client, *resty.Request) error {
	return func(client *resty.Client, req *resty.Request) error {
		// Wait for rate limiter permission
		return limiter.Wait(req.Context())
	}
}
