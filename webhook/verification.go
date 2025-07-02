package webhook

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"
)

// VerificationHandler handles webhook verification challenges
type VerificationHandler struct {
	verifyToken string
}

// NewVerificationHandler creates a new verification handler
func NewVerificationHandler(verifyToken string) *VerificationHandler {
	return &VerificationHandler{
		verifyToken: verifyToken,
	}
}

// HandleVerification handles the webhook verification challenge
func (v *VerificationHandler) HandleVerification(w http.ResponseWriter, r *http.Request) error {
	// Parse query parameters
	query := r.URL.Query()

	// Extract verification parameters
	mode := query.Get("hub.mode")
	token := query.Get("hub.verify_token")
	challenge := query.Get("hub.challenge")

	// Validate required parameters
	if mode == "" {
		return &VerificationError{
			Code:    "missing_mode",
			Message: "hub.mode parameter is required",
		}
	}

	if token == "" {
		return &VerificationError{
			Code:    "missing_token",
			Message: "hub.verify_token parameter is required",
		}
	}

	if challenge == "" {
		return &VerificationError{
			Code:    "missing_challenge",
			Message: "hub.challenge parameter is required",
		}
	}

	// Verify mode is "subscribe"
	if mode != "subscribe" {
		return &VerificationError{
			Code:    "invalid_mode",
			Message: fmt.Sprintf("expected mode 'subscribe', got '%s'", mode),
		}
	}

	// Verify token matches
	if token != v.verifyToken {
		return &VerificationError{
			Code:    "invalid_token",
			Message: "verify token does not match",
		}
	}

	// Respond with challenge
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(challenge))

	return nil
}

// VerificationError represents a webhook verification error
type VerificationError struct {
	Code    string
	Message string
}

func (e *VerificationError) Error() string {
	return fmt.Sprintf("verification error [%s]: %s", e.Code, e.Message)
}

// IsVerificationError checks if an error is a verification error
func IsVerificationError(err error) bool {
	_, ok := err.(*VerificationError)
	return ok
}

// ValidateWebhookURL validates a webhook URL for common issues
func ValidateWebhookURL(webhookURL string) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}

	parsedURL, err := url.Parse(webhookURL)
	if err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	// Must use HTTPS
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("webhook URL must use HTTPS, got %s", parsedURL.Scheme)
	}

	// Must have a host
	if parsedURL.Host == "" {
		return fmt.Errorf("webhook URL must have a valid host")
	}

	// Check for common problematic hosts
	problematicHosts := []string{
		"localhost",
		"127.0.0.1",
		"0.0.0.0",
	}

	for _, host := range problematicHosts {
		if parsedURL.Hostname() == host {
			return fmt.Errorf("webhook URL cannot use %s - use a publicly accessible domain", host)
		}
	}

	return nil
}

// WebhookResponse represents the expected response format
type WebhookResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

// CreateVerificationResponse creates a proper verification response
func CreateVerificationResponse(challenge string) *WebhookResponse {
	return &WebhookResponse{
		StatusCode: http.StatusOK,
		Body:       challenge,
	}
}

// WebhookHealthCheck performs a basic health check for webhook functionality
type WebhookHealthCheck struct {
	URL         string `json:"url"`
	VerifyToken string `json:"verify_token"`
	Status      string `json:"status"`
	LastCheck   string `json:"last_check"`
	Error       string `json:"error,omitempty"`
}

// PerformHealthCheck performs a health check on the webhook configuration
func PerformHealthCheck(webhookURL, verifyToken string) *WebhookHealthCheck {
	check := &WebhookHealthCheck{
		URL:         webhookURL,
		VerifyToken: verifyToken,
		LastCheck:   fmt.Sprintf("%d", time.Now().Unix()),
	}

	// Validate URL
	if err := ValidateWebhookURL(webhookURL); err != nil {
		check.Status = "error"
		check.Error = err.Error()
		return check
	}

	// Validate token
	if verifyToken == "" {
		check.Status = "error"
		check.Error = "verify token cannot be empty"
		return check
	}

	// Basic validation passed
	check.Status = "healthy"
	return check
}

// RetryPolicy represents webhook retry configuration
type RetryPolicy struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []int         `json:"retryable_errors"`
}

// DefaultRetryPolicy returns the default retry policy based on Meta's documentation
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:    7, // Meta retries for up to 7 days
		InitialDelay:  time.Second,
		MaxDelay:      time.Hour,
		BackoffFactor: 2.0, // Exponential backoff
		RetryableErrors: []int{
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
	}
}

// ShouldRetry determines if a response should be retried
func (rp *RetryPolicy) ShouldRetry(statusCode int, attempt int) bool {
	if attempt >= rp.MaxRetries {
		return false
	}

	// Always retry on non-200 status codes that are retryable
	for _, retryableCode := range rp.RetryableErrors {
		if statusCode == retryableCode {
			return true
		}
	}

	// Don't retry on 200 OK
	if statusCode == http.StatusOK {
		return false
	}

	// Retry on any 5xx error
	return statusCode >= 500
}

// CalculateDelay calculates the delay for the next retry attempt
func (rp *RetryPolicy) CalculateDelay(attempt int) time.Duration {
	delay := time.Duration(float64(rp.InitialDelay) * math.Pow(rp.BackoffFactor, float64(attempt)))
	if delay > rp.MaxDelay {
		delay = rp.MaxDelay
	}
	return delay
}
