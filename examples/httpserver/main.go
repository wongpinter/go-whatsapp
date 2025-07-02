package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp/internal/httpserver"
	"github.com/wongpinter/go-whatsapp/webhook"
)

// Real-world integration example with actual Gin
func realGinIntegrationExample() {
	// This example shows how to integrate with a real Gin application
	// Uncomment the following code when you have Gin installed:

	/*
		import "github.com/gin-gonic/gin"

		func setupRealGinServer() {
			// Create Gin engine
			gin.SetMode(gin.ReleaseMode)
			r := gin.New()
			r.Use(gin.Logger(), gin.Recovery())

			// Create WhatsApp webhook adapter
			logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
			webhookAdapter := webhook.NewServerAdapter(
				os.Getenv("WHATSAPP_APP_SECRET"),
				os.Getenv("WHATSAPP_VERIFY_TOKEN"),
				&logger,
			)

			// Create server with existing Gin engine
			factory := httpserver.NewServerFactory()
			server, err := factory.CreateServer(
				httpserver.FrameworkGin,
				httpserver.WithNativeEngine(r),
			)
			if err != nil {
				log.Fatal(err)
			}

			// Register webhook routes
			webhookAdapter.RegisterRoutes(server.Router())

			// Add custom application routes
			router := server.Router()
			router.GET("/", func(ctx httpserver.HTTPContext) error {
				return ctx.JSON(200, gin.H{
					"service": "WhatsApp Integration",
					"framework": "gin",
					"endpoints": []string{"/webhook", "/health", "/metrics"},
				})
			})

			// API group with authentication
			apiGroup := router.Group("/api/v1", authMiddleware())
			apiGroup.GET("/messages", func(ctx httpserver.HTTPContext) error {
				// Get messages from database
				return ctx.JSON(200, gin.H{"messages": []string{}})
			})

			apiGroup.POST("/send", func(ctx httpserver.HTTPContext) error {
				// Send WhatsApp message
				return ctx.JSON(200, gin.H{"status": "sent"})
			})

			// Start server
			log.Println("Starting Gin server on :8080")
			log.Fatal(server.Start(":8080"))
		}
	*/
}

func main() {
	// Setup logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	fmt.Println("=== WhatsApp HTTP Server Framework Examples ===")

	// Example 1: Standard net/http server
	fmt.Println("1. Standard net/http Server Example")
	standardServerExample(&logger)

	// Example 2: Gin server (mock)
	fmt.Println("\n2. Gin Server Example")
	ginServerExample(&logger)

	// Example 3: Echo server (mock)
	fmt.Println("\n3. Echo Server Example")
	echoServerExample(&logger)

	// Example 4: Framework detection
	fmt.Println("\n4. Framework Detection Example")
	frameworkDetectionExample(&logger)

	// Example 5: Webhook integration
	fmt.Println("\n5. Webhook Integration Example")
	webhookIntegrationExample(&logger)

	fmt.Println("\nAll examples completed successfully!")
}

func standardServerExample(logger *zerolog.Logger) {
	// Create server factory
	factory := httpserver.NewServerFactory()

	// Create standard HTTP server
	server, err := factory.CreateServer(httpserver.FrameworkStandard)
	if err != nil {
		log.Fatalf("Failed to create standard server: %v", err)
	}

	// Get router and register routes
	router := server.Router()

	// Register some example routes
	router.GET("/", func(ctx httpserver.HTTPContext) error {
		return ctx.String(200, "Hello from Standard HTTP Server!")
	})

	router.POST("/api/data", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, map[string]string{
			"message":   "Data received",
			"framework": "net/http",
		})
	})

	// Add middleware
	router.Use(loggingMiddleware(logger))

	// Create a group with prefix
	apiGroup := router.Group("/api/v1")
	apiGroup.GET("/status", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, map[string]string{
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	fmt.Printf("Standard server created with framework: %T\n", server.Native())
	fmt.Println("Routes registered: /, /api/data, /api/v1/status")
}

func ginServerExample(logger *zerolog.Logger) {
	// Create server factory
	factory := httpserver.NewServerFactory()

	// Create Gin server (using mock engine)
	server, err := factory.CreateServer(httpserver.FrameworkGin,
		httpserver.WithDebug(true))
	if err != nil {
		log.Fatalf("Failed to create Gin server: %v", err)
	}

	// Get router and register routes
	router := server.Router()

	// Register routes
	router.GET("/", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, map[string]string{
			"message":   "Hello from Gin Server!",
			"framework": "gin",
		})
	})

	router.POST("/api/users", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(201, map[string]string{
			"message":   "User created",
			"framework": "gin",
		})
	})

	// Add middleware
	router.Use(corsMiddleware())

	// Create API group
	apiGroup := router.Group("/api/v1", authMiddleware())
	apiGroup.GET("/protected", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, map[string]string{
			"message": "Protected resource",
			"user":    "authenticated",
		})
	})

	fmt.Printf("Gin server created with framework: %T\n", server.Native())
	fmt.Println("Routes registered with Gin-style routing")
}

func echoServerExample(logger *zerolog.Logger) {
	// Create server factory
	factory := httpserver.NewServerFactory()

	// Create Echo server (using mock engine)
	server, err := factory.CreateServer(httpserver.FrameworkEcho,
		httpserver.WithDebug(false))
	if err != nil {
		log.Fatalf("Failed to create Echo server: %v", err)
	}

	// Get router and register routes
	router := server.Router()

	// Register routes
	router.GET("/", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, map[string]interface{}{
			"message":   "Hello from Echo Server!",
			"framework": "echo",
			"timestamp": time.Now().Unix(),
		})
	})

	router.PUT("/api/items/:id", func(ctx httpserver.HTTPContext) error {
		// In a real implementation, you'd extract the ID from the path
		return ctx.JSON(200, map[string]string{
			"message":   "Item updated",
			"framework": "echo",
		})
	})

	// Add middleware
	router.Use(timeoutMiddleware(30 * time.Second))

	// Create nested groups
	apiGroup := router.Group("/api")
	v1Group := apiGroup.Group("/v1")
	v1Group.GET("/health", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, map[string]string{
			"status":  "healthy",
			"service": "echo-api",
		})
	})

	fmt.Printf("Echo server created with framework: %T\n", server.Native())
	fmt.Println("Routes registered with Echo-style routing")
}

func frameworkDetectionExample(logger *zerolog.Logger) {
	// Test framework detection
	frameworks := []httpserver.Framework{
		httpserver.FrameworkStandard,
		httpserver.FrameworkGin,
		httpserver.FrameworkEcho,
	}

	factory := httpserver.NewServerFactory()

	for _, framework := range frameworks {
		server, err := factory.CreateServer(framework)
		if err != nil {
			fmt.Printf("❌ Failed to create %s server: %v\n", framework, err)
			continue
		}

		nativeEngine := server.Native()
		detectedFramework := httpserver.DetectFramework(nativeEngine)

		fmt.Printf("✅ Framework: %s, Detected: %s, Match: %v\n",
			framework, detectedFramework, framework == detectedFramework)
	}

	// Test supported frameworks
	supported := httpserver.GetSupportedFrameworks()
	fmt.Printf("Supported frameworks: %v\n", supported)
}

func webhookIntegrationExample(logger *zerolog.Logger) {
	// Create webhook server adapter
	appSecret := "your-app-secret"
	verifyToken := "your-verify-token"

	webhookAdapter := webhook.NewServerAdapter(appSecret, verifyToken, logger)

	// Test with different frameworks
	frameworks := []httpserver.Framework{
		httpserver.FrameworkStandard,
		httpserver.FrameworkGin,
		httpserver.FrameworkEcho,
	}

	factory := httpserver.NewServerFactory()

	for _, framework := range frameworks {
		fmt.Printf("\n--- Webhook Integration with %s ---\n", framework)

		server, err := factory.CreateServer(framework)
		if err != nil {
			fmt.Printf("❌ Failed to create %s server: %v\n", framework, err)
			continue
		}

		router := server.Router()

		// Register webhook routes
		webhookAdapter.RegisterRoutes(router)

		// Add additional routes
		router.GET("/", func(ctx httpserver.HTTPContext) error {
			return ctx.JSON(200, map[string]string{
				"service":   "WhatsApp Webhook Server",
				"framework": string(framework),
				"endpoints": "/webhook, /health, /metrics",
			})
		})

		fmt.Printf("✅ Webhook routes registered on %s server\n", framework)
		fmt.Println("Available endpoints:")
		fmt.Println("  GET  / - Service info")
		fmt.Println("  GET  /webhook - Webhook verification")
		fmt.Println("  POST /webhook - Webhook events")
		fmt.Println("  GET  /health - Health check")
		fmt.Println("  GET  /metrics - Metrics")
	}
}

// Example middleware functions

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

func corsMiddleware() httpserver.Middleware {
	return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
		return func(ctx httpserver.HTTPContext) error {
			ctx.SetHeader("Access-Control-Allow-Origin", "*")
			ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if ctx.Method() == "OPTIONS" {
				return ctx.String(200, "OK")
			}

			return next(ctx)
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
			return next(ctx)
		}
	}
}

func timeoutMiddleware(timeout time.Duration) httpserver.Middleware {
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

// Example of running a server (commented out to avoid blocking in examples)
func runServerExample() {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Create server
	factory := httpserver.NewServerFactory()
	server, err := factory.CreateServer(httpserver.FrameworkStandard)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup routes
	router := server.Router()
	router.GET("/", func(ctx httpserver.HTTPContext) error {
		return ctx.String(200, "Hello World!")
	})

	// Setup graceful shutdown
	go func() {
		if err := server.Start(":8080"); err != nil {
			logger.Error().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server exited")
}
