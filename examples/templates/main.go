package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/wongpinter/go-whatsapp/bm"
	"github.com/wongpinter/go-whatsapp/cloudapi"
)

func main() {
	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Get configuration from environment
	wabaID := os.Getenv("WABA_ID")
	phoneNumberID := os.Getenv("PHONE_NUMBER_ID")
	accessToken := os.Getenv("ACCESS_TOKEN")
	userPhoneNumber := os.Getenv("USER_PHONE_NUMBER")

	if wabaID == "" || phoneNumberID == "" || accessToken == "" || userPhoneNumber == "" {
		log.Fatal("Please set WABA_ID, PHONE_NUMBER_ID, ACCESS_TOKEN, and USER_PHONE_NUMBER environment variables")
	}

	// Initialize clients
	bmClient := bm.NewClient(accessToken, bm.WithWABAID(wabaID), bm.WithLogger(logger))
	cloudClient := cloudapi.NewClient(phoneNumberID, accessToken, cloudapi.WithLogger(logger))

	ctx := context.Background()

	// Example 1: Create a simple marketing template
	fmt.Println("=== Creating Marketing Template ===")
	if err := createMarketingTemplate(ctx, bmClient); err != nil {
		logger.Error().Err(err).Msg("Failed to create marketing template")
	}

	// Example 2: Create an order confirmation template
	fmt.Println("\n=== Creating Order Confirmation Template ===")
	if err := createOrderConfirmationTemplate(ctx, bmClient); err != nil {
		logger.Error().Err(err).Msg("Failed to create order confirmation template")
	}

	// Example 3: Create an appointment reminder template
	fmt.Println("\n=== Creating Appointment Reminder Template ===")
	if err := createAppointmentReminderTemplate(ctx, bmClient); err != nil {
		logger.Error().Err(err).Msg("Failed to create appointment reminder template")
	}

	// Example 4: List all templates
	fmt.Println("\n=== Listing All Templates ===")
	if err := listTemplates(ctx, bmClient); err != nil {
		logger.Error().Err(err).Msg("Failed to list templates")
	}

	// Example 5: Send template messages
	fmt.Println("\n=== Sending Template Messages ===")
	if err := sendTemplateMessages(ctx, cloudClient, userPhoneNumber); err != nil {
		logger.Error().Err(err).Msg("Failed to send template messages")
	}

	// Example 6: Template management workflow
	fmt.Println("\n=== Template Management Workflow ===")
	if err := templateManagementWorkflow(ctx, bmClient); err != nil {
		logger.Error().Err(err).Msg("Failed to execute template management workflow")
	}
}

// createMarketingTemplate demonstrates creating a marketing template with buttons.
func createMarketingTemplate(ctx context.Context, client *bm.Client) error {
	// Build template using fluent API
	template := bm.NewMarketingTemplate(
		"summer_sale_2024",
		"en_US",
		"🌞 Summer Sale Alert!",
		"Hi {{1}}, enjoy {{2}} off on all summer items! Valid until {{3}}. Shop now and save big!",
		"Terms and conditions apply",
	).WithCategoryChange(true)

	// Add buttons
	buttons := bm.NewButtons().
		AddURL("Shop Now", "https://example.com/summer-sale").
		AddQuickReply("Not Interested").
		Build()

	template.AddButtons(buttons...)

	// Add examples for placeholders
	template = bm.NewTemplate("summer_sale_2024", "en_US", bm.CategoryMarketing).
		WithCategoryChange(true).
		AddHeaderWithExample(bm.FormatText, "🌞 Summer Sale Alert!", nil).
		AddBodyWithExample(
			"Hi {{1}}, enjoy {{2}} off on all summer items! Valid until {{3}}. Shop now and save big!",
			[][]string{{"John", "20%", "August 31st"}},
		).
		AddFooter("Terms and conditions apply").
		AddButtons(buttons...)

	// Validate template before creating
	if result := template.Validate(); !result.Valid {
		fmt.Println("Template validation failed:")
		for _, err := range result.Errors {
			fmt.Printf("  - %s: %s\n", err.Field, err.Message)
		}
		return fmt.Errorf("template validation failed")
	}

	// Create template
	response, err := client.CreateTemplate(ctx, template.Build())
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	fmt.Printf("Marketing template created successfully:\n")
	fmt.Printf("  ID: %s\n", response.ID)
	fmt.Printf("  Status: %s\n", response.Status)
	fmt.Printf("  Category: %s\n", response.Category)

	return nil
}

// createOrderConfirmationTemplate demonstrates creating a utility template.
func createOrderConfirmationTemplate(ctx context.Context, client *bm.Client) error {
	// Use convenience function for order confirmation
	template := bm.NewOrderConfirmationTemplate("order_confirmation_v2", "en_US")

	// Add examples
	template = bm.NewTemplate("order_confirmation_v2", "en_US", bm.CategoryUtility).
		AddHeader(bm.FormatText, "Order Confirmation").
		AddBodyWithExample(
			"Hi {{1}}, your order #{{2}} has been confirmed and will be delivered by {{3}}.",
			[][]string{{"Alice", "ORD-12345", "December 25th"}},
		).
		AddFooter("Thank you for your business!")

	// Create template
	response, err := client.CreateTemplate(ctx, template.Build())
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	fmt.Printf("Order confirmation template created successfully:\n")
	fmt.Printf("  ID: %s\n", response.ID)
	fmt.Printf("  Status: %s\n", response.Status)

	return nil
}

// createAppointmentReminderTemplate demonstrates creating a template with multiple button types.
func createAppointmentReminderTemplate(ctx context.Context, client *bm.Client) error {
	// Build buttons with different types
	buttons := bm.NewButtons().
		AddQuickReply("Confirm").
		AddQuickReply("Reschedule").
		AddPhoneNumber("Call Us", "+1234567890").
		Build()

	// Create template
	template := bm.NewTemplate("appointment_reminder_v3", "en_US", bm.CategoryUtility).
		AddBodyWithExample(
			"Hi {{1}}, this is a reminder for your appointment on {{2}} at {{3}}.",
			[][]string{{"Bob", "December 20th", "2:00 PM"}},
		).
		AddButtons(buttons...)

	// Create template
	response, err := client.CreateTemplate(ctx, template.Build())
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	fmt.Printf("Appointment reminder template created successfully:\n")
	fmt.Printf("  ID: %s\n", response.ID)
	fmt.Printf("  Status: %s\n", response.Status)

	return nil
}

// listTemplates demonstrates listing and filtering templates.
func listTemplates(ctx context.Context, client *bm.Client) error {
	// List all templates
	response, err := client.ListTemplates(ctx,
		bm.WithTemplateFields("id", "name", "status", "category"),
		bm.WithTemplateLimit(10),
	)
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	fmt.Printf("Found %d templates:\n", len(response.Data))
	for _, template := range response.Data {
		fmt.Printf("  - %s (%s) - Status: %s, Category: %s\n",
			template.Name, template.ID, template.Status, template.Category)
	}

	// List only approved templates
	approved, err := client.GetApprovedTemplates(ctx)
	if err != nil {
		return fmt.Errorf("failed to get approved templates: %w", err)
	}

	fmt.Printf("\nApproved templates: %d\n", len(approved.Data))

	// List pending templates
	pending, err := client.GetPendingTemplates(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pending templates: %w", err)
	}

	fmt.Printf("Pending templates: %d\n", len(pending.Data))

	return nil
}

// sendTemplateMessages demonstrates sending various types of template messages.
func sendTemplateMessages(ctx context.Context, client *cloudapi.Client, userPhone string) error {
	// Send simple template (assuming it exists and is approved)
	fmt.Println("Sending simple template message...")
	_, err := client.SendTemplate(ctx, userPhone, "hello_world", "en_US")
	if err != nil {
		fmt.Printf("Failed to send simple template: %v\n", err)
	}

	// Send template with text parameters
	fmt.Println("Sending template with parameters...")
	_, err = client.SendTemplateWithParams(ctx, userPhone, "order_confirmation_v2", "en_US",
		"John Doe", "ORD-12345", "December 25th")
	if err != nil {
		fmt.Printf("Failed to send template with params: %v\n", err)
	}

	// Send template with header image
	fmt.Println("Sending template with header image...")
	_, err = client.SendTemplateWithHeaderImage(ctx, userPhone, "summer_sale_2024", "en_US",
		"https://example.com/summer-banner.jpg", "John", "20%", "August 31st")
	if err != nil {
		fmt.Printf("Failed to send template with header image: %v\n", err)
	}

	// Send template with currency
	fmt.Println("Sending template with currency...")
	_, err = client.SendTemplateWithCurrency(ctx, userPhone, "price_alert", "en_US",
		2500, "USD", "$25.00", "Premium Plan")
	if err != nil {
		fmt.Printf("Failed to send template with currency: %v\n", err)
	}

	return nil
}

// templateManagementWorkflow demonstrates a complete template management workflow.
func templateManagementWorkflow(ctx context.Context, client *bm.Client) error {
	templateName := "workflow_test_template"

	// Step 1: Create template
	fmt.Println("Step 1: Creating template...")
	template := bm.NewTemplate(templateName, "en_US", bm.CategoryUtility).
		AddBody("This is a test template created at {{1}}.").
		AddBodyWithExample("This is a test template created at {{1}}.",
			[][]string{{time.Now().Format("2006-01-02 15:04:05")}})

	response, err := client.CreateTemplate(ctx, template.Build())
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	templateID := response.ID
	fmt.Printf("Template created with ID: %s\n", templateID)

	// Step 2: Get template details
	fmt.Println("Step 2: Getting template details...")
	templateDetails, err := client.GetTemplate(ctx, templateID, "id", "name", "status", "category")
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	fmt.Printf("Template details: %s - Status: %s\n", templateDetails.Name, templateDetails.Status)

	// Step 3: Update template (if allowed)
	if templateDetails.Status == "REJECTED" || templateDetails.Status == "APPROVED" {
		fmt.Println("Step 3: Updating template category...")
		newCategory := bm.CategoryMarketing
		updateRequest := &bm.UpdateTemplateRequest{
			Category: &newCategory,
		}

		err = client.UpdateTemplate(ctx, templateID, updateRequest)
		if err != nil {
			fmt.Printf("Failed to update template: %v\n", err)
		} else {
			fmt.Println("Template updated successfully")
		}
	}

	// Step 4: Delete template (cleanup)
	fmt.Println("Step 4: Deleting template...")
	err = client.DeleteTemplate(ctx, templateName, templateID)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	fmt.Println("Template deleted successfully")

	return nil
}
