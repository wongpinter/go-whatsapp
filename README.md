# Go WhatsApp Cloud API SDK

A production-grade Golang SDK for the WhatsApp Business Cloud API. This SDK provides a comprehensive, type-safe, and developer-friendly interface for integrating WhatsApp messaging capabilities into your Go applications.

## Features

* 🚀 **Complete API Coverage**: Send messages, handle webhooks, manage business accounts
* 🔒 **Security First**: Built-in webhook signature verification and secure defaults
* 🏗️ **Clean Architecture**: Follows SOLID principles with clear separation of concerns
* 📝 **Type Safety**: Comprehensive type definitions for all API entities
* 🔄 **Robust Error Handling**: Structured error types with helper functions
* 📊 **Observability**: Structured logging with zerolog integration
* ⚡ **Performance**: Built-in rate limiting and retry mechanisms
* 🧪 **Testable**: Interface-based design for easy mocking and testing

## Installation

```bash
go get github.com/wongpinter/go-whatsapp
```

## Quick Start

### 1. Sending Messages

```go
package main

import (
    "context"
    "log"
    
    "github.com/wongpinter/go-whatsapp/cloudapi"
)

func main() {
    client := cloudapi.NewClient("YOUR_PHONE_NUMBER_ID", "YOUR_ACCESS_TOKEN")
    
    // Send a text message
    resp, err := client.SendText(context.Background(), "+1234567890", "Hello from Go!")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Message sent with ID: %s", resp.GetMessageID())
}
```

### 2. Handling Webhooks

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/rs/zerolog"
    "github.com/wongpinter/go-whatsapp/webhook"
)

type MyHandler struct{}

func (h *MyHandler) OnTextMessage(ctx context.Context, msg *webhook.Message, metadata *webhook.Metadata) error {
    log.Printf("Received: %s from %s", msg.Text.Body, msg.From)
    return nil
}

func main() {
    logger := zerolog.New(os.Stdout)
    handler := webhook.NewHandler("YOUR_APP_SECRET", "YOUR_VERIFY_TOKEN", logger)
    
    // Register your message handler
    handler.GetDispatcher().RegisterTextMessageHandler(&MyHandler{})
    
    http.Handle("/webhook", handler)
    http.ListenAndServe(":8080", nil)
}
```

### 3. Business Management

```go
package main

import (
    "context"
    "log"
    
    "github.com/wongpinter/go-whatsapp/bm"
)

func main() {
    client := bm.NewClient("YOUR_ACCESS_TOKEN", bm.WithWABAID("YOUR_WABA_ID"))
    
    // Get business account details
    account, err := client.GetBusinessAccount(context.Background(), "YOUR_WABA_ID")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Business: %s", account.Name)
}
```

## Package Structure

```
github.com/wongpinter/go-whatsapp/
├── cloudapi/          # Message sending functionality
├── webhook/           # Webhook handling and event processing
├── bm/               # Business Management API
├── flows/            # WhatsApp Flows builder and management
├── internal/         # Internal shared components
└── examples/         # Usage examples
```

## Core Packages

### CloudAPI Package

The `cloudapi` package handles all outbound message sending:

```go
import "github.com/wongpinter/go-whatsapp/cloudapi"

client := cloudapi.NewClient(phoneNumberID, accessToken,
    cloudapi.WithLogger(logger),
    cloudapi.WithRateLimiting(80.0), // 80 requests per second
)

// Send different message types
// Send different message types
client.SendText(ctx, to, "Hello!")
client.SendImageFromURL(ctx, to, imageURL, "Caption")
client.SendDocumentFromID(ctx, to, mediaID, "filename.pdf")

// Send template messages
client.SendTemplate(ctx, to, "hello_world", "en_US")
client.SendTemplateWithParams(ctx, to, "order_update", "en_US", "12345", "$29.99")

// Send interactive messages
buttons := map[string]string{"yes": "Yes", "no": "No"}
client.SendInteractiveButtons(ctx, to, "Are you satisfied?", buttons)

// Send location
client.SendLocation(ctx, to, 37.7749, -122.4194, "SF", "San Francisco, CA")
```

**Supported Message Types:**
* Text messages with URL preview
* Image messages (URL or media ID)
* Document messages (URL or media ID)
* Template messages with parameters
* Interactive messages (buttons and lists)
* Location messages

### Webhook Package

The `webhook` package provides comprehensive webhook handling with advanced monitoring:

```go
import "github.com/wongpinter/go-whatsapp/webhook"

// Create webhook handler with built-in monitoring
handler := webhook.NewHandler(appSecret, verifyToken, logger)
dispatcher := handler.GetDispatcher()

// Register handlers for different event types
dispatcher.RegisterTextMessageHandler(textHandler)
dispatcher.RegisterImageMessageHandler(imageHandler)
dispatcher.RegisterStatusUpdateHandler(statusHandler)

// Access monitoring features
statusMonitor := handler.GetStatusMonitor()
metrics := handler.GetMetrics()
```

**Core Features:**
* Automatic signature verification with enhanced error handling
* Event dispatching to registered handlers
* Support for all WhatsApp message types (text, media, interactive, reactions)
* Structured error handling with retry logic

**Advanced Monitoring Features:**
* **Real-time Status Tracking**: Monitor message delivery lifecycle (sent → delivered → read)
* **Comprehensive Metrics**: Request rates, processing latency, error counts
* **Delivery Analytics**: Delivery rates, read rates, failure analysis
* **Failed Message Detection**: Automatic identification of delivery failures
* **Performance Monitoring**: Webhook processing performance and bottlenecks

**Status Monitoring:**

```go
// Get delivery statistics
stats := statusMonitor.GetMessageStats()
fmt.Printf("Delivery Rate: %.1f%%", stats.DeliveryRate)
fmt.Printf("Read Rate: %.1f%%", stats.ReadRate)

// Get failed messages for analysis
failedMessages := statusMonitor.GetFailedMessages()
for _, msg := range failedMessages {
    fmt.Printf("Failed: %s - Error: %s", msg.MessageID, *msg.ErrorMessage)
}

// Get real-time metrics
metrics := handler.GetMetrics().GetMetrics()
statusCounts := metrics["status_counts"].(map[string]uint64)
```

**Supported Events:**
* Text, image, audio, video, document messages
* Interactive button and list replies
* Message reactions and unsupported message types
* Message status updates (sent, delivered, read, failed) with pricing info
* Error notifications with detailed error codes
* System messages

### Flows Package

The `flows` package provides comprehensive WhatsApp Flows support with a fluent builder API:

```go
import "github.com/wongpinter/go-whatsapp/flows"

// Create a survey Flow using the builder API
flow := flows.NewFlow().
    WithRouting("START", "QUESTIONS", "END").
    AddScreen(
        flows.NewScreen("START").
            WithTitle("Customer Survey").
            AddComponent(flows.NewTextHeading("Welcome to our survey!")).
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

// Create and upload Flow to WhatsApp
flowClient := flows.NewClient(wabaID, accessToken)
flowID, err := flowClient.CreateAndUploadFlow(ctx, &flows.CreateFlowRequest{
    Name:       "Customer Survey",
    Categories: []string{flows.CategorySurvey},
}, flow)

// Send Flow message to user
cloudClient := cloudapi.NewClient(phoneNumberID, accessToken)
flowToken, _ := flows.GenerateFlowToken(flowID, userID, nil)
cloudClient.SendFlow(ctx, userPhoneNumber, "Please take our survey", flowID, flowToken, "Take Survey")
```

**Key Features:**
* **Fluent Builder API**: Intuitive Flow construction with method chaining
* **Complete Flow JSON Support**: Full WhatsApp Flow JSON schema implementation
* **Flow Management**: Create, update, publish, and manage Flows via Graph API
* **Data Exchange Handling**: Built-in support for dynamic Flow interactions
* **Flow Token Management**: Secure token generation and validation
* **Webhook Integration**: Automatic Flow completion event handling
* **Validation Framework**: Comprehensive Flow validation with detailed error reporting

**Supported Components:**
* Text components (heading, body, caption)
* Input components (text, email, number, phone, textarea)
* Selection components (radio buttons, checkboxes, dropdown)
* Interactive components (buttons, footers, opt-in)
* Media components (images, embedded links)
* Date picker and other specialized components

### Business Management Package

The `bm` package provides comprehensive WhatsApp Business Management API support with a focus on message template management:

```go
import "github.com/wongpinter/go-whatsapp/bm"

// Initialize Business Management client
bmClient := bm.NewClient(wabaID, accessToken)

// Create a marketing template with fluent builder API
template := bm.NewMarketingTemplate(
    "summer_sale_2024",
    "en_US",
    "🌞 Summer Sale Alert!",
    "Hi {{1}}, enjoy {{2}} off on all summer items! Valid until {{3}}.",
    "Terms and conditions apply",
).WithCategoryChange(true)

// Add buttons using builder pattern
buttons := bm.NewButtons().
    AddURL("Shop Now", "https://example.com/summer-sale").
    AddQuickReply("Not Interested").
    Build()

template.AddButtons(buttons...)

// Add examples for template variables
template = bm.NewTemplate("summer_sale_2024", "en_US", bm.CategoryMarketing).
    AddHeaderWithExample(bm.FormatText, "🌞 Summer Sale Alert!", nil).
    AddBodyWithExample(
        "Hi {{1}}, enjoy {{2}} off on all summer items! Valid until {{3}}.",
        [][]string{{"John", "20%", "August 31st"}},
    ).
    AddFooter("Terms and conditions apply").
    AddButtons(buttons...)

// Validate template before submission
if result := template.Validate(); !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("Validation error: %s\n", err.Message)
    }
}

// Create template via Business Management API
response, err := bmClient.CreateTemplate(ctx, template.Build())
if err != nil {
    log.Fatal(err)
}

// List and manage templates
templates, err := bmClient.ListTemplates(ctx,
    bm.WithTemplateStatus(bm.TemplateStatusApproved),
    bm.WithTemplateCategory(bm.CategoryMarketing),
    bm.WithTemplateLimit(50),
)

// Send approved templates via CloudAPI
cloudClient := cloudapi.NewClient(phoneNumberID, accessToken)
_, err = cloudClient.SendTemplateWithParams(ctx, userPhone, "summer_sale_2024", "en_US",
    "John Doe", "20%", "August 31st")

// Get comprehensive analytics and monitoring data
endDate := time.Now().Format("2006-01-02")
startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

// Cost analytics with granular reporting
costAnalytics, err := bmClient.GetCostAnalytics(ctx, startDate, endDate,
    bm.WithAnalyticsGranularity(bm.GranularityDaily),
    bm.WithAnalyticsMetricTypes(bm.MetricTypeCost),
)
fmt.Printf("Total Cost: %.2f %s\n", costAnalytics.TotalCost.TotalCost, costAnalytics.TotalCost.Currency)

// Account quality metrics and compliance monitoring
qualityMetrics, err := bmClient.GetAccountQualityMetrics(ctx, startDate, endDate)
fmt.Printf("Quality Score: %s (Trend: %s)\n",
    qualityMetrics.QualityScore.Current, qualityMetrics.QualityScore.Trend)
fmt.Printf("Delivery Rate: %.2f%%\n", qualityMetrics.DeliveryMetrics.DeliveryRate)

// Phone number analytics and performance monitoring
phoneAnalytics, err := bmClient.GetPhoneNumberAnalytics(ctx, phoneNumberID, startDate, endDate)
fmt.Printf("Total Messages: %d (Success Rate: %.2f%%)\n",
    phoneAnalytics.PerformanceMetrics.MessageVolume.TotalMessages,
    phoneAnalytics.PerformanceMetrics.DeliveryPerformance.SuccessRate)

// Phone number status and health monitoring
phoneStatus, err := bmClient.GetPhoneNumberStatus(ctx, phoneNumberID)
fmt.Printf("Phone Status: %s (Health: %s)\n",
    phoneStatus.Status, phoneStatus.HealthStatus.Overall)
```

**Key Features:**
* **Template Builder API**: Fluent interface for creating message templates
* **Complete Template Management**: CRUD operations via Business Management API
* **Template Validation**: Comprehensive validation with detailed error reporting
* **Component Support**: Headers, body, footer, buttons with all parameter types
* **Template Categories**: Marketing, Utility, and Authentication templates
* **Button Builder**: Support for quick reply, URL, and phone number buttons
* **Convenience Functions**: Pre-built templates for common use cases
* **CloudAPI Integration**: Seamless template sending with parameter substitution
* **Analytics & Monitoring**: Comprehensive cost analytics, quality metrics, and performance monitoring
* **Phone Number Management**: Advanced phone number analytics, status monitoring, and configuration
* **Account Quality Tracking**: Quality score monitoring, delivery metrics, and compliance status
* **Cost Analytics**: Message costs, conversation costs, and template usage analytics with granular reporting

**Supported Template Components:**
* Text headers with placeholders (max 1 variable)
* Media headers (image, video, document)
* Body text with unlimited variables and examples
* Footer text (no variables allowed)
* Interactive buttons (quick reply, URL, phone number)
* Currency and date/time parameters
* Dynamic URL buttons with variables

### Business Management Package

The `bm` package manages WhatsApp Business Account resources:

```go
import "github.com/wongpinter/go-whatsapp/bm"

client := bm.NewClient(accessToken, bm.WithWABAID(wabaID))

// Manage business resources
account, _ := client.GetBusinessAccount(ctx, wabaID)
phoneNumbers, _ := client.GetPhoneNumbers(ctx, wabaID)
profile, _ := client.GetBusinessProfile(ctx, phoneNumberID)
```

## Error Handling

The SDK provides structured error handling with helper functions:

```go
resp, err := client.SendText(ctx, to, message)
if err != nil {
    // Check for specific error types
    if whatsapp.IsRateLimitError(err) {
        // Handle rate limiting
    } else if whatsapp.IsUndeliverableError(err) {
        // Handle undeliverable message
    } else if apiErr, ok := err.(*whatsapp.APIError); ok {
        // Access structured error details
        log.Printf("API Error: %d - %s", apiErr.Code(), apiErr.Message())
    }
}
```

## Configuration

### Environment Variables

```bash
# Required for sending messages
WHATSAPP_ACCESS_TOKEN=your_permanent_access_token
WHATSAPP_PHONE_NUMBER_ID=your_phone_number_id

# Required for webhooks
WHATSAPP_WEBHOOK_SECRET=your_app_secret
WHATSAPP_VERIFY_TOKEN=your_verify_token

# Optional for business management
WHATSAPP_WABA_ID=your_business_account_id
```

### Client Options

Both CloudAPI and Business Management clients support functional options:

```go
client := cloudapi.NewClient(phoneNumberID, accessToken,
    cloudapi.WithLogger(logger),
    cloudapi.WithAPIVersion("v19.0"),
    cloudapi.WithTimeout(30*time.Second),
    cloudapi.WithRateLimiting(80.0),
    cloudapi.WithRetryConfig(3, time.Second, 10*time.Second),
)
```

## Security

### Webhook Signature Verification

The SDK automatically verifies webhook signatures using HMAC-SHA256:

```go
handler := webhook.NewHandler(appSecret, verifyToken, logger)
// Signature verification is automatic and mandatory
```

### Rate Limiting

Built-in rate limiting prevents API quota exhaustion:

```go
client := cloudapi.NewClient(phoneNumberID, accessToken,
    cloudapi.WithRateLimiting(80.0), // Requests per second
)
```

## Testing

The SDK is designed for testability with interface-based dependencies:

```go
// Mock the HTTP client for testing
type MockHTTPClient struct{}
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    // Return mock response
}

client := cloudapi.NewClient(phoneNumberID, accessToken,
    cloudapi.WithHTTPClient(&MockHTTPClient{}),
)
```

## Examples

See the `examples/` directory for complete working examples:

* `examples/basic/` - Basic usage of all packages
* `examples/webhook-server/` - Complete webhook server implementation
* `examples/business-management/` - Business account management

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

* 📖 [WhatsApp Cloud API Documentation](https://developers.facebook.com/docs/whatsapp/cloud-api)
* 🐛 [Report Issues](https://github.com/wongpinter/go-whatsapp/issues)
* 💬 [Discussions](https://github.com/wongpinter/go-whatsapp/discussions)

## Roadmap

* [x] Template message builder with fluent API
* [x] Interactive message support (buttons, lists)
* [x] Location message support
* [x] Enhanced Business Management features
* [ ] WhatsApp Flows package
* [ ] Media upload utilities
* [ ] Advanced webhook features (metrics, queuing)
* [ ] Multi-tenant webhook handling
* [ ] Comprehensive test suite
* [ ] Performance benchmarks
