package webhook

import (
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector collects metrics for webhook processing
type MetricsCollector struct {
	// Request metrics
	totalRequests     uint64
	successfulEvents  uint64
	failedEvents      uint64
	processingLatency atomic.Value // *LatencyStats

	// Rate metrics
	requestsPerMinute atomic.Value // *RateStats
	eventsPerMinute   atomic.Value // *RateStats

	// Status metrics
	statusCounts map[string]uint64
	statusMu     sync.RWMutex

	// Message type metrics
	messageTypeCounts map[string]uint64
	messageTypeMu     sync.RWMutex

	// Reset timer
	resetTicker *time.Ticker
	stopChan    chan struct{}
}

// LatencyStats tracks latency statistics
type LatencyStats struct {
	Min     time.Duration
	Max     time.Duration
	Average time.Duration
	LastAt  time.Time
}

// RateStats tracks rate statistics
type RateStats struct {
	Current   uint64
	Average   float64
	Peak      uint64
	LastReset time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	m := &MetricsCollector{
		statusCounts:      make(map[string]uint64),
		messageTypeCounts: make(map[string]uint64),
		resetTicker:       time.NewTicker(time.Minute),
		stopChan:          make(chan struct{}),
	}

	m.processingLatency.Store(&LatencyStats{})
	m.requestsPerMinute.Store(&RateStats{LastReset: time.Now()})
	m.eventsPerMinute.Store(&RateStats{LastReset: time.Now()})

	go m.resetLoop()
	return m
}

// RecordRequest records metrics for a webhook request
func (m *MetricsCollector) RecordRequest(start time.Time, successful bool, eventCount int) {
	atomic.AddUint64(&m.totalRequests, 1)
	if successful {
		atomic.AddUint64(&m.successfulEvents, uint64(eventCount))
	} else {
		atomic.AddUint64(&m.failedEvents, uint64(eventCount))
	}

	// Update latency stats
	latency := time.Since(start)
	m.updateLatency(latency)

	// Update rate stats
	m.updateRate(m.requestsPerMinute.Load().(*RateStats), 1)
	m.updateRate(m.eventsPerMinute.Load().(*RateStats), uint64(eventCount))
}

// RecordStatus records metrics for a message status
func (m *MetricsCollector) RecordStatus(status string) {
	m.statusMu.Lock()
	m.statusCounts[status]++
	m.statusMu.Unlock()
}

// RecordMessageType records metrics for a message type
func (m *MetricsCollector) RecordMessageType(messageType string) {
	m.messageTypeMu.Lock()
	m.messageTypeCounts[messageType]++
	m.messageTypeMu.Unlock()
}

// GetMetrics returns the current metrics
func (m *MetricsCollector) GetMetrics() map[string]interface{} {
	latency := m.processingLatency.Load().(*LatencyStats)
	reqRate := m.requestsPerMinute.Load().(*RateStats)
	eventRate := m.eventsPerMinute.Load().(*RateStats)

	m.statusMu.RLock()
	statusCounts := make(map[string]uint64)
	for k, v := range m.statusCounts {
		statusCounts[k] = v
	}
	m.statusMu.RUnlock()

	m.messageTypeMu.RLock()
	messageTypeCounts := make(map[string]uint64)
	for k, v := range m.messageTypeCounts {
		messageTypeCounts[k] = v
	}
	m.messageTypeMu.RUnlock()

	return map[string]interface{}{
		"total_requests":    atomic.LoadUint64(&m.totalRequests),
		"successful_events": atomic.LoadUint64(&m.successfulEvents),
		"failed_events":     atomic.LoadUint64(&m.failedEvents),
		"latency": map[string]interface{}{
			"min":     latency.Min.String(),
			"max":     latency.Max.String(),
			"average": latency.Average.String(),
			"last_at": latency.LastAt,
		},
		"rates": map[string]interface{}{
			"requests_per_minute": map[string]interface{}{
				"current":    reqRate.Current,
				"average":    reqRate.Average,
				"peak":       reqRate.Peak,
				"last_reset": reqRate.LastReset,
			},
			"events_per_minute": map[string]interface{}{
				"current":    eventRate.Current,
				"average":    eventRate.Average,
				"peak":       eventRate.Peak,
				"last_reset": eventRate.LastReset,
			},
		},
		"status_counts":       statusCounts,
		"message_type_counts": messageTypeCounts,
	}
}

// GetStatusMetrics returns status-specific metrics
func (m *MetricsCollector) GetStatusMetrics() map[string]uint64 {
	m.statusMu.RLock()
	defer m.statusMu.RUnlock()

	statusCounts := make(map[string]uint64)
	for k, v := range m.statusCounts {
		statusCounts[k] = v
	}
	return statusCounts
}

// GetMessageTypeMetrics returns message type-specific metrics
func (m *MetricsCollector) GetMessageTypeMetrics() map[string]uint64 {
	m.messageTypeMu.RLock()
	defer m.messageTypeMu.RUnlock()

	messageTypeCounts := make(map[string]uint64)
	for k, v := range m.messageTypeCounts {
		messageTypeCounts[k] = v
	}
	return messageTypeCounts
}

// Stop stops the metrics collector
func (m *MetricsCollector) Stop() {
	close(m.stopChan)
	m.resetTicker.Stop()
}

// resetLoop periodically resets rate statistics
func (m *MetricsCollector) resetLoop() {
	for {
		select {
		case <-m.resetTicker.C:
			m.resetRates()
		case <-m.stopChan:
			return
		}
	}
}

// resetRates resets the rate statistics
func (m *MetricsCollector) resetRates() {
	reqRate := m.requestsPerMinute.Load().(*RateStats)
	m.requestsPerMinute.Store(&RateStats{
		Average:   (float64(reqRate.Current) + reqRate.Average) / 2,
		Peak:      max(reqRate.Peak, reqRate.Current),
		LastReset: time.Now(),
	})

	eventRate := m.eventsPerMinute.Load().(*RateStats)
	m.eventsPerMinute.Store(&RateStats{
		Average:   (float64(eventRate.Current) + eventRate.Average) / 2,
		Peak:      max(eventRate.Peak, eventRate.Current),
		LastReset: time.Now(),
	})
}

// updateLatency updates latency statistics
func (m *MetricsCollector) updateLatency(latency time.Duration) {
	for {
		old := m.processingLatency.Load().(*LatencyStats)
		new := &LatencyStats{
			Min:     minDuration(old.Min, latency),
			Max:     maxDuration(old.Max, latency),
			Average: (old.Average + latency) / 2,
			LastAt:  time.Now(),
		}
		if old.Min == 0 {
			new.Min = latency
		}
		if m.processingLatency.CompareAndSwap(old, new) {
			break
		}
	}
}

// updateRate updates a rate statistic
func (m *MetricsCollector) updateRate(stats *RateStats, count uint64) {
	atomic.AddUint64(&stats.Current, count)
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func minDuration(a, b time.Duration) time.Duration {
	if a == 0 || b < a {
		return b
	}
	return a
}

func maxDuration(a, b time.Duration) time.Duration {
	if b > a {
		return b
	}
	return a
}
