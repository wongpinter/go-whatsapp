package webhook

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// RateLimitType represents different types of rate limits
type RateLimitType string

const (
	RateLimitTypeAPI        RateLimitType = "api_calls"      // 200/hour default, 5000/hour with phone
	RateLimitTypeThroughput RateLimitType = "throughput"     // 80 MPS default, 1000 MPS upgraded
	RateLimitTypeMessaging  RateLimitType = "messaging"      // Daily conversation limits by tier
	RateLimitTypePair       RateLimitType = "pair"           // Messages to same recipient
	RateLimitTypeSpam       RateLimitType = "spam"           // Quality-based spam detection
)

// RateLimitTier represents messaging limit tiers
type RateLimitTier string

const (
	TierInitial RateLimitTier = "initial" // 250 conversations/day
	Tier1       RateLimitTier = "tier1"   // 1,000 conversations/day
	Tier2       RateLimitTier = "tier2"   // 10,000 conversations/day
	Tier3       RateLimitTier = "tier3"   // 100,000 conversations/day
	Tier4       RateLimitTier = "tier4"   // Unlimited conversations/day
)

// QualityRating represents the quality rating levels
type QualityRating string

const (
	QualityHigh   QualityRating = "high"   // Green - excellent reception
	QualityMedium QualityRating = "medium" // Yellow - acceptable quality
	QualityLow    QualityRating = "low"    // Red - poor quality, flagged
)

// RateLimitInfo contains information about current rate limits
type RateLimitInfo struct {
	Type            RateLimitType
	CurrentUsage    int64
	Limit           int64
	ResetTime       time.Time
	Tier            RateLimitTier
	QualityRating   QualityRating
	ThroughputMPS   int
	LastUpdated     time.Time
}

// RateLimiter manages rate limiting with exponential backoff
type RateLimiter struct {
	// Rate limit tracking
	limits map[RateLimitType]*RateLimitInfo
	mu     sync.RWMutex

	// Backoff configuration
	baseDelay    time.Duration
	maxDelay     time.Duration
	backoffFactor float64
	jitterFactor  float64

	// Metrics
	retryAttempts map[RateLimitType]int64
	retryMu       sync.RWMutex

	// Configuration
	enabled bool
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits:        make(map[RateLimitType]*RateLimitInfo),
		retryAttempts: make(map[RateLimitType]int64),
		baseDelay:     1 * time.Second,
		maxDelay:      5 * time.Minute,
		backoffFactor: 2.0,
		jitterFactor:  0.1,
		enabled:       true,
	}
}

// UpdateRateLimit updates rate limit information
func (rl *RateLimiter) UpdateRateLimit(limitType RateLimitType, usage, limit int64, resetTime time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.limits[limitType] == nil {
		rl.limits[limitType] = &RateLimitInfo{
			Type: limitType,
		}
	}

	info := rl.limits[limitType]
	info.CurrentUsage = usage
	info.Limit = limit
	info.ResetTime = resetTime
	info.LastUpdated = time.Now()
}

// UpdateQualityInfo updates quality rating and tier information
func (rl *RateLimiter) UpdateQualityInfo(tier RateLimitTier, quality QualityRating, throughputMPS int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Update messaging limit info
	if rl.limits[RateLimitTypeMessaging] == nil {
		rl.limits[RateLimitTypeMessaging] = &RateLimitInfo{
			Type: RateLimitTypeMessaging,
		}
	}

	info := rl.limits[RateLimitTypeMessaging]
	info.Tier = tier
	info.QualityRating = quality
	info.ThroughputMPS = throughputMPS
	info.LastUpdated = time.Now()

	// Set limits based on tier
	switch tier {
	case TierInitial:
		info.Limit = 250
	case Tier1:
		info.Limit = 1000
	case Tier2:
		info.Limit = 10000
	case Tier3:
		info.Limit = 100000
	case Tier4:
		info.Limit = -1 // Unlimited
	}
}

// CheckRateLimit checks if a request would exceed rate limits
func (rl *RateLimiter) CheckRateLimit(limitType RateLimitType) (bool, time.Duration) {
	if !rl.enabled {
		return true, 0
	}

	rl.mu.RLock()
	defer rl.mu.RUnlock()

	info, exists := rl.limits[limitType]
	if !exists {
		return true, 0 // No limit info, allow request
	}

	// Check if limit is exceeded
	if info.Limit > 0 && info.CurrentUsage >= info.Limit {
		// Calculate wait time until reset
		waitTime := time.Until(info.ResetTime)
		if waitTime < 0 {
			waitTime = 0
		}
		return false, waitTime
	}

	return true, 0
}

// HandleRateLimitError processes rate limit errors and calculates backoff
func (rl *RateLimiter) HandleRateLimitError(ctx context.Context, errorCode int, limitType RateLimitType) (time.Duration, error) {
	// Map error codes to rate limit types
	switch errorCode {
	case 80007:
		limitType = RateLimitTypeAPI
	case 130429:
		limitType = RateLimitTypeThroughput
	case 131048:
		limitType = RateLimitTypeSpam
	}

	// Increment retry attempts
	rl.retryMu.Lock()
	rl.retryAttempts[limitType]++
	attempts := rl.retryAttempts[limitType]
	rl.retryMu.Unlock()

	// Calculate exponential backoff with jitter
	delay := rl.calculateBackoffDelay(attempts)

	// Check if we should continue retrying
	if attempts > 10 { // Max 10 retries
		return 0, fmt.Errorf("max retry attempts exceeded for rate limit type %s", limitType)
	}

	return delay, nil
}

// calculateBackoffDelay calculates exponential backoff delay with jitter
func (rl *RateLimiter) calculateBackoffDelay(attempts int64) time.Duration {
	// Exponential backoff: baseDelay * backoffFactor^attempts
	delay := float64(rl.baseDelay) * math.Pow(rl.backoffFactor, float64(attempts-1))

	// Add jitter to prevent thundering herd
	jitter := delay * rl.jitterFactor * (rand.Float64()*2 - 1) // ±10% jitter
	delay += jitter

	// Cap at maximum delay
	if delay > float64(rl.maxDelay) {
		delay = float64(rl.maxDelay)
	}

	return time.Duration(delay)
}

// ResetRetryAttempts resets retry attempts for a rate limit type
func (rl *RateLimiter) ResetRetryAttempts(limitType RateLimitType) {
	rl.retryMu.Lock()
	defer rl.retryMu.Unlock()
	delete(rl.retryAttempts, limitType)
}

// GetRateLimitInfo returns current rate limit information
func (rl *RateLimiter) GetRateLimitInfo(limitType RateLimitType) (*RateLimitInfo, bool) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	info, exists := rl.limits[limitType]
	if !exists {
		return nil, false
	}

	// Return a copy
	infoCopy := *info
	return &infoCopy, true
}

// GetAllRateLimits returns all current rate limit information
func (rl *RateLimiter) GetAllRateLimits() map[RateLimitType]*RateLimitInfo {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	result := make(map[RateLimitType]*RateLimitInfo)
	for limitType, info := range rl.limits {
		infoCopy := *info
		result[limitType] = &infoCopy
	}
	return result
}

// IsQualityAtRisk checks if quality rating is at risk
func (rl *RateLimiter) IsQualityAtRisk() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	info, exists := rl.limits[RateLimitTypeMessaging]
	if !exists {
		return false
	}

	return info.QualityRating == QualityLow
}

// GetQualityRecommendations returns recommendations for improving quality
func (rl *RateLimiter) GetQualityRecommendations() []string {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	info, exists := rl.limits[RateLimitTypeMessaging]
	if !exists {
		return []string{"No quality information available"}
	}

	var recommendations []string

	switch info.QualityRating {
	case QualityLow:
		recommendations = append(recommendations,
			"URGENT: Quality rating is low - you have 7 days to improve",
			"Ensure all recipients have opted in to receive messages",
			"Send only relevant, valuable content to reduce blocks",
			"Avoid sending repetitive or spam-like messages",
			"Provide clear opt-out instructions in messages",
			"Review message timing - send during business hours",
		)
	case QualityMedium:
		recommendations = append(recommendations,
			"Quality rating is acceptable but can be improved",
			"Focus on message relevance and timing",
			"Monitor user engagement and feedback",
			"Consider personalizing messages for better reception",
		)
	case QualityHigh:
		recommendations = append(recommendations,
			"Excellent quality rating - maintain current practices",
			"Continue focusing on user consent and message relevance",
			"Monitor for any changes in user behavior patterns",
		)
	}

	return recommendations
}

// GetThroughputUpgradeEligibility checks eligibility for throughput upgrade
func (rl *RateLimiter) GetThroughputUpgradeEligibility() (bool, []string) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	info, exists := rl.limits[RateLimitTypeMessaging]
	if !exists {
		return false, []string{"No messaging limit information available"}
	}

	var requirements []string
	eligible := true

	// Check quality rating
	if info.QualityRating == QualityLow {
		eligible = false
		requirements = append(requirements, "❌ Quality rating must be medium or higher")
	} else {
		requirements = append(requirements, "✅ Quality rating is acceptable")
	}

	// Check throughput level
	if info.ThroughputMPS < 1000 {
		requirements = append(requirements, "📈 Current throughput: "+fmt.Sprintf("%d MPS", info.ThroughputMPS))
		requirements = append(requirements, "🎯 Target: 1000 MPS (automatic upgrade available)")
	} else {
		requirements = append(requirements, "✅ Already at upgraded throughput level")
	}

	// Additional requirements
	requirements = append(requirements,
		"📋 Requirements for 1000 MPS upgrade:",
		"   • Phone number must be on Cloud API",
		"   • Must initiate conversations with unlimited unique customers in 24h",
		"   • Must maintain medium or higher quality rating",
	)

	return eligible, requirements
}

// Enable enables rate limiting
func (rl *RateLimiter) Enable() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.enabled = true
}

// Disable disables rate limiting (for testing)
func (rl *RateLimiter) Disable() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.enabled = false
}
