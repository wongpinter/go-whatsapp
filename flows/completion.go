package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

// FlowCompletionHandler handles Flow completion events from webhooks.
type FlowCompletionHandler struct {
	completionHandlers map[string]CompletionHandler
	stateManager       *FlowStateManager
	tokenManager       *FlowTokenManager
	logger             *zerolog.Logger
}

// CompletionHandler defines the interface for handling Flow completions.
type CompletionHandler interface {
	HandleCompletion(ctx context.Context, completion *FlowCompletion, state *FlowState) error
}

// CompletionHandlerFunc is a function adapter for CompletionHandler.
type CompletionHandlerFunc func(ctx context.Context, completion *FlowCompletion, state *FlowState) error

// HandleCompletion implements CompletionHandler.
func (f CompletionHandlerFunc) HandleCompletion(ctx context.Context, completion *FlowCompletion, state *FlowState) error {
	return f(ctx, completion, state)
}

// NewFlowCompletionHandler creates a new Flow completion handler.
func NewFlowCompletionHandler(stateManager *FlowStateManager, tokenManager *FlowTokenManager, logger *zerolog.Logger) *FlowCompletionHandler {
	if logger == nil {
		nopLogger := zerolog.Nop()
		logger = &nopLogger
	}

	return &FlowCompletionHandler{
		completionHandlers: make(map[string]CompletionHandler),
		stateManager:       stateManager,
		tokenManager:       tokenManager,
		logger:             logger,
	}
}

// RegisterCompletionHandler registers a completion handler for a specific Flow ID.
func (h *FlowCompletionHandler) RegisterCompletionHandler(flowID string, handler CompletionHandler) {
	h.completionHandlers[flowID] = handler
}

// RegisterCompletionHandlerFunc registers a completion handler function for a specific Flow ID.
func (h *FlowCompletionHandler) RegisterCompletionHandlerFunc(flowID string, handler CompletionHandlerFunc) {
	h.completionHandlers[flowID] = handler
}

// ProcessCompletion processes a Flow completion event.
func (h *FlowCompletionHandler) ProcessCompletion(ctx context.Context, completion *FlowCompletion) error {
	h.logger.Info().
		Str("flow_token", completion.FlowToken).
		Interface("response", completion.Response).
		Msg("Processing Flow completion")

	// Validate flow token
	tokenInfo, err := h.tokenManager.ValidateToken(completion.FlowToken)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("flow_token", completion.FlowToken).
			Msg("Invalid flow token in completion")
		return fmt.Errorf("invalid flow token: %w", err)
	}

	// Get flow state
	state, err := h.stateManager.GetState(completion.FlowToken)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Str("flow_token", completion.FlowToken).
			Msg("Flow state not found, creating new state")

		// Create a new state from token info
		state = &FlowState{
			FlowToken: completion.FlowToken,
			FlowID:    tokenInfo.FlowID,
			UserID:    tokenInfo.UserID,
			Data:      make(map[string]interface{}),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// Update state with completion response
	if state.Data == nil {
		state.Data = make(map[string]interface{})
	}
	state.Data["completion_response"] = completion.Response
	state.Data["completed_at"] = time.Now()
	state.UpdatedAt = time.Now()

	// Save updated state
	h.stateManager.SetState(state)

	// Find and execute completion handler
	handler, exists := h.completionHandlers[tokenInfo.FlowID]
	if !exists {
		h.logger.Warn().
			Str("flow_id", tokenInfo.FlowID).
			Msg("No completion handler registered for Flow")
		return fmt.Errorf("no completion handler for Flow: %s", tokenInfo.FlowID)
	}

	// Execute completion handler
	if err := handler.HandleCompletion(ctx, completion, state); err != nil {
		h.logger.Error().
			Err(err).
			Str("flow_id", tokenInfo.FlowID).
			Str("flow_token", completion.FlowToken).
			Msg("Completion handler failed")
		return fmt.Errorf("completion handler failed: %w", err)
	}

	// Mark token as used
	if err := h.tokenManager.UseToken(completion.FlowToken); err != nil {
		h.logger.Warn().
			Err(err).
			Str("flow_token", completion.FlowToken).
			Msg("Failed to mark token as used")
	}

	h.logger.Info().
		Str("flow_id", tokenInfo.FlowID).
		Str("flow_token", completion.FlowToken).
		Msg("Flow completion processed successfully")

	return nil
}

// FlowCompletionResult represents the result of a Flow completion.
type FlowCompletionResult struct {
	FlowID       string                 `json:"flow_id"`
	UserID       string                 `json:"user_id"`
	FlowToken    string                 `json:"flow_token"`
	Response     map[string]interface{} `json:"response"`
	CompletedAt  time.Time              `json:"completed_at"`
	ProcessedAt  time.Time              `json:"processed_at"`
	Success      bool                   `json:"success"`
	Error        string                 `json:"error,omitempty"`
}

// FlowCompletionStore defines an interface for storing Flow completion results.
type FlowCompletionStore interface {
	StoreCompletion(ctx context.Context, result *FlowCompletionResult) error
	GetCompletion(ctx context.Context, flowToken string) (*FlowCompletionResult, error)
	GetCompletionsByUser(ctx context.Context, userID string) ([]*FlowCompletionResult, error)
	GetCompletionsByFlow(ctx context.Context, flowID string) ([]*FlowCompletionResult, error)
}

// InMemoryCompletionStore is an in-memory implementation of FlowCompletionStore.
type InMemoryCompletionStore struct {
	completions map[string]*FlowCompletionResult
	userIndex   map[string][]string // userID -> []flowToken
	flowIndex   map[string][]string // flowID -> []flowToken
}

// NewInMemoryCompletionStore creates a new in-memory completion store.
func NewInMemoryCompletionStore() *InMemoryCompletionStore {
	return &InMemoryCompletionStore{
		completions: make(map[string]*FlowCompletionResult),
		userIndex:   make(map[string][]string),
		flowIndex:   make(map[string][]string),
	}
}

// StoreCompletion stores a Flow completion result.
func (s *InMemoryCompletionStore) StoreCompletion(ctx context.Context, result *FlowCompletionResult) error {
	s.completions[result.FlowToken] = result

	// Update user index
	userTokens := s.userIndex[result.UserID]
	userTokens = append(userTokens, result.FlowToken)
	s.userIndex[result.UserID] = userTokens

	// Update flow index
	flowTokens := s.flowIndex[result.FlowID]
	flowTokens = append(flowTokens, result.FlowToken)
	s.flowIndex[result.FlowID] = flowTokens

	return nil
}

// GetCompletion retrieves a Flow completion result by token.
func (s *InMemoryCompletionStore) GetCompletion(ctx context.Context, flowToken string) (*FlowCompletionResult, error) {
	result, exists := s.completions[flowToken]
	if !exists {
		return nil, fmt.Errorf("completion not found for token: %s", flowToken)
	}
	return result, nil
}

// GetCompletionsByUser retrieves all Flow completion results for a user.
func (s *InMemoryCompletionStore) GetCompletionsByUser(ctx context.Context, userID string) ([]*FlowCompletionResult, error) {
	tokens, exists := s.userIndex[userID]
	if !exists {
		return []*FlowCompletionResult{}, nil
	}

	results := make([]*FlowCompletionResult, 0, len(tokens))
	for _, token := range tokens {
		if result, exists := s.completions[token]; exists {
			results = append(results, result)
		}
	}

	return results, nil
}

// GetCompletionsByFlow retrieves all Flow completion results for a Flow.
func (s *InMemoryCompletionStore) GetCompletionsByFlow(ctx context.Context, flowID string) ([]*FlowCompletionResult, error) {
	tokens, exists := s.flowIndex[flowID]
	if !exists {
		return []*FlowCompletionResult{}, nil
	}

	results := make([]*FlowCompletionResult, 0, len(tokens))
	for _, token := range tokens {
		if result, exists := s.completions[token]; exists {
			results = append(results, result)
		}
	}

	return results, nil
}

// FlowCompletionProcessor provides high-level Flow completion processing.
type FlowCompletionProcessor struct {
	handler *FlowCompletionHandler
	store   FlowCompletionStore
	logger  *zerolog.Logger
}

// NewFlowCompletionProcessor creates a new Flow completion processor.
func NewFlowCompletionProcessor(handler *FlowCompletionHandler, store FlowCompletionStore, logger *zerolog.Logger) *FlowCompletionProcessor {
	if logger == nil {
		nopLogger := zerolog.Nop()
		logger = &nopLogger
	}

	return &FlowCompletionProcessor{
		handler: handler,
		store:   store,
		logger:  logger,
	}
}

// ProcessAndStore processes a Flow completion and stores the result.
func (p *FlowCompletionProcessor) ProcessAndStore(ctx context.Context, completion *FlowCompletion) (*FlowCompletionResult, error) {
	startTime := time.Now()

	// Validate flow token to get flow info
	tokenInfo, err := p.handler.tokenManager.ValidateToken(completion.FlowToken)
	if err != nil {
		return nil, fmt.Errorf("invalid flow token: %w", err)
	}

	// Create completion result
	result := &FlowCompletionResult{
		FlowID:      tokenInfo.FlowID,
		UserID:      tokenInfo.UserID,
		FlowToken:   completion.FlowToken,
		Response:    completion.Response,
		CompletedAt: time.Now(),
		ProcessedAt: startTime,
		Success:     false,
	}

	// Process completion
	if err := p.handler.ProcessCompletion(ctx, completion); err != nil {
		result.Error = err.Error()
		p.logger.Error().
			Err(err).
			Str("flow_token", completion.FlowToken).
			Msg("Failed to process Flow completion")
	} else {
		result.Success = true
		p.logger.Info().
			Str("flow_token", completion.FlowToken).
			Msg("Flow completion processed successfully")
	}

	result.ProcessedAt = time.Now()

	// Store result
	if err := p.store.StoreCompletion(ctx, result); err != nil {
		p.logger.Error().
			Err(err).
			Str("flow_token", completion.FlowToken).
			Msg("Failed to store completion result")
		return result, fmt.Errorf("failed to store completion result: %w", err)
	}

	return result, nil
}

// ParseFlowCompletionFromJSON parses a Flow completion from JSON.
func ParseFlowCompletionFromJSON(jsonData []byte) (*FlowCompletion, error) {
	var completion FlowCompletion
	if err := json.Unmarshal(jsonData, &completion); err != nil {
		return nil, fmt.Errorf("failed to parse Flow completion JSON: %w", err)
	}
	return &completion, nil
}
