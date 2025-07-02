package config

import (
	"time"
)

// Config holds common configuration values used across the SDK.
type Config struct {
	// API configuration
	BaseURL    string
	APIVersion string
	UserAgent  string

	// Timeout configuration
	RequestTimeout time.Duration
	RetryCount     int
	RetryWaitTime  time.Duration
	RetryMaxWait   time.Duration

	// Rate limiting
	RateLimitEnabled bool
	RequestsPerSecond float64
	BurstSize        int
}

// DefaultConfig returns the default configuration for the WhatsApp Cloud API.
func DefaultConfig() *Config {
	return &Config{
		BaseURL:           "https://graph.facebook.com",
		APIVersion:        "v19.0",
		UserAgent:         "go-whatsapp-sdk/1.0.0",
		RequestTimeout:    30 * time.Second,
		RetryCount:        3,
		RetryWaitTime:     1 * time.Second,
		RetryMaxWait:      10 * time.Second,
		RateLimitEnabled:  false,
		RequestsPerSecond: 80.0, // Default WhatsApp rate limit
		BurstSize:         10,
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.BaseURL == "" {
		return ErrInvalidBaseURL
	}
	if c.APIVersion == "" {
		return ErrInvalidAPIVersion
	}
	if c.RequestTimeout <= 0 {
		return ErrInvalidTimeout
	}
	if c.RetryCount < 0 {
		return ErrInvalidRetryCount
	}
	if c.RequestsPerSecond <= 0 {
		return ErrInvalidRateLimit
	}
	return nil
}

// Configuration errors
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}

var (
	ErrInvalidBaseURL     = &ConfigError{"BaseURL", "base URL cannot be empty"}
	ErrInvalidAPIVersion  = &ConfigError{"APIVersion", "API version cannot be empty"}
	ErrInvalidTimeout     = &ConfigError{"RequestTimeout", "request timeout must be positive"}
	ErrInvalidRetryCount  = &ConfigError{"RetryCount", "retry count cannot be negative"}
	ErrInvalidRateLimit   = &ConfigError{"RequestsPerSecond", "requests per second must be positive"}
)

// Environment represents different deployment environments.
type Environment string

const (
	// Production environment using the live WhatsApp Cloud API
	Production Environment = "production"
	// Sandbox environment for testing (if available)
	Sandbox Environment = "sandbox"
	// Development environment for local testing
	Development Environment = "development"
)

// GetBaseURLForEnvironment returns the appropriate base URL for the given environment.
func GetBaseURLForEnvironment(env Environment) string {
	switch env {
	case Production:
		return "https://graph.facebook.com"
	case Sandbox:
		// WhatsApp doesn't have a separate sandbox URL, but this allows for future expansion
		return "https://graph.facebook.com"
	case Development:
		// Could be used for local mock servers
		return "http://localhost:8080"
	default:
		return "https://graph.facebook.com"
	}
}

// APIEndpoints contains the various API endpoints used by the SDK.
type APIEndpoints struct {
	Messages    string
	Media       string
	Templates   string
	PhoneNumbers string
	BusinessProfile string
	Analytics   string
}

// GetEndpoints returns the API endpoints for the given API version.
func GetEndpoints(apiVersion string) *APIEndpoints {
	return &APIEndpoints{
		Messages:        "messages",
		Media:          "media",
		Templates:      "message_templates",
		PhoneNumbers:   "phone_numbers",
		BusinessProfile: "whatsapp_business_profile",
		Analytics:      "conversation_analytics",
	}
}

// MediaLimits defines the size and type limits for media uploads.
type MediaLimits struct {
	MaxImageSize    int64    // in bytes
	MaxVideoSize    int64    // in bytes
	MaxAudioSize    int64    // in bytes
	MaxDocumentSize int64    // in bytes
	AllowedImageTypes []string
	AllowedVideoTypes []string
	AllowedAudioTypes []string
	AllowedDocumentTypes []string
}

// GetMediaLimits returns the media upload limits for WhatsApp.
func GetMediaLimits() *MediaLimits {
	return &MediaLimits{
		MaxImageSize:    5 * 1024 * 1024,   // 5MB
		MaxVideoSize:    16 * 1024 * 1024,  // 16MB
		MaxAudioSize:    16 * 1024 * 1024,  // 16MB
		MaxDocumentSize: 100 * 1024 * 1024, // 100MB
		AllowedImageTypes: []string{
			"image/jpeg", "image/png", "image/webp",
		},
		AllowedVideoTypes: []string{
			"video/mp4", "video/3gpp",
		},
		AllowedAudioTypes: []string{
			"audio/aac", "audio/mp4", "audio/mpeg", "audio/amr", "audio/ogg",
		},
		AllowedDocumentTypes: []string{
			"text/plain", "application/pdf", "application/vnd.ms-powerpoint",
			"application/msword", "application/vnd.ms-excel",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/vnd.openxmlformats-officedocument.presentationml.presentation",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		},
	}
}
