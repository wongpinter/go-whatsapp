package webhook

import (
	"time"
)

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	ErrorCategoryAuthentication ErrorCategory = "authentication"
	ErrorCategoryParameter      ErrorCategory = "parameter"
	ErrorCategoryRateLimit      ErrorCategory = "rate_limit"
	ErrorCategoryPolicy         ErrorCategory = "policy"
	ErrorCategoryTemplate       ErrorCategory = "template"
	ErrorCategoryMedia          ErrorCategory = "media"
	ErrorCategoryNetwork        ErrorCategory = "network"
	ErrorCategoryUnknown        ErrorCategory = "unknown"
)

// ErrorSeverity represents the severity of an error
type ErrorSeverity string

const (
	ErrorSeverityTransient ErrorSeverity = "transient" // Retryable
	ErrorSeverityPermanent ErrorSeverity = "permanent" // Not retryable
	ErrorSeverityCritical  ErrorSeverity = "critical"  // Requires immediate attention
)

// ErrorClassification contains detailed information about an error
type ErrorClassification struct {
	Code        int
	Category    ErrorCategory
	Severity    ErrorSeverity
	Retryable   bool
	RetryDelay  time.Duration
	MaxRetries  int
	Description string
	Action      string
}

// ErrorClassifier classifies WhatsApp API errors
type ErrorClassifier struct {
	classifications map[int]*ErrorClassification
}

// NewErrorClassifier creates a new error classifier with predefined classifications
func NewErrorClassifier() *ErrorClassifier {
	classifier := &ErrorClassifier{
		classifications: make(map[int]*ErrorClassification),
	}
	
	classifier.initializeClassifications()
	return classifier
}

// ClassifyError classifies an error by its code
func (ec *ErrorClassifier) ClassifyError(code int) *ErrorClassification {
	if classification, exists := ec.classifications[code]; exists {
		return classification
	}
	
	// Return default classification for unknown errors
	return &ErrorClassification{
		Code:        code,
		Category:    ErrorCategoryUnknown,
		Severity:    ErrorSeverityPermanent,
		Retryable:   false,
		RetryDelay:  0,
		MaxRetries:  0,
		Description: "Unknown error code",
		Action:      "Check WhatsApp API documentation for this error code",
	}
}

// IsRetryable checks if an error is retryable
func (ec *ErrorClassifier) IsRetryable(code int) bool {
	classification := ec.ClassifyError(code)
	return classification.Retryable
}

// GetRetryDelay returns the recommended retry delay for an error
func (ec *ErrorClassifier) GetRetryDelay(code int, attempt int) time.Duration {
	classification := ec.ClassifyError(code)
	if !classification.Retryable || attempt >= classification.MaxRetries {
		return 0
	}
	
	// Exponential backoff
	delay := classification.RetryDelay
	for i := 0; i < attempt; i++ {
		delay *= 2
	}
	
	// Cap at 5 minutes
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}
	
	return delay
}

// initializeClassifications sets up the error classifications based on WhatsApp documentation
func (ec *ErrorClassifier) initializeClassifications() {
	// Authentication & Authorization Errors
	ec.classifications[0] = &ErrorClassification{
		Code:        0,
		Category:    ErrorCategoryAuthentication,
		Severity:    ErrorSeverityPermanent,
		Retryable:   false,
		Description: "Authentication exception",
		Action:      "Check access token and app permissions",
	}
	
	ec.classifications[190] = &ErrorClassification{
		Code:        190,
		Category:    ErrorCategoryAuthentication,
		Severity:    ErrorSeverityPermanent,
		Retryable:   false,
		Description: "Access token expired",
		Action:      "Refresh or regenerate access token",
	}

	// Parameter & Validation Errors
	ec.classifications[100] = &ErrorClassification{
		Code:        100,
		Category:    ErrorCategoryParameter,
		Severity:    ErrorSeverityPermanent,
		Retryable:   false,
		Description: "Invalid parameter",
		Action:      "Validate request parameters against API specification",
	}

	// Rate Limiting Errors
	ec.classifications[130429] = &ErrorClassification{
		Code:        130429,
		Category:    ErrorCategoryRateLimit,
		Severity:    ErrorSeverityTransient,
		Retryable:   true,
		RetryDelay:  30 * time.Second,
		MaxRetries:  5,
		Description: "Rate limit exceeded",
		Action:      "Implement exponential backoff and reduce request rate",
	}

	ec.classifications[80007] = &ErrorClassification{
		Code:        80007,
		Category:    ErrorCategoryRateLimit,
		Severity:    ErrorSeverityTransient,
		Retryable:   true,
		RetryDelay:  60 * time.Second,
		MaxRetries:  3,
		Description: "API rate limit exceeded",
		Action:      "Wait before retrying and implement rate limiting",
	}

	// Template Errors
	ec.classifications[131014] = &ErrorClassification{
		Code:        131014,
		Category:    ErrorCategoryTemplate,
		Severity:    ErrorSeverityPermanent,
		Retryable:   false,
		Description: "Template does not exist",
		Action:      "Verify template name and ensure it's approved",
	}

	ec.classifications[131005] = &ErrorClassification{
		Code:        131005,
		Category:    ErrorCategoryTemplate,
		Severity:    ErrorSeverityPermanent,
		Retryable:   false,
		Description: "Template format character policy violated",
		Action:      "Review template content for policy compliance",
	}

	// Message Delivery Errors
	ec.classifications[131026] = &ErrorClassification{
		Code:        131026,
		Category:    ErrorCategoryNetwork,
		Severity:    ErrorSeverityPermanent,
		Retryable:   false,
		Description: "Message undeliverable",
		Action:      "Verify recipient phone number and try alternative contact method",
	}

	ec.classifications[131051] = &ErrorClassification{
		Code:        131051,
		Category:    ErrorCategoryParameter,
		Severity:    ErrorSeverityPermanent,
		Retryable:   false,
		Description: "Unsupported message type",
		Action:      "Use supported message types or update API version",
	}

	// Media Errors
	ec.classifications[131052] = &ErrorClassification{
		Code:        131052,
		Category:    ErrorCategoryMedia,
		Severity:    ErrorSeverityTransient,
		Retryable:   true,
		RetryDelay:  10 * time.Second,
		MaxRetries:  3,
		Description: "Media download error",
		Action:      "Verify media URL accessibility and format",
	}

	ec.classifications[131053] = &ErrorClassification{
		Code:        131053,
		Category:    ErrorCategoryMedia,
		Severity:    ErrorSeverityTransient,
		Retryable:   true,
		RetryDelay:  10 * time.Second,
		MaxRetries:  3,
		Description: "Media upload error",
		Action:      "Check media file size and format, retry upload",
	}

	// Policy Violations
	ec.classifications[131047] = &ErrorClassification{
		Code:        131047,
		Category:    ErrorCategoryPolicy,
		Severity:    ErrorSeverityPermanent,
		Retryable:   false,
		Description: "Re-engagement message",
		Action:      "Use approved template for re-engagement",
	}

	ec.classifications[131048] = &ErrorClassification{
		Code:        131048,
		Category:    ErrorCategoryPolicy,
		Severity:    ErrorSeverityPermanent,
		Retryable:   false,
		Description: "Re-engagement message",
		Action:      "Use approved template for re-engagement",
	}

	// Network & Service Errors
	ec.classifications[131000] = &ErrorClassification{
		Code:        131000,
		Category:    ErrorCategoryNetwork,
		Severity:    ErrorSeverityTransient,
		Retryable:   true,
		RetryDelay:  5 * time.Second,
		MaxRetries:  3,
		Description: "Something went wrong",
		Action:      "Retry the request after a short delay",
	}

	ec.classifications[131008] = &ErrorClassification{
		Code:        131008,
		Category:    ErrorCategoryParameter,
		Severity:    ErrorSeverityTransient,
		Retryable:   true,
		RetryDelay:  1 * time.Second,
		MaxRetries:  2,
		Description: "Required parameter is missing",
		Action:      "Check request payload for missing required parameters",
	}

	ec.classifications[131009] = &ErrorClassification{
		Code:        131009,
		Category:    ErrorCategoryParameter,
		Severity:    ErrorSeverityTransient,
		Retryable:   true,
		RetryDelay:  1 * time.Second,
		MaxRetries:  2,
		Description: "Parameter value is not valid",
		Action:      "Validate parameter values against API specification",
	}
}

// GetErrorsByCategory returns all error classifications for a specific category
func (ec *ErrorClassifier) GetErrorsByCategory(category ErrorCategory) []*ErrorClassification {
	var errors []*ErrorClassification
	for _, classification := range ec.classifications {
		if classification.Category == category {
			errors = append(errors, classification)
		}
	}
	return errors
}

// GetRetryableErrors returns all retryable error classifications
func (ec *ErrorClassifier) GetRetryableErrors() []*ErrorClassification {
	var errors []*ErrorClassification
	for _, classification := range ec.classifications {
		if classification.Retryable {
			errors = append(errors, classification)
		}
	}
	return errors
}

// GetCriticalErrors returns all critical error classifications
func (ec *ErrorClassifier) GetCriticalErrors() []*ErrorClassification {
	var errors []*ErrorClassification
	for _, classification := range ec.classifications {
		if classification.Severity == ErrorSeverityCritical {
			errors = append(errors, classification)
		}
	}
	return errors
}

// AddCustomClassification allows adding custom error classifications
func (ec *ErrorClassifier) AddCustomClassification(code int, classification *ErrorClassification) {
	classification.Code = code
	ec.classifications[code] = classification
}

// GetAllClassifications returns all error classifications
func (ec *ErrorClassifier) GetAllClassifications() map[int]*ErrorClassification {
	// Return a copy to prevent external modification
	result := make(map[int]*ErrorClassification)
	for code, classification := range ec.classifications {
		result[code] = classification
	}
	return result
}
