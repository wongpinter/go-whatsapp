package whatsapp

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	// Test valid client creation
	client, err := NewClient("123456789", "test_token")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if client.GetPhoneNumberID() != "123456789" {
		t.Errorf("Expected phone number ID '123456789', got '%s'", client.GetPhoneNumberID())
	}
	
	if client.GetAPIVersion() != "v19.0" {
		t.Errorf("Expected API version 'v19.0', got '%s'", client.GetAPIVersion())
	}
}

func TestNewClientWithInvalidParams(t *testing.T) {
	// Test with empty phone number ID
	_, err := NewClient("", "test_token")
	if err == nil {
		t.Error("Expected error for empty phone number ID")
	}
	
	// Test with empty access token
	_, err = NewClient("123456789", "")
	if err == nil {
		t.Error("Expected error for empty access token")
	}
}

func TestAPIErrorMethods(t *testing.T) {
	apiErr := NewAPIError(131026, "Message Undeliverable", "OAuthException", "trace123")
	
	if apiErr.Code() != 131026 {
		t.Errorf("Expected code 131026, got %d", apiErr.Code())
	}
	
	if apiErr.Message() != "Message Undeliverable" {
		t.Errorf("Expected message 'Message Undeliverable', got '%s'", apiErr.Message())
	}
	
	if apiErr.Type() != "OAuthException" {
		t.Errorf("Expected type 'OAuthException', got '%s'", apiErr.Type())
	}
	
	if apiErr.TraceID() != "trace123" {
		t.Errorf("Expected trace ID 'trace123', got '%s'", apiErr.TraceID())
	}
}

func TestErrorHelperFunctions(t *testing.T) {
	// Test rate limit error
	rateLimitErr := NewAPIError(130429, "Rate limit hit", "OAuthException", "trace123")
	if !IsRateLimitError(rateLimitErr) {
		t.Error("Expected IsRateLimitError to return true for rate limit error")
	}
	
	// Test undeliverable error
	undeliverableErr := NewAPIError(131026, "Message Undeliverable", "OAuthException", "trace123")
	if !IsUndeliverableError(undeliverableErr) {
		t.Error("Expected IsUndeliverableError to return true for undeliverable error")
	}
	
	// Test auth error
	authErr := NewAPIError(190, "Access token expired", "OAuthException", "trace123")
	if !IsAuthError(authErr) {
		t.Error("Expected IsAuthError to return true for auth error")
	}
	
	// Test re-engagement error
	reEngagementErr := NewAPIError(131047, "Re-engagement Message", "OAuthException", "trace123")
	if !IsReEngagementError(reEngagementErr) {
		t.Error("Expected IsReEngagementError to return true for re-engagement error")
	}
}

func TestClientOptions(t *testing.T) {
	client, err := NewClient("123456789", "test_token",
		WithAPIVersion("v18.0"),
		WithWABAID("waba123"),
	)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if client.GetAPIVersion() != "v18.0" {
		t.Errorf("Expected API version 'v18.0', got '%s'", client.GetAPIVersion())
	}
	
	if client.GetWABAID() != "waba123" {
		t.Errorf("Expected WABA ID 'waba123', got '%s'", client.GetWABAID())
	}
}
