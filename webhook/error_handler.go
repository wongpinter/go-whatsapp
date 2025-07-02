package webhook

import (
	"context"
	"fmt"
	"time"
)

// WebhookErrorHandler handles webhook-specific errors and implements retry logic
type WebhookErrorHandler struct {
	retryPolicy *RetryPolicy
	metrics     *MetricsCollector
	logger      Logger
}

// Logger interface for error handling
type Logger interface {
	Error() LogEvent
	Warn() LogEvent
	Info() LogEvent
}

// LogEvent interface for logging
type LogEvent interface {
	Err(error) LogEvent
	Str(string, string) LogEvent
	Int(string, int) LogEvent
	Msg(string)
}

// NewWebhookErrorHandler creates a new error handler
func NewWebhookErrorHandler(retryPolicy *RetryPolicy, metrics *MetricsCollector, logger Logger) *WebhookErrorHandler {
	if retryPolicy == nil {
		retryPolicy = DefaultRetryPolicy()
	}

	return &WebhookErrorHandler{
		retryPolicy: retryPolicy,
		metrics:     metrics,
		logger:      logger,
	}
}

// HandleWebhookError processes webhook-specific errors
func (eh *WebhookErrorHandler) HandleWebhookError(ctx context.Context, err error, payload *WebhookPayload) error {
	if err == nil {
		return nil
	}

	// Log the error
	eh.logger.Error().
		Err(err).
		Str("error_type", "webhook_processing").
		Msg("Webhook processing error occurred")

	// Check for specific error types
	switch e := err.(type) {
	case *VerificationError:
		return eh.handleVerificationError(ctx, e)
	case *MessageError:
		return eh.handleMessageError(ctx, e, payload)
	case *StatusError:
		return eh.handleStatusError(ctx, e, payload)
	default:
		return eh.handleGenericError(ctx, err, payload)
	}
}

// MessageError represents an error processing a message
type MessageError struct {
	MessageID string
	ErrorCode int
	Code      int       // Error code
	Title     string    // Error title
	Message   string    // Error message
	Timestamp time.Time // When the error occurred
	Source    string    // Source of the error ("api" or "webhook")
	Retryable bool      // Whether the error is retryable
}

func (e *MessageError) Error() string {
	return fmt.Sprintf("message error [%d]: %s (message_id: %s)", e.ErrorCode, e.Message, e.MessageID)
}

// StatusError represents an error processing a status update
type StatusError struct {
	MessageID string
	Status    string
	ErrorCode int
	Message   string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("status error [%d]: %s (message_id: %s, status: %s)", e.ErrorCode, e.Message, e.MessageID, e.Status)
}

// handleVerificationError handles webhook verification errors
func (eh *WebhookErrorHandler) handleVerificationError(_ context.Context, err *VerificationError) error {
	eh.logger.Error().
		Str("error_code", err.Code).
		Str("error_message", err.Message).
		Msg("Webhook verification failed")

	// Verification errors are not retryable
	return err
}

// handleMessageError handles message processing errors
func (eh *WebhookErrorHandler) handleMessageError(_ context.Context, err *MessageError, _ *WebhookPayload) error {
	eh.logger.Error().
		Str("message_id", err.MessageID).
		Int("error_code", err.ErrorCode).
		Str("error_message", err.Message).
		Msg("Message processing error")

	// Record metrics
	if eh.metrics != nil {
		eh.metrics.RecordRequest(time.Now(), false, 1)
	}

	// Check if error is retryable
	if !err.Retryable {
		return err
	}

	// For retryable errors, we would typically queue for retry
	// In this implementation, we just log and return
	eh.logger.Warn().
		Str("message_id", err.MessageID).
		Msg("Message error is retryable but no retry mechanism configured")

	return err
}

// handleStatusError handles status processing errors
func (eh *WebhookErrorHandler) handleStatusError(_ context.Context, err *StatusError, _ *WebhookPayload) error {
	eh.logger.Error().
		Str("message_id", err.MessageID).
		Str("status", err.Status).
		Int("error_code", err.ErrorCode).
		Str("error_message", err.Message).
		Msg("Status processing error")

	// Record metrics
	if eh.metrics != nil {
		eh.metrics.RecordRequest(time.Now(), false, 1)
	}

	return err
}

// handleGenericError handles generic errors
func (eh *WebhookErrorHandler) handleGenericError(_ context.Context, err error, _ *WebhookPayload) error {
	eh.logger.Error().
		Err(err).
		Msg("Generic webhook processing error")

	// Record metrics
	if eh.metrics != nil {
		eh.metrics.RecordRequest(time.Now(), false, 1)
	}

	return err
}

// ProcessMessageErrors processes errors from message payloads
func (eh *WebhookErrorHandler) ProcessMessageErrors(ctx context.Context, message *Message, metadata *Metadata) []error {
	var errors []error

	if len(message.Errors) == 0 {
		return errors
	}

	for _, webhookError := range message.Errors {
		err := &MessageError{
			MessageID: message.ID,
			ErrorCode: webhookError.Code,
			Message:   webhookError.Message,
			Retryable: eh.isRetryableError(webhookError.Code),
		}

		eh.logger.Error().
			Str("message_id", message.ID).
			Int("error_code", webhookError.Code).
			Str("error_title", webhookError.Title).
			Str("error_message", webhookError.Message).
			Msg("Message contains error")

		errors = append(errors, err)
	}

	return errors
}

// ProcessStatusErrors processes errors from status payloads
func (eh *WebhookErrorHandler) ProcessStatusErrors(ctx context.Context, status *Status, metadata *Metadata) []error {
	var errors []error

	if len(status.Errors) == 0 {
		return errors
	}

	for _, webhookError := range status.Errors {
		err := &StatusError{
			MessageID: status.ID,
			Status:    status.Status,
			ErrorCode: webhookError.Code,
			Message:   webhookError.Message,
		}

		eh.logger.Error().
			Str("message_id", status.ID).
			Str("status", status.Status).
			Int("error_code", webhookError.Code).
			Str("error_title", webhookError.Title).
			Str("error_message", webhookError.Message).
			Msg("Status contains error")

		errors = append(errors, err)
	}

	return errors
}

// isRetryableError determines if an error code is retryable
func (eh *WebhookErrorHandler) isRetryableError(errorCode int) bool {
	// Based on WhatsApp error codes documentation
	retryableErrors := map[int]bool{
		// Rate limiting errors
		130429: true, // Rate limit hit
		80007:  true, // Rate limit exceeded

		// Temporary service errors
		131000: true, // Something went wrong
		131005: true, // Request timeout
		131008: true, // Required parameter is missing
		131009: true, // Parameter value is not valid

		// Network/connectivity errors
		131047: true, // Re-engagement message
		131048: true, // Re-engagement message

		// Non-retryable errors
		131014: false, // Template does not exist
		131026: false, // Message undeliverable
		131051: false, // Unsupported message type
		131052: false, // Media download error
		131053: false, // Media upload error
	}

	if retryable, exists := retryableErrors[errorCode]; exists {
		return retryable
	}

	// Default to non-retryable for unknown errors
	return false
}

// GetErrorSummary returns a summary of errors for monitoring
func (eh *WebhookErrorHandler) GetErrorSummary() map[string]interface{} {
	// This would typically be implemented with persistent storage
	// For now, return a basic summary
	return map[string]interface{}{
		"total_errors":     0,
		"retryable_errors": 0,
		"fatal_errors":     0,
		"last_error_time":  nil,
	}
}

// IsTemporaryError checks if an error is temporary and should be retried
func IsTemporaryError(err error) bool {
	if msgErr, ok := err.(*MessageError); ok {
		return msgErr.Retryable
	}
	return false
}

// IsFatalError checks if an error is fatal and should not be retried
func IsFatalError(err error) bool {
	if _, ok := err.(*VerificationError); ok {
		return true
	}
	if msgErr, ok := err.(*MessageError); ok {
		return !msgErr.Retryable
	}
	return false
}
