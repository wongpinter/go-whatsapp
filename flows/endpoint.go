package flows

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// DataExchangeHandler handles Flow data exchange requests.
type DataExchangeHandler struct {
	privateKey   *rsa.PrivateKey
	actionRouter *ActionRouter
	tokenManager *FlowTokenManager
	stateManager *FlowStateManager
	logger       *zerolog.Logger
}

// DataExchangeOption configures the data exchange handler.
type DataExchangeOption func(*DataExchangeHandler)

// WithPrivateKey sets the RSA private key for decryption.
func WithPrivateKey(privateKey *rsa.PrivateKey) DataExchangeOption {
	return func(h *DataExchangeHandler) {
		h.privateKey = privateKey
	}
}

// WithPrivateKeyPEM sets the RSA private key from PEM string.
func WithPrivateKeyPEM(pemData string) DataExchangeOption {
	return func(h *DataExchangeHandler) {
		block, _ := pem.Decode([]byte(pemData))
		if block == nil {
			return
		}

		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return
		}

		h.privateKey = privateKey
	}
}

// WithActionRouter sets the action router.
func WithActionRouter(router *ActionRouter) DataExchangeOption {
	return func(h *DataExchangeHandler) {
		h.actionRouter = router
	}
}

// WithTokenManager sets the token manager.
func WithTokenManager(manager *FlowTokenManager) DataExchangeOption {
	return func(h *DataExchangeHandler) {
		h.tokenManager = manager
	}
}

// WithStateManager sets the state manager.
func WithStateManager(manager *FlowStateManager) DataExchangeOption {
	return func(h *DataExchangeHandler) {
		h.stateManager = manager
	}
}

// WithDataExchangeLogger sets the logger.
func WithDataExchangeLogger(logger *zerolog.Logger) DataExchangeOption {
	return func(h *DataExchangeHandler) {
		h.logger = logger
	}
}

// NewDataExchangeHandler creates a new data exchange handler.
func NewDataExchangeHandler(opts ...DataExchangeOption) *DataExchangeHandler {
	nopLogger := zerolog.Nop()
	h := &DataExchangeHandler{
		actionRouter: NewActionRouter(),
		tokenManager: DefaultFlowTokenManager,
		stateManager: NewFlowStateManager(),
		logger:       &nopLogger,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

// ServeHTTP implements http.Handler for data exchange requests.
func (h *DataExchangeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.Warn().Str("method", r.Method).Msg("Invalid method for data exchange")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to read request body")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Process the data exchange request
	response, err := h.processDataExchange(r.Context(), body)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to process data exchange")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// processDataExchange processes a data exchange request.
func (h *DataExchangeHandler) processDataExchange(ctx context.Context, encryptedData []byte) (*DataExchangeResponse, error) {
	// Decrypt the request
	request, err := h.decryptRequest(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt request: %w", err)
	}

	h.logger.Info().
		Str("action", request.Action).
		Str("screen", request.Screen).
		Str("flow_token", request.FlowToken).
		Msg("Processing data exchange request")

	// Validate flow token
	tokenInfo, err := h.tokenManager.ValidateToken(request.FlowToken)
	if err != nil {
		h.logger.Error().Err(err).Str("flow_token", request.FlowToken).Msg("Invalid flow token")
		return &DataExchangeResponse{
			Version: request.Version,
			ErrorDetails: &ErrorDetails{
				Code:    "INVALID_TOKEN",
				Message: "Invalid or expired flow token",
			},
		}, nil
	}

	// Create context with token info
	ctx = context.WithValue(ctx, "token_info", tokenInfo)
	ctx = context.WithValue(ctx, "flow_id", tokenInfo.FlowID)
	ctx = context.WithValue(ctx, "user_id", tokenInfo.UserID)

	// Route the action
	response, err := h.actionRouter.RouteAction(ctx, request)
	if err != nil {
		h.logger.Error().Err(err).Str("action", request.Action).Msg("Action routing failed")
		return &DataExchangeResponse{
			Version: request.Version,
			ErrorDetails: &ErrorDetails{
				Code:    "ACTION_FAILED",
				Message: "Failed to process action",
			},
		}, nil
	}

	h.logger.Info().
		Str("action", request.Action).
		Str("screen", request.Screen).
		Msg("Data exchange processed successfully")

	return response, nil
}

// decryptRequest decrypts an encrypted data exchange request.
func (h *DataExchangeHandler) decryptRequest(encryptedData []byte) (*DataExchangeRequest, error) {
	if h.privateKey == nil {
		return nil, fmt.Errorf("private key not configured")
	}

	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(string(encryptedData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// The encrypted data format from WhatsApp:
	// - First part: RSA-encrypted AES key
	// - Second part: AES-encrypted payload

	// Extract RSA-encrypted AES key (first 256 bytes for RSA-2048)
	keySize := h.privateKey.Size()
	if len(ciphertext) < keySize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	encryptedAESKey := ciphertext[:keySize]
	encryptedPayload := ciphertext[keySize:]

	// Decrypt AES key with RSA private key
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, h.privateKey, encryptedAESKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
	}

	// Decrypt payload with AES key
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Extract IV (first 16 bytes)
	if len(encryptedPayload) < aes.BlockSize {
		return nil, fmt.Errorf("encrypted payload too short")
	}

	iv := encryptedPayload[:aes.BlockSize]
	ciphertext = encryptedPayload[aes.BlockSize:]

	// Decrypt with AES-CBC
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Remove PKCS7 padding
	plaintext, err := removePKCS7Padding(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to remove padding: %w", err)
	}

	// Parse JSON
	var request DataExchangeRequest
	if err := json.Unmarshal(plaintext, &request); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &request, nil
}

// encryptResponse encrypts a data exchange response (if needed).
func (h *DataExchangeHandler) encryptResponse(response *DataExchangeResponse) ([]byte, error) {
	// For now, return unencrypted JSON
	// In production, you might want to encrypt responses as well
	return json.Marshal(response)
}

// removePKCS7Padding removes PKCS7 padding from decrypted data.
func removePKCS7Padding(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}

	// Verify padding
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}

// ActionHandler defines the interface for handling Flow actions.
type ActionHandler interface {
	HandleAction(ctx context.Context, request *DataExchangeRequest) (*DataExchangeResponse, error)
}

// ActionHandlerFunc is a function adapter for ActionHandler.
type ActionHandlerFunc func(ctx context.Context, request *DataExchangeRequest) (*DataExchangeResponse, error)

// HandleAction implements ActionHandler.
func (f ActionHandlerFunc) HandleAction(ctx context.Context, request *DataExchangeRequest) (*DataExchangeResponse, error) {
	return f(ctx, request)
}

// ActionRouter routes actions to their handlers.
type ActionRouter struct {
	handlers map[string]ActionHandler
}

// NewActionRouter creates a new action router.
func NewActionRouter() *ActionRouter {
	return &ActionRouter{
		handlers: make(map[string]ActionHandler),
	}
}

// RegisterHandler registers an action handler.
func (r *ActionRouter) RegisterHandler(action string, handler ActionHandler) {
	r.handlers[action] = handler
}

// RegisterHandlerFunc registers an action handler function.
func (r *ActionRouter) RegisterHandlerFunc(action string, handler ActionHandlerFunc) {
	r.handlers[action] = handler
}

// RouteAction routes an action to its handler.
func (r *ActionRouter) RouteAction(ctx context.Context, request *DataExchangeRequest) (*DataExchangeResponse, error) {
	handler, exists := r.handlers[request.Action]
	if !exists {
		return &DataExchangeResponse{
			Version: request.Version,
			ErrorDetails: &ErrorDetails{
				Code:    "UNKNOWN_ACTION",
				Message: fmt.Sprintf("Unknown action: %s", request.Action),
			},
		}, nil
	}

	return handler.HandleAction(ctx, request)
}

// FlowStateManager manages Flow state across interactions.
type FlowStateManager struct {
	states map[string]*FlowState
	mutex  sync.RWMutex
}

// FlowState represents the state of a Flow interaction.
type FlowState struct {
	FlowToken string                 `json:"flow_token"`
	FlowID    string                 `json:"flow_id"`
	UserID    string                 `json:"user_id"`
	Screen    string                 `json:"screen"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// NewFlowStateManager creates a new Flow state manager.
func NewFlowStateManager() *FlowStateManager {
	return &FlowStateManager{
		states: make(map[string]*FlowState),
	}
}

// GetState retrieves Flow state by token.
func (m *FlowStateManager) GetState(flowToken string) (*FlowState, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	state, exists := m.states[flowToken]
	if !exists {
		return nil, fmt.Errorf("state not found for token: %s", flowToken)
	}

	return state, nil
}

// SetState stores Flow state.
func (m *FlowStateManager) SetState(state *FlowState) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	state.UpdatedAt = time.Now()
	if state.CreatedAt.IsZero() {
		state.CreatedAt = state.UpdatedAt
	}

	m.states[state.FlowToken] = state
}

// UpdateState updates specific fields in Flow state.
func (m *FlowStateManager) UpdateState(flowToken string, updates map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	state, exists := m.states[flowToken]
	if !exists {
		return fmt.Errorf("state not found for token: %s", flowToken)
	}

	// Update fields
	for key, value := range updates {
		switch key {
		case "screen":
			if screen, ok := value.(string); ok {
				state.Screen = screen
			}
		case "data":
			if data, ok := value.(map[string]interface{}); ok {
				if state.Data == nil {
					state.Data = make(map[string]interface{})
				}
				for k, v := range data {
					state.Data[k] = v
				}
			}
		}
	}

	state.UpdatedAt = time.Now()
	return nil
}

// DeleteState removes Flow state.
func (m *FlowStateManager) DeleteState(flowToken string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.states, flowToken)
}

// HandleDataExchange provides framework-agnostic data exchange handling
// This method works with the HTTP server abstraction
func (h *DataExchangeHandler) HandleDataExchange(ctx interface{}) error {
	// For now, we'll use a type assertion approach
	// In a real implementation, this would use the httpserver.HTTPContext interface

	// This is a placeholder implementation that delegates to the existing ServeHTTP method
	// The actual implementation would extract the request data from the context
	// and process it using the existing processDataExchange method

	return fmt.Errorf("framework-agnostic data exchange handler not yet fully implemented")
}
