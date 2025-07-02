# WhatsApp Flows Usage Guide

This guide provides comprehensive examples and best practices for using the WhatsApp Flows package.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Building Your First Flow](#building-your-first-flow)
3. [Flow Management](#flow-management)
4. [Sending Flow Messages](#sending-flow-messages)
5. [Handling Data Exchange](#handling-data-exchange)
6. [Processing Flow Completions](#processing-flow-completions)
7. [Advanced Features](#advanced-features)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

## Getting Started

### Installation

```bash
go get github.com/wongpinter/go-whatsapp
```

### Basic Setup

```go
import (
    "github.com/wongpinter/go-whatsapp/flows"
    "github.com/wongpinter/go-whatsapp/cloudapi"
    "github.com/wongpinter/go-whatsapp/webhook"
)

// Initialize clients
flowClient := flows.NewClient(wabaID, accessToken)
cloudClient := cloudapi.NewClient(phoneNumberID, accessToken)
```

## Building Your First Flow

### Simple Contact Form Flow

```go
func createContactFlow() *flows.Flow {
    return flows.NewFlow().
        WithRouting("START", "FORM").
        WithRouting("FORM", "CONFIRMATION").
        AddScreen(
            flows.NewScreen("START").
                WithTitle("Contact Us").
                AddComponent(flows.NewTextHeading("Get in Touch")).
                AddComponent(flows.NewTextBody("We'd love to hear from you!")).
                AddComponent(flows.NewFooter("Continue")).
                Build(),
        ).
        AddScreen(
            flows.NewScreen("FORM").
                WithTitle("Contact Information").
                AddComponent(flows.NewTextHeading("Your Details")).
                AddComponent(flows.NewTextInput("name", "Full Name").AsRequired().Build()).
                AddComponent(flows.NewEmailInput("email", "Email Address").AsRequired().Build()).
                AddComponent(flows.NewTextArea("message", "Your Message").AsRequired().Build()).
                AddComponent(flows.NewDataExchangeFooter("Send Message", map[string]interface{}{
                    "action": "submit_contact",
                })).
                Build(),
        ).
        AddScreen(
            flows.NewScreen("CONFIRMATION").
                AsTerminal().
                AsSuccess().
                AddComponent(flows.NewTextHeading("Message Sent!")).
                AddComponent(flows.NewTextBody("Thank you for contacting us. We'll get back to you soon.")).
                AddComponent(flows.NewCompleteFooter("Close")).
                Build(),
        ).
        Build()
}
```

### Multi-Step Survey Flow

```go
func createSurveyFlow() *flows.Flow {
    return flows.NewFlow().
        WithRouting("INTRO", "DEMOGRAPHICS").
        WithRouting("DEMOGRAPHICS", "PREFERENCES").
        WithRouting("PREFERENCES", "FEEDBACK").
        WithRouting("FEEDBACK", "THANK_YOU").
        AddScreen(
            flows.NewScreen("INTRO").
                WithTitle("Customer Survey").
                AddComponent(flows.NewTextHeading("Help Us Improve")).
                AddComponent(flows.NewTextBody("This survey takes about 3 minutes to complete.")).
                AddComponent(flows.NewOptIn("I agree to participate in this survey")).
                AddComponent(flows.NewFooter("Start Survey")).
                Build(),
        ).
        AddScreen(
            flows.NewScreen("DEMOGRAPHICS").
                WithTitle("About You").
                AddComponent(flows.NewTextHeading("Tell us about yourself")).
                AddComponent(flows.NewTextInput("name", "Your Name").AsRequired().Build()).
                AddComponent(flows.NewRadioButtonsGroup("age_group", "Age Group").
                    AddOption("18-25", "18-25 years").
                    AddOption("26-35", "26-35 years").
                    AddOption("36-45", "36-45 years").
                    AddOption("46-55", "46-55 years").
                    AddOption("55+", "55+ years").
                    AsRequired().
                    Build()).
                AddComponent(flows.NewFooter("Next")).
                Build(),
        ).
        // Add more screens...
        Build()
}
```

## Flow Management

### Creating and Publishing Flows

```go
func deployFlow(client *flows.Client, flow *flows.Flow) (string, error) {
    ctx := context.Background()
    
    // Create Flow
    flowID, err := client.CreateAndUploadFlow(ctx, &flows.CreateFlowRequest{
        Name:       "Contact Form",
        Categories: []string{flows.CategoryContactUs},
    }, flow)
    if err != nil {
        return "", fmt.Errorf("failed to create flow: %w", err)
    }
    
    // Get preview URL for testing
    preview, err := client.GetPreviewURL(ctx, flowID, false)
    if err != nil {
        log.Printf("Warning: Could not get preview URL: %v", err)
    } else {
        log.Printf("Preview URL: %s", preview.PreviewURL)
    }
    
    // Publish Flow
    err = client.PublishFlow(ctx, flowID)
    if err != nil {
        return "", fmt.Errorf("failed to publish flow: %w", err)
    }
    
    return flowID, nil
}
```

### Flow Lifecycle Management

```go
func manageFlowLifecycle(client *flows.Client, flowID string) error {
    ctx := context.Background()
    
    // Check Flow status
    flowInfo, err := client.GetFlow(ctx, flowID, "status", "health_status")
    if err != nil {
        return err
    }
    
    switch flowInfo.Status {
    case flows.FlowStatusDraft:
        log.Println("Flow is in draft - ready to publish")
    case flows.FlowStatusPublished:
        log.Println("Flow is published and active")
        if flowInfo.HealthStatus == flows.FlowHealthStatusUnhealthy {
            log.Println("Warning: Flow has health issues")
        }
    case flows.FlowStatusDeprecated:
        log.Println("Flow is deprecated")
    }
    
    return nil
}
```

## Sending Flow Messages

### Basic Flow Message

```go
func sendContactFlow(client *cloudapi.Client, flowID, userPhone string) error {
    ctx := context.Background()
    
    // Generate secure token
    flowToken, err := flows.GenerateFlowToken(flowID, userPhone, map[string]interface{}{
        "source": "website",
        "timestamp": time.Now().Unix(),
    })
    if err != nil {
        return err
    }
    
    // Send Flow message
    _, err = client.SendFlow(
        ctx,
        userPhone,
        "We'd love to hear from you! Please fill out our contact form.",
        flowID,
        flowToken,
        "Contact Us",
    )
    
    return err
}
```

### Flow Message with Initial Data

```go
func sendPersonalizedSurvey(client *cloudapi.Client, flowID, userPhone, userName string) error {
    ctx := context.Background()
    
    // Generate token with user data
    flowToken, err := flows.GenerateFlowToken(flowID, userPhone, map[string]interface{}{
        "user_name": userName,
        "survey_type": "satisfaction",
        "pre_filled": true,
    })
    if err != nil {
        return err
    }
    
    // Send personalized Flow message
    _, err = client.SendFlowWithData(
        ctx,
        userPhone,
        fmt.Sprintf("Hi %s! Help us improve by taking our quick survey.", userName),
        flowID,
        flowToken,
        "Take Survey",
        map[string]interface{}{
            "user_name": userName,
        },
    )
    
    return err
}
```

## Handling Data Exchange

### Setting Up Data Exchange Endpoint

```go
func setupDataExchange() *flows.DataExchangeHandler {
    router := flows.NewActionRouter()
    
    // Register action handlers
    router.RegisterHandlerFunc("submit_contact", handleContactSubmission)
    router.RegisterHandlerFunc("validate_email", handleEmailValidation)
    router.RegisterHandlerFunc("load_preferences", handlePreferencesLoad)
    
    return flows.NewDataExchangeHandler(
        flows.WithActionRouter(router),
    )
}

func handleContactSubmission(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
    // Extract form data
    name := request.Data["name"].(string)
    email := request.Data["email"].(string)
    message := request.Data["message"].(string)
    
    // Validate data
    if !isValidEmail(email) {
        return &flows.DataExchangeResponse{
            Version: request.Version,
            ErrorDetails: &flows.ErrorDetails{
                Code:    "INVALID_EMAIL",
                Message: "Please enter a valid email address",
            },
        }, nil
    }
    
    // Save to database
    contactID, err := saveContactSubmission(name, email, message)
    if err != nil {
        return &flows.DataExchangeResponse{
            Version: request.Version,
            ErrorDetails: &flows.ErrorDetails{
                Code:    "SAVE_FAILED",
                Message: "Failed to save your message. Please try again.",
            },
        }, nil
    }
    
    // Send confirmation email
    go sendConfirmationEmail(email, name, contactID)
    
    return &flows.DataExchangeResponse{
        Version: request.Version,
        Screen:  "CONFIRMATION",
        Data: map[string]interface{}{
            "contact_id": contactID,
            "name":       name,
        },
    }, nil
}
```

### Dynamic Data Loading

```go
func handlePreferencesLoad(ctx context.Context, request *flows.DataExchangeRequest) (*flows.DataExchangeResponse, error) {
    // Get user preferences from token
    tokenInfo := ctx.Value("token_info").(*flows.FlowTokenInfo)
    userID := tokenInfo.UserID
    
    // Load user's previous preferences
    preferences, err := loadUserPreferences(userID)
    if err != nil {
        // Return default preferences
        preferences = getDefaultPreferences()
    }
    
    return &flows.DataExchangeResponse{
        Version: request.Version,
        Screen:  "PREFERENCES",
        Data: map[string]interface{}{
            "current_preferences": preferences,
            "available_options":   getAvailableOptions(),
        },
    }, nil
}
```

## Processing Flow Completions

### Webhook Integration

```go
func setupWebhookWithFlows() *webhook.Handler {
    handler := webhook.NewHandler(appSecret, verifyToken, logger)
    
    // Register Flow completion handler
    dispatcher := handler.GetDispatcher()
    dispatcher.RegisterFlowReplyHandler(&FlowCompletionHandler{})
    
    return handler
}

type FlowCompletionHandler struct{}

func (h *FlowCompletionHandler) OnFlowReply(ctx context.Context, msg *webhook.Message, metadata *webhook.Metadata) error {
    flowReply := msg.Interactive.FlowReply
    
    // Validate token
    tokenInfo, err := flows.ValidateFlowToken(flowReply.FlowToken)
    if err != nil {
        log.Printf("Invalid flow token: %v", err)
        return nil // Don't return error to avoid webhook retries
    }
    
    // Process based on Flow type
    switch tokenInfo.FlowID {
    case "contact_flow_id":
        return h.processContactCompletion(flowReply, tokenInfo)
    case "survey_flow_id":
        return h.processSurveyCompletion(flowReply, tokenInfo)
    default:
        log.Printf("Unknown flow ID: %s", tokenInfo.FlowID)
    }
    
    return nil
}

func (h *FlowCompletionHandler) processContactCompletion(reply *flows.FlowReply, tokenInfo *flows.FlowTokenInfo) error {
    // Extract final response data
    finalData := reply.Response
    
    // Update contact record with completion status
    contactID := finalData["contact_id"].(string)
    err := markContactCompleted(contactID)
    if err != nil {
        log.Printf("Failed to mark contact completed: %v", err)
    }
    
    // Trigger follow-up actions
    go scheduleFollowUp(contactID, tokenInfo.UserID)
    
    return nil
}
```

## Advanced Features

### Flow State Management

```go
func setupAdvancedFlowHandling() {
    stateManager := flows.NewFlowStateManager()
    tokenManager := flows.NewFlowTokenManager(24 * time.Hour)
    
    // Create completion handler with state management
    completionHandler := flows.NewFlowCompletionHandler(stateManager, tokenManager, logger)
    
    // Register completion handlers for different flows
    completionHandler.RegisterCompletionHandlerFunc("multi_step_flow", func(ctx context.Context, completion *flows.FlowCompletion, state *flows.FlowState) error {
        // Access flow state across multiple interactions
        log.Printf("Flow completed for user %s", state.UserID)
        log.Printf("Flow data: %+v", state.Data)
        log.Printf("Final response: %+v", completion.Response)
        
        // Process completion based on accumulated state
        return processMultiStepCompletion(state, completion)
    })
}
```

### Custom Validation

```go
func validateFlowBeforeUpload(flow *flows.Flow) error {
    // Use built-in validator
    validator := flows.NewFlowValidator()
    errors := validator.ValidateFlow(flow)
    
    if len(errors) > 0 {
        for _, err := range errors {
            log.Printf("Validation error: %s (line %d)", err.Message, err.Line)
        }
        return fmt.Errorf("flow validation failed with %d errors", len(errors))
    }
    
    // Custom business logic validation
    if err := validateBusinessRules(flow); err != nil {
        return fmt.Errorf("business rule validation failed: %w", err)
    }
    
    return nil
}

func validateBusinessRules(flow *flows.Flow) error {
    // Example: Ensure all flows have at least one required field
    hasRequiredField := false
    for _, screen := range flow.Screens {
        for _, component := range screen.Layout.Children {
            if component.Required {
                hasRequiredField = true
                break
            }
        }
    }
    
    if !hasRequiredField {
        return fmt.Errorf("flow must have at least one required field")
    }
    
    return nil
}
```

## Best Practices

### 1. Flow Design

- **Keep flows short**: Aim for 3-5 screens maximum
- **Use clear navigation**: Always provide clear next steps
- **Validate early**: Validate input as soon as possible
- **Provide feedback**: Show progress and confirmation

### 2. Security

- **Secure tokens**: Use appropriate TTL for flow tokens
- **Validate inputs**: Always validate data in exchange handlers
- **Handle errors gracefully**: Don't expose internal errors to users

### 3. Performance

- **Async processing**: Handle completions asynchronously
- **Cache data**: Cache frequently accessed data
- **Monitor metrics**: Track flow completion rates and errors

### 4. User Experience

- **Mobile-first**: Design for mobile screens
- **Clear labels**: Use descriptive field labels
- **Error messages**: Provide helpful error messages
- **Progress indicators**: Show users where they are in the flow

## Troubleshooting

### Common Issues

1. **Flow validation errors**: Check the validation output for specific issues
2. **Token validation failures**: Ensure tokens haven't expired
3. **Data exchange errors**: Verify action handlers are registered correctly
4. **Webhook delivery issues**: Check webhook endpoint accessibility

### Debugging Tips

```go
// Enable debug logging
logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
flowClient := flows.NewClient(wabaID, accessToken, flows.WithLogger(&logger))

// Validate flows before uploading
if !flows.IsValidFlow(flow) {
    validator := flows.NewFlowValidator()
    errors := validator.ValidateFlow(flow)
    for _, err := range errors {
        log.Printf("Error: %s", err.Message)
    }
}

// Test with preview URLs
preview, _ := flowClient.GetPreviewURL(ctx, flowID, false)
log.Printf("Test your flow: %s", preview.PreviewURL)
```

For more examples, see the `examples/flows/` directory in the repository.
