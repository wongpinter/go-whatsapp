package whatsapp

import (
	"fmt"
	"strconv"
)

// APIError represents a structured error response from the WhatsApp Cloud API.
// It implements the error interface and provides detailed information about
// API failures that can be programmatically inspected.
type APIError struct {
	ErrorInfo struct {
		Message      string `json:"message"`
		Type         string `json:"type"`
		Code         int    `json:"code"`
		ErrorSubcode int    `json:"error_subcode"`
		FBTraceID    string `json:"fbtrace_id"`
	} `json:"error"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("whatsapp api error: code=%d, type=%s, message=%s",
		e.ErrorInfo.Code, e.ErrorInfo.Type, e.ErrorInfo.Message)
}

// Code returns the primary error code.
func (e *APIError) Code() int {
	return e.ErrorInfo.Code
}

// Subcode returns the secondary error code.
func (e *APIError) Subcode() int {
	return e.ErrorInfo.ErrorSubcode
}

// Message returns the human-readable error message.
func (e *APIError) Message() string {
	return e.ErrorInfo.Message
}

// Type returns the error category.
func (e *APIError) Type() string {
	return e.ErrorInfo.Type
}

// TraceID returns the Facebook trace ID for debugging.
func (e *APIError) TraceID() string {
	return e.ErrorInfo.FBTraceID
}

// Error category helper functions for common error handling scenarios

// IsAuthError checks if the error is related to authentication.
func IsAuthError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		code := apiErr.Code()
		return code == 10 || code == 190 || (code >= 200 && code <= 299)
	}
	return false
}

// IsRateLimitError checks if the error is due to rate limiting.
func IsRateLimitError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		code := apiErr.Code()
		return code == 130429 || code == 131056
	}
	return false
}

// IsUndeliverableError checks if the message could not be delivered.
func IsUndeliverableError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code() == 131026
	}
	return false
}

// IsReEngagementError checks if a re-engagement message is required.
func IsReEngagementError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code() == 131047
	}
	return false
}

// IsTemplateError checks if the error is related to message templates.
func IsTemplateError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code() == 132001
	}
	return false
}

// IsInvalidParameterError checks if the error is due to invalid parameters.
func IsInvalidParameterError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code() == 100
	}
	return false
}

// IsFlowError checks if the error is related to WhatsApp Flows.
func IsFlowError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code() == 139004
	}
	return false
}

// ErrCustomerWindowClosed is returned when attempting to send a free-form message
// outside the 24-hour customer service window.
type ErrCustomerWindowClosed struct {
	LastMessageTime int64
}

func (e *ErrCustomerWindowClosed) Error() string {
	return fmt.Sprintf("customer service window closed: last message was at %d, use template message instead",
		e.LastMessageTime)
}

// Common validation errors

// ErrInvalidPhoneNumber is returned when a phone number is invalid.
type ErrInvalidPhoneNumber struct {
	PhoneNumber string
}

func (e *ErrInvalidPhoneNumber) Error() string {
	return fmt.Sprintf("invalid phone number: %s", e.PhoneNumber)
}

// ErrEmptyMessage is returned when attempting to send an empty message.
type ErrEmptyMessage struct{}

func (e *ErrEmptyMessage) Error() string {
	return "message body cannot be empty"
}

// ErrInvalidAccessToken is returned when the access token is invalid or missing.
type ErrInvalidAccessToken struct{}

func (e *ErrInvalidAccessToken) Error() string {
	return "access token is required and must be a valid WhatsApp Business API token"
}

// ErrInvalidPhoneNumberID is returned when the phone number ID is invalid or missing.
type ErrInvalidPhoneNumberID struct{}

func (e *ErrInvalidPhoneNumberID) Error() string {
	return "phone number ID is required and must be a valid WhatsApp Business phone number ID"
}

// Helper function to create APIError from response
func NewAPIError(code int, message, errorType, traceID string) *APIError {
	apiErr := &APIError{}
	apiErr.ErrorInfo.Code = code
	apiErr.ErrorInfo.Message = message
	apiErr.ErrorInfo.Type = errorType
	apiErr.ErrorInfo.FBTraceID = traceID
	return apiErr
}

// ParseErrorCode safely parses an error code from string
func ParseErrorCode(codeStr string) int {
	if code, err := strconv.Atoi(codeStr); err == nil {
		return code
	}
	return 0
}
