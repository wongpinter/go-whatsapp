// Package flows provides functionality for building and managing WhatsApp Flows.
package flows

import (
	"encoding/json"
)

// Flow represents a complete WhatsApp Flow definition following the Flow JSON schema.
type Flow struct {
	Version        string              `json:"version"`
	DataAPIVersion string              `json:"data_api_version"`
	RoutingModel   map[string][]string `json:"routing_model,omitempty"`
	Screens        []Screen            `json:"screens"`
}

// Screen represents a screen in a WhatsApp Flow.
type Screen struct {
	ID            string               `json:"id"`
	Title         string               `json:"title,omitempty"`
	Terminal      bool                 `json:"terminal,omitempty"`
	Success       bool                 `json:"success,omitempty"`
	Data          map[string]DataField `json:"data,omitempty"`
	Layout        Layout               `json:"layout"`
	RefreshOnBack bool                 `json:"refresh_on_back,omitempty"`
}

// DataField represents a data field definition in a screen.
type DataField struct {
	Type     string      `json:"type"`
	Example  interface{} `json:"__example__,omitempty"`
	Required bool        `json:"required,omitempty"`
}

// Layout represents the layout of a screen.
type Layout struct {
	Type     string      `json:"type"`
	Children []Component `json:"children"`
}

// Component represents a UI component in a Flow screen.
type Component struct {
	Type              string           `json:"type"`
	Text              string           `json:"text,omitempty"`
	Label             string           `json:"label,omitempty"`
	Name              string           `json:"name,omitempty"`
	InputType         string           `json:"input-type,omitempty"`
	Required          bool             `json:"required,omitempty"`
	Enabled           bool             `json:"enabled,omitempty"`
	Visible           bool             `json:"visible,omitempty"`
	OnClickAction     *Action          `json:"on-click-action,omitempty"`
	OnSelectionChange *Action          `json:"on-selection-change,omitempty"`
	Children          []Component      `json:"children,omitempty"`
	DataSource        []DataSourceItem `json:"data-source,omitempty"`
	MinSelectedItems  int              `json:"min-selected-items,omitempty"`
	MaxSelectedItems  int              `json:"max-selected-items,omitempty"`
	InitValue         interface{}      `json:"init-value,omitempty"`
	HelperText        string           `json:"helper-text,omitempty"`
	ErrorMessage      string           `json:"error-message,omitempty"`
}

// Action represents an action that can be triggered in a Flow.
type Action struct {
	Name    string                 `json:"name"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// DataSourceItem represents an item in a data source for components like dropdown.
type DataSourceItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Component types constants
const (
	ComponentTypeTextHeading       = "TextHeading"
	ComponentTypeTextSubheading    = "TextSubheading"
	ComponentTypeTextBody          = "TextBody"
	ComponentTypeTextCaption       = "TextCaption"
	ComponentTypeTextInput         = "TextInput"
	ComponentTypeTextArea          = "TextArea"
	ComponentTypeCheckboxGroup     = "CheckboxGroup"
	ComponentTypeRadioButtonsGroup = "RadioButtonsGroup"
	ComponentTypeDropdown          = "Dropdown"
	ComponentTypeFooter            = "Footer"
	ComponentTypeOptIn             = "OptIn"
	ComponentTypeEmbeddedLink      = "EmbeddedLink"
	ComponentTypeDatePicker        = "DatePicker"
	ComponentTypeImage             = "Image"
)

// Layout types constants
const (
	LayoutTypeSingleColumn = "SingleColumnLayout"
)

// Input types constants
const (
	InputTypeText     = "text"
	InputTypeEmail    = "email"
	InputTypeNumber   = "number"
	InputTypePassword = "password"
	InputTypePhone    = "phone"
)

// Action names constants
const (
	ActionNavigate     = "navigate"
	ActionDataExchange = "data_exchange"
	ActionComplete     = "complete"
)

// Flow categories constants
const (
	CategorySignUp             = "SIGN_UP"
	CategorySignIn             = "SIGN_IN"
	CategoryAppointmentBooking = "APPOINTMENT_BOOKING"
	CategoryLeadGeneration     = "LEAD_GENERATION"
	CategoryContactUs          = "CONTACT_US"
	CategoryCustomerSupport    = "CUSTOMER_SUPPORT"
	CategorySurvey             = "SURVEY"
	CategoryOther              = "OTHER"
)

// FlowMetadata represents metadata for Flow management operations.
type FlowMetadata struct {
	Name        string   `json:"name"`
	Categories  []string `json:"categories"`
	EndpointURI string   `json:"endpoint_uri,omitempty"`
}

// FlowStatus represents the status of a Flow.
type FlowStatus string

const (
	FlowStatusDraft      FlowStatus = "DRAFT"
	FlowStatusPublished  FlowStatus = "PUBLISHED"
	FlowStatusDeprecated FlowStatus = "DEPRECATED"
)

// FlowHealthStatus represents the health status of a Flow.
type FlowHealthStatus string

const (
	FlowHealthStatusHealthy   FlowHealthStatus = "HEALTHY"
	FlowHealthStatusUnhealthy FlowHealthStatus = "UNHEALTHY"
)

// FlowInfo represents Flow information returned by the Graph API.
type FlowInfo struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Status           FlowStatus        `json:"status"`
	Categories       []string          `json:"categories"`
	HealthStatus     FlowHealthStatus  `json:"health_status"`
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
	JSONVersion      string            `json:"json_version,omitempty"`
	DataAPIVersion   string            `json:"data_api_version,omitempty"`
	EndpointURI      string            `json:"endpoint_uri,omitempty"`
	PreviewURL       string            `json:"preview_url,omitempty"`
}

// ValidationError represents a Flow validation error.
type ValidationError struct {
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

// FlowAssetType represents the type of Flow asset.
type FlowAssetType string

const (
	FlowAssetTypeFlowJSON FlowAssetType = "FLOW_JSON"
)

// DataExchangeRequest represents a data exchange request from WhatsApp.
type DataExchangeRequest struct {
	Version   string                 `json:"version"`
	Action    string                 `json:"action"`
	Screen    string                 `json:"screen"`
	Data      map[string]interface{} `json:"data"`
	FlowToken string                 `json:"flow_token"`
}

// DataExchangeResponse represents a data exchange response to WhatsApp.
type DataExchangeResponse struct {
	Version      string                 `json:"version"`
	Screen       string                 `json:"screen,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	ErrorDetails *ErrorDetails          `json:"error_details,omitempty"`
}

// ErrorDetails represents error details in a data exchange response.
type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// FlowCompletion represents a Flow completion event from webhook.
type FlowCompletion struct {
	FlowToken string                 `json:"flow_token"`
	Response  map[string]interface{} `json:"response"`
}

// ToJSON converts the Flow to JSON string.
func (f *Flow) ToJSON() (string, error) {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON creates a Flow from JSON string.
func FromJSON(jsonStr string) (*Flow, error) {
	var flow Flow
	err := json.Unmarshal([]byte(jsonStr), &flow)
	if err != nil {
		return nil, err
	}
	return &flow, nil
}

// Validate performs basic validation on the Flow structure.
func (f *Flow) Validate() error {
	if f.Version == "" {
		return ErrMissingVersion
	}
	if f.DataAPIVersion == "" {
		return ErrMissingDataAPIVersion
	}
	if len(f.Screens) == 0 {
		return ErrNoScreens
	}

	// Validate screens
	for _, screen := range f.Screens {
		if err := screen.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate performs basic validation on the Screen structure.
func (s *Screen) Validate() error {
	if s.ID == "" {
		return ErrMissingScreenID
	}
	if s.Layout.Type == "" {
		return ErrMissingLayoutType
	}
	return nil
}
