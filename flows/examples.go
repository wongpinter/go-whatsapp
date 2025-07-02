package flows

import (
	"context"
	"fmt"
	"log"
)

// ExampleSurveyFlow demonstrates creating a simple survey Flow.
func ExampleSurveyFlow() *Flow {
	// Create a survey Flow with multiple screens
	flow := NewFlow().
		WithRouting("SURVEY_START", "SURVEY_QUESTIONS").
		WithRouting("SURVEY_QUESTIONS", "SURVEY_END").
		AddScreen(
			NewScreen("SURVEY_START").
				WithTitle("Customer Survey").
				WithData("user_name", "string", "John Doe").
				AddComponent(NewTextHeading("Welcome to our Customer Survey")).
				AddComponent(NewTextBody("Help us improve our service by answering a few questions.")).
				AddComponent(NewDataExchangeFooter("Start Survey", map[string]interface{}{
					"action": "start_survey",
				})).
				Build(),
		).
		AddScreen(
			NewScreen("SURVEY_QUESTIONS").
				WithTitle("Survey Questions").
				WithRefreshOnBack().
				AddComponent(NewTextHeading("Please answer the following questions:")).
				AddComponent(NewTextInput("name", "Your Name").AsRequired().Build()).
				AddComponent(NewEmailInput("email", "Email Address").AsRequired().Build()).
				AddComponent(NewRadioButtonsGroup("satisfaction", "How satisfied are you?").
					AddOption("very_satisfied", "Very Satisfied").
					AddOption("satisfied", "Satisfied").
					AddOption("neutral", "Neutral").
					AddOption("dissatisfied", "Dissatisfied").
					AddOption("very_dissatisfied", "Very Dissatisfied").
					AsRequired().
					Build()).
				AddComponent(NewTextArea("feedback", "Additional Feedback").Build()).
				AddComponent(NewDataExchangeFooter("Submit Survey", map[string]interface{}{
					"action": "submit_survey",
				})).
				Build(),
		).
		AddScreen(
			NewScreen("SURVEY_END").
				WithTitle("Thank You").
				AsSuccess().
				AddComponent(NewTextHeading("Thank you for your feedback!")).
				AddComponent(NewTextBody("Your responses have been recorded and will help us improve our service.")).
				AddComponent(NewCompleteFooter("Close")).
				Build(),
		).
		Build()

	return flow
}

// ExampleLeadGenerationFlow demonstrates creating a lead generation Flow.
func ExampleLeadGenerationFlow() *Flow {
	flow := NewFlow().
		WithRouting("LEAD_START", "LEAD_FORM").
		WithRouting("LEAD_FORM", "LEAD_CONFIRMATION").
		AddScreen(
			NewScreen("LEAD_START").
				WithTitle("Get a Free Quote").
				AddComponent(NewTextHeading("Get Your Free Quote Today!")).
				AddComponent(NewTextBody("Fill out our quick form to receive a personalized quote.")).
				AddComponent(NewOptIn("I agree to receive marketing communications")).
				AddComponent(NewFooter("Get Started")).
				Build(),
		).
		AddScreen(
			NewScreen("LEAD_FORM").
				WithTitle("Contact Information").
				AddComponent(NewTextHeading("Tell us about yourself")).
				AddComponent(NewTextInput("first_name", "First Name").AsRequired().Build()).
				AddComponent(NewTextInput("last_name", "Last Name").AsRequired().Build()).
				AddComponent(NewEmailInput("email", "Email Address").AsRequired().Build()).
				AddComponent(NewPhoneInput("phone", "Phone Number").AsRequired().Build()).
				AddComponent(NewDropdown("company_size", "Company Size", []DataSourceItem{
					{ID: "1-10", Title: "1-10 employees"},
					{ID: "11-50", Title: "11-50 employees"},
					{ID: "51-200", Title: "51-200 employees"},
					{ID: "201-1000", Title: "201-1000 employees"},
					{ID: "1000+", Title: "1000+ employees"},
				}).AsRequired().Build()).
				AddComponent(NewDataExchangeFooter("Submit", map[string]interface{}{
					"action": "submit_lead",
				})).
				Build(),
		).
		AddScreen(
			NewScreen("LEAD_CONFIRMATION").
				WithTitle("Quote Request Submitted").
				AsSuccess().
				AddComponent(NewTextHeading("Thank you for your interest!")).
				AddComponent(NewTextBody("We've received your information and will contact you within 24 hours with your personalized quote.")).
				AddComponent(NewCompleteFooter("Done")).
				Build(),
		).
		Build()

	return flow
}

// ExampleAppointmentBookingFlow demonstrates creating an appointment booking Flow.
func ExampleAppointmentBookingFlow() *Flow {
	flow := NewFlow().
		WithRouting("BOOKING_START", "SELECT_SERVICE").
		WithRouting("SELECT_SERVICE", "SELECT_DATE").
		WithRouting("SELECT_DATE", "BOOKING_CONFIRMATION").
		AddScreen(
			NewScreen("BOOKING_START").
				WithTitle("Book an Appointment").
				WithData("available_services", "array", []string{"consultation", "treatment", "follow_up"}).
				AddComponent(NewTextHeading("Book Your Appointment")).
				AddComponent(NewTextBody("Schedule a convenient time for your visit.")).
				AddComponent(NewFooter("Continue")).
				Build(),
		).
		AddScreen(
			NewScreen("SELECT_SERVICE").
				WithTitle("Select Service").
				WithRefreshOnBack().
				AddComponent(NewTextHeading("What service do you need?")).
				AddComponent(NewRadioButtonsGroup("service_type", "Service Type").
					AddOption("consultation", "Initial Consultation (60 min)").
					AddOption("treatment", "Treatment Session (90 min)").
					AddOption("follow_up", "Follow-up Visit (30 min)").
					AsRequired().
					Build()).
				AddComponent(NewDataExchangeFooter("Next", map[string]interface{}{
					"action": "load_available_dates",
				})).
				Build(),
		).
		AddScreen(
			NewScreen("SELECT_DATE").
				WithTitle("Select Date & Time").
				WithData("available_slots", "array", []string{}).
				WithRefreshOnBack().
				AddComponent(NewTextHeading("Choose your preferred date and time")).
				AddComponent(NewDatePicker("appointment_date", "Date").AsRequired().Build()).
				AddComponent(NewDropdown("appointment_time", "Time Slot", []DataSourceItem{}).AsRequired().Build()).
				AddComponent(NewTextInput("special_requests", "Special Requests or Notes").Build()).
				AddComponent(NewDataExchangeFooter("Book Appointment", map[string]interface{}{
					"action": "book_appointment",
				})).
				Build(),
		).
		AddScreen(
			NewScreen("BOOKING_CONFIRMATION").
				WithTitle("Appointment Confirmed").
				AsSuccess().
				AddComponent(NewTextHeading("Your appointment is confirmed!")).
				AddComponent(NewTextBody("We've sent you a confirmation email with all the details.")).
				AddComponent(NewTextBody("If you need to reschedule, please contact us at least 24 hours in advance.")).
				AddComponent(NewCompleteFooter("Done")).
				Build(),
		).
		Build()

	return flow
}

// ExampleFlowActionHandlers demonstrates implementing action handlers for data exchange.
func ExampleFlowActionHandlers() {
	// Create action router
	router := NewActionRouter()

	// Register survey action handlers
	router.RegisterHandlerFunc("start_survey", func(ctx context.Context, request *DataExchangeRequest) (*DataExchangeResponse, error) {
		// Load user data or initialize survey
		return &DataExchangeResponse{
			Version: request.Version,
			Screen:  "SURVEY_QUESTIONS",
			Data: map[string]interface{}{
				"user_name": "John Doe", // Could be loaded from database
			},
		}, nil
	})

	router.RegisterHandlerFunc("submit_survey", func(ctx context.Context, request *DataExchangeRequest) (*DataExchangeResponse, error) {
		// Process survey submission
		name := request.Data["name"]
		email := request.Data["email"]
		satisfaction := request.Data["satisfaction"]
		feedback := request.Data["feedback"]

		// Save to database
		log.Printf("Survey submitted: name=%v, email=%v, satisfaction=%v, feedback=%v",
			name, email, satisfaction, feedback)

		return &DataExchangeResponse{
			Version: request.Version,
			Screen:  "SURVEY_END",
			Data: map[string]interface{}{
				"submission_id": "12345",
			},
		}, nil
	})

	// Register lead generation action handlers
	router.RegisterHandlerFunc("submit_lead", func(ctx context.Context, request *DataExchangeRequest) (*DataExchangeResponse, error) {
		// Process lead submission
		firstName := request.Data["first_name"]
		lastName := request.Data["last_name"]
		email := request.Data["email"]
		phone := request.Data["phone"]
		companySize := request.Data["company_size"]

		// Save lead to CRM
		log.Printf("Lead submitted: %v %v, email=%v, phone=%v, company_size=%v",
			firstName, lastName, email, phone, companySize)

		return &DataExchangeResponse{
			Version: request.Version,
			Screen:  "LEAD_CONFIRMATION",
			Data: map[string]interface{}{
				"lead_id": "LEAD-67890",
			},
		}, nil
	})

	// Register appointment booking action handlers
	router.RegisterHandlerFunc("load_available_dates", func(ctx context.Context, request *DataExchangeRequest) (*DataExchangeResponse, error) {
		serviceType := request.Data["service_type"]

		// Load available time slots based on service type
		var timeSlots []DataSourceItem
		switch serviceType {
		case "consultation":
			timeSlots = []DataSourceItem{
				{ID: "09:00", Title: "9:00 AM"},
				{ID: "11:00", Title: "11:00 AM"},
				{ID: "14:00", Title: "2:00 PM"},
				{ID: "16:00", Title: "4:00 PM"},
			}
		case "treatment":
			timeSlots = []DataSourceItem{
				{ID: "09:00", Title: "9:00 AM"},
				{ID: "13:00", Title: "1:00 PM"},
			}
		case "follow_up":
			timeSlots = []DataSourceItem{
				{ID: "10:00", Title: "10:00 AM"},
				{ID: "11:00", Title: "11:00 AM"},
				{ID: "15:00", Title: "3:00 PM"},
				{ID: "16:00", Title: "4:00 PM"},
				{ID: "17:00", Title: "5:00 PM"},
			}
		}

		return &DataExchangeResponse{
			Version: request.Version,
			Screen:  "SELECT_DATE",
			Data: map[string]interface{}{
				"available_slots": timeSlots,
			},
		}, nil
	})

	router.RegisterHandlerFunc("book_appointment", func(ctx context.Context, request *DataExchangeRequest) (*DataExchangeResponse, error) {
		// Process appointment booking
		serviceType := request.Data["service_type"]
		appointmentDate := request.Data["appointment_date"]
		appointmentTime := request.Data["appointment_time"]
		specialRequests := request.Data["special_requests"]

		// Save appointment to calendar system
		log.Printf("Appointment booked: service=%v, date=%v, time=%v, notes=%v",
			serviceType, appointmentDate, appointmentTime, specialRequests)

		return &DataExchangeResponse{
			Version: request.Version,
			Screen:  "BOOKING_CONFIRMATION",
			Data: map[string]interface{}{
				"appointment_id":    "APPT-54321",
				"confirmation_code": "ABC123",
			},
		}, nil
	})

	fmt.Println("Flow action handlers registered successfully")
}

// ExampleFlowCompletionHandlers demonstrates implementing Flow completion handlers.
func ExampleFlowCompletionHandlers() {
	stateManager := NewFlowStateManager()
	tokenManager := NewFlowTokenManager(0) // Use default TTL
	completionHandler := NewFlowCompletionHandler(stateManager, tokenManager, nil)

	// Register completion handler for survey Flow
	completionHandler.RegisterCompletionHandlerFunc("survey_flow_id", func(ctx context.Context, completion *FlowCompletion, state *FlowState) error {
		log.Printf("Survey Flow completed by user %s", state.UserID)
		log.Printf("Final response: %+v", completion.Response)

		// Process final survey data
		// Send thank you email, update user profile, etc.

		return nil
	})

	// Register completion handler for lead generation Flow
	completionHandler.RegisterCompletionHandlerFunc("lead_flow_id", func(ctx context.Context, completion *FlowCompletion, state *FlowState) error {
		log.Printf("Lead generation Flow completed by user %s", state.UserID)

		// Trigger lead nurturing workflow
		// Send to CRM, schedule follow-up, etc.

		return nil
	})

	// Register completion handler for appointment booking Flow
	completionHandler.RegisterCompletionHandlerFunc("booking_flow_id", func(ctx context.Context, completion *FlowCompletion, state *FlowState) error {
		log.Printf("Appointment booking Flow completed by user %s", state.UserID)

		// Send confirmation email and calendar invite
		// Update appointment system, etc.

		return nil
	})

	fmt.Println("Flow completion handlers registered successfully")
}
