package bm

// TemplateBuilder provides a fluent API for building message templates.
type TemplateBuilder struct {
	request *CreateTemplateRequest
}

// NewTemplate creates a new template builder.
func NewTemplate(name, language string, category TemplateCategory) *TemplateBuilder {
	return &TemplateBuilder{
		request: &CreateTemplateRequest{
			Name:       name,
			Language:   language,
			Category:   category,
			Components: make([]TemplateComponent, 0),
		},
	}
}

// WithCategoryChange allows automatic category changes during review.
func (b *TemplateBuilder) WithCategoryChange(allow bool) *TemplateBuilder {
	b.request.AllowCategoryChange = allow
	return b
}

// AddHeader adds a header component to the template.
func (b *TemplateBuilder) AddHeader(format TemplateFormat, text string) *TemplateBuilder {
	component := TemplateComponent{
		Type:   string(ComponentTypeHeader),
		Format: string(format),
		Text:   text,
	}
	b.request.Components = append(b.request.Components, component)
	return b
}

// AddHeaderWithExample adds a header component with example.
func (b *TemplateBuilder) AddHeaderWithExample(format TemplateFormat, text string, example []string) *TemplateBuilder {
	component := TemplateComponent{
		Type:   string(ComponentTypeHeader),
		Format: string(format),
		Text:   text,
		Example: &TemplateExample{
			HeaderText: example,
		},
	}
	b.request.Components = append(b.request.Components, component)
	return b
}

// AddBody adds a body component to the template.
func (b *TemplateBuilder) AddBody(text string) *TemplateBuilder {
	component := TemplateComponent{
		Type: string(ComponentTypeBody),
		Text: text,
	}
	b.request.Components = append(b.request.Components, component)
	return b
}

// AddBodyWithExample adds a body component with example.
func (b *TemplateBuilder) AddBodyWithExample(text string, examples [][]string) *TemplateBuilder {
	component := TemplateComponent{
		Type: string(ComponentTypeBody),
		Text: text,
		Example: &TemplateExample{
			BodyText: examples,
		},
	}
	b.request.Components = append(b.request.Components, component)
	return b
}

// AddFooter adds a footer component to the template.
func (b *TemplateBuilder) AddFooter(text string) *TemplateBuilder {
	component := TemplateComponent{
		Type: string(ComponentTypeFooter),
		Text: text,
	}
	b.request.Components = append(b.request.Components, component)
	return b
}

// AddButtons adds a buttons component to the template.
func (b *TemplateBuilder) AddButtons(buttons ...TemplateButton) *TemplateBuilder {
	component := TemplateComponent{
		Type:    string(ComponentTypeButtons),
		Buttons: buttons,
	}
	b.request.Components = append(b.request.Components, component)
	return b
}

// Build returns the completed template request.
func (b *TemplateBuilder) Build() *CreateTemplateRequest {
	return b.request
}

// Validate validates the template and returns any errors.
func (b *TemplateBuilder) Validate() *TemplateValidationResult {
	validator := NewTemplateValidator()
	return validator.ValidateTemplate(b.request)
}

// ButtonBuilder provides a fluent API for building template buttons.
type ButtonBuilder struct {
	buttons []TemplateButton
}

// NewButtons creates a new button builder.
func NewButtons() *ButtonBuilder {
	return &ButtonBuilder{
		buttons: make([]TemplateButton, 0),
	}
}

// AddQuickReply adds a quick reply button.
func (b *ButtonBuilder) AddQuickReply(text string) *ButtonBuilder {
	button := TemplateButton{
		Type: string(ButtonTypeQuickReply),
		Text: text,
	}
	b.buttons = append(b.buttons, button)
	return b
}

// AddURL adds a URL button.
func (b *ButtonBuilder) AddURL(text, url string) *ButtonBuilder {
	button := TemplateButton{
		Type: string(ButtonTypeURL),
		Text: text,
		URL:  url,
	}
	b.buttons = append(b.buttons, button)
	return b
}

// AddURLWithExample adds a URL button with dynamic URL example.
func (b *ButtonBuilder) AddURLWithExample(text, url string, example []string) *ButtonBuilder {
	button := TemplateButton{
		Type:    string(ButtonTypeURL),
		Text:    text,
		URL:     url,
		Example: example,
	}
	b.buttons = append(b.buttons, button)
	return b
}

// AddPhoneNumber adds a phone number button.
func (b *ButtonBuilder) AddPhoneNumber(text, phoneNumber string) *ButtonBuilder {
	button := TemplateButton{
		Type:        string(ButtonTypePhoneNumber),
		Text:        text,
		PhoneNumber: phoneNumber,
	}
	b.buttons = append(b.buttons, button)
	return b
}

// Build returns the completed buttons array.
func (b *ButtonBuilder) Build() []TemplateButton {
	return b.buttons
}

// TemplateMessageBuilder provides a fluent API for building template messages for sending.
type TemplateMessageBuilder struct {
	payload *TemplateMessagePayload
}

// NewTemplateMessage creates a new template message builder.
func NewTemplateMessage(name, languageCode string) *TemplateMessageBuilder {
	return &TemplateMessageBuilder{
		payload: &TemplateMessagePayload{
			Name: name,
			Language: TemplateLanguage{
				Code: languageCode,
			},
			Components: make([]TemplateComponentParam, 0),
		},
	}
}

// AddBodyParameters adds body parameters for template variables.
func (b *TemplateMessageBuilder) AddBodyParameters(parameters ...string) *TemplateMessageBuilder {
	var params []TemplateParameter
	for _, param := range parameters {
		params = append(params, TemplateParameter{
			Type: string(ParameterTypeText),
			Text: param,
		})
	}

	component := TemplateComponentParam{
		Type:       ComponentTypeBody,
		Parameters: params,
	}
	b.payload.Components = append(b.payload.Components, component)
	return b
}

// AddHeaderParameter adds a header parameter for template variables.
func (b *TemplateMessageBuilder) AddHeaderParameter(paramType TemplateParameterType, value string) *TemplateMessageBuilder {
	param := TemplateParameter{
		Type: string(paramType),
		Text: value,
	}

	component := TemplateComponentParam{
		Type:       ComponentTypeHeader,
		Parameters: []TemplateParameter{param},
	}
	b.payload.Components = append(b.payload.Components, component)
	return b
}

// AddButtonParameter adds button parameters for dynamic buttons.
func (b *TemplateMessageBuilder) AddButtonParameter(index int, paramType TemplateParameterType, value string) *TemplateMessageBuilder {
	param := TemplateParameter{
		Type: string(paramType),
		Text: value,
	}

	component := TemplateComponentParam{
		Type:       ComponentTypeButtons,
		SubType:    "url", // For URL buttons with dynamic parameters
		Index:      index,
		Parameters: []TemplateParameter{param},
	}
	b.payload.Components = append(b.payload.Components, component)
	return b
}

// Build returns the completed template message payload.
func (b *TemplateMessageBuilder) Build() *TemplateMessagePayload {
	return b.payload
}

// Validate validates the template message and returns any errors.
func (b *TemplateMessageBuilder) Validate() *TemplateValidationResult {
	validator := NewTemplateValidator()
	return validator.ValidateTemplateForSending(b.payload)
}

// Convenience functions for common template patterns

// NewMarketingTemplate creates a marketing template with common structure.
func NewMarketingTemplate(name, language, headerText, bodyText, footerText string) *TemplateBuilder {
	return NewTemplate(name, language, CategoryMarketing).
		AddHeader(FormatText, headerText).
		AddBody(bodyText).
		AddFooter(footerText)
}

// NewUtilityTemplate creates a utility template with common structure.
func NewUtilityTemplate(name, language, bodyText string) *TemplateBuilder {
	return NewTemplate(name, language, CategoryUtility).
		AddBody(bodyText)
}

// NewAuthenticationTemplate creates an authentication template with common structure.
func NewAuthenticationTemplate(name, language, bodyText string) *TemplateBuilder {
	return NewTemplate(name, language, CategoryAuthentication).
		AddBody(bodyText)
}

// NewPromoTemplate creates a promotional template with buttons.
func NewPromoTemplate(name, language, headerText, bodyText, footerText, buttonText, buttonURL string) *TemplateBuilder {
	buttons := NewButtons().
		AddURL(buttonText, buttonURL).
		AddQuickReply("Not interested").
		Build()

	return NewTemplate(name, language, CategoryMarketing).
		AddHeader(FormatText, headerText).
		AddBody(bodyText).
		AddFooter(footerText).
		AddButtons(buttons...)
}

// NewOrderConfirmationTemplate creates an order confirmation template.
func NewOrderConfirmationTemplate(name, language string) *TemplateBuilder {
	return NewTemplate(name, language, CategoryUtility).
		AddHeader(FormatText, "Order Confirmation").
		AddBody("Hi {{1}}, your order #{{2}} has been confirmed and will be delivered by {{3}}.").
		AddFooter("Thank you for your business!")
}

// NewAppointmentReminderTemplate creates an appointment reminder template.
func NewAppointmentReminderTemplate(name, language string) *TemplateBuilder {
	buttons := NewButtons().
		AddQuickReply("Confirm").
		AddQuickReply("Reschedule").
		AddPhoneNumber("Call Us", "+1234567890").
		Build()

	return NewTemplate(name, language, CategoryUtility).
		AddBody("Hi {{1}}, this is a reminder for your appointment on {{2}} at {{3}}.").
		AddButtons(buttons...)
}
