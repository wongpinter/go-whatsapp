package flows

import "errors"

// Flow validation errors
var (
	ErrMissingVersion        = errors.New("flow version is required")
	ErrMissingDataAPIVersion = errors.New("flow data API version is required")
	ErrNoScreens             = errors.New("flow must have at least one screen")
	ErrMissingScreenID       = errors.New("screen ID is required")
	ErrMissingLayoutType     = errors.New("screen layout type is required")
	ErrInvalidComponentType  = errors.New("invalid component type")
	ErrMissingComponentName  = errors.New("component name is required for input components")
	ErrInvalidActionName     = errors.New("invalid action name")
	ErrMissingFlowID         = errors.New("flow ID is required")
	ErrMissingFlowToken      = errors.New("flow token is required")
	ErrInvalidFlowStatus     = errors.New("invalid flow status")
)

// Flow management errors
var (
	ErrFlowNotFound         = errors.New("flow not found")
	ErrFlowAlreadyPublished = errors.New("flow is already published")
	ErrFlowNotPublished     = errors.New("flow is not published")
	ErrInvalidFlowJSON      = errors.New("invalid flow JSON")
	ErrFlowValidationFailed = errors.New("flow validation failed")
	ErrUnauthorized         = errors.New("unauthorized access")
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")
)

// Data exchange errors
var (
	ErrInvalidDataExchangeRequest = errors.New("invalid data exchange request")
	ErrMissingAction              = errors.New("action is required in data exchange request")
	ErrMissingScreen              = errors.New("screen is required in data exchange request")
	ErrInvalidEncryption          = errors.New("invalid encryption in data exchange request")
	ErrDecryptionFailed           = errors.New("failed to decrypt data exchange request")
	ErrEncryptionFailed           = errors.New("failed to encrypt data exchange response")
)

// FlowError represents a structured error with additional context.
type FlowError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *FlowError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// NewFlowError creates a new FlowError.
func NewFlowError(code, message, details string) *FlowError {
	return &FlowError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Common error codes
const (
	ErrorCodeValidation     = "VALIDATION_ERROR"
	ErrorCodeNotFound       = "NOT_FOUND"
	ErrorCodeUnauthorized   = "UNAUTHORIZED"
	ErrorCodeRateLimit      = "RATE_LIMIT"
	ErrorCodeInternalError  = "INTERNAL_ERROR"
	ErrorCodeInvalidRequest = "INVALID_REQUEST"
)
