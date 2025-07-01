package flows

import (
	"fmt"
	"regexp"
	"strings"
)

// FlowValidator provides comprehensive validation for Flow definitions.
type FlowValidator struct {
	errors []ValidationError
}

// NewFlowValidator creates a new Flow validator.
func NewFlowValidator() *FlowValidator {
	return &FlowValidator{
		errors: make([]ValidationError, 0),
	}
}

// ValidateFlow performs comprehensive validation on a Flow definition.
func (v *FlowValidator) ValidateFlow(flow *Flow) []ValidationError {
	v.errors = make([]ValidationError, 0)

	// Validate basic Flow structure
	v.validateFlowStructure(flow)

	// Validate screens
	v.validateScreens(flow.Screens)

	// Validate routing model
	v.validateRoutingModel(flow.RoutingModel, flow.Screens)

	return v.errors
}

// validateFlowStructure validates the basic Flow structure.
func (v *FlowValidator) validateFlowStructure(flow *Flow) {
	if flow.Version == "" {
		v.addError("Flow version is required", 0, 0)
	} else if !isValidVersion(flow.Version) {
		v.addError("Invalid Flow version format", 0, 0)
	}

	if flow.DataAPIVersion == "" {
		v.addError("Data API version is required", 0, 0)
	} else if !isValidDataAPIVersion(flow.DataAPIVersion) {
		v.addError("Invalid data API version format", 0, 0)
	}

	if len(flow.Screens) == 0 {
		v.addError("Flow must have at least one screen", 0, 0)
	}
}

// validateScreens validates all screens in the Flow.
func (v *FlowValidator) validateScreens(screens []Screen) {
	screenIDs := make(map[string]bool)
	hasTerminalScreen := false

	for i, screen := range screens {
		// Check for duplicate screen IDs
		if screenIDs[screen.ID] {
			v.addError(fmt.Sprintf("Duplicate screen ID: %s", screen.ID), i+1, 0)
		}
		screenIDs[screen.ID] = true

		// Validate individual screen
		v.validateScreen(screen, i+1)

		// Check for terminal screens
		if screen.Terminal {
			hasTerminalScreen = true
		}
	}

	if !hasTerminalScreen {
		v.addError("Flow must have at least one terminal screen", 0, 0)
	}
}

// validateScreen validates an individual screen.
func (v *FlowValidator) validateScreen(screen Screen, lineNumber int) {
	if screen.ID == "" {
		v.addError("Screen ID is required", lineNumber, 0)
	} else if !isValidScreenID(screen.ID) {
		v.addError("Invalid screen ID format", lineNumber, 0)
	}

	if screen.Layout.Type == "" {
		v.addError("Screen layout type is required", lineNumber, 0)
	} else if !isValidLayoutType(screen.Layout.Type) {
		v.addError("Invalid layout type", lineNumber, 0)
	}

	// Validate data fields
	v.validateDataFields(screen.Data, lineNumber)

	// Validate components
	v.validateComponents(screen.Layout.Children, lineNumber)
}

// validateDataFields validates screen data fields.
func (v *FlowValidator) validateDataFields(data map[string]DataField, lineNumber int) {
	for name, field := range data {
		if !isValidDataFieldName(name) {
			v.addError(fmt.Sprintf("Invalid data field name: %s", name), lineNumber, 0)
		}

		if !isValidDataFieldType(field.Type) {
			v.addError(fmt.Sprintf("Invalid data field type: %s", field.Type), lineNumber, 0)
		}
	}
}

// validateComponents validates screen components.
func (v *FlowValidator) validateComponents(components []Component, lineNumber int) {
	componentNames := make(map[string]bool)
	hasFooter := false

	for _, component := range components {
		// Validate component structure
		if err := ValidateComponent(component); err != nil {
			v.addError(err.Error(), lineNumber, 0)
		}

		// Check for duplicate component names
		if component.Name != "" {
			if componentNames[component.Name] {
				v.addError(fmt.Sprintf("Duplicate component name: %s", component.Name), lineNumber, 0)
			}
			componentNames[component.Name] = true
		}

		// Check for footer component
		if component.Type == ComponentTypeFooter {
			if hasFooter {
				v.addError("Screen can have only one footer component", lineNumber, 0)
			}
			hasFooter = true
		}

		// Validate component-specific rules
		v.validateComponentSpecific(component, lineNumber)
	}
}

// validateComponentSpecific validates component-specific rules.
func (v *FlowValidator) validateComponentSpecific(component Component, lineNumber int) {
	switch component.Type {
	case ComponentTypeTextInput, ComponentTypeTextArea:
		if component.InputType != "" && !isValidInputType(component.InputType) {
			v.addError(fmt.Sprintf("Invalid input type: %s", component.InputType), lineNumber, 0)
		}

	case ComponentTypeCheckboxGroup:
		if len(component.DataSource) == 0 {
			v.addError("Checkbox group must have at least one option", lineNumber, 0)
		}
		if component.MaxSelectedItems > 0 && component.MaxSelectedItems < component.MinSelectedItems {
			v.addError("Max selected items cannot be less than min selected items", lineNumber, 0)
		}

	case ComponentTypeRadioButtonsGroup, ComponentTypeDropdown:
		if len(component.DataSource) == 0 {
			v.addError(fmt.Sprintf("%s must have at least one option", component.Type), lineNumber, 0)
		}

	case ComponentTypeFooter:
		if component.OnClickAction == nil {
			v.addError("Footer component must have an on-click action", lineNumber, 0)
		}

	case ComponentTypeOptIn:
		if component.Text == "" {
			v.addError("OptIn component must have text", lineNumber, 0)
		}
	}
}

// validateRoutingModel validates the routing model.
func (v *FlowValidator) validateRoutingModel(routingModel map[string][]string, screens []Screen) {
	screenIDs := make(map[string]bool)
	for _, screen := range screens {
		screenIDs[screen.ID] = true
	}

	for from, toList := range routingModel {
		if !screenIDs[from] {
			v.addError(fmt.Sprintf("Routing model references non-existent screen: %s", from), 0, 0)
		}

		for _, to := range toList {
			if !screenIDs[to] {
				v.addError(fmt.Sprintf("Routing model references non-existent screen: %s", to), 0, 0)
			}
		}
	}
}

// addError adds a validation error.
func (v *FlowValidator) addError(message string, line, column int) {
	v.errors = append(v.errors, ValidationError{
		Message: message,
		Line:    line,
		Column:  column,
	})
}

// Validation helper functions

// isValidVersion checks if the Flow version is valid.
func isValidVersion(version string) bool {
	// Version should be in format "X.Y" where X and Y are numbers
	matched, _ := regexp.MatchString(`^\d+\.\d+$`, version)
	return matched
}

// isValidDataAPIVersion checks if the data API version is valid.
func isValidDataAPIVersion(version string) bool {
	// Data API version should be in format "X.Y" where X and Y are numbers
	matched, _ := regexp.MatchString(`^\d+\.\d+$`, version)
	return matched
}

// isValidScreenID checks if the screen ID is valid.
func isValidScreenID(id string) bool {
	// Screen ID should be alphanumeric with underscores, no spaces
	matched, _ := regexp.MatchString(`^[A-Z][A-Z0-9_]*$`, id)
	return matched
}

// isValidLayoutType checks if the layout type is valid.
func isValidLayoutType(layoutType string) bool {
	validTypes := map[string]bool{
		LayoutTypeSingleColumn: true,
	}
	return validTypes[layoutType]
}

// isValidDataFieldName checks if the data field name is valid.
func isValidDataFieldName(name string) bool {
	// Data field name should be alphanumeric with underscores, no spaces
	matched, _ := regexp.MatchString(`^[a-z][a-z0-9_]*$`, name)
	return matched
}

// isValidDataFieldType checks if the data field type is valid.
func isValidDataFieldType(fieldType string) bool {
	validTypes := map[string]bool{
		"string":  true,
		"number":  true,
		"boolean": true,
		"array":   true,
		"object":  true,
	}
	return validTypes[fieldType]
}

// isValidInputType checks if the input type is valid.
func isValidInputType(inputType string) bool {
	validTypes := map[string]bool{
		InputTypeText:     true,
		InputTypeEmail:    true,
		InputTypeNumber:   true,
		InputTypePassword: true,
		InputTypePhone:    true,
	}
	return validTypes[inputType]
}

// ValidateFlowJSON validates a Flow JSON string.
func ValidateFlowJSON(jsonStr string) ([]ValidationError, error) {
	flow, err := FromJSON(jsonStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	validator := NewFlowValidator()
	return validator.ValidateFlow(flow), nil
}

// IsValidFlow checks if a Flow is valid (has no validation errors).
func IsValidFlow(flow *Flow) bool {
	validator := NewFlowValidator()
	errors := validator.ValidateFlow(flow)
	return len(errors) == 0
}

// GetValidationSummary returns a summary of validation errors.
func GetValidationSummary(errors []ValidationError) string {
	if len(errors) == 0 {
		return "Flow validation passed"
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Flow validation failed with %d error(s):\n", len(errors)))

	for i, err := range errors {
		if err.Line > 0 {
			summary.WriteString(fmt.Sprintf("%d. Line %d: %s\n", i+1, err.Line, err.Message))
		} else {
			summary.WriteString(fmt.Sprintf("%d. %s\n", i+1, err.Message))
		}
	}

	return summary.String()
}
