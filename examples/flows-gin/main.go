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

	"github.com/wongpinter/go-whatsapp/flows"
	"github.com/wongpinter/go-whatsapp/internal/httpclient"
	"github.com/wongpinter/go-whatsapp/internal/httpserver"
)

func main() {
	// Setup logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	logger.Info().Msg("🚀 WhatsApp Flows with Gin Framework Example")

	// Get configuration from environment
	appSecret := getEnvOrDefault("WHATSAPP_APP_SECRET", "demo-app-secret")
	verifyToken := getEnvOrDefault("WHATSAPP_VERIFY_TOKEN", "demo-verify-token")
	port := getEnvOrDefault("PORT", "8080")

	// Create HTTP client manager
	clientManager := httpclient.NewManager(nil, &logger)

	// Create Gin engine with custom configuration
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	
	// Add Gin middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// Create Flows server factory
	flowsFactory := flows.NewServerFactory(clientManager, &logger)

	// Create Flows server with Gin
	server, flowsAdapter, err := flowsFactory.CreateFullFlowsServer(
		httpserver.FrameworkGin,
		appSecret,
		verifyToken,
		flows.WithFramework(httpserver.FrameworkGin),
		flows.WithServerOptions(httpserver.WithNativeEngine(r)),
		flows.WithRoutePrefix("/api/v1"),
		flows.WithRateLimit(true, 100), // 100 requests per minute
		flows.WithMetrics(true),
		flows.WithSecurity(true),
		flows.WithLogger(&logger),
	)
	if err != nil {
		log.Fatalf("Failed to create Flows server: %v", err)
	}

	// Add custom Gin routes
	router := server.Router()
	
	// Service info endpoint
	router.GET("/", func(ctx httpserver.HTTPContext) error {
		return ctx.JSON(200, gin.H{
			"service":     "WhatsApp Flows with Gin",
			"version":     "1.0.0",
			"framework":   "gin",
			"description": "Production-ready WhatsApp Flows server using Gin framework",
			"endpoints": gin.H{
				"flows": []string{
					"POST /api/v1/data-exchange - Flow data exchange",
					"GET  /api/v1/health - Health check",
					"GET  /api/v1/metrics - Performance metrics",
					"POST /api/v1/send/* - Send Flows",
				},
				"webhook": []string{
					"GET  /api/v1/webhook - Webhook verification",
					"POST /api/v1/webhook - Webhook events",
				},
			},
		})
	})

	// Custom Flow action handlers
	setupFlowActions(flowsAdapter)

	// Start server
	logger.Info().
		Str("port", port).
		Str("framework", "gin").
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

func setupFlowActions(adapter *flows.ServerAdapter) {
	// Get the action router from the adapter
	// In a real implementation, you would register actual action handlers
	
	// Example action handlers would be registered here:
	// router := adapter.GetActionRouter()
	// router.RegisterHandlerFunc("submit_survey", handleSurveySubmission)
	// router.RegisterHandlerFunc("book_appointment", handleAppointmentBooking)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Hub-Signature-256")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Example action handlers (commented out as they require full implementation)

/*
func handleSurveySubmission(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
	// Extract survey data
	name := request.Data["name"]
	email := request.Data["email"]
	satisfaction := request.Data["satisfaction"]
	
	// Process survey (save to database, send notifications, etc.)
	log.Printf("Survey submitted: name=%v, email=%v, satisfaction=%v", name, email, satisfaction)
	
	return &flows.DataExchangeResponse{
		Version: request.Version,
		Screen:  "SURVEY_COMPLETE",
		Data: map[string]interface{}{
			"submission_id": "SURVEY-" + generateID(),
			"thank_you_message": "Thank you for your feedback!",
		},
	}, nil
}

func handleAppointmentBooking(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
	// Extract appointment data
	date := request.Data["date"]
	time := request.Data["time"]
	service := request.Data["service"]
	
	// Book appointment (check availability, save to calendar, etc.)
	log.Printf("Appointment booked: date=%v, time=%v, service=%v", date, time, service)
	
	return &flows.DataExchangeResponse{
		Version: request.Version,
		Screen:  "BOOKING_CONFIRMATION",
		Data: map[string]interface{}{
			"appointment_id": "APPT-" + generateID(),
			"confirmation_code": generateConfirmationCode(),
			"appointment_details": map[string]interface{}{
				"date": date,
				"time": time,
				"service": service,
			},
		},
	}, nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

func generateConfirmationCode() string {
	return fmt.Sprintf("CONF%d", time.Now().Unix()%10000)
}
*/
