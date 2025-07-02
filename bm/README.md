# WhatsApp Business Management Package

The `bm` package provides comprehensive support for managing WhatsApp Business API resources through the Business Management API, with a primary focus on message template management.

## Features

- 🏗️ **Template Builder API**: Fluent interface for creating message templates
- 📋 **Complete Template Management**: Create, read, update, delete templates via Business Management API
- ✅ **Template Validation**: Comprehensive validation framework with detailed error reporting
- 🔧 **Template Categories**: Support for Marketing, Utility, and Authentication templates
- 🎯 **Component Support**: Headers, body, footer, and buttons with all parameter types
- 📱 **Template Sending**: Integration with CloudAPI for sending approved templates
- 🧩 **Convenience Functions**: Pre-built templates for common use cases

## Quick Start

### 1. Initialize Business Management Client

```go
import "github.com/wongpinter/go-whatsapp/bm"

// Initialize client with WABA ID and access token
client := bm.NewClient(wabaID, accessToken)
```

### 2. Create a Simple Template

```go
// Create a basic utility template
template := bm.NewTemplate("order_confirmation", "en_US", bm.CategoryUtility).
    AddBody("Hi {{1}}, your order #{{2}} has been confirmed!").
    AddBodyWithExample("Hi {{1}}, your order #{{2}} has been confirmed!",
        [][]string{{"John", "ORD-12345"}}).
    Build()

// Create the template
response, err := client.CreateTemplate(ctx, template)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Template created with ID: %s\n", response.ID)
```

### 3. Create a Marketing Template with Buttons

```go
// Build buttons
buttons := bm.NewButtons().
    AddURL("Shop Now", "https://example.com/shop").
    AddQuickReply("Not Interested").
    Build()

// Create marketing template
template := bm.NewMarketingTemplate(
    "summer_sale",
    "en_US",
    "🌞 Summer Sale!",
    "Get {{1}} off all items! Valid until {{2}}.",
    "Terms apply",
).AddButtons(buttons...).
AddBodyWithExample("Get {{1}} off all items! Valid until {{2}}.",
    [][]string{{"20%", "August 31st"}})

response, err := client.CreateTemplate(ctx, template.Build())
```

### 4. List and Manage Templates

```go
// List all templates
templates, err := client.ListTemplates(ctx,
    bm.WithTemplateFields("id", "name", "status", "category"),
    bm.WithTemplateLimit(50),
)

// Get approved templates only
approved, err := client.GetApprovedTemplates(ctx)

// Get specific template
template, err := client.GetTemplate(ctx, templateID)

// Update template category
updateRequest := &bm.UpdateTemplateRequest{
    Category: &bm.CategoryUtility,
}
err = client.UpdateTemplate(ctx, templateID, updateRequest)

// Delete template
err = client.DeleteTemplate(ctx, "template_name")
```

### 5. Send Template Messages

```go
import "github.com/wongpinter/go-whatsapp/cloudapi"

// Initialize CloudAPI client
cloudClient := cloudapi.NewClient(phoneNumberID, accessToken)

// Send template with parameters
_, err := cloudClient.SendTemplateWithParams(ctx, userPhone, "order_confirmation", "en_US",
    "John Doe", "ORD-12345")

// Send template with header image
_, err := cloudClient.SendTemplateWithHeaderImage(ctx, userPhone, "summer_sale", "en_US",
    "https://example.com/banner.jpg", "20%", "August 31st")
```

## Template Components

### Header Components

```go
// Text header
template.AddHeader(bm.FormatText, "Welcome!")

// Header with placeholder
template.AddHeaderWithExample(bm.FormatText, "Hello {{1}}!", []string{"John"})

// Image header (for templates with media)
template.AddHeader(bm.FormatImage, "")

// Document header
template.AddHeader(bm.FormatDocument, "")

// Video header
template.AddHeader(bm.FormatVideo, "")
```

### Body Components

```go
// Simple body
template.AddBody("Thank you for your order!")

// Body with placeholders and examples
template.AddBodyWithExample(
    "Hi {{1}}, your order #{{2}} will arrive on {{3}}.",
    [][]string{{"Alice", "12345", "Dec 25th"}},
)
```

### Footer Components

```go
// Simple footer (no placeholders allowed)
template.AddFooter("Terms and conditions apply")
```

### Button Components

```go
// Quick reply buttons
buttons := bm.NewButtons().
    AddQuickReply("Yes").
    AddQuickReply("No").
    Build()

// URL buttons
buttons := bm.NewButtons().
    AddURL("Visit Website", "https://example.com").
    AddURLWithExample("Visit {{1}}", "https://example.com/{{1}}", []string{"shop"}).
    Build()

// Phone number buttons
buttons := bm.NewButtons().
    AddPhoneNumber("Call Support", "+1234567890").
    Build()

// Mixed buttons (respecting limits)
buttons := bm.NewButtons().
    AddQuickReply("Confirm").
    AddQuickReply("Cancel").
    AddURL("Learn More", "https://example.com").
    AddPhoneNumber("Call Us", "+1234567890").
    Build()
```

## Template Categories

### Marketing Templates
For promotional content, offers, and marketing communications:

```go
template := bm.NewMarketingTemplate(name, language, header, body, footer)
// or
template := bm.NewTemplate(name, language, bm.CategoryMarketing)
```

### Utility Templates
For transactional messages, confirmations, and updates:

```go
template := bm.NewUtilityTemplate(name, language, body)
// or
template := bm.NewTemplate(name, language, bm.CategoryUtility)
```

### Authentication Templates
For OTP, verification codes, and security messages:

```go
template := bm.NewAuthenticationTemplate(name, language, body)
// or
template := bm.NewTemplate(name, language, bm.CategoryAuthentication)
```

## Template Validation

### Automatic Validation

```go
// Validate during build
template := bm.NewTemplate("test", "en_US", bm.CategoryUtility).
    AddBody("Test message")

result := template.Validate()
if !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("Error: %s - %s\n", err.Field, err.Message)
    }
}
```

### Manual Validation

```go
validator := bm.NewTemplateValidator()
result := validator.ValidateTemplate(templateRequest)

// Convenience function
if bm.IsValidTemplate(templateRequest) {
    // Template is valid
}
```

### Validation Rules

- **Template Name**: Lowercase letters, numbers, underscores only (max 512 chars)
- **Language**: Format `xx_XX` (e.g., `en_US`, `es_ES`)
- **Components**: Must have exactly one BODY component
- **Headers**: Max 60 characters, max 1 placeholder
- **Body**: Max 1024 characters, examples required for placeholders
- **Footer**: Max 60 characters, no placeholders allowed
- **Buttons**: Max 10 total (max 1 phone, max 2 URL, max 10 quick reply)

## Convenience Templates

### Order Confirmation

```go
template := bm.NewOrderConfirmationTemplate("order_confirm", "en_US")
// Creates: Header + Body with order placeholders + Footer
```

### Appointment Reminder

```go
template := bm.NewAppointmentReminderTemplate("appointment", "en_US")
// Creates: Body with appointment placeholders + Confirm/Reschedule buttons
```

### Promotional Template

```go
template := bm.NewPromoTemplate("promo", "en_US", 
    "Special Offer", "Get 20% off!", "Limited time", "Shop Now", "https://shop.com")
// Creates: Header + Body + Footer + URL button + Quick reply
```

## Error Handling

```go
// API errors
response, err := client.CreateTemplate(ctx, template)
if err != nil {
    if apiErr, ok := err.(*whatsapp.APIError); ok {
        fmt.Printf("API Error: %s (code: %d)\n", apiErr.Err.Message, apiErr.Err.Code)
    }
}

// Validation errors
result := template.Validate()
if !result.Valid {
    for _, validationErr := range result.Errors {
        fmt.Printf("Validation Error: %s\n", validationErr.Error())
    }
}
```

## Template Lifecycle

1. **Create** → Template enters `PENDING` status
2. **Review** → WhatsApp reviews template (automated + manual)
3. **Approval** → Template becomes `APPROVED` and can be sent
4. **Rejection** → Template becomes `REJECTED` (can be edited)
5. **Sending** → Use CloudAPI to send approved templates
6. **Management** → Update category or components, delete when needed

## Best Practices

1. **Use descriptive names** in lowercase with underscores
2. **Provide clear examples** for all placeholders
3. **Keep text concise** within character limits
4. **Test templates** using preview URLs before approval
5. **Handle rejections** by reviewing and updating templates
6. **Monitor template health** and performance
7. **Use appropriate categories** for better approval rates

## Examples

See the `examples/templates/` directory for complete working examples:

- **Basic Template Creation**: Simple utility and marketing templates
- **Advanced Templates**: Templates with buttons, media, and complex parameters
- **Template Management**: Complete CRUD operations workflow
- **Template Sending**: Integration with CloudAPI for message delivery

## License

This package is part of the go-whatsapp library and follows the same license terms.
