package flows

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// FlowTokenManager manages Flow tokens for secure Flow interactions.
type FlowTokenManager struct {
	tokens map[string]*FlowTokenInfo
	mutex  sync.RWMutex
	ttl    time.Duration
}

// FlowTokenInfo contains information about a Flow token.
type FlowTokenInfo struct {
	Token     string                 `json:"token"`
	FlowID    string                 `json:"flow_id"`
	UserID    string                 `json:"user_id"`
	Data      map[string]interface{} `json:"data,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
	Used      bool                   `json:"used"`
}

// NewFlowTokenManager creates a new Flow token manager.
func NewFlowTokenManager(ttl time.Duration) *FlowTokenManager {
	if ttl == 0 {
		ttl = 24 * time.Hour // Default TTL of 24 hours
	}

	manager := &FlowTokenManager{
		tokens: make(map[string]*FlowTokenInfo),
		ttl:    ttl,
	}

	// Start cleanup goroutine
	go manager.cleanupExpiredTokens()

	return manager
}

// GenerateToken generates a new Flow token.
func (m *FlowTokenManager) GenerateToken(flowID, userID string, data map[string]interface{}) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)
	now := time.Now()

	// Store token info
	m.tokens[token] = &FlowTokenInfo{
		Token:     token,
		FlowID:    flowID,
		UserID:    userID,
		Data:      data,
		CreatedAt: now,
		ExpiresAt: now.Add(m.ttl),
		Used:      false,
	}

	return token, nil
}

// ValidateToken validates a Flow token and returns its information.
func (m *FlowTokenManager) ValidateToken(token string) (*FlowTokenInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	tokenInfo, exists := m.tokens[token]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}

	if time.Now().After(tokenInfo.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}

	if tokenInfo.Used {
		return nil, fmt.Errorf("token already used")
	}

	return tokenInfo, nil
}

// UseToken marks a token as used (for one-time use tokens).
func (m *FlowTokenManager) UseToken(token string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	tokenInfo, exists := m.tokens[token]
	if !exists {
		return fmt.Errorf("token not found")
	}

	if time.Now().After(tokenInfo.ExpiresAt) {
		return fmt.Errorf("token expired")
	}

	if tokenInfo.Used {
		return fmt.Errorf("token already used")
	}

	tokenInfo.Used = true
	return nil
}

// RefreshToken extends the expiration time of a token.
func (m *FlowTokenManager) RefreshToken(token string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	tokenInfo, exists := m.tokens[token]
	if !exists {
		return fmt.Errorf("token not found")
	}

	if time.Now().After(tokenInfo.ExpiresAt) {
		return fmt.Errorf("token expired")
	}

	tokenInfo.ExpiresAt = time.Now().Add(m.ttl)
	return nil
}

// RevokeToken revokes a token immediately.
func (m *FlowTokenManager) RevokeToken(token string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.tokens[token]; !exists {
		return fmt.Errorf("token not found")
	}

	delete(m.tokens, token)
	return nil
}

// GetTokensByUser returns all active tokens for a user.
func (m *FlowTokenManager) GetTokensByUser(userID string) []*FlowTokenInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var userTokens []*FlowTokenInfo
	now := time.Now()

	for _, tokenInfo := range m.tokens {
		if tokenInfo.UserID == userID && now.Before(tokenInfo.ExpiresAt) && !tokenInfo.Used {
			userTokens = append(userTokens, tokenInfo)
		}
	}

	return userTokens
}

// GetTokensByFlow returns all active tokens for a Flow.
func (m *FlowTokenManager) GetTokensByFlow(flowID string) []*FlowTokenInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var flowTokens []*FlowTokenInfo
	now := time.Now()

	for _, tokenInfo := range m.tokens {
		if tokenInfo.FlowID == flowID && now.Before(tokenInfo.ExpiresAt) && !tokenInfo.Used {
			flowTokens = append(flowTokens, tokenInfo)
		}
	}

	return flowTokens
}

// GetStats returns statistics about the token manager.
func (m *FlowTokenManager) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	now := time.Now()
	totalTokens := len(m.tokens)
	activeTokens := 0
	expiredTokens := 0
	usedTokens := 0

	for _, tokenInfo := range m.tokens {
		if tokenInfo.Used {
			usedTokens++
		} else if now.After(tokenInfo.ExpiresAt) {
			expiredTokens++
		} else {
			activeTokens++
		}
	}

	return map[string]interface{}{
		"total_tokens":   totalTokens,
		"active_tokens":  activeTokens,
		"expired_tokens": expiredTokens,
		"used_tokens":    usedTokens,
		"ttl_hours":      m.ttl.Hours(),
	}
}

// cleanupExpiredTokens periodically removes expired tokens.
func (m *FlowTokenManager) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour) // Cleanup every hour
	defer ticker.Stop()

	for range ticker.C {
		m.mutex.Lock()
		now := time.Now()
		expiredTokens := make([]string, 0)

		for token, tokenInfo := range m.tokens {
			if now.After(tokenInfo.ExpiresAt) {
				expiredTokens = append(expiredTokens, token)
			}
		}

		for _, token := range expiredTokens {
			delete(m.tokens, token)
		}

		m.mutex.Unlock()

		if len(expiredTokens) > 0 {
			// Log cleanup if logger is available
			fmt.Printf("Cleaned up %d expired Flow tokens\n", len(expiredTokens))
		}
	}
}

// FlowTokenBuilder helps build Flow tokens with specific configurations.
type FlowTokenBuilder struct {
	manager *FlowTokenManager
	flowID  string
	userID  string
	data    map[string]interface{}
}

// NewFlowTokenBuilder creates a new Flow token builder.
func NewFlowTokenBuilder(manager *FlowTokenManager) *FlowTokenBuilder {
	return &FlowTokenBuilder{
		manager: manager,
		data:    make(map[string]interface{}),
	}
}

// ForFlow sets the Flow ID for the token.
func (b *FlowTokenBuilder) ForFlow(flowID string) *FlowTokenBuilder {
	b.flowID = flowID
	return b
}

// ForUser sets the user ID for the token.
func (b *FlowTokenBuilder) ForUser(userID string) *FlowTokenBuilder {
	b.userID = userID
	return b
}

// WithData adds data to the token.
func (b *FlowTokenBuilder) WithData(key string, value interface{}) *FlowTokenBuilder {
	b.data[key] = value
	return b
}

// WithDataMap sets the entire data map for the token.
func (b *FlowTokenBuilder) WithDataMap(data map[string]interface{}) *FlowTokenBuilder {
	b.data = data
	return b
}

// Generate generates the token with the configured parameters.
func (b *FlowTokenBuilder) Generate() (string, error) {
	if b.flowID == "" {
		return "", fmt.Errorf("flow ID is required")
	}
	if b.userID == "" {
		return "", fmt.Errorf("user ID is required")
	}

	return b.manager.GenerateToken(b.flowID, b.userID, b.data)
}

// DefaultFlowTokenManager is a global default token manager.
var DefaultFlowTokenManager = NewFlowTokenManager(24 * time.Hour)

// GenerateFlowToken is a convenience function using the default token manager.
func GenerateFlowToken(flowID, userID string, data map[string]interface{}) (string, error) {
	return DefaultFlowTokenManager.GenerateToken(flowID, userID, data)
}

// ValidateFlowToken is a convenience function using the default token manager.
func ValidateFlowToken(token string) (*FlowTokenInfo, error) {
	return DefaultFlowTokenManager.ValidateToken(token)
}

// UseFlowToken is a convenience function using the default token manager.
func UseFlowToken(token string) error {
	return DefaultFlowTokenManager.UseToken(token)
}
