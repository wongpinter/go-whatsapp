//go:build echo
// +build echo

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp/flows"
	"github.com/wongpinter/go-whatsapp/internal/httpclient"
	"github.com/wongpinter/go-whatsapp/internal/httpserver"
)

func main() {
	// Setup logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	logger.Info().Msg("🚀 WhatsApp Flows with Echo Framework Example")

	// Get configuration from environment
	appSecret := getEnvOrDefault("WHATSAPP_APP_SECRET", "demo-app-secret")
	verifyToken := getEnvOrDefault("WHATSAPP_VERIFY_TOKEN", "demo-verify-token")
	port := getEnvOrDefault("PORT", "8080")

	// Create HTTP client manager
	clientManager := httpclient.NewManager(nil, &logger)

	// Create Echo instance with custom configuration
	e := echo.New()
	
	// Configure Echo
	e.HideBanner = true
	e.HidePort = true
	
	// Add Echo middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// Create Flows server factory
	flowsFactory := flows.NewServerFactory(clientManager, &logger)

	// Create Flows server with Echo
	server, flowsAdapter, err := flowsFactory.CreateFullFlowsServer(
		httpserver.FrameworkEcho,
		appSecret,
		verifyToken,
		flows.WithFramework(httpserver.FrameworkEcho),
		flows.WithServerOptions(httpserver.WithNativeEngine(e)),
		flows.WithRoutePrefix("/api/v1"),
		flows.WithRateLimit(true, 120), // 120 requests per minute
		flows.WithMetrics(true),
		flows.WithSecurity(true),
		flows.WithLogger(&logger),
	)
	if err != nil {
		log.Fatalf("Failed to create Flows server: %v", err)
	}

	// Add custom Echo routes
	router := server.Router()
	
	// Service info endpoint
	router.GET("/", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"service":     "WhatsApp Flows with Echo",
			"version":     "1.0.0",
			"framework":   "echo",
			"description": "High-performance WhatsApp Flows server using Echo framework",
			"features": []string{
				"Real-time Flow data exchange",
				"Webhook event processing",
				"Health monitoring",
				"Performance metrics",
				"Rate limiting",
				"Security middleware",
			},
			"endpoints": map[string][]string{
				"flows": {
					"POST /api/v1/data-exchange - Flow data exchange",
					"GET  /api/v1/health - Health check",
					"GET  /api/v1/metrics - Performance metrics",
					"POST /api/v1/send/survey - Send survey Flow",
					"POST /api/v1/send/lead - Send lead generation Flow",
					"POST /api/v1/send/custom - Send custom Flow",
				},
				"webhook": {
					"GET  /api/v1/webhook - Webhook verification",
					"POST /api/v1/webhook - Webhook events",
					"GET  /api/v1/health - Health check",
				},
			},
		})
	})

	// Development endpoints
	router.GET("/debug/metrics", func(ctx httpserver.HTTPContext) error {
		metrics := flowsAdapter.GetMetrics()
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"debug": true,
			"metrics": map[string]interface{}{
				"total_requests":      metrics.totalRequests,
				"successful_flows":    metrics.successfulFlows,
				"failed_flows":        metrics.failedFlows,
				"active_flows":        metrics.activeFlows,
				"token_validations":   metrics.tokenValidations,
				"encryption_requests": metrics.encryptionRequests,
				"uptime_seconds":      time.Since(metrics.startTime).Seconds(),
			},
		})
	})

	// Custom Flow action handlers
	setupAdvancedFlowActions(flowsAdapter, &logger)

	// Start server
	logger.Info().
		Str("port", port).
		Str("framework", "echo").
		Msg("Starting Flows server")

	go func() {
		if err := server.Start(":" + port); err != nil {
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

func setupAdvancedFlowActions(adapter *flows.ServerAdapter, logger *zerolog.Logger) {
	// Get the data exchange handler from the adapter
	handler := adapter.GetDataExchangeHandler()
	
	// In a real implementation, you would register actual action handlers
	// This demonstrates the pattern for Echo-specific Flow handling
	
	logger.Info().Msg("Flow action handlers configured for Echo framework")
	
	// Example of how you would register handlers:
	// router := adapter.GetActionRouter()
	// router.RegisterHandlerFunc("process_order", handleOrderProcessing)
	// router.RegisterHandlerFunc("schedule_delivery", handleDeliveryScheduling)
	// router.RegisterHandlerFunc("customer_feedback", handleCustomerFeedback)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Example advanced action handlers for Echo (commented out as they require full implementation)

/*
func handleOrderProcessing(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
	// Extract order data
	items := request.Data["items"]
	customerInfo := request.Data["customer"]
	paymentMethod := request.Data["payment_method"]
	
	// Process order (validate items, calculate total, process payment, etc.)
	orderID := generateOrderID()
	
	log.Printf("Order processed: ID=%s, items=%v, customer=%v", orderID, items, customerInfo)
	
	return &flows.DataExchangeResponse{
		Version: request.Version,
		Screen:  "ORDER_CONFIRMATION",
		Data: map[string]interface{}{
			"order_id": orderID,
			"total_amount": calculateTotal(items),
			"estimated_delivery": calculateDeliveryTime(),
			"tracking_url": fmt.Sprintf("https://example.com/track/%s", orderID),
		},
	}, nil
}

func handleDeliveryScheduling(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
	// Extract delivery preferences
	address := request.Data["address"]
	preferredDate := request.Data["preferred_date"]
	timeSlot := request.Data["time_slot"]
	
	// Schedule delivery (check availability, assign driver, etc.)
	deliveryID := generateDeliveryID()
	
	log.Printf("Delivery scheduled: ID=%s, address=%v, date=%v, time=%v", 
		deliveryID, address, preferredDate, timeSlot)
	
	return &flows.DataExchangeResponse{
		Version: request.Version,
		Screen:  "DELIVERY_SCHEDULED",
		Data: map[string]interface{}{
			"delivery_id": deliveryID,
			"scheduled_date": preferredDate,
			"time_slot": timeSlot,
			"driver_contact": "+1234567890",
		},
	}, nil
}

func handleCustomerFeedback(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
	// Extract feedback data
	rating := request.Data["rating"]
	comments := request.Data["comments"]
	category := request.Data["category"]
	
	// Process feedback (save to database, trigger notifications, etc.)
	feedbackID := generateFeedbackID()
	
	log.Printf("Feedback received: ID=%s, rating=%v, category=%v", feedbackID, rating, category)
	
	return &flows.DataExchangeResponse{
		Version: request.Version,
		Screen:  "FEEDBACK_THANK_YOU",
		Data: map[string]interface{}{
			"feedback_id": feedbackID,
			"thank_you_message": "Thank you for your valuable feedback!",
			"follow_up": rating.(float64) < 3.0, // Follow up for low ratings
		},
	}, nil
}

func generateOrderID() string {
	return fmt.Sprintf("ORD-%d", time.Now().Unix())
}

func generateDeliveryID() string {
	return fmt.Sprintf("DEL-%d", time.Now().Unix())
}

func generateFeedbackID() string {
	return fmt.Sprintf("FB-%d", time.Now().Unix())
}

func calculateTotal(items interface{}) float64 {
	// Simplified calculation
	return 99.99
}

func calculateDeliveryTime() string {
	// Simplified calculation
	return time.Now().Add(24 * time.Hour).Format("2006-01-02")
}
*/
