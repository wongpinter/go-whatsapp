# WhatsApp Flows Package

The `flows` package provides comprehensive support for building, managing, and handling WhatsApp Flows with a fluent builder API, Flow JSON generation, and complete lifecycle management.

## Features

- 🏗️ **Fluent Builder API**: Intuitive Flow construction with method chaining
- 📋 **Complete Flow JSON Support**: Full WhatsApp Flow JSON schema implementation
- 🔧 **Flow Management**: Create, update, publish, and manage Flows via Graph API
- 🔄 **Data Exchange Handling**: Built-in support for dynamic Flow interactions
- 🔐 **Flow Token Management**: Secure token generation and validation
- 📡 **Webhook Integration**: Automatic Flow completion event handling
- ✅ **Validation Framework**: Comprehensive Flow validation with detailed error reporting
- 🧩 **Component Library**: Rich set of pre-built Flow components

## Quick Start

### 1. Building a Flow

```go
import "github.com/wongpinter/go-whatsapp/flows"

// Create a survey Flow
flow := flows.NewFlow().
    WithRouting("START", "QUESTIONS").
    WithRouting("QUESTIONS", "END").
    AddScreen(
        flows.NewScreen("START").
            WithTitle("Customer Survey").
            AddComponent(flows.NewTextHeading("Welcome!")).
            AddComponent(flows.NewTextBody("Help us improve our service.")).
            AddComponent(flows.NewFooter("Start Survey")).
            Build(),
    ).
    AddScreen(
        flows.NewScreen("QUESTIONS").
            WithTitle("Survey Questions").
            AddComponent(flows.NewTextInput("name", "Your Name").AsRequired()).
            AddComponent(flows.NewEmailInput("email", "Email Address").AsRequired()).
            AddComponent(flows.NewRadioButtonsGroup("satisfaction", "How satisfied are you?").
                AddOption("very_satisfied", "Very Satisfied").
                AddOption("satisfied", "Satisfied").
                AddOption("neutral", "Neutral").
                AsRequired().
                Build()).
            AddComponent(flows.NewDataExchangeFooter("Submit", map[string]interface{}{
                "action": "submit_survey",
            })).
            Build(),
    ).
    AddScreen(
        flows.NewScreen("END").
            AsTerminal().
            AsSuccess().
            AddComponent(flows.NewTextHeading("Thank you!")).
            AddComponent(flows.NewCompleteFooter("Close")).
            Build(),
    ).
    Build()
```

### 2. Flow Management

```go
// Create Flow management client
flowClient := flows.NewClient(wabaID, accessToken)

// Create and upload Flow
flowID, err := flowClient.CreateAndUploadFlow(ctx, &flows.CreateFlowRequest{
    Name:       "Customer Survey",
    Categories: []string{flows.CategorySurvey},
}, flow)

// Publish Flow
err = flowClient.PublishFlow(ctx, flowID)

// Get Flow information
flowInfo, err := flowClient.GetFlow(ctx, flowID, "id", "name", "status")
```

### 3. Sending Flow Messages

```go
import "github.com/wongpinter/go-whatsapp/cloudapi"

// Initialize CloudAPI client
cloudClient := cloudapi.NewClient(phoneNumberID, accessToken)

// Generate Flow token
flowToken, err := flows.GenerateFlowToken(flowID, userID, map[string]interface{}{
    "survey_type": "customer_satisfaction",
})

// Send Flow message
response, err := cloudClient.SendFlow(
    ctx,
    userPhoneNumber,
    "Please take our quick survey",
    flowID,
    flowToken,
    "Take Survey",
)
```

### 4. Data Exchange Handling

```go
// Create action router
router := flows.NewActionRouter()

// Register action handler
router.RegisterHandlerFunc("submit_survey", func(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
    // Process survey data
    name := request.Data["name"]
    email := request.Data["email"]
    satisfaction := request.Data["satisfaction"]
    
    // Save to database
    saveSurveyResponse(name, email, satisfaction)
    
    return &flows.DataExchangeResponse{
        Version: request.Version,
        Screen:  "END",
        Data: map[string]interface{}{
            "submission_id": "12345",
        },
    }, nil
})

// Create data exchange handler
handler := flows.NewDataExchangeHandler(
    flows.WithActionRouter(router),
)

// Use as HTTP handler
http.Handle("/flows/data-exchange", handler)
```

### 5. Webhook Integration

```go
import "github.com/wongpinter/go-whatsapp/webhook"

// Create webhook handler
webhookHandler := webhook.NewHandler(appSecret, verifyToken, logger)

// Register Flow reply handler
dispatcher := webhookHandler.GetDispatcher()
dispatcher.RegisterFlowReplyHandler(&FlowReplyHandler{})

type FlowReplyHandler struct{}

func (h *FlowReplyHandler) OnFlowReply(ctx context.Context, msg *webhook.Message, metadata *webhook.Metadata) error {
    flowReply := msg.Interactive.FlowReply
    
    // Process Flow completion
    completion := &flows.FlowCompletion{
        FlowToken: flowReply.FlowToken,
        Response:  flowReply.Response,
    }
    
    // Handle completion based on Flow type
    return processFlowCompletion(completion)
}
```

## Components

### Text Components

```go
// Text components
flows.NewTextHeading("Main Title")
flows.NewTextSubheading("Subtitle")
flows.NewTextBody("Body text content")
flows.NewTextCaption("Caption text")
```

### Input Components

```go
// Input components
flows.NewTextInput("name", "Your Name").AsRequired()
flows.NewEmailInput("email", "Email Address").AsRequired()
flows.NewNumberInput("age", "Your Age")
flows.NewPhoneInput("phone", "Phone Number")
flows.NewTextArea("feedback", "Your Feedback")
flows.NewDatePicker("date", "Select Date")
```

### Selection Components

```go
// Radio buttons
flows.NewRadioButtonsGroup("choice", "Select One").
    AddOption("option1", "Option 1").
    AddOption("option2", "Option 2").
    AsRequired().
    Build()

// Checkboxes
flows.NewCheckboxGroup("interests", "Select Interests").
    AddOption("tech", "Technology").
    AddOption("sports", "Sports").
    WithMinSelected(1).
    WithMaxSelected(3).
    Build()

// Dropdown
options := []flows.DataSourceItem{
    {ID: "small", Title: "Small (1-10 employees)"},
    {ID: "medium", Title: "Medium (11-50 employees)"},
    {ID: "large", Title: "Large (50+ employees)"},
}
flows.NewDropdown("company_size", "Company Size", options)
```

### Interactive Components

```go
// Footer buttons
flows.NewFooter("Continue")
flows.NewDataExchangeFooter("Submit", map[string]interface{}{
    "action": "submit_form",
})
flows.NewCompleteFooter("Finish")

// Opt-in
flows.NewOptIn("I agree to receive marketing communications")
```

## Flow Categories

Available Flow categories:

- `flows.CategorySignUp` - User registration flows
- `flows.CategorySignIn` - User authentication flows
- `flows.CategoryAppointmentBooking` - Appointment scheduling flows
- `flows.CategoryLeadGeneration` - Lead capture flows
- `flows.CategoryContactUs` - Contact form flows
- `flows.CategoryCustomerSupport` - Support ticket flows
- `flows.CategorySurvey` - Survey and feedback flows
- `flows.CategoryOther` - Other use cases

## Validation

```go
// Validate Flow before uploading
validator := flows.NewFlowValidator()
errors := validator.ValidateFlow(flow)

if len(errors) > 0 {
    for _, err := range errors {
        fmt.Printf("Validation error: %s (line %d)\n", err.Message, err.Line)
    }
}

// Check if Flow is valid
if flows.IsValidFlow(flow) {
    fmt.Println("Flow is valid!")
}
```

## Examples

See the `examples/flows/` directory for complete working examples:

- **Survey Flow**: Customer satisfaction survey with multiple question types
- **Lead Generation Flow**: Lead capture form with validation
- **Appointment Booking Flow**: Multi-step appointment scheduling
- **Data Exchange Handlers**: Complete action handler implementations
- **Webhook Integration**: Flow completion event handling

## Error Handling

The package provides structured error handling:

```go
// Flow validation errors
if err := flow.Validate(); err != nil {
    switch err {
    case flows.ErrMissingVersion:
        // Handle missing version
    case flows.ErrNoScreens:
        // Handle missing screens
    default:
        // Handle other validation errors
    }
}

// API errors
if err := flowClient.CreateFlow(ctx, request); err != nil {
    if apiErr, ok := err.(*flows.APIError); ok {
        fmt.Printf("API Error: %s (code: %d)", apiErr.Err.Message, apiErr.Err.Code)
    }
}
```

## Best Practices

1. **Always validate Flows** before uploading to WhatsApp
2. **Use meaningful screen IDs** in UPPER_CASE format
3. **Implement proper error handling** in data exchange handlers
4. **Secure Flow tokens** with appropriate TTL values
5. **Test Flows thoroughly** using the preview URL feature
6. **Handle Flow completions** asynchronously for better performance
7. **Use structured logging** for debugging and monitoring

## License

This package is part of the go-whatsapp library and follows the same license terms.
