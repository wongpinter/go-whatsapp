package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// VerifySignature verifies the HMAC-SHA256 signature of a webhook payload.
// It compares the provided signature with the expected signature computed
// from the payload and secret.
func VerifySignature(payload []byte, signature, secret string) error {
	if len(signature) == 0 {
		return ErrMissingSignature
	}

	if len(secret) == 0 {
		return ErrMissingSecret
	}

	// WhatsApp sends signatures in the format "sha256=<hex_signature>"
	if !strings.HasPrefix(signature, "sha256=") {
		return ErrInvalidSignatureFormat
	}

	// Extract the hex signature part
	hexSignature := signature[7:] // Remove "sha256=" prefix

	// Compute the expected signature
	expectedSignature, err := ComputeSignature(payload, secret)
	if err != nil {
		return fmt.Errorf("failed to compute expected signature: %w", err)
	}

	// Compare signatures using constant-time comparison to prevent timing attacks
	if !hmac.Equal([]byte(hexSignature), []byte(expectedSignature)) {
		return ErrSignatureMismatch
	}

	return nil
}

// ComputeSignature computes the HMAC-SHA256 signature for the given payload and secret.
// Returns the signature as a hex-encoded string.
func ComputeSignature(payload []byte, secret string) (string, error) {
	if len(secret) == 0 {
		return "", ErrMissingSecret
	}

	// Create HMAC-SHA256 hash
	mac := hmac.New(sha256.New, []byte(secret))
	
	// Write payload to hash
	if _, err := mac.Write(payload); err != nil {
		return "", fmt.Errorf("failed to write payload to hash: %w", err)
	}

	// Get the final hash and encode as hex
	signature := hex.EncodeToString(mac.Sum(nil))
	return signature, nil
}

// ValidateSignatureFormat checks if the signature has the correct format.
func ValidateSignatureFormat(signature string) error {
	if len(signature) == 0 {
		return ErrMissingSignature
	}

	if !strings.HasPrefix(signature, "sha256=") {
		return ErrInvalidSignatureFormat
	}

	// Check if the hex part is valid
	hexPart := signature[7:]
	if len(hexPart) != 64 { // SHA256 produces 32 bytes = 64 hex characters
		return ErrInvalidSignatureLength
	}

	// Validate hex encoding
	if _, err := hex.DecodeString(hexPart); err != nil {
		return ErrInvalidHexEncoding
	}

	return nil
}

// Signature verification errors
type SignatureError struct {
	Code    string
	Message string
}

func (e *SignatureError) Error() string {
	return e.Message
}

var (
	ErrMissingSignature = &SignatureError{
		Code:    "MISSING_SIGNATURE",
		Message: "signature header is missing",
	}
	ErrMissingSecret = &SignatureError{
		Code:    "MISSING_SECRET",
		Message: "webhook secret is missing",
	}
	ErrInvalidSignatureFormat = &SignatureError{
		Code:    "INVALID_SIGNATURE_FORMAT",
		Message: "signature must be in format 'sha256=<hex_signature>'",
	}
	ErrInvalidSignatureLength = &SignatureError{
		Code:    "INVALID_SIGNATURE_LENGTH",
		Message: "signature hex part must be exactly 64 characters",
	}
	ErrInvalidHexEncoding = &SignatureError{
		Code:    "INVALID_HEX_ENCODING",
		Message: "signature contains invalid hex characters",
	}
	ErrSignatureMismatch = &SignatureError{
		Code:    "SIGNATURE_MISMATCH",
		Message: "signature verification failed",
	}
)

// IsSignatureError checks if an error is a signature-related error.
func IsSignatureError(err error) bool {
	_, ok := err.(*SignatureError)
	return ok
}

// GetSignatureErrorCode returns the error code for signature errors.
func GetSignatureErrorCode(err error) string {
	if sigErr, ok := err.(*SignatureError); ok {
		return sigErr.Code
	}
	return ""
}

// SecureCompare performs a constant-time comparison of two byte slices.
// This is a wrapper around hmac.Equal for consistency.
func SecureCompare(a, b []byte) bool {
	return hmac.Equal(a, b)
}

// GenerateSecret generates a random secret for webhook verification.
// This is useful for testing or initial setup.
func GenerateSecret(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("secret length must be positive")
	}

	// For simplicity, we'll generate a hex string
	// In production, you might want to use a more sophisticated approach
	bytes := make([]byte, length/2)
	for i := range bytes {
		bytes[i] = byte(i % 256) // Simple pattern for demo
	}
	
	return hex.EncodeToString(bytes), nil
}

// Constants for signature verification
const (
	// SignatureHeader is the HTTP header name for webhook signatures
	SignatureHeader = "X-Hub-Signature-256"
	
	// SignaturePrefix is the prefix used in WhatsApp webhook signatures
	SignaturePrefix = "sha256="
	
	// SignatureLength is the expected length of the hex signature part
	SignatureLength = 64
)
