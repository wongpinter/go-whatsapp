package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp/cloudapi"
	"github.com/wongpinter/go-whatsapp/flows"
	"github.com/wongpinter/go-whatsapp/webhook"
)

func main() {
	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Get configuration from environment
	accessToken := os.Getenv("WHATSAPP_ACCESS_TOKEN")
	wabaID := os.Getenv("WHATSAPP_WABA_ID")
	phoneNumberID := os.Getenv("WHATSAPP_PHONE_NUMBER_ID")
	appSecret := os.Getenv("WHATSAPP_APP_SECRET")
	verifyToken := os.Getenv("WHATSAPP_VERIFY_TOKEN")

	if accessToken == "" || wabaID == "" || phoneNumberID == "" {
		log.Fatal("Missing required environment variables")
	}

	// Initialize Flow management client
	flowClient := flows.NewClient(wabaID, accessToken, flows.WithLogger(&logger))

	// Initialize CloudAPI client
	cloudClient := cloudapi.NewClient(phoneNumberID, accessToken, cloudapi.WithLogger(logger))

	// Create and upload a survey Flow
	surveyFlow := flows.ExampleSurveyFlow()
	flowID, err := createAndUploadFlow(context.Background(), flowClient, "Customer Survey", []string{flows.CategorySurvey}, surveyFlow)
	if err != nil {
		log.Fatalf("Failed to create survey Flow: %v", err)
	}
	logger.Info().Str("flow_id", flowID).Msg("Survey Flow created successfully")

	// Create and upload a lead generation Flow
	leadFlow := flows.ExampleLeadGenerationFlow()
	leadFlowID, err := createAndUploadFlow(context.Background(), flowClient, "Lead Generation", []string{flows.CategoryLeadGeneration}, leadFlow)
	if err != nil {
		log.Fatalf("Failed to create lead Flow: %v", err)
	}
	logger.Info().Str("flow_id", leadFlowID).Msg("Lead generation Flow created successfully")

	// Set up data exchange endpoint
	dataExchangeHandler := setupDataExchangeHandler(&logger)

	// Set up webhook handler with Flow support
	webhookHandler := setupWebhookHandler(appSecret, verifyToken, &logger)

	// Set up HTTP routes
	http.Handle("/webhook", webhookHandler)
	http.Handle("/flows/data-exchange", dataExchangeHandler)
	http.HandleFunc("/flows/send-survey", func(w http.ResponseWriter, r *http.Request) {
		sendSurveyFlow(w, r, cloudClient, flowID)
	})
	http.HandleFunc("/flows/send-lead", func(w http.ResponseWriter, r *http.Request) {
		sendLeadFlow(w, r, cloudClient, leadFlowID)
	})

	// Start server
	logger.Info().Msg("Starting Flow example server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// createAndUploadFlow creates a Flow and uploads its JSON definition.
func createAndUploadFlow(ctx context.Context, client *flows.Client, name string, categories []string, flow *flows.Flow) (string, error) {
	// Create Flow metadata
	metadata := &flows.CreateFlowRequest{
		Name:       name,
		Categories: categories,
	}

	// Create and upload Flow
	workflow := flows.NewFlowWorkflow(client)
	return workflow.CreateCompleteFlow(ctx, metadata, flow, false) // Don't auto-publish
}

// setupDataExchangeHandler sets up the data exchange endpoint handler.
func setupDataExchangeHandler(logger *zerolog.Logger) *flows.DataExchangeHandler {
	// Create action router
	router := flows.NewActionRouter()

	// Register action handlers
	registerSurveyHandlers(router)
	registerLeadHandlers(router)
	registerAppointmentHandlers(router)

	// Create data exchange handler
	handler := flows.NewDataExchangeHandler(
		flows.WithActionRouter(router),
		flows.WithDataExchangeLogger(logger),
	)

	return handler
}

// setupWebhookHandler sets up the webhook handler with Flow support.
func setupWebhookHandler(appSecret, verifyToken string, logger *zerolog.Logger) *webhook.Handler {
	// Create webhook handler
	handler := webhook.NewHandler(appSecret, verifyToken, *logger)

	// Register Flow reply handler
	dispatcher := handler.GetDispatcher()
	dispatcher.RegisterFlowReplyHandler(&FlowReplyHandler{logger: logger})

	return handler
}

// FlowReplyHandler handles Flow completion events from webhooks.
type FlowReplyHandler struct {
	logger *zerolog.Logger
}

// OnFlowReply handles Flow reply events.
func (h *FlowReplyHandler) OnFlowReply(ctx context.Context, msg *webhook.Message, metadata *webhook.Metadata) error {
	if msg.Interactive == nil || msg.Interactive.FlowReply == nil {
		return fmt.Errorf("invalid Flow reply message")
	}

	flowReply := msg.Interactive.FlowReply
	h.logger.Info().
		Str("flow_token", flowReply.FlowToken).
		Interface("response", flowReply.Response).
		Str("from", msg.From).
		Msg("Flow completed")

	// Process Flow completion
	completion := &flows.FlowCompletion{
		FlowToken: flowReply.FlowToken,
		Response:  flowReply.Response,
	}

	// Here you would typically:
	// 1. Validate the Flow token
	// 2. Process the completion based on Flow type
	// 3. Store results in database
	// 4. Trigger follow-up actions

	h.logger.Info().
		Str("flow_token", completion.FlowToken).
		Interface("response", completion.Response).
		Msg("Flow completion processed")
	return nil
}

// sendSurveyFlow sends a survey Flow to a user.
func sendSurveyFlow(w http.ResponseWriter, r *http.Request, client *cloudapi.Client, flowID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		To string `json:"to"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate Flow token
	flowToken, err := flows.GenerateFlowToken(flowID, req.To, map[string]interface{}{
		"survey_type": "customer_satisfaction",
	})
	if err != nil {
		http.Error(w, "Failed to generate Flow token", http.StatusInternalServerError)
		return
	}

	// Send Flow message
	_, err = client.SendFlow(
		context.Background(),
		req.To,
		"Help us improve our service by taking a quick survey.",
		flowID,
		flowToken,
		"Take Survey",
	)

	if err != nil {
		http.Error(w, "Failed to send Flow message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"flow_id":    flowID,
		"flow_token": flowToken,
	})
}

// sendLeadFlow sends a lead generation Flow to a user.
func sendLeadFlow(w http.ResponseWriter, r *http.Request, client *cloudapi.Client, flowID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		To string `json:"to"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate Flow token
	flowToken, err := flows.GenerateFlowToken(flowID, req.To, map[string]interface{}{
		"source": "website",
	})
	if err != nil {
		http.Error(w, "Failed to generate Flow token", http.StatusInternalServerError)
		return
	}

	// Send Flow message
	_, err = client.SendFlow(
		context.Background(),
		req.To,
		"Get your free personalized quote in just 2 minutes!",
		flowID,
		flowToken,
		"Get Quote",
	)

	if err != nil {
		http.Error(w, "Failed to send Flow message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"flow_id":    flowID,
		"flow_token": flowToken,
	})
}

// registerSurveyHandlers registers action handlers for survey Flow.
func registerSurveyHandlers(router *flows.ActionRouter) {
	router.RegisterHandlerFunc("start_survey", func(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
		return &flows.DataExchangeResponse{
			Version: request.Version,
			Screen:  "SURVEY_QUESTIONS",
			Data: map[string]interface{}{
				"user_name": "Valued Customer",
			},
		}, nil
	})

	router.RegisterHandlerFunc("submit_survey", func(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
		// Process survey data
		log.Printf("Survey submitted: %+v", request.Data)

		return &flows.DataExchangeResponse{
			Version: request.Version,
			Screen:  "SURVEY_END",
			Data: map[string]interface{}{
				"submission_id": "SURVEY-12345",
			},
		}, nil
	})
}

// registerLeadHandlers registers action handlers for lead generation Flow.
func registerLeadHandlers(router *flows.ActionRouter) {
	router.RegisterHandlerFunc("submit_lead", func(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
		// Process lead data
		log.Printf("Lead submitted: %+v", request.Data)

		return &flows.DataExchangeResponse{
			Version: request.Version,
			Screen:  "LEAD_CONFIRMATION",
			Data: map[string]interface{}{
				"lead_id": "LEAD-67890",
			},
		}, nil
	})
}

// registerAppointmentHandlers registers action handlers for appointment booking Flow.
func registerAppointmentHandlers(router *flows.ActionRouter) {
	router.RegisterHandlerFunc("load_available_dates", func(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
		// Load available time slots
		timeSlots := []flows.DataSourceItem{
			{ID: "09:00", Title: "9:00 AM"},
			{ID: "11:00", Title: "11:00 AM"},
			{ID: "14:00", Title: "2:00 PM"},
			{ID: "16:00", Title: "4:00 PM"},
		}

		return &flows.DataExchangeResponse{
			Version: request.Version,
			Screen:  "SELECT_DATE",
			Data: map[string]interface{}{
				"available_slots": timeSlots,
			},
		}, nil
	})

	router.RegisterHandlerFunc("book_appointment", func(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
		// Process appointment booking
		log.Printf("Appointment booked: %+v", request.Data)

		return &flows.DataExchangeResponse{
			Version: request.Version,
			Screen:  "BOOKING_CONFIRMATION",
			Data: map[string]interface{}{
				"appointment_id":    "APPT-54321",
				"confirmation_code": "ABC123",
			},
		}, nil
	})
}
