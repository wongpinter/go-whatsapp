package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp"
	"github.com/wongpinter/go-whatsapp/bm"
	"github.com/wongpinter/go-whatsapp/cloudapi"
	"github.com/wongpinter/go-whatsapp/webhook"
)

// Example implementation of webhook handlers
type MyWebhookHandlers struct {
	logger          zerolog.Logger
	statusMonitor   *webhook.StatusMonitor
	messageTracker  *webhook.MessageLifecycleTracker
	errorClassifier *webhook.ErrorClassifier
	rateLimiter     *webhook.RateLimiter
	messageQueue    *webhook.MessageQueueManager
}

// OnTextMessage handles incoming text messages
func (h *MyWebhookHandlers) OnTextMessage(ctx context.Context, msg *webhook.Message, metadata *webhook.Metadata) error {
	h.logger.Info().
		Str("from", msg.From).
		Str("text", msg.Text.Body).
		Str("phone_number_id", metadata.PhoneNumberID).
		Msg("Received text message")

	// Echo the message back (in a real app, you'd implement your business logic here)
	fmt.Printf("Received text from %s: %s\n", msg.From, msg.Text.Body)
	return nil
}

// OnStatusUpdate handles message status updates
func (h *MyWebhookHandlers) OnStatusUpdate(ctx context.Context, status *webhook.Status, metadata *webhook.Metadata) error {
	h.logger.Info().
		Str("message_id", status.ID).
		Str("status", status.Status).
		Str("recipient_id", status.RecipientID).
		Str("phone_number_id", metadata.PhoneNumberID).
		Msg("Received status update")

	fmt.Printf("Message %s status: %s\n", status.ID, status.Status)

	// Show pricing information if available
	if status.Pricing != nil {
		fmt.Printf("💰 Pricing: %s category, billable: %t\n",
			status.Pricing.Category, status.Pricing.Billable)
	}

	// Show conversation information if available
	if status.Conversation != nil {
		fmt.Printf("💬 Conversation: %s", status.Conversation.ID)
		if status.Conversation.Origin != nil {
			fmt.Printf(" (origin: %s)", status.Conversation.Origin.Type)
		}
		fmt.Println()
	}

	// Handle any errors in the status
	if len(status.Errors) > 0 {
		for _, err := range status.Errors {
			classification := h.errorClassifier.ClassifyError(err.Code)
			fmt.Printf("❌ Error %d: %s\n", err.Code, err.Message)
			fmt.Printf("   Category: %s, Severity: %s, Retryable: %t\n",
				classification.Category, classification.Severity, classification.Retryable)
			fmt.Printf("   Action: %s\n", classification.Action)

			// Handle rate limit errors specifically
			if err.Code == 130429 || err.Code == 80007 || err.Code == 131048 {
				fmt.Printf("🚦 Rate limit detected! Checking current limits...\n")

				// Show current rate limit status
				allLimits := h.rateLimiter.GetAllRateLimits()
				for limitType, info := range allLimits {
					fmt.Printf("   %s: %d/%d (resets at %s)\n",
						limitType, info.CurrentUsage, info.Limit, info.ResetTime.Format("15:04:05"))
				}

				// Show quality recommendations if quality is at risk
				if h.rateLimiter.IsQualityAtRisk() {
					fmt.Printf("⚠️  Quality rating is at risk! Recommendations:\n")
					recommendations := h.rateLimiter.GetQualityRecommendations()
					for _, rec := range recommendations {
						fmt.Printf("   • %s\n", rec)
					}
				}
			}
		}
	}

	// Get detailed message lifecycle information
	if lifecycle, exists := h.messageTracker.GetMessageLifecycle(status.ID); exists {
		fmt.Printf("📋 Message Lifecycle:\n")
		fmt.Printf("   Type: %s, Sent: %s\n", lifecycle.MessageType, lifecycle.SentAt.Format("15:04:05"))
		if lifecycle.DeliveredAt != nil {
			deliveryTime := lifecycle.DeliveredAt.Sub(lifecycle.SentAt)
			fmt.Printf("   Delivered: %s (took %v)\n", lifecycle.DeliveredAt.Format("15:04:05"), deliveryTime)
		}
		if lifecycle.ReadAt != nil && lifecycle.DeliveredAt != nil {
			readTime := lifecycle.ReadAt.Sub(*lifecycle.DeliveredAt)
			fmt.Printf("   Read: %s (took %v after delivery)\n", lifecycle.ReadAt.Format("15:04:05"), readTime)
		}
		fmt.Printf("   Status History: %d events\n", len(lifecycle.StatusHistory))
	}

	// Print delivery statistics when a message is delivered or read
	if status.Status == "delivered" || status.Status == "read" {
		stats := h.statusMonitor.GetMessageStats()
		fmt.Printf("📊 Delivery Stats: %.1f%% delivered, %.1f%% read, %.1f%% failed\n",
			stats.DeliveryRate, stats.ReadRate, stats.FailureRate)
	}

	return nil
}

func main() {
	// Set up logging
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Get configuration from environment variables
	accessToken := os.Getenv("WHATSAPP_ACCESS_TOKEN")
	phoneNumberID := os.Getenv("WHATSAPP_PHONE_NUMBER_ID")
	wabaID := os.Getenv("WHATSAPP_WABA_ID")
	webhookSecret := os.Getenv("WHATSAPP_WEBHOOK_SECRET")
	verifyToken := os.Getenv("WHATSAPP_VERIFY_TOKEN")

	if accessToken == "" || phoneNumberID == "" {
		log.Fatal("WHATSAPP_ACCESS_TOKEN and WHATSAPP_PHONE_NUMBER_ID environment variables are required")
	}

	ctx := context.Background()

	// Example 1: Sending Messages with CloudAPI
	fmt.Println("=== CloudAPI Example ===")

	cloudClient := cloudapi.NewClient(phoneNumberID, accessToken,
		cloudapi.WithLogger(logger))

	// Health check
	if err := cloudClient.Health(ctx); err != nil {
		logger.Error().Err(err).Msg("CloudAPI health check failed")
	} else {
		logger.Info().Msg("CloudAPI health check passed")
	}

	// Send a text message (replace with a real phone number for testing)
	recipientPhone := "+1234567890" // Replace with actual phone number
	if recipientPhone != "+1234567890" {
		// Example 1: Send a simple text message
		resp, err := cloudClient.SendText(ctx, recipientPhone, "Hello from Go WhatsApp SDK!")
		if err != nil {
			if apiErr, ok := err.(*whatsapp.APIError); ok {
				logger.Error().
					Int("code", apiErr.Code()).
					Str("message", apiErr.Message()).
					Str("type", apiErr.Type()).
					Msg("Failed to send message")
			} else {
				logger.Error().Err(err).Msg("Failed to send message")
			}
		} else {
			logger.Info().
				Str("message_id", resp.GetMessageID()).
				Str("wa_id", resp.GetWAID()).
				Msg("Text message sent successfully")
		}

		// Example 2: Send a template message
		resp, err = cloudClient.SendTemplateWithParams(ctx, recipientPhone, "hello_world", "en_US", "John")
		if err != nil {
			logger.Error().Err(err).Msg("Failed to send template message")
		} else {
			logger.Info().
				Str("message_id", resp.GetMessageID()).
				Msg("Template message sent successfully")
		}

		// Example 3: Send an interactive button message
		buttons := map[string]string{
			"yes": "Yes, I'm interested",
			"no":  "No, thanks",
		}
		resp, err = cloudClient.SendInteractiveButtons(ctx, recipientPhone, "Are you interested in our new product?", buttons)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to send interactive message")
		} else {
			logger.Info().
				Str("message_id", resp.GetMessageID()).
				Msg("Interactive message sent successfully")
		}

		// Example 4: Send a location message
		resp, err = cloudClient.SendLocation(ctx, recipientPhone, 37.7749, -122.4194, "San Francisco", "San Francisco, CA, USA")
		if err != nil {
			logger.Error().Err(err).Msg("Failed to send location message")
		} else {
			logger.Info().
				Str("message_id", resp.GetMessageID()).
				Msg("Location message sent successfully")
		}

		// Example 5: Send an image message
		resp, err = cloudClient.SendImageFromURL(ctx, recipientPhone, "https://via.placeholder.com/300x200", "Sample image")
		if err != nil {
			logger.Error().Err(err).Msg("Failed to send image message")
		} else {
			logger.Info().
				Str("message_id", resp.GetMessageID()).
				Msg("Image message sent successfully")

		}
	}

	// Example 2: Business Management API
	if wabaID != "" {
		fmt.Println("\n=== Business Management API Example ===")

		bmClient := bm.NewClient(accessToken,
			bm.WithWABAID(wabaID),
			bm.WithLogger(logger))

		// Get business account details
		account, err := bmClient.GetBusinessAccount(ctx, wabaID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get business account")
		} else {
			logger.Info().
				Str("account_name", account.Name).
				Str("timezone", account.TimezoneID).
				Msg("Business account retrieved")

			fmt.Printf("Business Account: %s (ID: %s)\n", account.Name, account.ID)
		}

		// Get phone numbers
		phoneNumbers, err := bmClient.GetPhoneNumbers(ctx, wabaID)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get phone numbers")
		} else {
			logger.Info().Int("count", len(phoneNumbers)).Msg("Phone numbers retrieved")

			for _, phone := range phoneNumbers {
				fmt.Printf("Phone: %s, Status: %s, Quality: %s\n",
					phone.DisplayPhoneNumber, phone.Status, phone.QualityRating)
			}
		}

		// Get business profile
		if len(phoneNumbers) > 0 {
			profile, err := bmClient.GetBusinessProfile(ctx, phoneNumbers[0].ID)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to get business profile")
			} else {
				logger.Info().
					Str("about", profile.About).
					Str("description", profile.Description).
					Msg("Business profile retrieved")

				fmt.Printf("Business Profile: %s\n", profile.About)
			}
		}
	}

	// Example 3: Webhook Handler
	if webhookSecret != "" && verifyToken != "" {
		fmt.Println("\n=== Webhook Handler Example ===")

		// Create webhook handler
		webhookHandler := webhook.NewHandler(webhookSecret, verifyToken, logger)

		// Create and register our custom handlers
		myHandlers := &MyWebhookHandlers{
			logger:          logger,
			statusMonitor:   webhookHandler.GetStatusMonitor(),
			messageTracker:  webhookHandler.GetMessageTracker(),
			errorClassifier: webhookHandler.GetErrorClassifier(),
			rateLimiter:     webhookHandler.GetRateLimiter(),
			messageQueue:    webhookHandler.GetMessageQueue(),
		}
		dispatcher := webhookHandler.GetDispatcher()
		dispatcher.RegisterTextMessageHandler(myHandlers)
		dispatcher.RegisterStatusUpdateHandler(myHandlers)

		// Example: Demonstrate rate limit management
		fmt.Println("\n📊 Rate Limiting & Quality Management Demo:")

		// Simulate updating rate limit information (normally from API responses)
		webhookHandler.UpdateRateLimitInfo(webhook.RateLimitTypeAPI, 150, 200, time.Now().Add(time.Hour))
		webhookHandler.UpdateRateLimitInfo(webhook.RateLimitTypeThroughput, 45, 80, time.Now().Add(time.Minute))

		// Simulate quality and tier information
		webhookHandler.UpdateQualityInfo(webhook.Tier1, webhook.QualityHigh, 80)

		// Show current rate limit status
		rateLimiter := webhookHandler.GetRateLimiter()
		allLimits := rateLimiter.GetAllRateLimits()
		for limitType, info := range allLimits {
			fmt.Printf("🚦 %s: %d/%d", limitType, info.CurrentUsage, info.Limit)
			if info.Limit > 0 {
				usage := float64(info.CurrentUsage) / float64(info.Limit) * 100
				fmt.Printf(" (%.1f%% used)", usage)
			}
			fmt.Printf(" - resets at %s\n", info.ResetTime.Format("15:04:05"))
		}

		// Check throughput upgrade eligibility
		if eligible, requirements := rateLimiter.GetThroughputUpgradeEligibility(); eligible {
			fmt.Println("✅ Eligible for throughput upgrade!")
		} else {
			fmt.Println("📋 Throughput upgrade requirements:")
			for _, req := range requirements {
				fmt.Printf("   %s\n", req)
			}
		}

		// Set up HTTP server
		mux := http.NewServeMux()
		mux.Handle("/webhook", webhookHandler)

		// Add a simple health check endpoint
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		// Add a metrics endpoint
		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			// Get webhook metrics
			metrics := webhookHandler.GetMetrics().GetMetrics()

			// Get status monitor stats
			statusMonitor := webhookHandler.GetStatusMonitor()
			messageStats := statusMonitor.GetMessageStats()
			failedMessages := statusMonitor.GetFailedMessages()

			// Get message tracker stats
			messageTracker := webhookHandler.GetMessageTracker()
			failedLifecycles := messageTracker.GetFailedMessages()
			undeliveredMessages := messageTracker.GetUndeliveredMessages(5 * time.Minute)

			// Get error classifier stats
			errorClassifier := webhookHandler.GetErrorClassifier()
			retryableErrors := errorClassifier.GetRetryableErrors()

			// Get rate limiting stats
			rateLimiter := webhookHandler.GetRateLimiter()
			allRateLimits := rateLimiter.GetAllRateLimits()
			qualityRecommendations := rateLimiter.GetQualityRecommendations()
			isQualityAtRisk := rateLimiter.IsQualityAtRisk()
			eligible, upgradeRequirements := rateLimiter.GetThroughputUpgradeEligibility()

			// Get message queue stats
			messageQueue := webhookHandler.GetMessageQueue()
			queueStats := messageQueue.GetQueueStats()

			response := map[string]interface{}{
				"webhook_metrics": metrics,
				"message_stats": map[string]interface{}{
					"total_messages":        messageStats.TotalMessages,
					"delivered_messages":    messageStats.DeliveredMessages,
					"read_messages":         messageStats.ReadMessages,
					"failed_messages":       messageStats.FailedMessages,
					"delivery_rate":         messageStats.DeliveryRate,
					"read_rate":             messageStats.ReadRate,
					"failure_rate":          messageStats.FailureRate,
					"average_delivery_time": messageStats.AverageDeliveryTime.String(),
					"average_read_time":     messageStats.AverageReadTime.String(),
				},
				"message_lifecycle": map[string]interface{}{
					"failed_messages_count":      len(failedLifecycles),
					"undelivered_messages_count": len(undeliveredMessages),
				},
				"error_analysis": map[string]interface{}{
					"retryable_error_types": len(retryableErrors),
					"failed_message_count":  len(failedMessages),
				},
				"rate_limiting": map[string]interface{}{
					"current_limits":              allRateLimits,
					"quality_at_risk":             isQualityAtRisk,
					"quality_recommendations":     qualityRecommendations,
					"throughput_upgrade_eligible": eligible,
					"upgrade_requirements":        upgradeRequirements,
				},
				"message_queue": queueStats,
			}

			json.NewEncoder(w).Encode(response)
		})

		logger.Info().Msg("Starting webhook server on :8080")
		fmt.Println("Webhook server starting on :8080")
		fmt.Println("Webhook endpoint: http://localhost:8080/webhook")
		fmt.Println("Health check: http://localhost:8080/health")
		fmt.Println("Metrics endpoint: http://localhost:8080/metrics")

		// Start server (this will block)
		if err := http.ListenAndServe(":8080", mux); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start webhook server")
		}
	} else {
		fmt.Println("\nTo test webhooks, set WHATSAPP_WEBHOOK_SECRET and WHATSAPP_VERIFY_TOKEN environment variables")
	}

	fmt.Println("\nExample completed successfully!")
}
