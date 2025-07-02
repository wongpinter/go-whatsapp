package httpclient

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"

	"github.com/wongpinter/go-whatsapp/internal/config"
)

// Manager manages shared HTTP clients to avoid creating multiple clients
// for the same configuration
type Manager struct {
	factory *Factory
	clients map[string]*resty.Client
	mutex   sync.RWMutex
	logger  *zerolog.Logger
}

// NewManager creates a new HTTP client manager
func NewManager(cfg *config.Config, logger *zerolog.Logger) *Manager {
	return &Manager{
		factory: NewFactory(cfg, logger),
		clients: make(map[string]*resty.Client),
		logger:  logger,
	}
}

// GetOrCreateClient returns an existing client or creates a new one
func (m *Manager) GetOrCreateClient(clientType ClientType, cfg *ClientConfig) (*resty.Client, error) {
	// Create a unique key for this configuration
	key := m.createClientKey(clientType, cfg)

	// Check if client already exists
	m.mutex.RLock()
	if client, exists := m.clients[key]; exists {
		m.mutex.RUnlock()
		m.logger.Debug().
			Str("client_type", string(clientType)).
			Str("key", key).
			Msg("Reusing existing HTTP client")
		return client, nil
	}
	m.mutex.RUnlock()

	// Create new client
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Double-check in case another goroutine created it
	if client, exists := m.clients[key]; exists {
		return client, nil
	}

	client, err := m.factory.CreateClient(clientType, cfg)
	if err != nil {
		return nil, err
	}

	m.clients[key] = client
	m.logger.Info().
		Str("client_type", string(clientType)).
		Str("key", key).
		Msg("Created and cached new HTTP client")

	return client, nil
}

// createClientKey creates a unique key for client configuration
func (m *Manager) createClientKey(clientType ClientType, cfg *ClientConfig) string {
	// Create a key based on important configuration parameters
	// This ensures we reuse clients with the same configuration
	return string(clientType) + ":" + cfg.AccessToken[:10] + ":" + cfg.APIVersion
}

// CloseAll closes all managed clients and clears the cache
func (m *Manager) CloseAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for key, client := range m.clients {
		// Close the underlying HTTP client if possible
		if httpClient := client.GetClient(); httpClient != nil {
			if transport, ok := httpClient.Transport.(*http.Transport); ok {
				transport.CloseIdleConnections()
			}
		}
		delete(m.clients, key)
	}

	m.logger.Info().Msg("All HTTP clients closed and cache cleared")
}

// GetClientCount returns the number of cached clients
func (m *Manager) GetClientCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.clients)
}

// DefaultClientOptions provides common client options for different client types
type DefaultClientOptions struct {
	CloudAPI *ClientConfig
	Business *ClientConfig
	Flows    *ClientConfig
}

// GetDefaultOptions returns default client configurations for each client type
func GetDefaultOptions(accessToken string) *DefaultClientOptions {
	baseConfig := &ClientConfig{
		AccessToken:   accessToken,
		Timeout:       30 * time.Second,
		RetryCount:    3,
		RetryWaitTime: 1 * time.Second,
		RetryMaxWait:  10 * time.Second,
		UserAgent:     "go-whatsapp-sdk/1.0.0",
		CustomHeaders: make(map[string]string),
	}

	return &DefaultClientOptions{
		CloudAPI: &ClientConfig{
			AccessToken:   baseConfig.AccessToken,
			APIVersion:    "v19.0",
			Timeout:       baseConfig.Timeout,
			RetryCount:    baseConfig.RetryCount,
			RetryWaitTime: baseConfig.RetryWaitTime,
			RetryMaxWait:  baseConfig.RetryMaxWait,
			UserAgent:     baseConfig.UserAgent,
			CustomHeaders: baseConfig.CustomHeaders,
		},
		Business: &ClientConfig{
			AccessToken:   baseConfig.AccessToken,
			APIVersion:    "v19.0",
			Timeout:       baseConfig.Timeout,
			RetryCount:    baseConfig.RetryCount,
			RetryWaitTime: baseConfig.RetryWaitTime,
			RetryMaxWait:  baseConfig.RetryMaxWait,
			UserAgent:     baseConfig.UserAgent,
			CustomHeaders: baseConfig.CustomHeaders,
		},
		Flows: &ClientConfig{
			AccessToken:   baseConfig.AccessToken,
			APIVersion:    "v18.0", // Flows use different API version
			Timeout:       baseConfig.Timeout,
			RetryCount:    baseConfig.RetryCount,
			RetryWaitTime: baseConfig.RetryWaitTime,
			RetryMaxWait:  baseConfig.RetryMaxWait,
			UserAgent:     baseConfig.UserAgent,
			CustomHeaders: baseConfig.CustomHeaders,
		},
	}
}

// WithRateLimiting adds rate limiting to client configuration
func WithRateLimiting(cfg *ClientConfig, requestsPerSecond float64, burstSize int) *ClientConfig {
	cfg.RateLimiter = rate.NewLimiter(rate.Limit(requestsPerSecond), burstSize)
	return cfg
}

// WithLogger adds logger to client configuration
func WithLogger(cfg *ClientConfig, logger *zerolog.Logger) *ClientConfig {
	cfg.Logger = logger
	return cfg
}

// WithTimeout sets timeout for client configuration
func WithTimeout(cfg *ClientConfig, timeout time.Duration) *ClientConfig {
	cfg.Timeout = timeout
	return cfg
}

// WithRetryConfig sets retry configuration
func WithRetryConfig(cfg *ClientConfig, retryCount int, waitTime, maxWaitTime time.Duration) *ClientConfig {
	cfg.RetryCount = retryCount
	cfg.RetryWaitTime = waitTime
	cfg.RetryMaxWait = maxWaitTime
	return cfg
}

// WithCustomHeader adds a custom header to client configuration
func WithCustomHeader(cfg *ClientConfig, key, value string) *ClientConfig {
	if cfg.CustomHeaders == nil {
		cfg.CustomHeaders = make(map[string]string)
	}
	cfg.CustomHeaders[key] = value
	return cfg
}

// WithAPIVersion sets API version for client configuration
func WithAPIVersion(cfg *ClientConfig, version string) *ClientConfig {
	cfg.APIVersion = version
	return cfg
}
