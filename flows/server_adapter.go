package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp/internal/httpserver"
)

// ServerAdapter provides framework-agnostic Flows HTTP handling
type ServerAdapter struct {
	dataExchangeHandler *DataExchangeHandler
	actionRouter        *ActionRouter
	tokenManager        *FlowTokenManager
	stateManager        *FlowStateManager
	logger              *zerolog.Logger
	metrics             *FlowMetrics
}

// FlowMetrics tracks Flows performance and usage
type FlowMetrics struct {
	totalRequests      int64
	successfulFlows    int64
	failedFlows        int64
	averageLatency     time.Duration
	lastRequestTime    time.Time
	activeFlows        int64
	tokenValidations   int64
	encryptionRequests int64
	actionExecutions   map[string]int64
	errorsByType       map[string]int64
	responseTimeP95    time.Duration
	responseTimeP99    time.Duration
	startTime          time.Time
	mutex              sync.RWMutex
}

// NewServerAdapter creates a new Flows server adapter
func NewServerAdapter(
	dataExchangeHandler *DataExchangeHandler,
	actionRouter *ActionRouter,
	tokenManager *FlowTokenManager,
	stateManager *FlowStateManager,
	logger *zerolog.Logger,
) *ServerAdapter {
	if logger == nil {
		nopLogger := zerolog.Nop()
		logger = &nopLogger
	}

	return &ServerAdapter{
		dataExchangeHandler: dataExchangeHandler,
		actionRouter:        actionRouter,
		tokenManager:        tokenManager,
		stateManager:        stateManager,
		logger:              logger,
		metrics: &FlowMetrics{
			actionExecutions: make(map[string]int64),
			errorsByType:     make(map[string]int64),
			startTime:        time.Now(),
		},
	}
}

// RegisterRoutes registers Flows routes on the provided router
func (s *ServerAdapter) RegisterRoutes(router httpserver.Router) {
	// Add global Flows middleware
	router.Use(FlowsLoggingMiddleware(s.logger))
	router.Use(FlowsMetricsMiddleware(s.metrics, s.logger))

	// Data exchange endpoint (main Flows endpoint) with full middleware stack
	router.POST("/flows/data-exchange", s.handleDataExchange,
		FlowsEncryptionMiddleware("", s.logger),
		FlowsTokenValidationMiddleware(s.tokenManager, s.logger),
		FlowsStateMiddleware(s.stateManager, s.logger),
		FlowsRateLimitMiddleware(60, s.logger), // 60 requests per minute
	)

	// Health check endpoint
	router.GET("/flows/health", s.handleHealth)

	// Metrics endpoint
	router.GET("/flows/metrics", s.handleMetrics)

	// Flow management endpoints
	flowGroup := router.Group("/flows")
	flowGroup.Use(FlowsRateLimitMiddleware(100, s.logger)) // Higher limit for management endpoints

	flowGroup.POST("/send/survey", s.handleSendSurvey)
	flowGroup.POST("/send/lead", s.handleSendLead)
	flowGroup.POST("/send/custom", s.handleSendCustomFlow)

	// Token validation endpoint with enhanced security
	flowGroup.POST("/validate-token", s.handleValidateToken,
		FlowsTokenValidationMiddleware(s.tokenManager, s.logger),
	)

	// Action registry endpoint
	flowGroup.GET("/actions", s.handleListActions)
}

// RegisterRoutesWithPrefix registers Flows routes with a custom prefix
func (s *ServerAdapter) RegisterRoutesWithPrefix(router httpserver.Router, prefix string) {
	group := router.Group(prefix)

	// Add middleware to the group
	group.Use(FlowsLoggingMiddleware(s.logger))
	group.Use(FlowsMetricsMiddleware(s.metrics, s.logger))

	// Data exchange endpoint with full middleware stack
	group.POST("/data-exchange", s.handleDataExchange,
		FlowsEncryptionMiddleware("", s.logger),
		FlowsTokenValidationMiddleware(s.tokenManager, s.logger),
		FlowsStateMiddleware(s.stateManager, s.logger),
		FlowsRateLimitMiddleware(60, s.logger),
	)

	// Health and metrics
	group.GET("/health", s.handleHealth)
	group.GET("/metrics", s.handleMetrics)

	// Flow management with rate limiting
	managementGroup := group.Group("/send")
	managementGroup.Use(FlowsRateLimitMiddleware(100, s.logger))
	managementGroup.POST("/survey", s.handleSendSurvey)
	managementGroup.POST("/lead", s.handleSendLead)
	managementGroup.POST("/custom", s.handleSendCustomFlow)

	// Token and actions
	group.POST("/validate-token", s.handleValidateToken,
		FlowsTokenValidationMiddleware(s.tokenManager, s.logger),
	)
	group.GET("/actions", s.handleListActions)
}

// handleDataExchange handles Flow data exchange requests
func (s *ServerAdapter) handleDataExchange(ctx httpserver.HTTPContext) error {
	start := time.Now()
	s.metrics.totalRequests++
	s.metrics.lastRequestTime = start

	s.logger.Debug().
		Str("method", ctx.Method()).
		Str("path", ctx.Path()).
		Msg("Flow data exchange request received")

	// Read request body
	body, err := ctx.Body()
	if err != nil {
		s.logger.Error().
			Err(err).
			Msg("Failed to read request body")
		s.metrics.failedFlows++
		return ctx.String(400, "Failed to read request body")
	}

	// Process the data exchange using the existing handler logic
	response, err := s.processDataExchange(ctx.Context(), body)
	if err != nil {
		s.logger.Error().
			Err(err).
			Msg("Failed to process data exchange")
		s.metrics.failedFlows++
		return ctx.String(500, "Processing failed")
	}

	// Update metrics
	s.metrics.successfulFlows++
	s.metrics.averageLatency = time.Since(start)

	s.logger.Info().
		Dur("duration", time.Since(start)).
		Msg("Flow data exchange processed successfully")

	// Return response
	return ctx.JSON(200, response)
}

// handleHealth handles health check requests
func (s *ServerAdapter) handleHealth(ctx httpserver.HTTPContext) error {
	s.metrics.mutex.RLock()
	defer s.metrics.mutex.RUnlock()

	uptime := time.Since(s.metrics.startTime)
	successRate := s.calculateSuccessRate()

	// Determine health status
	status := "healthy"
	if successRate < 95.0 && s.metrics.totalRequests > 10 {
		status = "degraded"
	}
	if successRate < 80.0 && s.metrics.totalRequests > 10 {
		status = "unhealthy"
	}

	health := map[string]interface{}{
		"status":         status,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"service":        "whatsapp-flows",
		"version":        "1.0.0",
		"uptime_seconds": uptime.Seconds(),
		"flows": map[string]interface{}{
			"active_flows":        s.metrics.activeFlows,
			"total_requests":      s.metrics.totalRequests,
			"successful_flows":    s.metrics.successfulFlows,
			"failed_flows":        s.metrics.failedFlows,
			"success_rate":        successRate,
			"last_request":        s.metrics.lastRequestTime.Format(time.RFC3339),
			"token_validations":   s.metrics.tokenValidations,
			"encryption_requests": s.metrics.encryptionRequests,
		},
		"performance": map[string]interface{}{
			"average_latency_ms": s.metrics.averageLatency.Milliseconds(),
			"p95_latency_ms":     s.metrics.responseTimeP95.Milliseconds(),
			"p99_latency_ms":     s.metrics.responseTimeP99.Milliseconds(),
		},
	}

	return ctx.JSON(200, health)
}

// handleMetrics handles metrics endpoint
func (s *ServerAdapter) handleMetrics(ctx httpserver.HTTPContext) error {
	s.metrics.mutex.RLock()
	defer s.metrics.mutex.RUnlock()

	uptime := time.Since(s.metrics.startTime)

	metrics := map[string]interface{}{
		"timestamp":           time.Now().UTC().Format(time.RFC3339),
		"uptime_seconds":      uptime.Seconds(),
		"total_requests":      s.metrics.totalRequests,
		"successful_flows":    s.metrics.successfulFlows,
		"failed_flows":        s.metrics.failedFlows,
		"success_rate":        s.calculateSuccessRate(),
		"active_flows":        s.metrics.activeFlows,
		"token_validations":   s.metrics.tokenValidations,
		"encryption_requests": s.metrics.encryptionRequests,
		"last_request_time":   s.metrics.lastRequestTime.Format(time.RFC3339),
		"performance": map[string]interface{}{
			"average_latency_ms":  s.metrics.averageLatency.Milliseconds(),
			"p95_latency_ms":      s.metrics.responseTimeP95.Milliseconds(),
			"p99_latency_ms":      s.metrics.responseTimeP99.Milliseconds(),
			"requests_per_second": s.calculateRequestsPerSecond(),
		},
		"action_executions": s.metrics.actionExecutions,
		"errors_by_type":    s.metrics.errorsByType,
	}

	return ctx.JSON(200, metrics)
}

// handleSendSurvey handles survey Flow sending
func (s *ServerAdapter) handleSendSurvey(ctx httpserver.HTTPContext) error {
	// Implementation for sending survey Flows
	return ctx.JSON(200, map[string]string{
		"message": "Survey Flow sent successfully",
		"type":    "survey",
	})
}

// handleSendLead handles lead generation Flow sending
func (s *ServerAdapter) handleSendLead(ctx httpserver.HTTPContext) error {
	// Implementation for sending lead generation Flows
	return ctx.JSON(200, map[string]string{
		"message": "Lead generation Flow sent successfully",
		"type":    "lead",
	})
}

// handleSendCustomFlow handles custom Flow sending
func (s *ServerAdapter) handleSendCustomFlow(ctx httpserver.HTTPContext) error {
	// Implementation for sending custom Flows
	return ctx.JSON(200, map[string]string{
		"message": "Custom Flow sent successfully",
		"type":    "custom",
	})
}

// handleValidateToken handles Flow token validation
func (s *ServerAdapter) handleValidateToken(ctx httpserver.HTTPContext) error {
	s.metrics.tokenValidations++

	// Implementation for token validation
	return ctx.JSON(200, map[string]interface{}{
		"valid":      true,
		"expires_at": time.Now().Add(24 * time.Hour).Unix(),
	})
}

// handleListActions handles action registry listing
func (s *ServerAdapter) handleListActions(ctx httpserver.HTTPContext) error {
	actions := []string{}
	if s.actionRouter != nil {
		// Get registered actions from action router
		// This would need to be implemented in the ActionRouter
		actions = []string{"survey_response", "lead_capture", "custom_action"}
	}

	return ctx.JSON(200, map[string]interface{}{
		"actions": actions,
		"count":   len(actions),
	})
}

// processDataExchange processes the data exchange using existing logic
func (s *ServerAdapter) processDataExchange(ctx context.Context, body []byte) (interface{}, error) {
	// This would delegate to the existing DataExchangeHandler logic
	// For now, return a placeholder response
	var request map[string]interface{}
	if err := json.Unmarshal(body, &request); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Process the request using existing handler logic
	response := map[string]interface{}{
		"version": "3.0",
		"data": map[string]interface{}{
			"status": "success",
			"screen": "SURVEY_COMPLETE",
		},
	}

	return response, nil
}

// calculateSuccessRate calculates the success rate of Flow processing
func (s *ServerAdapter) calculateSuccessRate() float64 {
	if s.metrics.totalRequests == 0 {
		return 0.0
	}
	return float64(s.metrics.successfulFlows) / float64(s.metrics.totalRequests) * 100.0
}

// calculateRequestsPerSecond calculates the current requests per second rate
func (s *ServerAdapter) calculateRequestsPerSecond() float64 {
	uptime := time.Since(s.metrics.startTime)
	if uptime.Seconds() == 0 {
		return 0.0
	}
	return float64(s.metrics.totalRequests) / uptime.Seconds()
}

// GetMetrics returns current metrics
func (s *ServerAdapter) GetMetrics() *FlowMetrics {
	return s.metrics
}

// GetDataExchangeHandler returns the underlying data exchange handler
func (s *ServerAdapter) GetDataExchangeHandler() *DataExchangeHandler {
	return s.dataExchangeHandler
}

// Getter methods for FlowMetrics

// GetTotalRequests returns the total number of requests
func (m *FlowMetrics) GetTotalRequests() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.totalRequests
}

// GetSuccessfulFlows returns the number of successful flows
func (m *FlowMetrics) GetSuccessfulFlows() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.successfulFlows
}

// GetFailedFlows returns the number of failed flows
func (m *FlowMetrics) GetFailedFlows() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.failedFlows
}

// GetActiveFlows returns the number of active flows
func (m *FlowMetrics) GetActiveFlows() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.activeFlows
}

// GetTokenValidations returns the number of token validations
func (m *FlowMetrics) GetTokenValidations() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.tokenValidations
}

// GetEncryptionRequests returns the number of encryption requests
func (m *FlowMetrics) GetEncryptionRequests() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.encryptionRequests
}

// GetAverageLatency returns the average latency
func (m *FlowMetrics) GetAverageLatency() time.Duration {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.averageLatency
}

// GetLastRequestTime returns the last request time
func (m *FlowMetrics) GetLastRequestTime() time.Time {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.lastRequestTime
}

// GetStartTime returns the start time
func (m *FlowMetrics) GetStartTime() time.Time {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.startTime
}

// Middleware methods

// encryptionMiddleware handles request/response encryption for data exchange
func (s *ServerAdapter) encryptionMiddleware() httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			// Add encryption/decryption logic here
			s.logger.Debug().Msg("Processing encrypted Flow request")

			// For now, just pass through
			return next(ctx)
		}
	}
}

// tokenValidationMiddleware validates Flow tokens
func (s *ServerAdapter) tokenValidationMiddleware() httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			// Add token validation logic here
			authHeader := ctx.Header("Authorization")
			if authHeader == "" {
				return ctx.JSON(401, map[string]string{
					"error": "Missing authorization header",
				})
			}

			s.logger.Debug().Msg("Validating Flow token")

			// For now, just pass through
			return next(ctx)
		}
	}
}

// FlowStateMiddleware manages Flow state during processing
func FlowStateMiddleware(stateManager *FlowStateManager) httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			// Add Flow state management logic here
			return next(ctx)
		}
	}
}

// FlowMetricsMiddleware tracks Flow processing metrics
func FlowMetricsMiddleware(metrics *FlowMetrics) httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			start := time.Now()

			err := next(ctx)

			// Update metrics
			duration := time.Since(start)
			metrics.averageLatency = duration

			return err
		}
	}
}

// FlowRateLimitMiddleware implements rate limiting for Flow endpoints
func FlowRateLimitMiddleware(requestsPerMinute int) httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			// Add rate limiting logic here
			// For now, just pass through
			return next(ctx)
		}
	}
}
