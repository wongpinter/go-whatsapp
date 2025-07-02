package webhook

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp/internal/crypto"
	"github.com/wongpinter/go-whatsapp/internal/httpserver"
)

// ServerAdapter provides framework-agnostic webhook handling
type ServerAdapter struct {
	handler     *Handler
	logger      *zerolog.Logger
	appSecret   string
	verifyToken string
}

// NewServerAdapter creates a new server adapter for webhooks
func NewServerAdapter(appSecret, verifyToken string, logger *zerolog.Logger) *ServerAdapter {
	if logger == nil {
		nopLogger := zerolog.Nop()
		logger = &nopLogger
	}

	// Create the main webhook handler
	handler := NewHandler(appSecret, verifyToken, *logger)

	return &ServerAdapter{
		handler:     handler,
		logger:      logger,
		appSecret:   appSecret,
		verifyToken: verifyToken,
	}
}

// GetHandler returns the underlying webhook handler
func (s *ServerAdapter) GetHandler() *Handler {
	return s.handler
}

// RegisterRoutes registers webhook routes on the provided router
func (s *ServerAdapter) RegisterRoutes(router httpserver.Router) {
	// Webhook verification endpoint (GET)
	router.GET("/webhook", s.handleVerification)

	// Webhook event endpoint (POST)
	router.POST("/webhook", s.handleWebhook, s.signatureValidationMiddleware())

	// Health check endpoint
	router.GET("/health", s.handleHealth)

	// Metrics endpoint (if metrics are enabled)
	if s.handler.metrics != nil {
		router.GET("/metrics", s.handleMetrics)
	}
}

// RegisterRoutesWithPrefix registers webhook routes with a custom prefix
func (s *ServerAdapter) RegisterRoutesWithPrefix(router httpserver.Router, prefix string) {
	group := router.Group(prefix)

	group.GET("/webhook", s.handleVerification)
	group.POST("/webhook", s.handleWebhook, s.signatureValidationMiddleware())
	group.GET("/health", s.handleHealth)

	if s.handler.metrics != nil {
		group.GET("/metrics", s.handleMetrics)
	}
}

// handleVerification handles webhook verification requests (GET)
func (s *ServerAdapter) handleVerification(ctx httpserver.HTTPContext) error {
	s.logger.Debug().
		Str("method", ctx.Method()).
		Str("path", ctx.Path()).
		Msg("Webhook verification request received")

	// Get verification parameters
	mode := ctx.Query("hub.mode")
	token := ctx.Query("hub.verify_token")
	challenge := ctx.Query("hub.challenge")

	// Validate verification request
	if mode != "subscribe" {
		s.logger.Warn().
			Str("mode", mode).
			Msg("Invalid verification mode")
		return ctx.String(400, "Invalid mode")
	}

	if token != s.verifyToken {
		s.logger.Warn().
			Str("provided_token", token).
			Msg("Invalid verify token")
		return ctx.String(403, "Invalid verify token")
	}

	if challenge == "" {
		s.logger.Warn().Msg("Missing challenge parameter")
		return ctx.String(400, "Missing challenge")
	}

	s.logger.Info().
		Str("challenge", challenge).
		Msg("Webhook verification successful")

	// Return the challenge to complete verification
	return ctx.String(200, challenge)
}

// handleWebhook handles webhook event requests (POST)
func (s *ServerAdapter) handleWebhook(ctx httpserver.HTTPContext) error {
	start := time.Now()

	s.logger.Debug().
		Str("method", ctx.Method()).
		Str("path", ctx.Path()).
		Msg("Webhook event request received")

	// Read request body
	body, err := ctx.Body()
	if err != nil {
		s.logger.Error().
			Err(err).
			Msg("Failed to read request body")
		return ctx.String(400, "Failed to read request body")
	}

	// Parse the webhook payload
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		s.logger.Error().
			Err(err).
			Str("body", string(body)).
			Msg("Failed to parse webhook payload")
		return ctx.String(400, "Invalid JSON")
	}

	// Count events in payload
	eventCount := 0
	for _, entry := range payload.Entry {
		eventCount += len(entry.Changes)
	}

	// Process the webhook payload
	if err := s.processPayload(ctx.Context(), &payload); err != nil {
		s.logger.Error().
			Err(err).
			Interface("payload", payload).
			Msg("Failed to process webhook payload")

		// Record failed processing
		if s.handler.metrics != nil {
			s.handler.metrics.RecordRequest(start, false, eventCount)
		}

		return ctx.String(500, "Processing failed")
	}

	// Record successful processing
	if s.handler.metrics != nil {
		s.handler.metrics.RecordRequest(start, true, eventCount)
	}

	s.logger.Info().
		Int("event_count", eventCount).
		Dur("duration", time.Since(start)).
		Msg("Webhook processed successfully")

	// Respond with 200 OK to acknowledge receipt
	return ctx.String(200, "OK")
}

// handleHealth handles health check requests
func (s *ServerAdapter) handleHealth(ctx httpserver.HTTPContext) error {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "whatsapp-webhook",
	}

	// Add metrics if available
	if s.handler.metrics != nil {
		metrics := s.handler.metrics.GetMetrics()
		health["metrics"] = map[string]interface{}{
			"total_requests":      metrics["total_requests"],
			"successful_requests": metrics["successful_events"],
			"failed_requests":     metrics["failed_events"],
		}
	}

	return ctx.JSON(200, health)
}

// handleMetrics handles metrics endpoint
func (s *ServerAdapter) handleMetrics(ctx httpserver.HTTPContext) error {
	if s.handler.metrics == nil {
		return ctx.String(404, "Metrics not enabled")
	}

	// Get all metrics from the collector
	metrics := s.handler.metrics.GetMetrics()

	return ctx.JSON(200, metrics)
}

// signatureValidationMiddleware validates webhook signatures
func (s *ServerAdapter) signatureValidationMiddleware() httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			// Only validate POST requests
			if ctx.Method() != "POST" {
				return next(ctx)
			}

			// Get signature from header
			signature := ctx.Header("X-Hub-Signature-256")
			if signature == "" {
				s.logger.Warn().Msg("Missing signature header")
				return ctx.String(401, "Missing signature")
			}

			// Read body for signature verification
			body, err := ctx.Body()
			if err != nil {
				s.logger.Error().
					Err(err).
					Msg("Failed to read body for signature verification")
				return ctx.String(400, "Failed to read body")
			}

			// Verify signature
			if err := crypto.VerifySignature(body, signature, s.appSecret); err != nil {
				s.logger.Warn().
					Err(err).
					Str("signature", signature).
					Msg("Signature verification failed")
				return ctx.String(401, "Invalid signature")
			}

			s.logger.Debug().Msg("Signature verification successful")
			return next(ctx)
		}
	}
}

// processPayload processes the webhook payload using the main handler
func (s *ServerAdapter) processPayload(ctx context.Context, payload *WebhookPayload) error {
	// Delegate to the main handler's processing logic
	return s.handler.processPayload(ctx, payload)
}

// Middleware helpers

// LoggingMiddleware logs webhook requests
func LoggingMiddleware(logger *zerolog.Logger) httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			start := time.Now()

			logger.Info().
				Str("method", ctx.Method()).
				Str("path", ctx.Path()).
				Msg("Request started")

			err := next(ctx)

			logger.Info().
				Str("method", ctx.Method()).
				Str("path", ctx.Path()).
				Dur("duration", time.Since(start)).
				Msg("Request completed")

			return err
		}
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware() httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			ctx.SetHeader("Access-Control-Allow-Origin", "*")
			ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, X-Hub-Signature-256")

			if ctx.Method() == "OPTIONS" {
				return ctx.String(200, "OK")
			}

			return next(ctx)
		}
	}
}

// TimeoutMiddleware adds request timeout
func TimeoutMiddleware(timeout time.Duration) httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			timeoutCtx, cancel := context.WithTimeout(ctx.Context(), timeout)
			defer cancel()

			ctx.WithContext(timeoutCtx)

			done := make(chan error, 1)
			go func() {
				done <- next(ctx)
			}()

			select {
			case err := <-done:
				return err
			case <-timeoutCtx.Done():
				return ctx.String(408, "Request timeout")
			}
		}
	}
}
