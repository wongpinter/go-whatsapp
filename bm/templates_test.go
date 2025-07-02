package bm

import (
	"testing"
)

func TestTemplateBuilder(t *testing.T) {
	// Test basic template creation
	template := NewTemplate("test_template", "en_US", CategoryMarketing).
		AddHeader(FormatText, "Test Header").
		AddBody("Test body with {{1}} parameter").
		AddFooter("Test footer").
		Build()

	if template.Name != "test_template" {
		t.Errorf("Expected name 'test_template', got '%s'", template.Name)
	}

	if template.Language != "en_US" {
		t.Errorf("Expected language 'en_US', got '%s'", template.Language)
	}

	if template.Category != CategoryMarketing {
		t.Errorf("Expected category '%s', got '%s'", CategoryMarketing, template.Category)
	}

	if len(template.Components) != 3 {
		t.Errorf("Expected 3 components, got %d", len(template.Components))
	}

	// Verify component types
	expectedTypes := []string{
		string(ComponentTypeHeader),
		string(ComponentTypeBody),
		string(ComponentTypeFooter),
	}

	for i, expectedType := range expectedTypes {
		if template.Components[i].Type != expectedType {
			t.Errorf("Expected component %d type '%s', got '%s'", i, expectedType, template.Components[i].Type)
		}
	}
}

func TestTemplateBuilderWithButtons(t *testing.T) {
	// Test template with buttons
	buttons := NewButtons().
		AddQuickReply("Yes").
		AddQuickReply("No").
		AddURL("Learn More", "https://example.com").
		AddPhoneNumber("Call Us", "+1234567890").
		Build()

	template := NewTemplate("button_template", "en_US", CategoryUtility).
		AddBody("Do you want to continue?").
		AddButtons(buttons...).
		Build()

	// Find buttons component
	var buttonsComponent *TemplateComponent
	for _, comp := range template.Components {
		if comp.Type == string(ComponentTypeButtons) {
			buttonsComponent = &comp
			break
		}
	}

	if buttonsComponent == nil {
		t.Fatal("Buttons component not found")
	}

	if len(buttonsComponent.Buttons) != 4 {
		t.Errorf("Expected 4 buttons, got %d", len(buttonsComponent.Buttons))
	}

	// Verify button types
	expectedButtonTypes := []string{
		string(ButtonTypeQuickReply),
		string(ButtonTypeQuickReply),
		string(ButtonTypeURL),
		string(ButtonTypePhoneNumber),
	}

	for i, expectedType := range expectedButtonTypes {
		if buttonsComponent.Buttons[i].Type != expectedType {
			t.Errorf("Expected button %d type '%s', got '%s'", i, expectedType, buttonsComponent.Buttons[i].Type)
		}
	}
}

func TestTemplateValidation(t *testing.T) {
	validator := NewTemplateValidator()

	// Test valid template
	validTemplate := &CreateTemplateRequest{
		Name:     "valid_template",
		Language: "en_US",
		Category: CategoryUtility,
		Components: []TemplateComponent{
			{
				Type: string(ComponentTypeBody),
				Text: "This is a valid template",
			},
		},
	}

	result := validator.ValidateTemplate(validTemplate)
	if !result.Valid {
		t.Errorf("Expected valid template, got errors: %v", result.Errors)
	}

	// Test invalid template name
	invalidNameTemplate := &CreateTemplateRequest{
		Name:     "Invalid-Template-Name",
		Language: "en_US",
		Category: CategoryUtility,
		Components: []TemplateComponent{
			{
				Type: string(ComponentTypeBody),
				Text: "This is a template",
			},
		},
	}

	result = validator.ValidateTemplate(invalidNameTemplate)
	if result.Valid {
		t.Error("Expected invalid template due to name format")
	}

	// Test missing body component
	missingBodyTemplate := &CreateTemplateRequest{
		Name:     "missing_body",
		Language: "en_US",
		Category: CategoryUtility,
		Components: []TemplateComponent{
			{
				Type: string(ComponentTypeHeader),
				Text: "Header only",
			},
		},
	}

	result = validator.ValidateTemplate(missingBodyTemplate)
	if result.Valid {
		t.Error("Expected invalid template due to missing body component")
	}
}

func TestTemplateValidationWithExamples(t *testing.T) {
	validator := NewTemplateValidator()

	// Test template with placeholders and examples
	templateWithExamples := &CreateTemplateRequest{
		Name:     "template_with_examples",
		Language: "en_US",
		Category: CategoryMarketing,
		Components: []TemplateComponent{
			{
				Type: string(ComponentTypeBody),
				Text: "Hello {{1}}, your order {{2}} is ready!",
				Example: &TemplateExample{
					BodyText: [][]string{{"John", "12345"}},
				},
			},
		},
	}

	result := validator.ValidateTemplate(templateWithExamples)
	if !result.Valid {
		t.Errorf("Expected valid template with examples, got errors: %v", result.Errors)
	}

	// Test template with mismatched examples
	templateWithMismatchedExamples := &CreateTemplateRequest{
		Name:     "mismatched_examples",
		Language: "en_US",
		Category: CategoryMarketing,
		Components: []TemplateComponent{
			{
				Type: string(ComponentTypeBody),
				Text: "Hello {{1}}, your order {{2}} is ready!",
				Example: &TemplateExample{
					BodyText: [][]string{{"John"}}, // Missing second parameter
				},
			},
		},
	}

	result = validator.ValidateTemplate(templateWithMismatchedExamples)
	if result.Valid {
		t.Error("Expected invalid template due to mismatched examples")
	}
}

func TestButtonValidation(t *testing.T) {
	validator := NewTemplateValidator()

	// Test template with too many buttons
	tooManyButtons := make([]TemplateButton, 11) // Max is 10
	for i := 0; i < 11; i++ {
		tooManyButtons[i] = TemplateButton{
			Type: string(ButtonTypeQuickReply),
			Text: "Button",
		}
	}

	templateWithTooManyButtons := &CreateTemplateRequest{
		Name:     "too_many_buttons",
		Language: "en_US",
		Category: CategoryUtility,
		Components: []TemplateComponent{
			{
				Type: string(ComponentTypeBody),
				Text: "Choose an option",
			},
			{
				Type:    string(ComponentTypeButtons),
				Buttons: tooManyButtons,
			},
		},
	}

	result := validator.ValidateTemplate(templateWithTooManyButtons)
	if result.Valid {
		t.Error("Expected invalid template due to too many buttons")
	}

	// Test template with invalid URL button
	invalidURLButton := &CreateTemplateRequest{
		Name:     "invalid_url_button",
		Language: "en_US",
		Category: CategoryUtility,
		Components: []TemplateComponent{
			{
				Type: string(ComponentTypeBody),
				Text: "Click the button",
			},
			{
				Type: string(ComponentTypeButtons),
				Buttons: []TemplateButton{
					{
						Type: string(ButtonTypeURL),
						Text: "Visit",
						URL:  "invalid-url", // Invalid URL format
					},
				},
			},
		},
	}

	result = validator.ValidateTemplate(invalidURLButton)
	if result.Valid {
		t.Error("Expected invalid template due to invalid URL format")
	}
}

func TestConvenienceTemplates(t *testing.T) {
	// Test marketing template
	marketingTemplate := NewMarketingTemplate(
		"marketing_test",
		"en_US",
		"Special Offer",
		"Get 50% off today!",
		"Limited time only",
	).Build()

	if marketingTemplate.Category != CategoryMarketing {
		t.Errorf("Expected marketing category, got %s", marketingTemplate.Category)
	}

	if len(marketingTemplate.Components) != 3 {
		t.Errorf("Expected 3 components, got %d", len(marketingTemplate.Components))
	}

	// Test utility template
	utilityTemplate := NewUtilityTemplate(
		"utility_test",
		"en_US",
		"Your order has been processed",
	).Build()

	if utilityTemplate.Category != CategoryUtility {
		t.Errorf("Expected utility category, got %s", utilityTemplate.Category)
	}

	// Test authentication template
	authTemplate := NewAuthenticationTemplate(
		"auth_test",
		"en_US",
		"Your verification code is 123456",
	).Build()

	if authTemplate.Category != CategoryAuthentication {
		t.Errorf("Expected authentication category, got %s", authTemplate.Category)
	}
}

func TestTemplateMessageBuilder(t *testing.T) {
	// Test template message builder
	templateMessage := NewTemplateMessage("test_template", "en_US").
		AddBodyParameters("John", "12345", "December 25th").
		AddHeaderParameter(ParameterTypeText, "Order Confirmation").
		Build()

	if templateMessage.Name != "test_template" {
		t.Errorf("Expected name 'test_template', got '%s'", templateMessage.Name)
	}

	if templateMessage.Language.Code != "en_US" {
		t.Errorf("Expected language 'en_US', got '%s'", templateMessage.Language.Code)
	}

	if len(templateMessage.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(templateMessage.Components))
	}

	// Verify body component has 3 parameters
	bodyComponent := templateMessage.Components[0]
	if bodyComponent.Type != ComponentTypeBody {
		t.Errorf("Expected first component to be body, got %s", bodyComponent.Type)
	}

	if len(bodyComponent.Parameters) != 3 {
		t.Errorf("Expected 3 body parameters, got %d", len(bodyComponent.Parameters))
	}

	// Verify header component
	headerComponent := templateMessage.Components[1]
	if headerComponent.Type != ComponentTypeHeader {
		t.Errorf("Expected second component to be header, got %s", headerComponent.Type)
	}
}

func TestTemplateMessageValidation(t *testing.T) {
	validator := NewTemplateValidator()

	// Test valid template message
	validMessage := &TemplateMessagePayload{
		Name: "valid_template",
		Language: TemplateLanguage{
			Code: "en_US",
		},
		Components: []TemplateComponentParam{
			{
				Type: ComponentTypeBody,
				Parameters: []TemplateParameter{
					{
						Type: string(ParameterTypeText),
						Text: "John",
					},
				},
			},
		},
	}

	result := validator.ValidateTemplateForSending(validMessage)
	if !result.Valid {
		t.Errorf("Expected valid template message, got errors: %v", result.Errors)
	}

	// Test invalid template message (missing name)
	invalidMessage := &TemplateMessagePayload{
		Language: TemplateLanguage{
			Code: "en_US",
		},
	}

	result = validator.ValidateTemplateForSending(invalidMessage)
	if result.Valid {
		t.Error("Expected invalid template message due to missing name")
	}
}

func TestIsValidTemplate(t *testing.T) {
	// Test convenience validation function
	validTemplate := &CreateTemplateRequest{
		Name:     "valid_template",
		Language: "en_US",
		Category: CategoryUtility,
		Components: []TemplateComponent{
			{
				Type: string(ComponentTypeBody),
				Text: "This is valid",
			},
		},
	}

	if !IsValidTemplate(validTemplate) {
		t.Error("Expected template to be valid")
	}

	// Test invalid template
	invalidTemplate := &CreateTemplateRequest{
		Name:       "",
		Language:   "en_US",
		Category:   CategoryUtility,
		Components: []TemplateComponent{},
	}

	if IsValidTemplate(invalidTemplate) {
		t.Error("Expected template to be invalid")
	}
}
