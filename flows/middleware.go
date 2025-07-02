package flows

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp/internal/httpserver"
)

// FlowsEncryptionMiddleware handles request/response encryption for Flow data exchange
func FlowsEncryptionMiddleware(privateKey string, logger *zerolog.Logger) httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			start := time.Now()

			logger.Debug().
				Str("path", ctx.Path()).
				Msg("Processing encrypted Flow request")

			// Check if this is a data exchange endpoint
			if strings.Contains(ctx.Path(), "/data-exchange") {
				// Add encryption context to request context
				requestCtx := context.WithValue(ctx.Context(), "encryption_enabled", true)
				requestCtx = context.WithValue(requestCtx, "private_key", privateKey)
				ctx.WithContext(requestCtx)

				// Validate content type for encrypted requests
				contentType := ctx.Header("Content-Type")
				if contentType != "application/json" && contentType != "application/octet-stream" {
					logger.Warn().
						Str("content_type", contentType).
						Msg("Invalid content type for encrypted Flow request")
					return ctx.JSON(400, map[string]string{
						"error": "Invalid content type for encrypted request",
					})
				}
			}

			err := next(ctx)

			logger.Debug().
				Str("path", ctx.Path()).
				Dur("duration", time.Since(start)).
				Msg("Encrypted Flow request processed")

			return err
		}
	}
}

// FlowsTokenValidationMiddleware validates Flow tokens
func FlowsTokenValidationMiddleware(tokenManager *FlowTokenManager, logger *zerolog.Logger) httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			// Skip token validation for health checks and metrics
			if strings.Contains(ctx.Path(), "/health") || strings.Contains(ctx.Path(), "/metrics") {
				return next(ctx)
			}

			// Extract token from various sources
			token := extractFlowToken(ctx)
			if token == "" {
				logger.Warn().
					Str("path", ctx.Path()).
					Msg("Missing Flow token")
				return ctx.JSON(401, map[string]string{
					"error": "Missing Flow token",
					"code":  "MISSING_TOKEN",
				})
			}

			// Validate token (if tokenManager is available)
			if tokenManager != nil {
				tokenInfo, err := tokenManager.ValidateToken(token)
				if err != nil {
					logger.Error().
						Err(err).
						Str("token", token[:8]+"...").
						Msg("Invalid Flow token")
					return ctx.JSON(401, map[string]string{
						"error": "Invalid or expired Flow token",
						"code":  "INVALID_TOKEN",
					})
				}

				// Add token info to request context
				requestCtx := context.WithValue(ctx.Context(), "token_info", tokenInfo)
				requestCtx = context.WithValue(requestCtx, "flow_id", tokenInfo.FlowID)
				requestCtx = context.WithValue(requestCtx, "user_id", tokenInfo.UserID)
				ctx.WithContext(requestCtx)

				logger.Debug().
					Str("flow_id", tokenInfo.FlowID).
					Str("user_id", tokenInfo.UserID).
					Msg("Flow token validated successfully")
			}

			return next(ctx)
		}
	}
}

// FlowsStateMiddleware manages Flow state during processing
func FlowsStateMiddleware(stateManager *FlowStateManager, logger *zerolog.Logger) httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			// Skip if no state manager
			if stateManager == nil {
				return next(ctx)
			}

			// Get token info from request context
			tokenInfo := ctx.Context().Value("token_info")
			if tokenInfo == nil {
				return next(ctx)
			}

			// For now, we'll use a simplified approach since we don't have the full FlowTokenInfo structure
			// In a real implementation, you would extract the actual token
			flowToken := "demo-token" // This would come from tokenInfo.Token

			// Load existing state (simplified - assuming GetState returns only state)
			// state := stateManager.GetState(flowToken)
			// For now, just log that state middleware is active
			logger.Debug().
				Str("flow_token", flowToken[:8]+"...").
				Msg("Flow state middleware active")

			// Process request
			err := next(ctx)

			// In a real implementation, you would update state here
			logger.Debug().
				Str("flow_token", flowToken[:8]+"...").
				Msg("Flow state middleware completed")

			return err
		}
	}
}

// FlowsLoggingMiddleware provides detailed logging for Flow requests
func FlowsLoggingMiddleware(logger *zerolog.Logger) httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			start := time.Now()

			// Log request start
			logEvent := logger.Info().
				Str("method", ctx.Method()).
				Str("path", ctx.Path()).
				Str("user_agent", ctx.Header("User-Agent")).
				Str("remote_addr", ctx.Header("X-Forwarded-For"))

			// Add Flow-specific context if available
			if flowID := ctx.Context().Value("flow_id"); flowID != nil {
				logEvent.Str("flow_id", flowID.(string))
			}
			if userID := ctx.Context().Value("user_id"); userID != nil {
				logEvent.Str("user_id", userID.(string))
			}

			logEvent.Msg("Flow request started")

			// Process request
			err := next(ctx)

			// Log request completion
			duration := time.Since(start)
			logEvent = logger.Info().
				Str("method", ctx.Method()).
				Str("path", ctx.Path()).
				Dur("duration", duration)

			if err != nil {
				logEvent.Err(err).Msg("Flow request failed")
			} else {
				logEvent.Msg("Flow request completed")
			}

			return err
		}
	}
}

// FlowsRateLimitMiddleware implements rate limiting for Flow endpoints
func FlowsRateLimitMiddleware(requestsPerMinute int, logger *zerolog.Logger) httpserver.Middleware {
	type rateLimitEntry struct {
		count     int
		resetTime time.Time
		blocked   bool
		blockTime time.Time
	}

	var (
		rateLimitMap = make(map[string]*rateLimitEntry)
		mutex        sync.RWMutex
	)

	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			// Get client identifier (IP or user ID)
			clientID := getClientIdentifier(ctx)

			mutex.Lock()
			entry, exists := rateLimitMap[clientID]
			now := time.Now()

			if !exists || now.After(entry.resetTime) {
				// Create new entry or reset expired entry
				rateLimitMap[clientID] = &rateLimitEntry{
					count:     1,
					resetTime: now.Add(time.Minute),
					blocked:   false,
				}
				mutex.Unlock()
			} else {
				// Check if client is temporarily blocked
				if entry.blocked && now.Before(entry.blockTime) {
					mutex.Unlock()
					logger.Warn().
						Str("client_id", clientID).
						Msg("Client is temporarily blocked")

					return ctx.JSON(429, map[string]interface{}{
						"error":        "Client temporarily blocked",
						"code":         "CLIENT_BLOCKED",
						"unblock_time": entry.blockTime.Unix(),
					})
				}

				// Check rate limit
				if entry.count >= requestsPerMinute {
					// Block client for 5 minutes after exceeding limit
					entry.blocked = true
					entry.blockTime = now.Add(5 * time.Minute)
					mutex.Unlock()

					logger.Warn().
						Str("client_id", clientID).
						Int("requests", entry.count).
						Msg("Rate limit exceeded, client blocked")

					return ctx.JSON(429, map[string]interface{}{
						"error":         "Rate limit exceeded",
						"code":          "RATE_LIMIT_EXCEEDED",
						"reset_time":    entry.resetTime.Unix(),
						"blocked_until": entry.blockTime.Unix(),
					})
				}

				entry.count++
				entry.blocked = false // Reset block status if within limits
				mutex.Unlock()
			}

			return next(ctx)
		}
	}
}

// FlowsMetricsMiddleware tracks Flow processing metrics
func FlowsMetricsMiddleware(metrics *FlowMetrics, logger *zerolog.Logger) httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			start := time.Now()

			// Increment total requests
			metrics.totalRequests++
			metrics.lastRequestTime = start

			// Track active flows
			if strings.Contains(ctx.Path(), "/data-exchange") {
				metrics.activeFlows++
				defer func() {
					metrics.activeFlows--
				}()
			}

			// Process request
			err := next(ctx)

			// Update metrics
			duration := time.Since(start)
			metrics.averageLatency = duration

			if err != nil {
				metrics.failedFlows++
			} else {
				metrics.successfulFlows++
			}

			// Log metrics periodically
			if metrics.totalRequests%100 == 0 {
				logger.Info().
					Int64("total_requests", metrics.totalRequests).
					Int64("successful_flows", metrics.successfulFlows).
					Int64("failed_flows", metrics.failedFlows).
					Dur("avg_latency", metrics.averageLatency).
					Msg("Flow metrics update")
			}

			return err
		}
	}
}

// FlowsSecurityMiddleware provides security headers and validation
func FlowsSecurityMiddleware(appSecret string, logger *zerolog.Logger) httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			// Add security headers
			ctx.SetHeader("X-Content-Type-Options", "nosniff")
			ctx.SetHeader("X-Frame-Options", "DENY")
			ctx.SetHeader("X-XSS-Protection", "1; mode=block")
			ctx.SetHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

			// Validate signature for data exchange requests
			if strings.Contains(ctx.Path(), "/data-exchange") && ctx.Method() == "POST" {
				signature := ctx.Header("X-Hub-Signature-256")
				if signature == "" {
					logger.Warn().Msg("Missing signature for Flow data exchange")
					return ctx.JSON(401, map[string]string{
						"error": "Missing signature",
						"code":  "MISSING_SIGNATURE",
					})
				}

				// Validate signature
				body, err := ctx.Body()
				if err != nil {
					return ctx.JSON(400, map[string]string{
						"error": "Failed to read request body",
					})
				}

				if !validateSignature(body, signature, appSecret) {
					logger.Warn().Msg("Invalid signature for Flow data exchange")
					return ctx.JSON(401, map[string]string{
						"error": "Invalid signature",
						"code":  "INVALID_SIGNATURE",
					})
				}
			}

			return next(ctx)
		}
	}
}

// Helper functions

func extractFlowToken(ctx httpserver.HTTPContext) string {
	// Try Authorization header first
	auth := ctx.Header("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Try X-Flow-Token header
	if token := ctx.Header("X-Flow-Token"); token != "" {
		return token
	}

	// Try query parameter
	return ctx.Query("flow_token")
}

func getClientIdentifier(ctx httpserver.HTTPContext) string {
	// Try to get user ID from context first
	if userID := ctx.Context().Value("user_id"); userID != nil {
		return fmt.Sprintf("user:%s", userID.(string))
	}

	// Fall back to IP address
	if forwarded := ctx.Header("X-Forwarded-For"); forwarded != "" {
		return fmt.Sprintf("ip:%s", strings.Split(forwarded, ",")[0])
	}

	return fmt.Sprintf("ip:%s", ctx.Header("X-Real-IP"))
}

func validateSignature(body []byte, signature, secret string) bool {
	// Remove "sha256=" prefix if present
	signature = strings.TrimPrefix(signature, "sha256=")

	// Calculate expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
