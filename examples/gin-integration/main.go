//go:build gin
// +build gin

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp/internal/httpserver"
	"github.com/wongpinter/go-whatsapp/webhook"
)

func main() {
	// Setup logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Get configuration from environment
	appSecret := os.Getenv("WHATSAPP_APP_SECRET")
	verifyToken := os.Getenv("WHATSAPP_VERIFY_TOKEN")
	
	if appSecret == "" || verifyToken == "" {
		log.Println("Warning: WHATSAPP_APP_SECRET and WHATSAPP_VERIFY_TOKEN not set")
		log.Println("Using default values for demonstration")
		appSecret = "your-app-secret"
		verifyToken = "your-verify-token"
	}

	// Example 1: Using existing Gin engine
	log.Println("=== Example 1: Integration with Existing Gin Engine ===")
	existingGinExample(&logger, appSecret, verifyToken)

	// Example 2: Creating new Gin server via factory
	log.Println("\n=== Example 2: Creating Gin Server via Factory ===")
	factoryGinExample(&logger, appSecret, verifyToken)

	// Example 3: Full application with graceful shutdown
	log.Println("\n=== Example 3: Full Application with Graceful Shutdown ===")
	fullApplicationExample(&logger, appSecret, verifyToken)
}

func existingGinExample(logger *zerolog.Logger, appSecret, verifyToken string) {
	// Create Gin engine with custom configuration
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Add some existing routes
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "My WhatsApp Integration Service",
			"status":  "running",
		})
	})

	// Create WhatsApp webhook adapter
	webhookAdapter := webhook.NewServerAdapter(appSecret, verifyToken, logger)

	// Create server with existing Gin engine
	factory := httpserver.NewServerFactory()
	server, err := factory.CreateServer(
		httpserver.FrameworkGin,
		httpserver.WithNativeEngine(r),
		httpserver.WithDebug(false),
	)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Register webhook routes
	webhookAdapter.RegisterRoutes(server.Router())

	// Add API routes using the abstraction
	router := server.Router()
	apiGroup := router.Group("/api/v1")
	
	apiGroup.GET("/status", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, map[string]interface{}{
			"status":    "healthy",
			"framework": "gin",
			"timestamp": time.Now().Unix(),
		})
	})

	apiGroup.POST("/send-message", func(ctx httpserver.HTTPContext) error {
		// In a real application, you would send a WhatsApp message here
		return ctx.JSON(200, map[string]string{
			"message": "Message sent successfully",
			"id":      "msg_123456",
		})
	})

	log.Printf("Gin server configured with webhook routes")
	log.Printf("Available endpoints:")
	log.Printf("  GET  / - Service info")
	log.Printf("  GET  /webhook - Webhook verification")
	log.Printf("  POST /webhook - Webhook events")
	log.Printf("  GET  /health - Health check")
	log.Printf("  GET  /api/v1/status - API status")
	log.Printf("  POST /api/v1/send-message - Send message")
}

func factoryGinExample(logger *zerolog.Logger, appSecret, verifyToken string) {
	// Create server via factory (will create new Gin engine)
	factory := httpserver.NewServerFactory()
	server, err := factory.CreateServer(
		httpserver.FrameworkGin,
		httpserver.WithDebug(true),
		httpserver.WithTrustedProxies([]string{"127.0.0.1"}),
	)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Create webhook adapter
	webhookAdapter := webhook.NewServerAdapter(appSecret, verifyToken, logger)

	// Register routes
	router := server.Router()

	// Add global middleware
	router.Use(func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			start := time.Now()
			err := next(ctx)
			logger.Info().
				Str("method", ctx.Method()).
				Str("path", ctx.Path()).
				Dur("duration", time.Since(start)).
				Msg("Request processed")
			return err
		}
	})

	// Register webhook routes
	webhookAdapter.RegisterRoutes(router)

	// Add custom routes
	router.GET("/", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, map[string]interface{}{
			"service":   "WhatsApp Webhook Service",
			"framework": "gin",
			"created":   "via factory",
		})
	})

	log.Printf("Factory-created Gin server configured")
}

func fullApplicationExample(logger *zerolog.Logger, appSecret, verifyToken string) {
	// This example shows a complete application with graceful shutdown
	
	// Create Gin engine
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Create server
	factory := httpserver.NewServerFactory()
	server, err := factory.CreateServer(
		httpserver.FrameworkGin,
		httpserver.WithNativeEngine(r),
	)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup webhook
	webhookAdapter := webhook.NewServerAdapter(appSecret, verifyToken, logger)
	router := server.Router()

	// Add middleware
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware(logger))

	// Register webhook routes
	webhookAdapter.RegisterRoutes(router)

	// Add application routes
	setupApplicationRoutes(router)

	// Start server in goroutine
	go func() {
		logger.Info().Msg("Starting server on :8080")
		if err := server.Start(":8080"); err != nil {
			logger.Error().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	} else {
		logger.Info().Msg("Server exited gracefully")
	}
}

func setupApplicationRoutes(router httpserver.Router) {
	// API v1 routes
	v1 := router.Group("/api/v1")
	
	v1.GET("/health", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "whatsapp-webhook",
		})
	})

	v1.GET("/info", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, map[string]interface{}{
			"service":     "WhatsApp Integration Service",
			"framework":   "gin",
			"version":     "1.0.0",
			"endpoints": []string{
				"GET /webhook - Webhook verification",
				"POST /webhook - Webhook events",
				"GET /health - Health check",
				"GET /api/v1/info - Service info",
			},
		})
	})

	// Protected routes (would require authentication in real app)
	protected := v1.Group("/protected", authMiddleware())
	
	protected.GET("/stats", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, map[string]interface{}{
			"total_webhooks": 0,
			"last_webhook":   nil,
			"uptime":         time.Since(time.Now()).String(),
		})
	})
}

// Middleware functions

func corsMiddleware() httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			ctx.SetHeader("Access-Control-Allow-Origin", "*")
			ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Hub-Signature-256")

			if ctx.Method() == "OPTIONS" {
				return ctx.String(200, "OK")
			}

			return next(ctx)
		}
	}
}

func loggingMiddleware(logger *zerolog.Logger) httpserver.Middleware {
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

func authMiddleware() httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			// Simple auth check
			authHeader := ctx.Header("Authorization")
			if authHeader == "" {
				return ctx.JSON(401, map[string]string{
					"error": "Missing authorization header",
				})
			}

			// In real implementation, validate the token
			if authHeader != "Bearer valid-token" {
				return ctx.JSON(401, map[string]string{
					"error": "Invalid authorization token",
				})
			}

			return next(ctx)
		}
	}
}
