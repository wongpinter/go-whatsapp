package bm

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// TemplateValidator provides validation for message templates.
type TemplateValidator struct {
	// Configuration options can be added here
}

// NewTemplateValidator creates a new template validator.
func NewTemplateValidator() *TemplateValidator {
	return &TemplateValidator{}
}

// ValidateTemplate validates a complete template request.
func (v *TemplateValidator) ValidateTemplate(request *CreateTemplateRequest) *TemplateValidationResult {
	var errors []TemplateValidationError

	// Validate template name
	if nameErrors := v.validateTemplateName(request.Name); len(nameErrors) > 0 {
		errors = append(errors, nameErrors...)
	}

	// Validate language
	if langErrors := v.validateLanguage(request.Language); len(langErrors) > 0 {
		errors = append(errors, langErrors...)
	}

	// Validate category
	if catErrors := v.validateCategory(request.Category); len(catErrors) > 0 {
		errors = append(errors, catErrors...)
	}

	// Validate components
	if compErrors := v.validateComponents(request.Components); len(compErrors) > 0 {
		errors = append(errors, compErrors...)
	}

	return &TemplateValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// validateTemplateName validates the template name.
func (v *TemplateValidator) validateTemplateName(name string) []TemplateValidationError {
	var errors []TemplateValidationError

	if name == "" {
		errors = append(errors, TemplateValidationError{
			Field:   "name",
			Message: "Template name is required",
			Code:    "REQUIRED_FIELD",
		})
		return errors
	}

	// Check length (max 512 characters)
	if utf8.RuneCountInString(name) > 512 {
		errors = append(errors, TemplateValidationError{
			Field:   "name",
			Message: "Template name must be 512 characters or less",
			Code:    "MAX_LENGTH_EXCEEDED",
		})
	}

	// Check format (lowercase, underscores, numbers only)
	namePattern := regexp.MustCompile(`^[a-z0-9_]+$`)
	if !namePattern.MatchString(name) {
		errors = append(errors, TemplateValidationError{
			Field:   "name",
			Message: "Template name can only contain lowercase letters, numbers, and underscores",
			Code:    "INVALID_FORMAT",
		})
	}

	return errors
}

// validateLanguage validates the language code.
func (v *TemplateValidator) validateLanguage(language string) []TemplateValidationError {
	var errors []TemplateValidationError

	if language == "" {
		errors = append(errors, TemplateValidationError{
			Field:   "language",
			Message: "Language is required",
			Code:    "REQUIRED_FIELD",
		})
		return errors
	}

	// Basic language code validation (e.g., en_US, es_ES)
	langPattern := regexp.MustCompile(`^[a-z]{2}_[A-Z]{2}$`)
	if !langPattern.MatchString(language) {
		errors = append(errors, TemplateValidationError{
			Field:   "language",
			Message: "Language must be in format 'xx_XX' (e.g., en_US, es_ES)",
			Code:    "INVALID_FORMAT",
		})
	}

	return errors
}

// validateCategory validates the template category.
func (v *TemplateValidator) validateCategory(category TemplateCategory) []TemplateValidationError {
	var errors []TemplateValidationError

	validCategories := map[TemplateCategory]bool{
		CategoryMarketing:      true,
		CategoryUtility:        true,
		CategoryAuthentication: true,
	}

	if !validCategories[category] {
		errors = append(errors, TemplateValidationError{
			Field:   "category",
			Message: fmt.Sprintf("Invalid category '%s'. Must be one of: MARKETING, UTILITY, AUTHENTICATION", category),
			Code:    "INVALID_VALUE",
		})
	}

	return errors
}

// validateComponents validates template components.
func (v *TemplateValidator) validateComponents(components []TemplateComponent) []TemplateValidationError {
	var errors []TemplateValidationError

	if len(components) == 0 {
		errors = append(errors, TemplateValidationError{
			Field:   "components",
			Message: "At least one component is required",
			Code:    "REQUIRED_FIELD",
		})
		return errors
	}

	// Check for required BODY component
	hasBody := false
	headerCount := 0
	footerCount := 0
	buttonCount := 0

	for i, component := range components {
		fieldPrefix := fmt.Sprintf("components[%d]", i)

		// Validate component type
		if compErrors := v.validateComponentType(component, fieldPrefix); len(compErrors) > 0 {
			errors = append(errors, compErrors...)
		}

		// Count component types
		switch TemplateComponentType(component.Type) {
		case ComponentTypeBody:
			hasBody = true
		case ComponentTypeHeader:
			headerCount++
		case ComponentTypeFooter:
			footerCount++
		case ComponentTypeButtons:
			buttonCount++
		}

		// Validate specific component
		if compErrors := v.validateComponent(component, fieldPrefix); len(compErrors) > 0 {
			errors = append(errors, compErrors...)
		}
	}

	// Validate component structure rules
	if !hasBody {
		errors = append(errors, TemplateValidationError{
			Field:   "components",
			Message: "Template must have exactly one BODY component",
			Code:    "MISSING_REQUIRED_COMPONENT",
		})
	}

	if headerCount > 1 {
		errors = append(errors, TemplateValidationError{
			Field:   "components",
			Message: "Template can have at most one HEADER component",
			Code:    "TOO_MANY_COMPONENTS",
		})
	}

	if footerCount > 1 {
		errors = append(errors, TemplateValidationError{
			Field:   "components",
			Message: "Template can have at most one FOOTER component",
			Code:    "TOO_MANY_COMPONENTS",
		})
	}

	if buttonCount > 1 {
		errors = append(errors, TemplateValidationError{
			Field:   "components",
			Message: "Template can have at most one BUTTONS component",
			Code:    "TOO_MANY_COMPONENTS",
		})
	}

	return errors
}

// validateComponentType validates the component type.
func (v *TemplateValidator) validateComponentType(component TemplateComponent, fieldPrefix string) []TemplateValidationError {
	var errors []TemplateValidationError

	validTypes := map[string]bool{
		string(ComponentTypeHeader):  true,
		string(ComponentTypeBody):    true,
		string(ComponentTypeFooter):  true,
		string(ComponentTypeButtons): true,
	}

	if !validTypes[component.Type] {
		errors = append(errors, TemplateValidationError{
			Field:   fieldPrefix + ".type",
			Message: fmt.Sprintf("Invalid component type '%s'", component.Type),
			Code:    "INVALID_VALUE",
		})
	}

	return errors
}

// validateComponent validates a specific component based on its type.
func (v *TemplateValidator) validateComponent(component TemplateComponent, fieldPrefix string) []TemplateValidationError {
	var errors []TemplateValidationError

	switch TemplateComponentType(component.Type) {
	case ComponentTypeHeader:
		errors = append(errors, v.validateHeaderComponent(component, fieldPrefix)...)
	case ComponentTypeBody:
		errors = append(errors, v.validateBodyComponent(component, fieldPrefix)...)
	case ComponentTypeFooter:
		errors = append(errors, v.validateFooterComponent(component, fieldPrefix)...)
	case ComponentTypeButtons:
		errors = append(errors, v.validateButtonsComponent(component, fieldPrefix)...)
	}

	return errors
}

// validateHeaderComponent validates a header component.
func (v *TemplateValidator) validateHeaderComponent(component TemplateComponent, fieldPrefix string) []TemplateValidationError {
	var errors []TemplateValidationError

	// Header can be TEXT, IMAGE, VIDEO, DOCUMENT, or LOCATION
	validFormats := map[string]bool{
		string(FormatText):     true,
		string(FormatImage):    true,
		string(FormatVideo):    true,
		string(FormatDocument): true,
		string(FormatLocation): true,
	}

	if component.Format != "" && !validFormats[component.Format] {
		errors = append(errors, TemplateValidationError{
			Field:   fieldPrefix + ".format",
			Message: fmt.Sprintf("Invalid header format '%s'", component.Format),
			Code:    "INVALID_VALUE",
		})
	}

	// For TEXT headers, validate text content
	if component.Format == string(FormatText) || component.Format == "" {
		if component.Text == "" {
			errors = append(errors, TemplateValidationError{
				Field:   fieldPrefix + ".text",
				Message: "Header text is required for TEXT format",
				Code:    "REQUIRED_FIELD",
			})
		} else if utf8.RuneCountInString(component.Text) > 60 {
			errors = append(errors, TemplateValidationError{
				Field:   fieldPrefix + ".text",
				Message: "Header text must be 60 characters or less",
				Code:    "MAX_LENGTH_EXCEEDED",
			})
		}

		// Validate placeholders (max 1 for header)
		placeholderCount := strings.Count(component.Text, "{{")
		if placeholderCount > 1 {
			errors = append(errors, TemplateValidationError{
				Field:   fieldPrefix + ".text",
				Message: "Header can have at most 1 placeholder",
				Code:    "TOO_MANY_PLACEHOLDERS",
			})
		}
	}

	return errors
}

// validateFooterComponent validates a footer component.
func (v *TemplateValidator) validateFooterComponent(component TemplateComponent, fieldPrefix string) []TemplateValidationError {
	var errors []TemplateValidationError

	if component.Text == "" {
		errors = append(errors, TemplateValidationError{
			Field:   fieldPrefix + ".text",
			Message: "Footer text is required",
			Code:    "REQUIRED_FIELD",
		})
		return errors
	}

	// Check length (max 60 characters)
	if utf8.RuneCountInString(component.Text) > 60 {
		errors = append(errors, TemplateValidationError{
			Field:   fieldPrefix + ".text",
			Message: "Footer text must be 60 characters or less",
			Code:    "MAX_LENGTH_EXCEEDED",
		})
	}

	// Footer cannot have placeholders
	if strings.Contains(component.Text, "{{") {
		errors = append(errors, TemplateValidationError{
			Field:   fieldPrefix + ".text",
			Message: "Footer cannot contain placeholders",
			Code:    "PLACEHOLDERS_NOT_ALLOWED",
		})
	}

	return errors
}

// validateButtonsComponent validates a buttons component.
func (v *TemplateValidator) validateButtonsComponent(component TemplateComponent, fieldPrefix string) []TemplateValidationError {
	var errors []TemplateValidationError

	if len(component.Buttons) == 0 {
		errors = append(errors, TemplateValidationError{
			Field:   fieldPrefix + ".buttons",
			Message: "At least one button is required",
			Code:    "REQUIRED_FIELD",
		})
		return errors
	}

	// Check maximum button count (10 total)
	if len(component.Buttons) > 10 {
		errors = append(errors, TemplateValidationError{
			Field:   fieldPrefix + ".buttons",
			Message: "Maximum 10 buttons allowed",
			Code:    "TOO_MANY_BUTTONS",
		})
	}

	// Count button types
	quickReplyCount := 0
	urlCount := 0
	phoneCount := 0

	for i, button := range component.Buttons {
		buttonPrefix := fmt.Sprintf("%s.buttons[%d]", fieldPrefix, i)

		// Validate button type
		switch TemplateButtonType(button.Type) {
		case ButtonTypeQuickReply:
			quickReplyCount++
		case ButtonTypeURL:
			urlCount++
		case ButtonTypePhoneNumber:
			phoneCount++
		default:
			errors = append(errors, TemplateValidationError{
				Field:   buttonPrefix + ".type",
				Message: fmt.Sprintf("Invalid button type '%s'", button.Type),
				Code:    "INVALID_VALUE",
			})
		}

		// Validate button-specific fields
		if buttonErrors := v.validateButton(button, buttonPrefix); len(buttonErrors) > 0 {
			errors = append(errors, buttonErrors...)
		}
	}

	// Validate button type limits
	if quickReplyCount > 10 {
		errors = append(errors, TemplateValidationError{
			Field:   fieldPrefix + ".buttons",
			Message: "Maximum 10 quick reply buttons allowed",
			Code:    "TOO_MANY_QUICK_REPLY_BUTTONS",
		})
	}

	if urlCount > 2 {
		errors = append(errors, TemplateValidationError{
			Field:   fieldPrefix + ".buttons",
			Message: "Maximum 2 URL buttons allowed",
			Code:    "TOO_MANY_URL_BUTTONS",
		})
	}

	if phoneCount > 1 {
		errors = append(errors, TemplateValidationError{
			Field:   fieldPrefix + ".buttons",
			Message: "Maximum 1 phone number button allowed",
			Code:    "TOO_MANY_PHONE_BUTTONS",
		})
	}

	return errors
}

// validateButton validates a specific button.
func (v *TemplateValidator) validateButton(button TemplateButton, fieldPrefix string) []TemplateValidationError {
	var errors []TemplateValidationError

	// Validate button text
	if button.Text == "" {
		errors = append(errors, TemplateValidationError{
			Field:   fieldPrefix + ".text",
			Message: "Button text is required",
			Code:    "REQUIRED_FIELD",
		})
	} else if utf8.RuneCountInString(button.Text) > 25 {
		errors = append(errors, TemplateValidationError{
			Field:   fieldPrefix + ".text",
			Message: "Button text must be 25 characters or less",
			Code:    "MAX_LENGTH_EXCEEDED",
		})
	}

	// Validate type-specific fields
	switch TemplateButtonType(button.Type) {
	case ButtonTypeURL:
		if button.URL == "" {
			errors = append(errors, TemplateValidationError{
				Field:   fieldPrefix + ".url",
				Message: "URL is required for URL buttons",
				Code:    "REQUIRED_FIELD",
			})
		} else {
			// Basic URL validation
			if !strings.HasPrefix(button.URL, "http://") && !strings.HasPrefix(button.URL, "https://") {
				errors = append(errors, TemplateValidationError{
					Field:   fieldPrefix + ".url",
					Message: "URL must start with http:// or https://",
					Code:    "INVALID_FORMAT",
				})
			}
		}

	case ButtonTypePhoneNumber:
		if button.PhoneNumber == "" {
			errors = append(errors, TemplateValidationError{
				Field:   fieldPrefix + ".phone_number",
				Message: "Phone number is required for phone number buttons",
				Code:    "REQUIRED_FIELD",
			})
		} else {
			// Basic phone number validation (starts with +)
			if !strings.HasPrefix(button.PhoneNumber, "+") {
				errors = append(errors, TemplateValidationError{
					Field:   fieldPrefix + ".phone_number",
					Message: "Phone number must start with + and country code",
					Code:    "INVALID_FORMAT",
				})
			}
		}
	}

	return errors
}

// ValidateTemplateForSending validates a template message payload for sending.
func (v *TemplateValidator) ValidateTemplateForSending(payload *TemplateMessagePayload) *TemplateValidationResult {
	var errors []TemplateValidationError

	// Validate template name
	if payload.Name == "" {
		errors = append(errors, TemplateValidationError{
			Field:   "template.name",
			Message: "Template name is required",
			Code:    "REQUIRED_FIELD",
		})
	}

	// Validate language
	if payload.Language.Code == "" {
		errors = append(errors, TemplateValidationError{
			Field:   "template.language.code",
			Message: "Language code is required",
			Code:    "REQUIRED_FIELD",
		})
	}

	// Validate components if provided
	for i, component := range payload.Components {
		fieldPrefix := fmt.Sprintf("template.components[%d]", i)

		// Validate component type
		validTypes := map[string]bool{
			string(ComponentTypeHeader):  true,
			string(ComponentTypeBody):    true,
			string(ComponentTypeButtons): true,
		}

		if !validTypes[string(component.Type)] {
			errors = append(errors, TemplateValidationError{
				Field:   fieldPrefix + ".type",
				Message: fmt.Sprintf("Invalid component type '%s' for sending", component.Type),
				Code:    "INVALID_VALUE",
			})
		}

		// Validate parameters
		for j, param := range component.Parameters {
			paramPrefix := fmt.Sprintf("%s.parameters[%d]", fieldPrefix, j)

			if param.Type == "" {
				errors = append(errors, TemplateValidationError{
					Field:   paramPrefix + ".type",
					Message: "Parameter type is required",
					Code:    "REQUIRED_FIELD",
				})
			}

			// For text parameters, validate text content
			if param.Type == string(ParameterTypeText) && param.Text == "" {
				errors = append(errors, TemplateValidationError{
					Field:   paramPrefix + ".text",
					Message: "Text is required for text parameters",
					Code:    "REQUIRED_FIELD",
				})
			}
		}
	}

	return &TemplateValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// IsValidTemplate is a convenience function to check if a template is valid.
func IsValidTemplate(request *CreateTemplateRequest) bool {
	validator := NewTemplateValidator()
	result := validator.ValidateTemplate(request)
	return result.Valid
}

// IsValidTemplateForSending is a convenience function to check if a template payload is valid for sending.
func IsValidTemplateForSending(payload *TemplateMessagePayload) bool {
	validator := NewTemplateValidator()
	result := validator.ValidateTemplateForSending(payload)
	return result.Valid
}

// validateBodyComponent validates a body component.
func (v *TemplateValidator) validateBodyComponent(component TemplateComponent, fieldPrefix string) []TemplateValidationError {
	var errors []TemplateValidationError

	if component.Text == "" {
		errors = append(errors, TemplateValidationError{
			Field:   fieldPrefix + ".text",
			Message: "Body text is required",
			Code:    "REQUIRED_FIELD",
		})
		return errors
	}

	// Check length (max 1024 characters)
	if utf8.RuneCountInString(component.Text) > 1024 {
		errors = append(errors, TemplateValidationError{
			Field:   fieldPrefix + ".text",
			Message: "Body text must be 1024 characters or less",
			Code:    "MAX_LENGTH_EXCEEDED",
		})
	}

	// Validate placeholders and examples
	placeholderCount := strings.Count(component.Text, "{{")
	if placeholderCount > 0 {
		if component.Example == nil || len(component.Example.BodyText) == 0 {
			errors = append(errors, TemplateValidationError{
				Field:   fieldPrefix + ".example",
				Message: "Example is required when using placeholders",
				Code:    "REQUIRED_FIELD",
			})
		} else {
			// Validate example matches placeholder count
			for i, exampleSet := range component.Example.BodyText {
				if len(exampleSet) != placeholderCount {
					errors = append(errors, TemplateValidationError{
						Field:   fmt.Sprintf("%s.example.body_text[%d]", fieldPrefix, i),
						Message: fmt.Sprintf("Example must have %d values to match %d placeholders", placeholderCount, placeholderCount),
						Code:    "EXAMPLE_MISMATCH",
					})
				}
			}
		}
	}

	return errors
}
