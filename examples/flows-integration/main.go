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

	"github.com/wongpinter/go-whatsapp/flows"
	"github.com/wongpinter/go-whatsapp/internal/httpserver"
	"github.com/wongpinter/go-whatsapp/webhook"
)

func main() {
	// Setup logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	fmt.Println("🚀 WhatsApp Flows HTTP Server Integration Example")
	fmt.Println("================================================")

	// Get configuration from environment
	appSecret := getEnvOrDefault("WHATSAPP_APP_SECRET", "demo-app-secret")
	verifyToken := getEnvOrDefault("WHATSAPP_VERIFY_TOKEN", "demo-verify-token")
	privateKeyPEM := getEnvOrDefault("FLOWS_PRIVATE_KEY", generateDemoPrivateKey())

	// Example 1: Standard HTTP server with Flows
	fmt.Println("\n📋 Example 1: Standard HTTP Server with Flows")
	standardHTTPExample(&logger, appSecret, verifyToken, privateKeyPEM)

	// Example 2: Framework-agnostic Flows server
	fmt.Println("\n📋 Example 2: Framework-Agnostic Flows Server")
	frameworkAgnosticExample(&logger, appSecret, verifyToken, privateKeyPEM)

	// Example 3: Full application with graceful shutdown
	fmt.Println("\n📋 Example 3: Full Application (Demo Mode)")
	fullApplicationExample(&logger, appSecret, verifyToken, privateKeyPEM)

	fmt.Println("\n🎉 All Flows integration examples completed successfully!")
	fmt.Println("\n📖 Usage Instructions:")
	fmt.Println("  • Set environment variables:")
	fmt.Println("    export WHATSAPP_APP_SECRET=your-app-secret")
	fmt.Println("    export WHATSAPP_VERIFY_TOKEN=your-verify-token")
	fmt.Println("    export FLOWS_PRIVATE_KEY=your-private-key-pem")
	fmt.Println("  • Run with different frameworks:")
	fmt.Println("    go run examples/flows-integration/main.go")
	fmt.Println("    go run -tags gin examples/flows-integration/main.go")
	fmt.Println("    go run -tags echo examples/flows-integration/main.go")
}

func standardHTTPExample(logger *zerolog.Logger, appSecret, verifyToken, privateKeyPEM string) {
	// Create Flows components
	dataExchangeHandler, err := createDataExchangeHandler(privateKeyPEM, logger)
	if err != nil {
		log.Printf("❌ Failed to create data exchange handler: %v", err)
		return
	}

	// Create Flows server adapter
	flowsAdapter := flows.NewServerAdapter(
		dataExchangeHandler,
		nil, // actionRouter
		nil, // tokenManager
		nil, // stateManager
		logger,
	)

	// Create HTTP server
	factory := httpserver.NewServerFactory()
	server, err := factory.CreateServer(httpserver.FrameworkStandard)
	if err != nil {
		log.Printf("❌ Failed to create server: %v", err)
		return
	}

	// Register Flows routes
	flowsAdapter.RegisterRoutes(server.Router())

	// Add webhook integration
	webhookAdapter := webhook.NewServerAdapter(appSecret, verifyToken, logger)
	webhookAdapter.RegisterRoutes(server.Router())

	fmt.Printf("✅ Standard HTTP server configured with Flows\n")
	fmt.Printf("   Framework: %T\n", server.Native())
	fmt.Printf("   Flows endpoints: /flows/data-exchange, /flows/health, /flows/metrics\n")
	fmt.Printf("   Webhook endpoints: /webhook\n")
}

func frameworkAgnosticExample(logger *zerolog.Logger, appSecret, verifyToken, privateKeyPEM string) {
	// Test with different frameworks
	frameworks := []httpserver.Framework{
		httpserver.FrameworkStandard,
		httpserver.FrameworkGin,
		httpserver.FrameworkEcho,
	}

	factory := httpserver.NewServerFactory()

	for _, framework := range frameworks {
		fmt.Printf("\n   Testing %s framework:\n", framework)

		server, err := factory.CreateServer(framework)
		if err != nil {
			fmt.Printf("   ❌ %s: %v\n", framework, err)
			continue
		}

		// Create Flows components
		dataExchangeHandler, err := createDataExchangeHandler(privateKeyPEM, logger)
		if err != nil {
			fmt.Printf("   ❌ %s: Failed to create handler: %v\n", framework, err)
			continue
		}

		// Create Flows server adapter
		flowsAdapter := flows.NewServerAdapter(
			dataExchangeHandler,
			nil, nil, nil,
			logger,
		)

		// Register routes
		router := server.Router()

		// Add middleware
		router.Use(loggingMiddleware(logger))

		// Register Flows routes with prefix
		flowsAdapter.RegisterRoutesWithPrefix(router, "/api/v1/flows")

		// Add webhook integration
		webhookAdapter := webhook.NewServerAdapter(appSecret, verifyToken, logger)
		webhookAdapter.RegisterRoutesWithPrefix(router, "/api/v1")

		fmt.Printf("   ✅ %s: Flows integration successful\n", framework)
		fmt.Printf("      Engine: %T\n", server.Native())
		fmt.Printf("      Endpoints: /api/v1/flows/*, /api/v1/webhook\n")
	}
}

func fullApplicationExample(logger *zerolog.Logger, appSecret, verifyToken, privateKeyPEM string) {
	// This example shows a complete application setup (demo mode)
	
	// Create server
	factory := httpserver.NewServerFactory()
	server, err := factory.CreateServer(httpserver.FrameworkStandard)
	if err != nil {
		log.Printf("❌ Failed to create server: %v", err)
		return
	}

	// Create Flows components
	dataExchangeHandler, err := createDataExchangeHandler(privateKeyPEM, logger)
	if err != nil {
		log.Printf("❌ Failed to create data exchange handler: %v", err)
		return
	}

	// Create adapters
	flowsAdapter := flows.NewServerAdapter(dataExchangeHandler, nil, nil, nil, logger)
	webhookAdapter := webhook.NewServerAdapter(appSecret, verifyToken, logger)

	// Setup routes
	router := server.Router()

	// Add global middleware
	router.Use(corsMiddleware())
	router.Use(loggingMiddleware(logger))

	// Register service routes
	router.GET("/", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, map[string]interface{}{
			"service":   "WhatsApp Flows Integration",
			"version":   "1.0.0",
			"framework": detectFramework(server),
			"endpoints": map[string][]string{
				"flows": {
					"POST /flows/data-exchange",
					"GET /flows/health",
					"GET /flows/metrics",
					"POST /flows/send/*",
				},
				"webhook": {
					"GET /webhook",
					"POST /webhook",
					"GET /health",
				},
			},
		})
	})

	// Register Flows and webhook routes
	flowsAdapter.RegisterRoutes(router)
	webhookAdapter.RegisterRoutes(router)

	fmt.Printf("✅ Full application configured\n")
	fmt.Printf("   Framework: %s\n", detectFramework(server))
	fmt.Printf("   Ready for production deployment\n")
	fmt.Printf("   Demo mode: Server configured but not started\n")
}

// Helper functions

func createDataExchangeHandler(privateKeyPEM string, logger *zerolog.Logger) (*flows.DataExchangeHandler, error) {
	// For demo purposes, create a minimal handler
	// In a real application, you would parse the actual private key
	return &flows.DataExchangeHandler{}, nil
}

func generateDemoPrivateKey() string {
	return `-----BEGIN RSA PRIVATE KEY-----
DEMO_KEY_FOR_TESTING_ONLY
-----END RSA PRIVATE KEY-----`
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func detectFramework(server httpserver.Server) string {
	nativeType := fmt.Sprintf("%T", server.Native())
	switch {
	case nativeType == "*gin.Engine":
		return "gin"
	case nativeType == "*echo.Echo":
		return "echo"
	default:
		return "standard"
	}
}

// Middleware functions

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
			ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Hub-Signature-256")

			if ctx.Method() == "OPTIONS" {
				return ctx.String(200, "OK")
			}

			return next(ctx)
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

	// Setup Flows
	dataExchangeHandler, _ := createDataExchangeHandler("", &logger)
	flowsAdapter := flows.NewServerAdapter(dataExchangeHandler, nil, nil, nil, &logger)
	flowsAdapter.RegisterRoutes(server.Router())

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
