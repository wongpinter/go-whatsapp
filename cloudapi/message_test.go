package cloudapi

import (
	"testing"
)

func TestNewTextMessage(t *testing.T) {
	msg := NewTextMessage("+1234567890", "Hello, World!")

	if msg.MessageType() != "text" {
		t.Errorf("Expected message type 'text', got '%s'", msg.MessageType())
	}

	if msg.GetTo() != "+1234567890" {
		t.Errorf("Expected recipient '+1234567890', got '%s'", msg.GetTo())
	}

	if msg.Text.Body != "Hello, World!" {
		t.Errorf("Expected body 'Hello, World!', got '%s'", msg.Text.Body)
	}

	if msg.MessagingProduct != "whatsapp" {
		t.Errorf("Expected messaging product 'whatsapp', got '%s'", msg.MessagingProduct)
	}

	if msg.RecipientType != "individual" {
		t.Errorf("Expected recipient type 'individual', got '%s'", msg.RecipientType)
	}
}

func TestTextMessageValidation(t *testing.T) {
	// Test valid message
	msg := NewTextMessage("+1234567890", "Hello")
	if err := msg.Validate(); err != nil {
		t.Errorf("Expected no error for valid message, got %v", err)
	}

	// Test empty recipient
	msg = NewTextMessage("", "Hello")
	if err := msg.Validate(); err == nil {
		t.Error("Expected error for empty recipient")
	}

	// Test empty body
	msg = NewTextMessage("+1234567890", "")
	if err := msg.Validate(); err == nil {
		t.Error("Expected error for empty body")
	}

	// Test body too long
	longBody := make([]byte, 5000)
	for i := range longBody {
		longBody[i] = 'a'
	}
	msg = NewTextMessage("+1234567890", string(longBody))
	if err := msg.Validate(); err == nil {
		t.Error("Expected error for body too long")
	}
}

func TestTextMessageWithPreviewURL(t *testing.T) {
	msg := NewTextMessage("+1234567890", "Check this out: https://example.com").
		WithPreviewURL(true)

	if !msg.Text.PreviewURL {
		t.Error("Expected preview URL to be enabled")
	}
}

func TestNewImageMessageFromURL(t *testing.T) {
	msg := NewImageMessageFromURL("+1234567890", "https://example.com/image.jpg")

	if msg.MessageType() != "image" {
		t.Errorf("Expected message type 'image', got '%s'", msg.MessageType())
	}

	if msg.Image.Link != "https://example.com/image.jpg" {
		t.Errorf("Expected image link 'https://example.com/image.jpg', got '%s'", msg.Image.Link)
	}

	if msg.Image.ID != "" {
		t.Error("Expected empty image ID when using URL")
	}
}

func TestNewImageMessageFromID(t *testing.T) {
	msg := NewImageMessageFromID("+1234567890", "media123")

	if msg.MessageType() != "image" {
		t.Errorf("Expected message type 'image', got '%s'", msg.MessageType())
	}

	if msg.Image.ID != "media123" {
		t.Errorf("Expected image ID 'media123', got '%s'", msg.Image.ID)
	}

	if msg.Image.Link != "" {
		t.Error("Expected empty image link when using ID")
	}
}

func TestImageMessageValidation(t *testing.T) {
	// Test valid message with URL
	msg := NewImageMessageFromURL("+1234567890", "https://example.com/image.jpg")
	if err := msg.Validate(); err != nil {
		t.Errorf("Expected no error for valid message, got %v", err)
	}

	// Test valid message with ID
	msg = NewImageMessageFromID("+1234567890", "media123")
	if err := msg.Validate(); err != nil {
		t.Errorf("Expected no error for valid message, got %v", err)
	}

	// Test message with both URL and ID (should fail)
	msg = NewImageMessageFromURL("+1234567890", "https://example.com/image.jpg")
	msg.Image.ID = "media123"
	if err := msg.Validate(); err == nil {
		t.Error("Expected error for message with both URL and ID")
	}

	// Test message with neither URL nor ID (should fail)
	msg = &ImageMessage{}
	msg.To = "+1234567890"
	msg.MessagingProduct = "whatsapp"
	msg.RecipientType = "individual"
	if err := msg.Validate(); err == nil {
		t.Error("Expected error for message with neither URL nor ID")
	}
}

func TestImageMessageWithCaption(t *testing.T) {
	msg := NewImageMessageFromURL("+1234567890", "https://example.com/image.jpg").
		WithCaption("Beautiful sunset")

	if msg.Image.Caption != "Beautiful sunset" {
		t.Errorf("Expected caption 'Beautiful sunset', got '%s'", msg.Image.Caption)
	}
}

func TestDocumentMessage(t *testing.T) {
	msg := NewDocumentMessageFromURL("+1234567890", "https://example.com/doc.pdf").
		WithCaption("Important document").
		WithFilename("report.pdf")

	if msg.MessageType() != "document" {
		t.Errorf("Expected message type 'document', got '%s'", msg.MessageType())
	}

	if msg.Document.Link != "https://example.com/doc.pdf" {
		t.Errorf("Expected document link 'https://example.com/doc.pdf', got '%s'", msg.Document.Link)
	}

	if msg.Document.Caption != "Important document" {
		t.Errorf("Expected caption 'Important document', got '%s'", msg.Document.Caption)
	}

	if msg.Document.Filename != "report.pdf" {
		t.Errorf("Expected filename 'report.pdf', got '%s'", msg.Document.Filename)
	}
}

func TestSendMessageResponse(t *testing.T) {
	resp := &SendMessageResponse{
		MessagingProduct: "whatsapp",
		Contacts: []struct {
			Input string `json:"input"`
			WAID  string `json:"wa_id"`
		}{
			{Input: "+1234567890", WAID: "1234567890"},
		},
		Messages: []struct {
			ID string `json:"id"`
		}{
			{ID: "wamid.123456"},
		},
	}

	if resp.GetMessageID() != "wamid.123456" {
		t.Errorf("Expected message ID 'wamid.123456', got '%s'", resp.GetMessageID())
	}

	if resp.GetWAID() != "1234567890" {
		t.Errorf("Expected WAID '1234567890', got '%s'", resp.GetWAID())
	}
}

func TestNewTemplateMessage(t *testing.T) {
	msg := NewTemplateMessage("+1234567890", "order_confirmation", "en_US")

	if msg.MessageType() != "template" {
		t.Errorf("Expected message type 'template', got '%s'", msg.MessageType())
	}

	if msg.Template.Name != "order_confirmation" {
		t.Errorf("Expected template name 'order_confirmation', got '%s'", msg.Template.Name)
	}

	if msg.Template.Language.Code != "en_US" {
		t.Errorf("Expected language code 'en_US', got '%s'", msg.Template.Language.Code)
	}
}

func TestTemplateMessageWithParameters(t *testing.T) {
	msg := NewTemplateMessage("+1234567890", "order_confirmation", "en_US").
		WithTextParameter("John").
		WithTextParameter("12345").
		WithCurrencyParameter(5000, "USD", "$50.00")

	if err := msg.Validate(); err != nil {
		t.Errorf("Expected no error for valid template message, got %v", err)
	}

	if len(msg.Template.Components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(msg.Template.Components))
	}

	if len(msg.Template.Components[0].Parameters) != 3 {
		t.Errorf("Expected 3 parameters, got %d", len(msg.Template.Components[0].Parameters))
	}

	// Check text parameters
	if msg.Template.Components[0].Parameters[0].Type != "text" {
		t.Errorf("Expected first parameter type 'text', got '%s'", msg.Template.Components[0].Parameters[0].Type)
	}

	if msg.Template.Components[0].Parameters[0].Text != "John" {
		t.Errorf("Expected first parameter text 'John', got '%s'", msg.Template.Components[0].Parameters[0].Text)
	}

	// Check currency parameter
	if msg.Template.Components[0].Parameters[2].Type != "currency" {
		t.Errorf("Expected third parameter type 'currency', got '%s'", msg.Template.Components[0].Parameters[2].Type)
	}

	if msg.Template.Components[0].Parameters[2].Currency.Amount1000 != 5000 {
		t.Errorf("Expected currency amount 5000, got %d", msg.Template.Components[0].Parameters[2].Currency.Amount1000)
	}
}

func TestNewInteractiveButtonMessage(t *testing.T) {
	msg := NewInteractiveButtonMessage("+1234567890", "Please choose an option:")

	if msg.MessageType() != "interactive" {
		t.Errorf("Expected message type 'interactive', got '%s'", msg.MessageType())
	}

	if msg.Interactive.Type != "button" {
		t.Errorf("Expected interactive type 'button', got '%s'", msg.Interactive.Type)
	}

	if msg.Interactive.Body.Text != "Please choose an option:" {
		t.Errorf("Expected body text 'Please choose an option:', got '%s'", msg.Interactive.Body.Text)
	}
}

func TestInteractiveButtonMessageWithButtons(t *testing.T) {
	msg := NewInteractiveButtonMessage("+1234567890", "Please choose an option:").
		WithHeader("Survey").
		WithFooter("Thank you").
		AddButton("yes", "Yes").
		AddButton("no", "No")

	if err := msg.Validate(); err != nil {
		t.Errorf("Expected no error for valid interactive message, got %v", err)
	}

	if msg.Interactive.Header.Text != "Survey" {
		t.Errorf("Expected header text 'Survey', got '%s'", msg.Interactive.Header.Text)
	}

	if msg.Interactive.Footer.Text != "Thank you" {
		t.Errorf("Expected footer text 'Thank you', got '%s'", msg.Interactive.Footer.Text)
	}

	if action, ok := msg.Interactive.Action.(*ButtonAction); ok {
		if len(action.Buttons) != 2 {
			t.Errorf("Expected 2 buttons, got %d", len(action.Buttons))
		}

		if action.Buttons[0].Reply.ID != "yes" {
			t.Errorf("Expected first button ID 'yes', got '%s'", action.Buttons[0].Reply.ID)
		}

		if action.Buttons[0].Reply.Title != "Yes" {
			t.Errorf("Expected first button title 'Yes', got '%s'", action.Buttons[0].Reply.Title)
		}
	} else {
		t.Error("Expected ButtonAction type")
	}
}

func TestNewInteractiveListMessage(t *testing.T) {
	msg := NewInteractiveListMessage("+1234567890", "Please select an option:", "Options")

	if msg.MessageType() != "interactive" {
		t.Errorf("Expected message type 'interactive', got '%s'", msg.MessageType())
	}

	if msg.Interactive.Type != "list" {
		t.Errorf("Expected interactive type 'list', got '%s'", msg.Interactive.Type)
	}

	if action, ok := msg.Interactive.Action.(*ListAction); ok {
		if action.Button != "Options" {
			t.Errorf("Expected button text 'Options', got '%s'", action.Button)
		}
	} else {
		t.Error("Expected ListAction type")
	}
}

func TestInteractiveListMessageWithRows(t *testing.T) {
	msg := NewInteractiveListMessage("+1234567890", "Please select an option:", "Options").
		AddRow("option1", "Option 1", "First option").
		AddRow("option2", "Option 2", "Second option")

	if err := msg.Validate(); err != nil {
		t.Errorf("Expected no error for valid interactive list message, got %v", err)
	}

	if action, ok := msg.Interactive.Action.(*ListAction); ok {
		if len(action.Sections) != 1 {
			t.Errorf("Expected 1 section, got %d", len(action.Sections))
		}

		if len(action.Sections[0].Rows) != 2 {
			t.Errorf("Expected 2 rows, got %d", len(action.Sections[0].Rows))
		}

		if action.Sections[0].Rows[0].ID != "option1" {
			t.Errorf("Expected first row ID 'option1', got '%s'", action.Sections[0].Rows[0].ID)
		}

		if action.Sections[0].Rows[0].Title != "Option 1" {
			t.Errorf("Expected first row title 'Option 1', got '%s'", action.Sections[0].Rows[0].Title)
		}
	} else {
		t.Error("Expected ListAction type")
	}
}

func TestNewLocationMessage(t *testing.T) {
	msg := NewLocationMessage("+1234567890", 37.7749, -122.4194).
		WithName("San Francisco").
		WithAddress("San Francisco, CA, USA")

	if msg.MessageType() != "location" {
		t.Errorf("Expected message type 'location', got '%s'", msg.MessageType())
	}

	if msg.Location.Latitude != 37.7749 {
		t.Errorf("Expected latitude 37.7749, got %f", msg.Location.Latitude)
	}

	if msg.Location.Longitude != -122.4194 {
		t.Errorf("Expected longitude -122.4194, got %f", msg.Location.Longitude)
	}

	if msg.Location.Name != "San Francisco" {
		t.Errorf("Expected name 'San Francisco', got '%s'", msg.Location.Name)
	}

	if msg.Location.Address != "San Francisco, CA, USA" {
		t.Errorf("Expected address 'San Francisco, CA, USA', got '%s'", msg.Location.Address)
	}

	if err := msg.Validate(); err != nil {
		t.Errorf("Expected no error for valid location message, got %v", err)
	}
}
