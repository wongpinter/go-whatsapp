package webhook

import (
	"context"
	"sync"
	"time"
)

// MessageLifecycleTracker tracks the complete lifecycle of messages from send to delivery
type MessageLifecycleTracker struct {
	// Message tracking
	messages map[string]*MessageLifecycle
	mu       sync.RWMutex

	// Deduplication
	processedEvents map[string]time.Time
	dedupMu         sync.RWMutex

	// Configuration
	retentionPeriod time.Duration
	dedupWindow     time.Duration
	staleThreshold  time.Duration
	cleanupTicker   *time.Ticker
	stopChan        chan struct{}
}

// MessageLifecycle represents the complete lifecycle of a message
type MessageLifecycle struct {
	// Message identification
	WAMID       string    // WhatsApp Message ID from API response
	MessageID   string    // Internal message ID
	RecipientID string    // Recipient phone number
	SentAt      time.Time // When message was sent

	// Status progression
	StatusHistory []StatusEvent
	CurrentStatus string

	// Delivery tracking
	DeliveredAt *time.Time
	ReadAt      *time.Time
	FailedAt    *time.Time

	// Error tracking
	Errors []MessageError

	// Conversation and pricing
	ConversationID string
	PricingInfo    *PricingInfo

	// Metadata
	MessageType string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// StatusEvent represents a single status change event
type StatusEvent struct {
	Status    string
	Timestamp time.Time
	Source    string // "api" or "webhook"
	Metadata  map[string]interface{}
}

// PricingInfo represents pricing information for a message
type PricingInfo struct {
	Billable     bool
	Category     string  // "service", "utility", "authentication", "marketing"
	PricingModel string  // "CBP"
	Cost         float64 // If available
	Currency     string  // If available
}

// NewMessageLifecycleTracker creates a new message lifecycle tracker
func NewMessageLifecycleTracker(retentionPeriod time.Duration) *MessageLifecycleTracker {
	tracker := &MessageLifecycleTracker{
		messages:        make(map[string]*MessageLifecycle),
		processedEvents: make(map[string]time.Time),
		retentionPeriod: retentionPeriod,
		dedupWindow:     10 * time.Minute, // Deduplicate events within 10 minutes
		staleThreshold:  10 * time.Minute, // Ignore events older than 10 minutes
		cleanupTicker:   time.NewTicker(time.Hour),
		stopChan:        make(chan struct{}),
	}

	go tracker.cleanupLoop()
	return tracker
}

// TrackMessageSent records when a message is sent via API
func (t *MessageLifecycleTracker) TrackMessageSent(ctx context.Context, wamid, messageID, recipientID, messageType string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	lifecycle := &MessageLifecycle{
		WAMID:       wamid,
		MessageID:   messageID,
		RecipientID: recipientID,
		MessageType: messageType,
		SentAt:      time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		StatusHistory: []StatusEvent{
			{
				Status:    "sent",
				Timestamp: time.Now(),
				Source:    "api",
				Metadata:  make(map[string]interface{}),
			},
		},
		CurrentStatus: "sent",
		Errors:        make([]MessageError, 0),
	}

	// Store by both WAMID and internal message ID for lookup flexibility
	t.messages[wamid] = lifecycle
	if messageID != "" && messageID != wamid {
		t.messages[messageID] = lifecycle
	}
}

// TrackStatusUpdate processes a status update from webhook
func (t *MessageLifecycleTracker) TrackStatusUpdate(ctx context.Context, status *Status, metadata *Metadata) error {
	// Check for duplicates
	eventKey := t.generateEventKey(status.ID, status.Status, status.Timestamp)
	if t.isDuplicate(eventKey) {
		return nil // Silently ignore duplicates
	}

	// Check for stale events
	statusTime := status.GetTimestamp()
	if time.Since(statusTime) > t.staleThreshold {
		return nil // Silently ignore stale events
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Find the message lifecycle
	lifecycle, exists := t.messages[status.ID]
	if !exists {
		// Create a new lifecycle if we receive a status before tracking the send
		lifecycle = &MessageLifecycle{
			WAMID:         status.ID,
			RecipientID:   status.RecipientID,
			SentAt:        statusTime, // Approximate
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			StatusHistory: make([]StatusEvent, 0),
			CurrentStatus: "",
			Errors:        make([]MessageError, 0),
		}
		t.messages[status.ID] = lifecycle
	}

	// Add status event
	statusEvent := StatusEvent{
		Status:    status.Status,
		Timestamp: statusTime,
		Source:    "webhook",
		Metadata: map[string]interface{}{
			"phone_number_id": metadata.PhoneNumberID,
		},
	}

	// Add conversation and pricing info if available
	if status.Conversation != nil {
		lifecycle.ConversationID = status.Conversation.ID
		statusEvent.Metadata["conversation_id"] = status.Conversation.ID
		if status.Conversation.Origin != nil {
			statusEvent.Metadata["conversation_origin"] = status.Conversation.Origin.Type
		}
	}

	if status.Pricing != nil {
		lifecycle.PricingInfo = &PricingInfo{
			Billable:     status.Pricing.Billable,
			Category:     status.Pricing.Category,
			PricingModel: status.Pricing.PricingModel,
		}
		statusEvent.Metadata["pricing"] = status.Pricing
	}

	lifecycle.StatusHistory = append(lifecycle.StatusHistory, statusEvent)
	lifecycle.CurrentStatus = status.Status
	lifecycle.UpdatedAt = time.Now()

	// Update specific status timestamps
	switch status.Status {
	case "delivered":
		lifecycle.DeliveredAt = &statusTime
	case "read":
		lifecycle.ReadAt = &statusTime
	case "failed":
		lifecycle.FailedAt = &statusTime
	}

	// Process any errors
	for _, err := range status.Errors {
		msgErr := MessageError{
			Code:      err.Code,
			Title:     err.Title,
			Message:   err.Message,
			Timestamp: statusTime,
			Source:    "webhook",
		}
		lifecycle.Errors = append(lifecycle.Errors, msgErr)
	}

	// Mark event as processed
	t.markProcessed(eventKey)

	return nil
}

// GetMessageLifecycle retrieves the lifecycle for a message
func (t *MessageLifecycleTracker) GetMessageLifecycle(messageID string) (*MessageLifecycle, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	lifecycle, exists := t.messages[messageID]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	lifecycleCopy := *lifecycle
	lifecycleCopy.StatusHistory = make([]StatusEvent, len(lifecycle.StatusHistory))
	copy(lifecycleCopy.StatusHistory, lifecycle.StatusHistory)
	lifecycleCopy.Errors = make([]MessageError, len(lifecycle.Errors))
	copy(lifecycleCopy.Errors, lifecycle.Errors)

	return &lifecycleCopy, true
}

// GetMessagesByStatus returns all messages with a specific status
func (t *MessageLifecycleTracker) GetMessagesByStatus(status string) []*MessageLifecycle {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var messages []*MessageLifecycle
	for _, lifecycle := range t.messages {
		if lifecycle.CurrentStatus == status {
			// Return a copy
			lifecycleCopy := *lifecycle
			lifecycleCopy.StatusHistory = make([]StatusEvent, len(lifecycle.StatusHistory))
			copy(lifecycleCopy.StatusHistory, lifecycle.StatusHistory)
			lifecycleCopy.Errors = make([]MessageError, len(lifecycle.Errors))
			copy(lifecycleCopy.Errors, lifecycle.Errors)
			messages = append(messages, &lifecycleCopy)
		}
	}
	return messages
}

// GetFailedMessages returns all failed messages with error details
func (t *MessageLifecycleTracker) GetFailedMessages() []*MessageLifecycle {
	return t.GetMessagesByStatus("failed")
}

// GetUndeliveredMessages returns messages that haven't been delivered within timeout
func (t *MessageLifecycleTracker) GetUndeliveredMessages(timeout time.Duration) []*MessageLifecycle {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var undelivered []*MessageLifecycle
	cutoff := time.Now().Add(-timeout)

	for _, lifecycle := range t.messages {
		if lifecycle.DeliveredAt == nil && lifecycle.FailedAt == nil && lifecycle.SentAt.Before(cutoff) {
			// Return a copy
			lifecycleCopy := *lifecycle
			lifecycleCopy.StatusHistory = make([]StatusEvent, len(lifecycle.StatusHistory))
			copy(lifecycleCopy.StatusHistory, lifecycle.StatusHistory)
			lifecycleCopy.Errors = make([]MessageError, len(lifecycle.Errors))
			copy(lifecycleCopy.Errors, lifecycle.Errors)
			undelivered = append(undelivered, &lifecycleCopy)
		}
	}
	return undelivered
}

// Stop stops the message lifecycle tracker
func (t *MessageLifecycleTracker) Stop() {
	close(t.stopChan)
	t.cleanupTicker.Stop()
}

// Helper methods for deduplication
func (t *MessageLifecycleTracker) generateEventKey(messageID, status, timestamp string) string {
	return messageID + ":" + status + ":" + timestamp
}

func (t *MessageLifecycleTracker) isDuplicate(eventKey string) bool {
	t.dedupMu.RLock()
	defer t.dedupMu.RUnlock()

	_, exists := t.processedEvents[eventKey]
	return exists
}

func (t *MessageLifecycleTracker) markProcessed(eventKey string) {
	t.dedupMu.Lock()
	defer t.dedupMu.Unlock()

	t.processedEvents[eventKey] = time.Now()
}

// cleanupLoop periodically removes old data
func (t *MessageLifecycleTracker) cleanupLoop() {
	for {
		select {
		case <-t.cleanupTicker.C:
			t.cleanup()
		case <-t.stopChan:
			return
		}
	}
}

// cleanup removes old message lifecycles and processed events
func (t *MessageLifecycleTracker) cleanup() {
	now := time.Now()

	// Cleanup old messages
	t.mu.Lock()
	cutoff := now.Add(-t.retentionPeriod)
	for messageID, lifecycle := range t.messages {
		if lifecycle.CreatedAt.Before(cutoff) {
			delete(t.messages, messageID)
		}
	}
	t.mu.Unlock()

	// Cleanup old processed events
	t.dedupMu.Lock()
	dedupCutoff := now.Add(-t.dedupWindow)
	for eventKey, processedAt := range t.processedEvents {
		if processedAt.Before(dedupCutoff) {
			delete(t.processedEvents, eventKey)
		}
	}
	t.dedupMu.Unlock()
}
