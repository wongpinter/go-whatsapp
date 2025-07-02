package webhook

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

// MessagePriority represents message priority levels
type MessagePriority int

const (
	PriorityOTP         MessagePriority = 0  // Highest priority - OTP/Authentication
	PriorityTransactional MessagePriority = 1  // Transaction confirmations
	PriorityService     MessagePriority = 2  // Service notifications
	PriorityUtility     MessagePriority = 5  // General utility messages
	PriorityMarketing   MessagePriority = 10 // Lowest priority - Marketing/Promotional
)

// QueuedMessage represents a message in the queue
type QueuedMessage struct {
	ID          string
	Priority    MessagePriority
	Payload     interface{}
	Timestamp   time.Time
	Retries     int
	MaxRetries  int
	NextRetry   time.Time
	Metadata    map[string]interface{}
}

// MessageQueue implements a priority queue for messages
type MessageQueue struct {
	messages []*QueuedMessage
	mu       sync.RWMutex
}

// Len returns the number of messages in the queue
func (mq *MessageQueue) Len() int {
	return len(mq.messages)
}

// Less compares two messages for priority ordering
func (mq *MessageQueue) Less(i, j int) bool {
	// Lower priority number = higher priority
	if mq.messages[i].Priority != mq.messages[j].Priority {
		return mq.messages[i].Priority < mq.messages[j].Priority
	}
	// If same priority, older messages first
	return mq.messages[i].Timestamp.Before(mq.messages[j].Timestamp)
}

// Swap swaps two messages in the queue
func (mq *MessageQueue) Swap(i, j int) {
	mq.messages[i], mq.messages[j] = mq.messages[j], mq.messages[i]
}

// Push adds a message to the queue
func (mq *MessageQueue) Push(x interface{}) {
	mq.messages = append(mq.messages, x.(*QueuedMessage))
}

// Pop removes and returns the highest priority message
func (mq *MessageQueue) Pop() interface{} {
	old := mq.messages
	n := len(old)
	item := old[n-1]
	mq.messages = old[0 : n-1]
	return item
}

// MessageQueueManager manages message queuing with rate limiting
type MessageQueueManager struct {
	queue       *MessageQueue
	rateLimiter *RateLimiter
	
	// Processing control
	processing  bool
	stopChan    chan struct{}
	mu          sync.RWMutex
	
	// Metrics
	processed   int64
	failed      int64
	retried     int64
	
	// Configuration
	maxWorkers     int
	processingRate time.Duration // Delay between processing messages
}

// NewMessageQueueManager creates a new message queue manager
func NewMessageQueueManager(rateLimiter *RateLimiter) *MessageQueueManager {
	return &MessageQueueManager{
		queue:          &MessageQueue{},
		rateLimiter:    rateLimiter,
		stopChan:       make(chan struct{}),
		maxWorkers:     5,
		processingRate: 100 * time.Millisecond, // 10 messages per second default
	}
}

// Enqueue adds a message to the queue
func (mqm *MessageQueueManager) Enqueue(ctx context.Context, message *QueuedMessage) error {
	mqm.mu.Lock()
	defer mqm.mu.Unlock()
	
	// Set defaults
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	if message.MaxRetries == 0 {
		message.MaxRetries = 3
	}
	if message.Metadata == nil {
		message.Metadata = make(map[string]interface{})
	}
	
	// Add to priority queue
	heap.Push(mqm.queue, message)
	
	return nil
}

// Dequeue removes and returns the highest priority message ready for processing
func (mqm *MessageQueueManager) Dequeue(ctx context.Context) (*QueuedMessage, error) {
	mqm.mu.Lock()
	defer mqm.mu.Unlock()
	
	if mqm.queue.Len() == 0 {
		return nil, nil
	}
	
	// Check if the highest priority message is ready for processing
	topMessage := mqm.queue.messages[0]
	if time.Now().Before(topMessage.NextRetry) {
		return nil, nil // Not ready for retry yet
	}
	
	// Remove from queue
	message := heap.Pop(mqm.queue).(*QueuedMessage)
	return message, nil
}

// StartProcessing starts the message processing workers
func (mqm *MessageQueueManager) StartProcessing(ctx context.Context, processor MessageProcessor) {
	mqm.mu.Lock()
	if mqm.processing {
		mqm.mu.Unlock()
		return
	}
	mqm.processing = true
	mqm.mu.Unlock()
	
	// Start worker goroutines
	for i := 0; i < mqm.maxWorkers; i++ {
		go mqm.worker(ctx, processor)
	}
}

// StopProcessing stops the message processing workers
func (mqm *MessageQueueManager) StopProcessing() {
	mqm.mu.Lock()
	defer mqm.mu.Unlock()
	
	if !mqm.processing {
		return
	}
	
	mqm.processing = false
	close(mqm.stopChan)
}

// MessageProcessor defines the interface for processing messages
type MessageProcessor interface {
	ProcessMessage(ctx context.Context, message *QueuedMessage) error
}

// worker processes messages from the queue
func (mqm *MessageQueueManager) worker(ctx context.Context, processor MessageProcessor) {
	ticker := time.NewTicker(mqm.processingRate)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-mqm.stopChan:
			return
		case <-ticker.C:
			mqm.processNextMessage(ctx, processor)
		}
	}
}

// processNextMessage processes the next message in the queue
func (mqm *MessageQueueManager) processNextMessage(ctx context.Context, processor MessageProcessor) {
	// Check rate limits before processing
	if allowed, waitTime := mqm.rateLimiter.CheckRateLimit(RateLimitTypeThroughput); !allowed {
		if waitTime > 0 {
			time.Sleep(waitTime)
		}
		return
	}
	
	// Get next message
	message, err := mqm.Dequeue(ctx)
	if err != nil || message == nil {
		return
	}
	
	// Process the message
	err = processor.ProcessMessage(ctx, message)
	if err != nil {
		mqm.handleProcessingError(ctx, message, err)
	} else {
		mqm.processed++
	}
}

// handleProcessingError handles errors during message processing
func (mqm *MessageQueueManager) handleProcessingError(ctx context.Context, message *QueuedMessage, err error) {
	message.Retries++
	
	// Check if we should retry
	if message.Retries < message.MaxRetries {
		// Calculate retry delay based on error type and attempt
		var retryDelay time.Duration
		
		// Check if it's a rate limit error
		if rateLimitErr, ok := err.(*RateLimitError); ok {
			delay, retryErr := mqm.rateLimiter.HandleRateLimitError(ctx, rateLimitErr.Code, rateLimitErr.Type)
			if retryErr != nil {
				mqm.failed++
				return
			}
			retryDelay = delay
		} else {
			// Exponential backoff for other errors
			retryDelay = time.Duration(message.Retries) * time.Second
		}
		
		// Schedule retry
		message.NextRetry = time.Now().Add(retryDelay)
		mqm.Enqueue(ctx, message)
		mqm.retried++
	} else {
		// Max retries exceeded
		mqm.failed++
	}
}

// RateLimitError represents a rate limiting error
type RateLimitError struct {
	Code int
	Type RateLimitType
	Message string
}

func (e *RateLimitError) Error() string {
	return e.Message
}

// GetQueueStats returns queue statistics
func (mqm *MessageQueueManager) GetQueueStats() map[string]interface{} {
	mqm.mu.RLock()
	defer mqm.mu.RUnlock()
	
	// Count messages by priority
	priorityCounts := make(map[MessagePriority]int)
	for _, msg := range mqm.queue.messages {
		priorityCounts[msg.Priority]++
	}
	
	return map[string]interface{}{
		"total_queued":    mqm.queue.Len(),
		"processed":       mqm.processed,
		"failed":          mqm.failed,
		"retried":         mqm.retried,
		"processing":      mqm.processing,
		"priority_counts": priorityCounts,
	}
}

// GetMessagesByPriority returns messages of a specific priority
func (mqm *MessageQueueManager) GetMessagesByPriority(priority MessagePriority) []*QueuedMessage {
	mqm.mu.RLock()
	defer mqm.mu.RUnlock()
	
	var messages []*QueuedMessage
	for _, msg := range mqm.queue.messages {
		if msg.Priority == priority {
			// Return a copy
			msgCopy := *msg
			messages = append(messages, &msgCopy)
		}
	}
	return messages
}

// SetProcessingRate sets the rate at which messages are processed
func (mqm *MessageQueueManager) SetProcessingRate(rate time.Duration) {
	mqm.mu.Lock()
	defer mqm.mu.Unlock()
	mqm.processingRate = rate
}

// SetMaxWorkers sets the maximum number of worker goroutines
func (mqm *MessageQueueManager) SetMaxWorkers(workers int) {
	mqm.mu.Lock()
	defer mqm.mu.Unlock()
	mqm.maxWorkers = workers
}

// Clear removes all messages from the queue
func (mqm *MessageQueueManager) Clear() {
	mqm.mu.Lock()
	defer mqm.mu.Unlock()
	mqm.queue.messages = mqm.queue.messages[:0]
}

// GetPriorityName returns a human-readable name for a priority level
func GetPriorityName(priority MessagePriority) string {
	switch priority {
	case PriorityOTP:
		return "OTP/Authentication"
	case PriorityTransactional:
		return "Transactional"
	case PriorityService:
		return "Service"
	case PriorityUtility:
		return "Utility"
	case PriorityMarketing:
		return "Marketing"
	default:
		return "Unknown"
	}
}
