package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp/internal/crypto"
)

// Handler is the main webhook processor. It implements http.Handler.
type Handler struct {
	logger              *zerolog.Logger
	appSecret           string
	verifyToken         string
	dispatcher          *EventDispatcher
	statusMonitor       *StatusMonitor
	metrics             *MetricsCollector
	verificationHandler *VerificationHandler
	errorHandler        *WebhookErrorHandler
	messageTracker      *MessageLifecycleTracker
	errorClassifier     *ErrorClassifier
	rateLimiter         *RateLimiter
	messageQueue        *MessageQueueManager
}

// LoggerAdapter adapts zerolog.Logger to the Logger interface
type LoggerAdapter struct {
	logger *zerolog.Logger
}

func (l *LoggerAdapter) Error() LogEvent {
	return &LogEventAdapter{event: l.logger.Error()}
}

func (l *LoggerAdapter) Warn() LogEvent {
	return &LogEventAdapter{event: l.logger.Warn()}
}

func (l *LoggerAdapter) Info() LogEvent {
	return &LogEventAdapter{event: l.logger.Info()}
}

// LogEventAdapter adapts zerolog.Event to the LogEvent interface
type LogEventAdapter struct {
	event *zerolog.Event
}

func (l *LogEventAdapter) Err(err error) LogEvent {
	l.event = l.event.Err(err)
	return l
}

func (l *LogEventAdapter) Str(key, val string) LogEvent {
	l.event = l.event.Str(key, val)
	return l
}

func (l *LogEventAdapter) Int(key string, val int) LogEvent {
	l.event = l.event.Int(key, val)
	return l
}

func (l *LogEventAdapter) Msg(msg string) {
	l.event.Msg(msg)
}

// NewHandler creates a new webhook handler.
func NewHandler(appSecret, verifyToken string, logger zerolog.Logger) *Handler {
	metrics := NewMetricsCollector()
	rateLimiter := NewRateLimiter()

	// Create a simple logger adapter for the error handler
	loggerAdapter := &LoggerAdapter{logger: &logger}

	return &Handler{
		logger:              &logger,
		appSecret:           appSecret,
		verifyToken:         verifyToken,
		dispatcher:          NewEventDispatcher(),
		statusMonitor:       NewStatusMonitor(24 * time.Hour), // Keep status for 24 hours
		metrics:             metrics,
		verificationHandler: NewVerificationHandler(verifyToken),
		errorHandler:        NewWebhookErrorHandler(DefaultRetryPolicy(), metrics, loggerAdapter),
		messageTracker:      NewMessageLifecycleTracker(24 * time.Hour), // Keep message lifecycle for 24 hours
		errorClassifier:     NewErrorClassifier(),
		rateLimiter:         rateLimiter,
		messageQueue:        NewMessageQueueManager(rateLimiter),
	}
}

// ServeHTTP handles incoming GET (verification) and POST (event) requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleVerification(w, r)
	case http.MethodPost:
		h.handleEvent(w, r)
	default:
		h.logger.Warn().
			Str("method", r.Method).
			Msg("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleVerification handles the webhook verification challenge.
func (h *Handler) handleVerification(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Msg("Received webhook verification request")

	// Use the enhanced verification handler
	if err := h.verificationHandler.HandleVerification(w, r); err != nil {
		h.logger.Error().
			Err(err).
			Msg("Webhook verification failed")

		// Handle verification error
		if verifyErr, ok := err.(*VerificationError); ok {
			h.logger.Error().
				Str("error_code", verifyErr.Code).
				Str("error_message", verifyErr.Message).
				Msg("Verification error details")
		}

		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	h.logger.Info().Msg("Webhook verification successful")
}

// handleEvent handles incoming webhook events.
func (h *Handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to read request body")
		h.metrics.RecordRequest(start, false, 0)
		http.Error(w, "Cannot read body", http.StatusInternalServerError)
		return
	}

	// Verify the signature if app secret is configured
	if h.appSecret != "" {
		signature := r.Header.Get(crypto.SignatureHeader)
		if err := crypto.VerifySignature(body, signature, h.appSecret); err != nil {
			h.logger.Error().
				Err(err).
				Str("signature", signature).
				Msg("Signature verification failed")
			h.metrics.RecordRequest(start, false, 0)
			http.Error(w, "Invalid signature", http.StatusForbidden)
			return
		}
		h.logger.Debug().Msg("Signature verification successful")
	}

	// Parse the webhook payload
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error().
			Err(err).
			Str("body", string(body)).
			Msg("Failed to parse webhook payload")
		h.metrics.RecordRequest(start, false, 0)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Count events in payload
	eventCount := 0
	for _, entry := range payload.Entry {
		eventCount += len(entry.Changes)
	}

	// Process the webhook payload
	ctx := r.Context()
	if err := h.processPayload(ctx, &payload); err != nil {
		h.logger.Error().
			Err(err).
			Interface("payload", payload).
			Msg("Failed to process webhook payload")
		h.metrics.RecordRequest(start, false, eventCount)
		http.Error(w, "Processing failed", http.StatusInternalServerError)
		return
	}

	// Record successful processing
	h.metrics.RecordRequest(start, true, eventCount)

	// Respond with 200 OK to acknowledge receipt
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// processPayload processes the webhook payload and dispatches events.
func (h *Handler) processPayload(ctx context.Context, payload *WebhookPayload) error {
	h.logger.Info().
		Str("object", payload.Object).
		Int("entries", len(payload.Entry)).
		Msg("Processing webhook payload")

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if err := h.processChange(ctx, &change); err != nil {
				h.logger.Error().
					Err(err).
					Str("field", change.Field).
					Msg("Failed to process change")
				// Continue processing other changes even if one fails
			}
		}
	}

	return nil
}

// processChange processes a single change notification.
func (h *Handler) processChange(ctx context.Context, change *Change) error {
	value := &change.Value

	h.logger.Debug().
		Str("field", change.Field).
		Str("messaging_product", value.MessagingProduct).
		Str("phone_number_id", value.Metadata.PhoneNumberID).
		Int("messages", len(value.Messages)).
		Int("statuses", len(value.Statuses)).
		Int("errors", len(value.Errors)).
		Msg("Processing change")

	// Process messages
	for _, message := range value.Messages {
		if err := h.processMessage(ctx, &message, &value.Metadata); err != nil {
			h.logger.Error().
				Err(err).
				Str("message_id", message.ID).
				Str("message_type", message.Type).
				Msg("Failed to process message")
		}
	}

	// Process status updates
	for _, status := range value.Statuses {
		if err := h.processStatus(ctx, &status, &value.Metadata); err != nil {
			h.logger.Error().
				Err(err).
				Str("message_id", status.ID).
				Str("status", status.Status).
				Msg("Failed to process status")
		}
	}

	// Process errors
	for _, webhookError := range value.Errors {
		if err := h.processError(ctx, &webhookError, &value.Metadata); err != nil {
			h.logger.Error().
				Err(err).
				Int("error_code", webhookError.Code).
				Str("error_title", webhookError.Title).
				Msg("Failed to process error")
		}
	}

	return nil
}

// processMessage processes an incoming message.
func (h *Handler) processMessage(ctx context.Context, message *Message, metadata *Metadata) error {
	eventType := message.GetEventType()

	h.logger.Info().
		Str("event_type", string(eventType)).
		Str("message_id", message.ID).
		Str("from", message.From).
		Str("phone_number_id", metadata.PhoneNumberID).
		Msg("Processing message")

	// Process any errors in the message
	if errors := h.errorHandler.ProcessMessageErrors(ctx, message, metadata); len(errors) > 0 {
		h.logger.Warn().
			Str("message_id", message.ID).
			Int("error_count", len(errors)).
			Msg("Message contains errors")

		// Log each error but continue processing
		for _, err := range errors {
			h.logger.Error().Err(err).Msg("Message error")
		}
	}

	// Record message type metrics
	h.metrics.RecordMessageType(message.Type)

	return h.dispatcher.DispatchMessage(ctx, message, metadata)
}

// processStatus processes a status update.
func (h *Handler) processStatus(ctx context.Context, status *Status, metadata *Metadata) error {
	h.logger.Info().
		Str("event_type", "status.update").
		Str("message_id", status.ID).
		Str("status", status.Status).
		Str("recipient_id", status.RecipientID).
		Str("phone_number_id", metadata.PhoneNumberID).
		Msg("Processing status update")

	// Track status in both monitors
	h.statusMonitor.TrackStatus(ctx, status, metadata)

	// Track in message lifecycle tracker (with deduplication and stale filtering)
	if err := h.messageTracker.TrackStatusUpdate(ctx, status, metadata); err != nil {
		h.logger.Warn().
			Err(err).
			Str("message_id", status.ID).
			Msg("Failed to track status in message lifecycle tracker")
	}

	// Process any errors in the status
	if len(status.Errors) > 0 {
		for _, statusError := range status.Errors {
			classification := h.errorClassifier.ClassifyError(statusError.Code)

			h.logger.Error().
				Int("error_code", statusError.Code).
				Str("error_title", statusError.Title).
				Str("error_message", statusError.Message).
				Str("error_category", string(classification.Category)).
				Str("error_severity", string(classification.Severity)).
				Bool("retryable", classification.Retryable).
				Str("recommended_action", classification.Action).
				Msg("Status contains error")

			// Handle rate limit errors specifically
			if statusError.Code == 130429 || statusError.Code == 80007 || statusError.Code == 131048 {
				var limitType RateLimitType
				switch statusError.Code {
				case 130429:
					limitType = RateLimitTypeThroughput
				case 80007:
					limitType = RateLimitTypeAPI
				case 131048:
					limitType = RateLimitTypeSpam
				}

				if delay, err := h.rateLimiter.HandleRateLimitError(ctx, statusError.Code, limitType); err != nil {
					h.logger.Error().
						Err(err).
						Int("error_code", statusError.Code).
						Msg("Failed to handle rate limit error")
				} else {
					h.logger.Warn().
						Int("error_code", statusError.Code).
						Str("limit_type", string(limitType)).
						Dur("retry_delay", delay).
						Msg("Rate limit error detected, applying backoff")
				}
			}
		}
	}

	return h.dispatcher.DispatchStatus(ctx, status, metadata)
}

// processError processes an error notification.
func (h *Handler) processError(ctx context.Context, webhookError *Error, metadata *Metadata) error {
	h.logger.Error().
		Str("event_type", "error").
		Int("error_code", webhookError.Code).
		Str("error_title", webhookError.Title).
		Str("error_message", webhookError.Message).
		Str("phone_number_id", metadata.PhoneNumberID).
		Msg("Processing error notification")

	return h.dispatcher.DispatchError(ctx, webhookError, metadata)
}

// GetDispatcher returns the event dispatcher for registering handlers.
func (h *Handler) GetDispatcher() *EventDispatcher {
	return h.dispatcher
}

// GetStatusMonitor returns the status monitor for tracking message delivery.
func (h *Handler) GetStatusMonitor() *StatusMonitor {
	return h.statusMonitor
}

// GetMetrics returns the metrics collector for webhook analytics.
func (h *Handler) GetMetrics() *MetricsCollector {
	return h.metrics
}

// GetMessageTracker returns the message lifecycle tracker for detailed message tracking.
func (h *Handler) GetMessageTracker() *MessageLifecycleTracker {
	return h.messageTracker
}

// GetErrorClassifier returns the error classifier for error analysis.
func (h *Handler) GetErrorClassifier() *ErrorClassifier {
	return h.errorClassifier
}

// GetRateLimiter returns the rate limiter for rate limiting management.
func (h *Handler) GetRateLimiter() *RateLimiter {
	return h.rateLimiter
}

// GetMessageQueue returns the message queue manager for message queuing.
func (h *Handler) GetMessageQueue() *MessageQueueManager {
	return h.messageQueue
}

// TrackSentMessage records when a message is sent via the Cloud API.
// This should be called after successfully sending a message to correlate with webhook status updates.
func (h *Handler) TrackSentMessage(ctx context.Context, wamid, messageID, recipientID, messageType string) {
	h.messageTracker.TrackMessageSent(ctx, wamid, messageID, recipientID, messageType)

	h.logger.Info().
		Str("wamid", wamid).
		Str("message_id", messageID).
		Str("recipient_id", recipientID).
		Str("message_type", messageType).
		Msg("Tracking sent message for lifecycle monitoring")
}

// UpdateRateLimitInfo updates rate limit information from API responses or external monitoring.
func (h *Handler) UpdateRateLimitInfo(limitType RateLimitType, usage, limit int64, resetTime time.Time) {
	h.rateLimiter.UpdateRateLimit(limitType, usage, limit, resetTime)

	h.logger.Info().
		Str("limit_type", string(limitType)).
		Int64("usage", usage).
		Int64("limit", limit).
		Time("reset_time", resetTime).
		Msg("Updated rate limit information")
}

// UpdateQualityInfo updates quality rating and messaging tier information.
func (h *Handler) UpdateQualityInfo(tier RateLimitTier, quality QualityRating, throughputMPS int) {
	h.rateLimiter.UpdateQualityInfo(tier, quality, throughputMPS)

	h.logger.Info().
		Str("tier", string(tier)).
		Str("quality", string(quality)).
		Int("throughput_mps", throughputMPS).
		Msg("Updated quality and tier information")

	// Log quality warnings
	if quality == QualityLow {
		h.logger.Warn().
			Msg("⚠️  Quality rating is LOW - you have 7 days to improve or face reduced limits")

		recommendations := h.rateLimiter.GetQualityRecommendations()
		for _, rec := range recommendations {
			h.logger.Warn().Str("recommendation", rec).Msg("Quality improvement recommendation")
		}
	}
}

// SetLogger updates the handler's logger.
func (h *Handler) SetLogger(logger zerolog.Logger) {
	h.logger = &logger
}

// GetAppSecret returns the configured app secret (for testing purposes).
func (h *Handler) GetAppSecret() string {
	return h.appSecret
}

// GetVerifyToken returns the configured verify token (for testing purposes).
func (h *Handler) GetVerifyToken() string {
	return h.verifyToken
}

// Health checks the health of the webhook handler.
func (h *Handler) Health() error {
	if strings.TrimSpace(h.verifyToken) == "" {
		return fmt.Errorf("verify token is not configured")
	}
	return nil
}

// Close performs cleanup when the handler is no longer needed.
func (h *Handler) Close() error {
	if h.statusMonitor != nil {
		h.statusMonitor.Stop()
	}
	if h.metrics != nil {
		h.metrics.Stop()
	}
	if h.messageTracker != nil {
		h.messageTracker.Stop()
	}
	return nil
}
