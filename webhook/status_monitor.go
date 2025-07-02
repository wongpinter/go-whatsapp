package webhook

import (
	"context"
	"sync"
	"time"
)

// StatusMonitor tracks and analyzes message status updates
type StatusMonitor struct {
	// Message tracking
	messages map[string]*MessageTracker
	mu       sync.RWMutex

	// Metrics
	metrics *MetricsCollector

	// Configuration
	retentionPeriod time.Duration
	cleanupTicker   *time.Ticker
	stopChan        chan struct{}
}

// MessageTracker tracks the lifecycle of a single message
type MessageTracker struct {
	MessageID     string
	RecipientID   string
	SentAt        time.Time
	DeliveredAt   *time.Time
	ReadAt        *time.Time
	FailedAt      *time.Time
	ErrorCode     *int
	ErrorMessage  *string
	LastStatus    string
	StatusHistory []StatusUpdate
}

// StatusUpdate represents a single status update
type StatusUpdate struct {
	Status    string
	Timestamp time.Time
	ErrorCode *int
	ErrorMsg  *string
}

// MessageStats provides statistics about message delivery
type MessageStats struct {
	TotalMessages     int
	SentMessages      int
	DeliveredMessages int
	ReadMessages      int
	FailedMessages    int

	// Rates
	DeliveryRate float64
	ReadRate     float64
	FailureRate  float64

	// Timing
	AverageDeliveryTime time.Duration
	AverageReadTime     time.Duration
}

// NewStatusMonitor creates a new status monitor
func NewStatusMonitor(retentionPeriod time.Duration) *StatusMonitor {
	sm := &StatusMonitor{
		messages:        make(map[string]*MessageTracker),
		metrics:         NewMetricsCollector(),
		retentionPeriod: retentionPeriod,
		cleanupTicker:   time.NewTicker(time.Hour), // Cleanup every hour
		stopChan:        make(chan struct{}),
	}

	go sm.cleanupLoop()
	return sm
}

// TrackStatus tracks a status update for a message
func (sm *StatusMonitor) TrackStatus(ctx context.Context, status *Status, metadata *Metadata) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	messageID := status.ID
	tracker, exists := sm.messages[messageID]

	if !exists {
		tracker = &MessageTracker{
			MessageID:     messageID,
			RecipientID:   status.RecipientID,
			SentAt:        time.Now(), // Approximate, since we don't have exact send time
			LastStatus:    status.Status,
			StatusHistory: make([]StatusUpdate, 0),
		}
		sm.messages[messageID] = tracker
	}

	// Update tracker based on status
	timestamp := status.GetTimestamp()
	statusUpdate := StatusUpdate{
		Status:    status.Status,
		Timestamp: timestamp,
	}

	// Handle errors
	if len(status.Errors) > 0 {
		errorCode := status.Errors[0].Code
		errorMsg := status.Errors[0].Message
		statusUpdate.ErrorCode = &errorCode
		statusUpdate.ErrorMsg = &errorMsg
		tracker.ErrorCode = &errorCode
		tracker.ErrorMessage = &errorMsg
	}

	// Update specific status timestamps
	switch status.Status {
	case "delivered":
		tracker.DeliveredAt = &timestamp
	case "read":
		tracker.ReadAt = &timestamp
	case "failed":
		tracker.FailedAt = &timestamp
	}

	tracker.LastStatus = status.Status
	tracker.StatusHistory = append(tracker.StatusHistory, statusUpdate)

	// Record metrics
	sm.metrics.RecordStatus(status.Status)
}

// GetMessageTracker returns the tracker for a specific message
func (sm *StatusMonitor) GetMessageTracker(messageID string) (*MessageTracker, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	tracker, exists := sm.messages[messageID]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	trackerCopy := *tracker
	trackerCopy.StatusHistory = make([]StatusUpdate, len(tracker.StatusHistory))
	copy(trackerCopy.StatusHistory, tracker.StatusHistory)

	return &trackerCopy, true
}

// GetMessageStats returns overall message statistics
func (sm *StatusMonitor) GetMessageStats() *MessageStats {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := &MessageStats{}
	var totalDeliveryTime, totalReadTime time.Duration
	var deliveredCount, readCount int

	for _, tracker := range sm.messages {
		stats.TotalMessages++

		switch tracker.LastStatus {
		case "sent":
			stats.SentMessages++
		case "delivered":
			stats.DeliveredMessages++
			if tracker.DeliveredAt != nil {
				deliveryTime := tracker.DeliveredAt.Sub(tracker.SentAt)
				totalDeliveryTime += deliveryTime
				deliveredCount++
			}
		case "read":
			stats.ReadMessages++
			stats.DeliveredMessages++ // Read implies delivered
			if tracker.ReadAt != nil && tracker.DeliveredAt != nil {
				readTime := tracker.ReadAt.Sub(*tracker.DeliveredAt)
				totalReadTime += readTime
				readCount++
			}
			if tracker.DeliveredAt != nil {
				deliveryTime := tracker.DeliveredAt.Sub(tracker.SentAt)
				totalDeliveryTime += deliveryTime
				deliveredCount++
			}
		case "failed":
			stats.FailedMessages++
		}
	}

	// Calculate rates
	if stats.TotalMessages > 0 {
		stats.DeliveryRate = float64(stats.DeliveredMessages) / float64(stats.TotalMessages) * 100
		stats.FailureRate = float64(stats.FailedMessages) / float64(stats.TotalMessages) * 100
	}

	if stats.DeliveredMessages > 0 {
		stats.ReadRate = float64(stats.ReadMessages) / float64(stats.DeliveredMessages) * 100
	}

	// Calculate average times
	if deliveredCount > 0 {
		stats.AverageDeliveryTime = totalDeliveryTime / time.Duration(deliveredCount)
	}

	if readCount > 0 {
		stats.AverageReadTime = totalReadTime / time.Duration(readCount)
	}

	return stats
}

// GetFailedMessages returns all messages that failed to deliver
func (sm *StatusMonitor) GetFailedMessages() []*MessageTracker {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var failed []*MessageTracker
	for _, tracker := range sm.messages {
		if tracker.LastStatus == "failed" {
			// Return a copy
			trackerCopy := *tracker
			trackerCopy.StatusHistory = make([]StatusUpdate, len(tracker.StatusHistory))
			copy(trackerCopy.StatusHistory, tracker.StatusHistory)
			failed = append(failed, &trackerCopy)
		}
	}
	return failed
}

// GetUndeliveredMessages returns messages that haven't been delivered within the timeout
func (sm *StatusMonitor) GetUndeliveredMessages(timeout time.Duration) []*MessageTracker {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var undelivered []*MessageTracker
	cutoff := time.Now().Add(-timeout)

	for _, tracker := range sm.messages {
		if tracker.DeliveredAt == nil && tracker.FailedAt == nil && tracker.SentAt.Before(cutoff) {
			// Return a copy
			trackerCopy := *tracker
			trackerCopy.StatusHistory = make([]StatusUpdate, len(tracker.StatusHistory))
			copy(trackerCopy.StatusHistory, tracker.StatusHistory)
			undelivered = append(undelivered, &trackerCopy)
		}
	}
	return undelivered
}

// GetMetrics returns the underlying metrics collector
func (sm *StatusMonitor) GetMetrics() *MetricsCollector {
	return sm.metrics
}

// Stop stops the status monitor
func (sm *StatusMonitor) Stop() {
	close(sm.stopChan)
	sm.cleanupTicker.Stop()
	sm.metrics.Stop()
}

// cleanupLoop periodically removes old message trackers
func (sm *StatusMonitor) cleanupLoop() {
	for {
		select {
		case <-sm.cleanupTicker.C:
			sm.cleanup()
		case <-sm.stopChan:
			return
		}
	}
}

// cleanup removes message trackers older than the retention period
func (sm *StatusMonitor) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cutoff := time.Now().Add(-sm.retentionPeriod)
	for messageID, tracker := range sm.messages {
		if tracker.SentAt.Before(cutoff) {
			delete(sm.messages, messageID)
		}
	}
}
