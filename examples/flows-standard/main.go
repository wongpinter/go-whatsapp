package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp/flows"
	"github.com/wongpinter/go-whatsapp/internal/httpclient"
	"github.com/wongpinter/go-whatsapp/internal/httpserver"
)

func main() {
	// Setup logger
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	logger.Info().Msg("🚀 WhatsApp Flows with Standard HTTP Example")

	// Get configuration from environment
	appSecret := getEnvOrDefault("WHATSAPP_APP_SECRET", "demo-app-secret")
	verifyToken := getEnvOrDefault("WHATSAPP_VERIFY_TOKEN", "demo-verify-token")
	port := getEnvOrDefault("PORT", "8080")

	// Create HTTP client manager
	clientManager := httpclient.NewManager(nil, &logger)

	// Create Flows server factory
	flowsFactory := flows.NewServerFactory(clientManager, &logger)

	// Create Flows server with standard HTTP
	server, flowsAdapter, err := flowsFactory.CreateFullFlowsServer(
		httpserver.FrameworkStandard,
		appSecret,
		verifyToken,
		flows.WithFramework(httpserver.FrameworkStandard),
		flows.WithRoutePrefix("/api/v1"),
		flows.WithRateLimit(true, 80), // 80 requests per minute
		flows.WithMetrics(true),
		flows.WithSecurity(true),
		flows.WithServerLogger(&logger),
	)
	if err != nil {
		log.Fatalf("Failed to create Flows server: %v", err)
	}

	// Add custom routes
	router := server.Router()

	// Service info endpoint
	router.GET("/", func(ctx httpserver.HTTPContext) error {
		response := map[string]interface{}{
			"service":     "WhatsApp Flows with Standard HTTP",
			"version":     "1.0.0",
			"framework":   "standard",
			"description": "Reliable WhatsApp Flows server using Go's standard HTTP library",
			"features": []string{
				"Zero external dependencies",
				"High reliability",
				"Standard HTTP compliance",
				"Production-ready",
			},
			"endpoints": map[string][]string{
				"flows": {
					"POST /api/v1/data-exchange - Flow data exchange",
					"GET  /api/v1/health - Health check",
					"GET  /api/v1/metrics - Performance metrics",
					"POST /api/v1/send/* - Send Flows",
					"POST /api/v1/validate-token - Token validation",
					"GET  /api/v1/actions - List actions",
				},
				"webhook": {
					"GET  /api/v1/webhook - Webhook verification",
					"POST /api/v1/webhook - Webhook events",
					"GET  /api/v1/health - Health check",
					"GET  /api/v1/metrics - Metrics",
				},
			},
			"documentation": map[string]string{
				"flows_api":   "https://developers.facebook.com/docs/whatsapp/flows",
				"webhook_api": "https://developers.facebook.com/docs/whatsapp/webhooks",
			},
		}

		return ctx.JSON(http.StatusOK, response)
	})

	// Status endpoint with detailed information
	router.GET("/status", func(ctx httpserver.HTTPContext) error {
		metrics := flowsAdapter.GetMetrics()

		status := map[string]interface{}{
			"server": map[string]interface{}{
				"status":    "running",
				"framework": "standard",
				"uptime":    time.Since(metrics.GetStartTime()).String(),
				"pid":       os.Getpid(),
			},
			"flows": map[string]interface{}{
				"total_requests":      metrics.GetTotalRequests(),
				"successful_flows":    metrics.GetSuccessfulFlows(),
				"failed_flows":        metrics.GetFailedFlows(),
				"active_flows":        metrics.GetActiveFlows(),
				"token_validations":   metrics.GetTokenValidations(),
				"encryption_requests": metrics.GetEncryptionRequests(),
			},
			"performance": map[string]interface{}{
				"average_latency_ms": metrics.GetAverageLatency().Milliseconds(),
				"last_request":       metrics.GetLastRequestTime().Format(time.RFC3339),
			},
		}

		return ctx.JSON(http.StatusOK, status)
	})

	// Configuration endpoint
	router.GET("/config", func(ctx httpserver.HTTPContext) error {
		config := map[string]interface{}{
			"server": map[string]interface{}{
				"framework":    "standard",
				"port":         port,
				"route_prefix": "/api/v1",
			},
			"features": map[string]bool{
				"webhook_integration": true,
				"rate_limiting":       true,
				"metrics_collection":  true,
				"security_middleware": true,
				"health_checks":       true,
			},
			"limits": map[string]interface{}{
				"rate_limit_rpm":   80,
				"max_request_size": "10MB",
				"request_timeout":  "30s",
				"shutdown_timeout": "30s",
			},
		}

		return ctx.JSON(http.StatusOK, config)
	})

	// Custom Flow action handlers
	setupStandardFlowActions(flowsAdapter, &logger)

	// Start server
	logger.Info().
		Str("port", port).
		Str("framework", "standard").
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

func setupStandardFlowActions(adapter *flows.ServerAdapter, logger *zerolog.Logger) {
	// Configure Flow action handlers for standard HTTP
	// This demonstrates the pattern for standard HTTP Flow handling

	logger.Info().Msg("Flow action handlers configured for standard HTTP")

	// Example of how you would register handlers:
	// router := adapter.GetActionRouter()
	// router.RegisterHandlerFunc("contact_form", handleContactForm)
	// router.RegisterHandlerFunc("newsletter_signup", handleNewsletterSignup)
	// router.RegisterHandlerFunc("support_ticket", handleSupportTicket)

	// Use adapter to avoid unused parameter warning
	_ = adapter
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Example action handlers for standard HTTP (commented out as they require full implementation)

/*
func handleContactForm(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
	// Extract contact form data
	name := request.Data["name"]
	email := request.Data["email"]
	message := request.Data["message"]
	subject := request.Data["subject"]

	// Process contact form (save to database, send email, etc.)
	ticketID := generateTicketID()

	log.Printf("Contact form submitted: ID=%s, name=%v, email=%v, subject=%v",
		ticketID, name, email, subject)

	return &flows.DataExchangeResponse{
		Version: request.Version,
		Screen:  "CONTACT_CONFIRMATION",
		Data: map[string]interface{}{
			"ticket_id": ticketID,
			"confirmation_message": "Thank you for contacting us! We'll get back to you soon.",
			"estimated_response_time": "24 hours",
		},
	}, nil
}

func handleNewsletterSignup(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
	// Extract newsletter signup data
	email := request.Data["email"]
	preferences := request.Data["preferences"]
	frequency := request.Data["frequency"]

	// Process newsletter signup (validate email, add to mailing list, etc.)
	subscriptionID := generateSubscriptionID()

	log.Printf("Newsletter signup: ID=%s, email=%v, preferences=%v",
		subscriptionID, email, preferences)

	return &flows.DataExchangeResponse{
		Version: request.Version,
		Screen:  "NEWSLETTER_CONFIRMATION",
		Data: map[string]interface{}{
			"subscription_id": subscriptionID,
			"welcome_message": "Welcome to our newsletter!",
			"first_newsletter": calculateNextNewsletterDate(frequency),
		},
	}, nil
}

func handleSupportTicket(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
	// Extract support ticket data
	category := request.Data["category"]
	priority := request.Data["priority"]
	description := request.Data["description"]
	attachments := request.Data["attachments"]

	// Process support ticket (create ticket, assign to team, etc.)
	ticketID := generateSupportTicketID()

	log.Printf("Support ticket created: ID=%s, category=%v, priority=%v",
		ticketID, category, priority)

	return &flows.DataExchangeResponse{
		Version: request.Version,
		Screen:  "TICKET_CREATED",
		Data: map[string]interface{}{
			"ticket_id": ticketID,
			"status": "open",
			"assigned_team": getAssignedTeam(category),
			"estimated_resolution": calculateResolutionTime(priority),
		},
	}, nil
}

func generateTicketID() string {
	return fmt.Sprintf("TICKET-%d", time.Now().Unix())
}

func generateSubscriptionID() string {
	return fmt.Sprintf("SUB-%d", time.Now().Unix())
}

func generateSupportTicketID() string {
	return fmt.Sprintf("SUPPORT-%d", time.Now().Unix())
}

func calculateNextNewsletterDate(frequency interface{}) string {
	// Simplified calculation
	return time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02")
}

func getAssignedTeam(category interface{}) string {
	// Simplified team assignment
	switch category {
	case "technical":
		return "Technical Support"
	case "billing":
		return "Billing Team"
	default:
		return "General Support"
	}
}

func calculateResolutionTime(priority interface{}) string {
	// Simplified resolution time calculation
	switch priority {
	case "high":
		return "4 hours"
	case "medium":
		return "24 hours"
	default:
		return "72 hours"
	}
}
*/
