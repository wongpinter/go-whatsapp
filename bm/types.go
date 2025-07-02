package bm

import (
	"fmt"
	"time"
)

// BusinessAccount represents a WhatsApp Business Account.
type BusinessAccount struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	TimezoneID       string `json:"timezone_id"`
	MessageTemplates struct {
		Data []MessageTemplate `json:"data"`
	} `json:"message_templates"`
	PhoneNumbers struct {
		Data []PhoneNumber `json:"data"`
	} `json:"phone_numbers"`
}

// PhoneNumber represents a phone number associated with a WhatsApp Business Account.
type PhoneNumber struct {
	ID                     string `json:"id"`
	DisplayPhoneNumber     string `json:"display_phone_number"`
	VerifiedName           string `json:"verified_name"`
	QualityRating          string `json:"quality_rating"`
	Status                 string `json:"status"`
	CodeVerificationStatus string `json:"code_verification_status,omitempty"`
	Certificate            string `json:"certificate,omitempty"`
}

// PhoneNumberStatus represents the possible statuses of a phone number.
type PhoneNumberStatus string

const (
	StatusConnected    PhoneNumberStatus = "CONNECTED"
	StatusDisconnected PhoneNumberStatus = "DISCONNECTED"
	StatusUnverified   PhoneNumberStatus = "UNVERIFIED"
	StatusPending      PhoneNumberStatus = "PENDING"
	StatusFlagged      PhoneNumberStatus = "FLAGGED"
	StatusRestricted   PhoneNumberStatus = "RESTRICTED"
)

// QualityRating represents the quality rating of a phone number.
type QualityRating string

const (
	QualityGreen   QualityRating = "GREEN"
	QualityYellow  QualityRating = "YELLOW"
	QualityRed     QualityRating = "RED"
	QualityUnknown QualityRating = "UNKNOWN"
)

// BusinessProfile represents the business profile information.
type BusinessProfile struct {
	MessagingProduct  string   `json:"messaging_product"`
	About             string   `json:"about"`
	Address           string   `json:"address"`
	Description       string   `json:"description"`
	Email             string   `json:"email"`
	ProfilePictureURL string   `json:"profile_picture_url"`
	Websites          []string `json:"websites"`
	Vertical          string   `json:"vertical"`
}

// MessageTemplate represents a message template.
type MessageTemplate struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Status     string              `json:"status"`
	Category   string              `json:"category"`
	Language   string              `json:"language"`
	Components []TemplateComponent `json:"components,omitempty"`
}

// TemplateStatus represents the possible statuses of a message template.
type TemplateStatus string

const (
	TemplateStatusApproved TemplateStatus = "APPROVED"
	TemplateStatusPending  TemplateStatus = "PENDING"
	TemplateStatusRejected TemplateStatus = "REJECTED"
	TemplateStatusDisabled TemplateStatus = "DISABLED"
)

// TemplateCategory represents the category of a message template.
type TemplateCategory string

const (
	CategoryMarketing      TemplateCategory = "MARKETING"
	CategoryUtility        TemplateCategory = "UTILITY"
	CategoryAuthentication TemplateCategory = "AUTHENTICATION"
)

// TemplateComponent represents a component of a message template.
type TemplateComponent struct {
	Type       string              `json:"type"`
	Format     string              `json:"format,omitempty"`
	Text       string              `json:"text,omitempty"`
	Example    *TemplateExample    `json:"example,omitempty"`
	Buttons    []TemplateButton    `json:"buttons,omitempty"`
	Parameters []TemplateParameter `json:"parameters,omitempty"`
}

// TemplateExample represents an example for a template component.
type TemplateExample struct {
	HeaderText   []string   `json:"header_text,omitempty"`
	BodyText     [][]string `json:"body_text,omitempty"`
	HeaderHandle []string   `json:"header_handle,omitempty"`
}

// TemplateButton represents a button in a template.
type TemplateButton struct {
	Type        string   `json:"type"`
	Text        string   `json:"text"`
	URL         string   `json:"url,omitempty"`
	PhoneNumber string   `json:"phone_number,omitempty"`
	Example     []string `json:"example,omitempty"`
}

// TemplateParameter represents a parameter in a template.
type TemplateParameter struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// CreateTemplateRequest represents a request to create a new message template.
type CreateTemplateRequest struct {
	Name                string              `json:"name"`
	Language            string              `json:"language"`
	Category            TemplateCategory    `json:"category"`
	AllowCategoryChange bool                `json:"allow_category_change,omitempty"`
	Components          []TemplateComponent `json:"components"`
}

// UpdateTemplateRequest represents a request to update an existing template.
type UpdateTemplateRequest struct {
	Category   *TemplateCategory   `json:"category,omitempty"`
	Components []TemplateComponent `json:"components,omitempty"`
}

// CreateTemplateResponse represents the response from creating a template.
type CreateTemplateResponse struct {
	ID       string           `json:"id"`
	Status   TemplateStatus   `json:"status"`
	Category TemplateCategory `json:"category"`
}

// ListTemplatesResponse represents the response from listing templates.
type ListTemplatesResponse struct {
	Data   []MessageTemplate `json:"data"`
	Paging *PagingInfo       `json:"paging,omitempty"`
}

// PagingInfo represents pagination information.
type PagingInfo struct {
	Cursors *struct {
		Before string `json:"before,omitempty"`
		After  string `json:"after,omitempty"`
	} `json:"cursors,omitempty"`
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
}

// DeleteTemplateResponse represents the response from deleting a template.
type DeleteTemplateResponse struct {
	Success bool `json:"success"`
}

// TemplateComponentType represents the type of a template component.
type TemplateComponentType string

const (
	ComponentTypeHeader  TemplateComponentType = "HEADER"
	ComponentTypeBody    TemplateComponentType = "BODY"
	ComponentTypeFooter  TemplateComponentType = "FOOTER"
	ComponentTypeButtons TemplateComponentType = "BUTTONS"
)

// TemplateFormat represents the format of a template component.
type TemplateFormat string

const (
	FormatText     TemplateFormat = "TEXT"
	FormatImage    TemplateFormat = "IMAGE"
	FormatVideo    TemplateFormat = "VIDEO"
	FormatDocument TemplateFormat = "DOCUMENT"
	FormatLocation TemplateFormat = "LOCATION"
)

// TemplateButtonType represents the type of a template button.
type TemplateButtonType string

const (
	ButtonTypeQuickReply  TemplateButtonType = "QUICK_REPLY"
	ButtonTypeURL         TemplateButtonType = "URL"
	ButtonTypePhoneNumber TemplateButtonType = "PHONE_NUMBER"
)

// TemplateParameterType represents the type of a template parameter.
type TemplateParameterType string

const (
	ParameterTypeText     TemplateParameterType = "text"
	ParameterTypeCurrency TemplateParameterType = "currency"
	ParameterTypeDateTime TemplateParameterType = "date_time"
	ParameterTypeImage    TemplateParameterType = "image"
	ParameterTypeDocument TemplateParameterType = "document"
	ParameterTypeVideo    TemplateParameterType = "video"
)

// TemplateValidationError represents a validation error for a template.
type TemplateValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Error implements the error interface.
func (e *TemplateValidationError) Error() string {
	return fmt.Sprintf("template validation error in field '%s': %s (code: %s)", e.Field, e.Message, e.Code)
}

// TemplateValidationResult represents the result of template validation.
type TemplateValidationResult struct {
	Valid  bool                      `json:"valid"`
	Errors []TemplateValidationError `json:"errors,omitempty"`
}

// SendTemplateRequest represents a request to send a template message.
type SendTemplateRequest struct {
	To       string                 `json:"to"`
	Template TemplateMessagePayload `json:"template"`
}

// TemplateMessagePayload represents the template payload for sending messages.
type TemplateMessagePayload struct {
	Name       string                   `json:"name"`
	Language   TemplateLanguage         `json:"language"`
	Components []TemplateComponentParam `json:"components,omitempty"`
}

// TemplateLanguage represents the language of a template.
type TemplateLanguage struct {
	Code string `json:"code"`
}

// TemplateComponentParam represents a component parameter when sending templates.
type TemplateComponentParam struct {
	Type       TemplateComponentType `json:"type"`
	Parameters []TemplateParameter   `json:"parameters,omitempty"`
	SubType    string                `json:"sub_type,omitempty"`
	Index      int                   `json:"index,omitempty"`
}

// TemplateListParams represents parameters for listing templates.
type TemplateListParams struct {
	Fields   string `json:"fields,omitempty"`
	Status   string `json:"status,omitempty"`
	Category string `json:"category,omitempty"`
	Language string `json:"language,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	After    string `json:"after,omitempty"`
	Before   string `json:"before,omitempty"`
}

// Analytics Types

// CostAnalytics represents cost analytics data.
type CostAnalytics struct {
	MessageCosts      []MessageCostData      `json:"message_costs,omitempty"`
	ConversationCosts []ConversationCostData `json:"conversation_costs,omitempty"`
	TemplateCosts     []TemplateCostData     `json:"template_costs,omitempty"`
	TotalCost         CostSummary            `json:"total_cost"`
	Period            AnalyticsPeriod        `json:"period"`
}

// MessageCostData represents cost data for individual messages.
type MessageCostData struct {
	Date        string  `json:"date"`
	MessageType string  `json:"message_type"`
	Count       int64   `json:"count"`
	Cost        float64 `json:"cost"`
	Currency    string  `json:"currency"`
	Destination string  `json:"destination,omitempty"`
}

// ConversationCostData represents cost data for conversations.
type ConversationCostData struct {
	Date                  string  `json:"date"`
	BusinessInitiated     int64   `json:"business_initiated"`
	UserInitiated         int64   `json:"user_initiated"`
	BusinessInitiatedCost float64 `json:"business_initiated_cost"`
	UserInitiatedCost     float64 `json:"user_initiated_cost"`
	TotalConversations    int64   `json:"total_conversations"`
	TotalCost             float64 `json:"total_cost"`
	Currency              string  `json:"currency"`
}

// TemplateCostData represents cost data for template usage.
type TemplateCostData struct {
	Date         string  `json:"date"`
	TemplateName string  `json:"template_name"`
	TemplateID   string  `json:"template_id"`
	Category     string  `json:"category"`
	Count        int64   `json:"count"`
	Cost         float64 `json:"cost"`
	Currency     string  `json:"currency"`
	SuccessRate  float64 `json:"success_rate"`
}

// CostSummary represents a summary of costs.
type CostSummary struct {
	TotalCost    float64 `json:"total_cost"`
	Currency     string  `json:"currency"`
	MessageCost  float64 `json:"message_cost"`
	TemplateCost float64 `json:"template_cost"`
	Period       string  `json:"period"`
}

// AnalyticsPeriod represents the time period for analytics.
type AnalyticsPeriod struct {
	Start       string `json:"start"`
	End         string `json:"end"`
	Granularity string `json:"granularity"` // HOURLY, DAILY, MONTHLY
}

// AccountQualityMetrics represents account quality and health metrics.
type AccountQualityMetrics struct {
	QualityScore     QualityScore     `json:"quality_score"`
	DeliveryMetrics  DeliveryMetrics  `json:"delivery_metrics"`
	ComplianceStatus ComplianceStatus `json:"compliance_status"`
	Period           AnalyticsPeriod  `json:"period"`
}

// QualityScore represents the account quality score.
type QualityScore struct {
	Current         string                  `json:"current"` // GREEN, YELLOW, RED
	Previous        string                  `json:"previous"`
	Trend           string                  `json:"trend"` // IMPROVING, STABLE, DECLINING
	LastUpdate      time.Time               `json:"last_update"`
	Factors         []QualityFactor         `json:"factors,omitempty"`
	History         []QualityScoreHistory   `json:"history,omitempty"`
	Recommendations []QualityRecommendation `json:"recommendations,omitempty"`
	Score           float64                 `json:"score,omitempty"` // Numeric score 0-100
	Threshold       QualityThresholds       `json:"threshold"`
}

// QualityFactor represents factors affecting quality score.
type QualityFactor struct {
	Factor      string  `json:"factor"`
	Impact      string  `json:"impact"` // HIGH, MEDIUM, LOW
	Description string  `json:"description"`
	Value       float64 `json:"value,omitempty"`
}

// QualityScoreHistory represents historical quality score data.
type QualityScoreHistory struct {
	Date   string         `json:"date"`
	Score  string         `json:"score"` // GREEN, YELLOW, RED
	Value  float64        `json:"value"` // Numeric score 0-100
	Events []QualityEvent `json:"events,omitempty"`
}

// QualityEvent represents events that affected quality score.
type QualityEvent struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"` // POSITIVE, NEGATIVE, NEUTRAL
	Timestamp   time.Time `json:"timestamp"`
}

// QualityRecommendation represents improvement recommendations.
type QualityRecommendation struct {
	Category    string `json:"category"`
	Priority    string `json:"priority"` // HIGH, MEDIUM, LOW
	Title       string `json:"title"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Impact      string `json:"impact"`   // Expected improvement
	Effort      string `json:"effort"`   // Implementation effort
	Timeline    string `json:"timeline"` // Expected timeline
}

// QualityThresholds represents quality score thresholds.
type QualityThresholds struct {
	Green  QualityThreshold `json:"green"`
	Yellow QualityThreshold `json:"yellow"`
	Red    QualityThreshold `json:"red"`
}

// QualityThreshold represents a quality threshold.
type QualityThreshold struct {
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
	Description string  `json:"description"`
}

// DeliveryMetrics represents message delivery performance.
type DeliveryMetrics struct {
	TotalMessages    int64                    `json:"total_messages"`
	DeliveredCount   int64                    `json:"delivered_count"`
	FailedCount      int64                    `json:"failed_count"`
	DeliveryRate     float64                  `json:"delivery_rate"`
	FailureRate      float64                  `json:"failure_rate"`
	AverageLatency   float64                  `json:"average_latency_ms"`
	FailureReasons   []FailureReason          `json:"failure_reasons,omitempty"`
	DeliveryTrends   []DeliveryTrend          `json:"delivery_trends,omitempty"`
	PerformanceGoals DeliveryPerformanceGoals `json:"performance_goals"`
	Benchmarks       DeliveryBenchmarks       `json:"benchmarks"`
}

// FailureReason represents reasons for message delivery failures.
type FailureReason struct {
	Reason          string                  `json:"reason"`
	Count           int64                   `json:"count"`
	Percentage      float64                 `json:"percentage"`
	Description     string                  `json:"description,omitempty"`
	ErrorCode       string                  `json:"error_code,omitempty"`
	Severity        string                  `json:"severity"` // HIGH, MEDIUM, LOW
	Category        string                  `json:"category"` // TECHNICAL, POLICY, USER_ERROR
	Trend           string                  `json:"trend"`    // INCREASING, STABLE, DECREASING
	Recommendations []FailureRecommendation `json:"recommendations,omitempty"`
	FirstOccurred   time.Time               `json:"first_occurred"`
	LastOccurred    time.Time               `json:"last_occurred"`
}

// FailureRecommendation represents recommendations for addressing failures.
type FailureRecommendation struct {
	Action      string `json:"action"`
	Priority    string `json:"priority"` // HIGH, MEDIUM, LOW
	Impact      string `json:"impact"`   // Expected reduction in failure rate
	Effort      string `json:"effort"`   // Implementation effort
	Timeline    string `json:"timeline"` // Expected timeline
	Description string `json:"description"`
}

// DeliveryTrend represents delivery performance trends over time.
type DeliveryTrend struct {
	Date         string  `json:"date"`
	DeliveryRate float64 `json:"delivery_rate"`
	FailureRate  float64 `json:"failure_rate"`
	Volume       int64   `json:"volume"`
	Latency      float64 `json:"latency_ms"`
}

// DeliveryPerformanceGoals represents delivery performance targets.
type DeliveryPerformanceGoals struct {
	TargetDeliveryRate float64         `json:"target_delivery_rate"`
	MaxFailureRate     float64         `json:"max_failure_rate"`
	MaxLatency         float64         `json:"max_latency_ms"`
	MinVolume          int64           `json:"min_volume"`
	Achievement        GoalAchievement `json:"achievement"`
}

// GoalAchievement represents goal achievement status.
type GoalAchievement struct {
	DeliveryRateAchieved bool    `json:"delivery_rate_achieved"`
	FailureRateAchieved  bool    `json:"failure_rate_achieved"`
	LatencyAchieved      bool    `json:"latency_achieved"`
	VolumeAchieved       bool    `json:"volume_achieved"`
	OverallScore         float64 `json:"overall_score"` // 0-100
}

// DeliveryBenchmarks represents industry benchmarks.
type DeliveryBenchmarks struct {
	IndustryAverage    BenchmarkData `json:"industry_average"`
	TopPerformers      BenchmarkData `json:"top_performers"`
	YourPerformance    BenchmarkData `json:"your_performance"`
	PerformanceRanking string        `json:"performance_ranking"` // TOP_10, TOP_25, AVERAGE, BELOW_AVERAGE
}

// BenchmarkData represents benchmark performance data.
type BenchmarkData struct {
	DeliveryRate float64 `json:"delivery_rate"`
	FailureRate  float64 `json:"failure_rate"`
	Latency      float64 `json:"latency_ms"`
	Volume       int64   `json:"volume"`
}

// Advanced Delivery Analytics Types

// DeliveryAnalytics represents comprehensive delivery analytics.
type DeliveryAnalytics struct {
	Summary                 DeliveryAnalyticsSummary `json:"summary"`
	FailureAnalysis         FailureAnalysis          `json:"failure_analysis"`
	OptimizationSuggestions []OptimizationSuggestion `json:"optimization_suggestions"`
	PerformanceInsights     PerformanceInsights      `json:"performance_insights"`
	Period                  AnalyticsPeriod          `json:"period"`
}

// DeliveryAnalyticsSummary represents a summary of delivery analytics.
type DeliveryAnalyticsSummary struct {
	TotalMessages        int64   `json:"total_messages"`
	SuccessfulDeliveries int64   `json:"successful_deliveries"`
	FailedDeliveries     int64   `json:"failed_deliveries"`
	SuccessRate          float64 `json:"success_rate"`
	FailureRate          float64 `json:"failure_rate"`
	AverageLatency       float64 `json:"average_latency_ms"`
	MedianLatency        float64 `json:"median_latency_ms"`
	P95Latency           float64 `json:"p95_latency_ms"`
	P99Latency           float64 `json:"p99_latency_ms"`
}

// FailureAnalysis represents detailed failure analysis.
type FailureAnalysis struct {
	TopFailureReasons  []FailureReason         `json:"top_failure_reasons"`
	FailuresByCategory map[string]int64        `json:"failures_by_category"`
	FailuresBySeverity map[string]int64        `json:"failures_by_severity"`
	FailureTrends      []FailureTrendData      `json:"failure_trends"`
	RecurringIssues    []RecurringIssue        `json:"recurring_issues"`
	ImpactAssessment   FailureImpactAssessment `json:"impact_assessment"`
}

// FailureTrendData represents failure trend information.
type FailureTrendData struct {
	Date          string           `json:"date"`
	TotalFailures int64            `json:"total_failures"`
	FailureRate   float64          `json:"failure_rate"`
	ByCategory    map[string]int64 `json:"by_category"`
	BySeverity    map[string]int64 `json:"by_severity"`
}

// RecurringIssue represents a recurring delivery issue.
type RecurringIssue struct {
	IssueType        string    `json:"issue_type"`
	Frequency        int64     `json:"frequency"`
	AffectedMessages int64     `json:"affected_messages"`
	FirstSeen        time.Time `json:"first_seen"`
	LastSeen         time.Time `json:"last_seen"`
	Pattern          string    `json:"pattern"`
	Severity         string    `json:"severity"`
	Status           string    `json:"status"` // ACTIVE, RESOLVED, INVESTIGATING
}

// FailureImpactAssessment represents the impact of failures.
type FailureImpactAssessment struct {
	BusinessImpact       string  `json:"business_impact"`        // HIGH, MEDIUM, LOW
	UserExperienceImpact string  `json:"user_experience_impact"` // HIGH, MEDIUM, LOW
	RevenueImpact        float64 `json:"revenue_impact"`
	ReputationRisk       string  `json:"reputation_risk"` // HIGH, MEDIUM, LOW
	ComplianceRisk       string  `json:"compliance_risk"` // HIGH, MEDIUM, LOW
}

// OptimizationSuggestion represents delivery optimization suggestions.
type OptimizationSuggestion struct {
	Category       string   `json:"category"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Priority       string   `json:"priority"`        // HIGH, MEDIUM, LOW
	Complexity     string   `json:"complexity"`      // HIGH, MEDIUM, LOW
	ExpectedImpact string   `json:"expected_impact"` // Expected improvement
	Implementation string   `json:"implementation"`  // How to implement
	Timeline       string   `json:"timeline"`        // Expected timeline
	Prerequisites  []string `json:"prerequisites,omitempty"`
	Metrics        []string `json:"metrics"`       // Metrics to track
	EstimatedROI   float64  `json:"estimated_roi"` // Return on investment
}

// PerformanceInsights represents performance insights and patterns.
type PerformanceInsights struct {
	BestPerformingHours    []int                    `json:"best_performing_hours"`
	WorstPerformingHours   []int                    `json:"worst_performing_hours"`
	BestPerformingDays     []int                    `json:"best_performing_days"`
	WorstPerformingDays    []int                    `json:"worst_performing_days"`
	MessageTypePerformance map[string]DeliveryStats `json:"message_type_performance"`
	RegionalPerformance    map[string]DeliveryStats `json:"regional_performance,omitempty"`
	SeasonalPatterns       []SeasonalPattern        `json:"seasonal_patterns,omitempty"`
	AnomalyDetection       []PerformanceAnomaly     `json:"anomaly_detection,omitempty"`
}

// SeasonalPattern represents seasonal performance patterns.
type SeasonalPattern struct {
	Period         string  `json:"period"`     // DAILY, WEEKLY, MONTHLY
	Pattern        string  `json:"pattern"`    // Description of the pattern
	Impact         string  `json:"impact"`     // Impact on performance
	Confidence     float64 `json:"confidence"` // Confidence level 0-100
	Recommendation string  `json:"recommendation"`
}

// PerformanceAnomaly represents detected performance anomalies.
type PerformanceAnomaly struct {
	DetectedAt     time.Time `json:"detected_at"`
	Type           string    `json:"type"`     // SPIKE, DROP, TREND_CHANGE
	Metric         string    `json:"metric"`   // Which metric was affected
	Severity       string    `json:"severity"` // HIGH, MEDIUM, LOW
	Description    string    `json:"description"`
	ExpectedValue  float64   `json:"expected_value"`
	ActualValue    float64   `json:"actual_value"`
	Deviation      float64   `json:"deviation"` // Percentage deviation
	PossibleCauses []string  `json:"possible_causes,omitempty"`
	Status         string    `json:"status"` // INVESTIGATING, RESOLVED, IGNORED
}

// ComplianceStatus represents compliance and policy status.
type ComplianceStatus struct {
	Status           string               `json:"status"` // COMPLIANT, WARNING, VIOLATION
	PolicyViolations []PolicyViolation    `json:"policy_violations,omitempty"`
	Restrictions     []AccountRestriction `json:"restrictions,omitempty"`
	LastReview       time.Time            `json:"last_review"`
}

// PolicyViolation represents a policy violation.
type PolicyViolation struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"` // HIGH, MEDIUM, LOW
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Status      string    `json:"status"` // ACTIVE, RESOLVED, APPEALED
}

// AccountRestriction represents account restrictions.
type AccountRestriction struct {
	Type        string     `json:"type"`
	Description string     `json:"description"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Active      bool       `json:"active"`
}

// Phone Number Analytics Types

// PhoneNumberAnalytics represents comprehensive phone number analytics.
type PhoneNumberAnalytics struct {
	PhoneNumberID      string                   `json:"phone_number_id"`
	DisplayPhoneNumber string                   `json:"display_phone_number"`
	PerformanceMetrics PhoneNumberPerformance   `json:"performance_metrics"`
	StatusInfo         PhoneNumberStatusInfo    `json:"status_info"`
	Configuration      PhoneNumberConfiguration `json:"configuration"`
	Period             AnalyticsPeriod          `json:"period"`
}

// PhoneNumberPerformance represents performance metrics for a phone number.
type PhoneNumberPerformance struct {
	MessageVolume       MessageVolumeMetrics       `json:"message_volume"`
	DeliveryPerformance DeliveryPerformanceMetrics `json:"delivery_performance"`
	UsagePatterns       UsagePatternMetrics        `json:"usage_patterns"`
	CostMetrics         PhoneNumberCostMetrics     `json:"cost_metrics"`
}

// MessageVolumeMetrics represents message volume statistics.
type MessageVolumeMetrics struct {
	TotalMessages    int64             `json:"total_messages"`
	InboundMessages  int64             `json:"inbound_messages"`
	OutboundMessages int64             `json:"outbound_messages"`
	MessagesByType   map[string]int64  `json:"messages_by_type"`
	DailyVolume      []DailyVolumeData `json:"daily_volume,omitempty"`
	PeakHours        []PeakHourData    `json:"peak_hours,omitempty"`
}

// DailyVolumeData represents daily message volume.
type DailyVolumeData struct {
	Date     string `json:"date"`
	Inbound  int64  `json:"inbound"`
	Outbound int64  `json:"outbound"`
	Total    int64  `json:"total"`
}

// PeakHourData represents peak usage hours.
type PeakHourData struct {
	Hour         int   `json:"hour"`
	MessageCount int64 `json:"message_count"`
	DayOfWeek    int   `json:"day_of_week"`
}

// DeliveryPerformanceMetrics represents delivery performance statistics.
type DeliveryPerformanceMetrics struct {
	SuccessRate      float64                  `json:"success_rate"`
	FailureRate      float64                  `json:"failure_rate"`
	AverageLatency   float64                  `json:"average_latency_ms"`
	DeliveryByType   map[string]DeliveryStats `json:"delivery_by_type"`
	FailureBreakdown []FailureBreakdownData   `json:"failure_breakdown,omitempty"`
}

// DeliveryStats represents delivery statistics for a message type.
type DeliveryStats struct {
	Sent      int64   `json:"sent"`
	Delivered int64   `json:"delivered"`
	Failed    int64   `json:"failed"`
	Rate      float64 `json:"rate"`
}

// FailureBreakdownData represents failure analysis data.
type FailureBreakdownData struct {
	ErrorCode   string  `json:"error_code"`
	ErrorType   string  `json:"error_type"`
	Count       int64   `json:"count"`
	Percentage  float64 `json:"percentage"`
	Description string  `json:"description"`
}

// UsagePatternMetrics represents usage pattern analysis.
type UsagePatternMetrics struct {
	ActiveHours      []int                   `json:"active_hours"`
	ActiveDays       []int                   `json:"active_days"`
	ConversationFlow ConversationFlowMetrics `json:"conversation_flow"`
	UserEngagement   UserEngagementMetrics   `json:"user_engagement"`
}

// ConversationFlowMetrics represents conversation flow analysis.
type ConversationFlowMetrics struct {
	AverageConversationLength float64 `json:"average_conversation_length"`
	ConversationStarters      int64   `json:"conversation_starters"`
	ConversationEnders        int64   `json:"conversation_enders"`
	ResponseTime              float64 `json:"average_response_time_minutes"`
}

// UserEngagementMetrics represents user engagement statistics.
type UserEngagementMetrics struct {
	UniqueUsers    int64   `json:"unique_users"`
	ReturningUsers int64   `json:"returning_users"`
	NewUsers       int64   `json:"new_users"`
	EngagementRate float64 `json:"engagement_rate"`
	RetentionRate  float64 `json:"retention_rate"`
}

// PhoneNumberCostMetrics represents cost metrics for a phone number.
type PhoneNumberCostMetrics struct {
	TotalCost      float64            `json:"total_cost"`
	Currency       string             `json:"currency"`
	CostPerMessage float64            `json:"cost_per_message"`
	CostByType     map[string]float64 `json:"cost_by_type"`
	MonthlyCosts   []MonthlyCostData  `json:"monthly_costs,omitempty"`
}

// MonthlyCostData represents monthly cost breakdown.
type MonthlyCostData struct {
	Month    string  `json:"month"`
	Cost     float64 `json:"cost"`
	Messages int64   `json:"messages"`
	AvgCost  float64 `json:"avg_cost_per_message"`
}

// PhoneNumberStatusInfo represents phone number status information.
type PhoneNumberStatusInfo struct {
	Status             string                   `json:"status"`              // CONNECTED, DISCONNECTED, PENDING
	VerificationStatus string                   `json:"verification_status"` // VERIFIED, UNVERIFIED, PENDING
	QualityRating      string                   `json:"quality_rating"`      // HIGH, MEDIUM, LOW
	HealthStatus       PhoneNumberHealth        `json:"health_status"`
	Capabilities       []string                 `json:"capabilities"`
	Restrictions       []PhoneNumberRestriction `json:"restrictions,omitempty"`
	LastStatusUpdate   time.Time                `json:"last_status_update"`
}

// PhoneNumberHealth represents health status details.
type PhoneNumberHealth struct {
	Overall   string        `json:"overall"` // HEALTHY, WARNING, CRITICAL
	Issues    []HealthIssue `json:"issues,omitempty"`
	LastCheck time.Time     `json:"last_check"`
	Metrics   HealthMetrics `json:"metrics"`
}

// HealthIssue represents a health issue.
type HealthIssue struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	DetectedAt  time.Time `json:"detected_at"`
	Status      string    `json:"status"` // ACTIVE, RESOLVED, INVESTIGATING
}

// HealthMetrics represents health-related metrics.
type HealthMetrics struct {
	UptimePercentage  float64 `json:"uptime_percentage"`
	ErrorRate         float64 `json:"error_rate"`
	ResponseTime      float64 `json:"response_time_ms"`
	ThroughputLimit   int64   `json:"throughput_limit"`
	CurrentThroughput int64   `json:"current_throughput"`
}

// PhoneNumberRestriction represents restrictions on a phone number.
type PhoneNumberRestriction struct {
	Type        string     `json:"type"`
	Description string     `json:"description"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Active      bool       `json:"active"`
	Reason      string     `json:"reason,omitempty"`
}

// PhoneNumberConfiguration represents phone number configuration.
type PhoneNumberConfiguration struct {
	DisplayName   string                `json:"display_name"`
	AboutText     string                `json:"about_text"`
	ProfilePhoto  *ProfilePhoto         `json:"profile_photo,omitempty"`
	BusinessHours *BusinessHours        `json:"business_hours,omitempty"`
	AutoReply     *AutoReplySettings    `json:"auto_reply,omitempty"`
	Webhooks      *WebhookConfiguration `json:"webhooks,omitempty"`
}

// ProfilePhoto represents profile photo information.
type ProfilePhoto struct {
	URL         string    `json:"url"`
	ID          string    `json:"id,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
}

// BusinessHours represents business hours configuration.
type BusinessHours struct {
	Timezone string        `json:"timezone"`
	Schedule []DaySchedule `json:"schedule"`
	Holidays []Holiday     `json:"holidays,omitempty"`
	Enabled  bool          `json:"enabled"`
}

// DaySchedule represents schedule for a day of the week.
type DaySchedule struct {
	DayOfWeek int        `json:"day_of_week"` // 0 = Sunday, 1 = Monday, etc.
	IsOpen    bool       `json:"is_open"`
	Hours     []TimeSlot `json:"hours,omitempty"`
}

// TimeSlot represents a time slot.
type TimeSlot struct {
	Start string `json:"start"` // HH:MM format
	End   string `json:"end"`   // HH:MM format
}

// Holiday represents a holiday.
type Holiday struct {
	Date        string `json:"date"` // YYYY-MM-DD format
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// AutoReplySettings represents auto-reply configuration.
type AutoReplySettings struct {
	Enabled        bool   `json:"enabled"`
	Message        string `json:"message"`
	OutOfHoursOnly bool   `json:"out_of_hours_only"`
	FirstTimeOnly  bool   `json:"first_time_only"`
}

// WebhookConfiguration represents webhook configuration.
type WebhookConfiguration struct {
	URL         string    `json:"url"`
	VerifyToken string    `json:"verify_token"`
	Events      []string  `json:"events"`
	Enabled     bool      `json:"enabled"`
	LastUpdated time.Time `json:"last_updated"`
}

// Analytics Request/Response Types

// AnalyticsRequest represents a request for analytics data.
type AnalyticsRequest struct {
	Start          string   `json:"start"`                 // YYYY-MM-DD format
	End            string   `json:"end"`                   // YYYY-MM-DD format
	Granularity    string   `json:"granularity,omitempty"` // HOURLY, DAILY, MONTHLY
	MetricTypes    []string `json:"metric_types,omitempty"`
	PhoneNumberIDs []string `json:"phone_number_ids,omitempty"`
	ProductTypes   []string `json:"product_types,omitempty"`
}

// AnalyticsResponse represents the response from analytics API.
type AnalyticsResponse struct {
	Data   []AnalyticsDataPoint `json:"data"`
	Paging *PagingInfo          `json:"paging,omitempty"`
}

// AnalyticsDataPoint represents a single analytics data point.
type AnalyticsDataPoint struct {
	Name        string           `json:"name"`
	Period      string           `json:"period"`
	Values      []AnalyticsValue `json:"values"`
	Title       string           `json:"title,omitempty"`
	Description string           `json:"description,omitempty"`
}

// AnalyticsValue represents a value in analytics data.
type AnalyticsValue struct {
	Value     interface{} `json:"value"`
	EndTime   string      `json:"end_time"`
	StartTime string      `json:"start_time,omitempty"`
}

// Business Profile Types

// BusinessProfileInfo represents a business profile information.
type BusinessProfileInfo struct {
	About             string   `json:"about,omitempty"`
	Address           string   `json:"address,omitempty"`
	Description       string   `json:"description,omitempty"`
	Email             string   `json:"email,omitempty"`
	ProfilePictureURL string   `json:"profile_picture_url,omitempty"`
	Websites          []string `json:"websites,omitempty"`
	Vertical          string   `json:"vertical,omitempty"`
}

// UpdateBusinessProfileRequest represents a request to update business profile.
type UpdateBusinessProfileRequest struct {
	About                *string  `json:"about,omitempty"`
	Address              *string  `json:"address,omitempty"`
	Description          *string  `json:"description,omitempty"`
	Email                *string  `json:"email,omitempty"`
	ProfilePictureHandle string   `json:"profile_picture_handle,omitempty"`
	Websites             []string `json:"websites,omitempty"`
	Vertical             *string  `json:"vertical,omitempty"`
}

// Analytics Options

// AnalyticsOption is a functional option for analytics requests.
type AnalyticsOption func(*AnalyticsRequest)

// WithAnalyticsGranularity sets the granularity for analytics.
func WithAnalyticsGranularity(granularity string) AnalyticsOption {
	return func(r *AnalyticsRequest) {
		r.Granularity = granularity
	}
}

// WithAnalyticsMetricTypes sets the metric types for analytics.
func WithAnalyticsMetricTypes(metricTypes ...string) AnalyticsOption {
	return func(r *AnalyticsRequest) {
		r.MetricTypes = metricTypes
	}
}

// WithAnalyticsPhoneNumbers sets the phone number IDs for analytics.
func WithAnalyticsPhoneNumbers(phoneNumberIDs ...string) AnalyticsOption {
	return func(r *AnalyticsRequest) {
		r.PhoneNumberIDs = phoneNumberIDs
	}
}

// WithAnalyticsProductTypes sets the product types for analytics.
func WithAnalyticsProductTypes(productTypes ...string) AnalyticsOption {
	return func(r *AnalyticsRequest) {
		r.ProductTypes = productTypes
	}
}

// Constants for analytics

// Granularity constants
const (
	GranularityHourly  = "HOURLY"
	GranularityDaily   = "DAILY"
	GranularityMonthly = "MONTHLY"
)

// Metric type constants
const (
	MetricTypeCost                = "cost"
	MetricTypeConversation        = "conversation"
	MetricTypeMessage             = "message"
	MetricTypePhoneNumberInsights = "phone_number_insights"
)

// Product type constants
const (
	ProductTypeWhatsApp = "whatsapp"
)

// Quality score constants
const (
	QualityScoreGreen  = "GREEN"
	QualityScoreYellow = "YELLOW"
	QualityScoreRed    = "RED"
)

// Quality trend constants
const (
	QualityTrendImproving = "IMPROVING"
	QualityTrendStable    = "STABLE"
	QualityTrendDeclining = "DECLINING"
)

// Compliance status constants
const (
	ComplianceStatusCompliant = "COMPLIANT"
	ComplianceStatusWarning   = "WARNING"
	ComplianceStatusViolation = "VIOLATION"
)

// Phone number health constants
const (
	HealthStatusHealthy  = "HEALTHY"
	HealthStatusWarning  = "WARNING"
	HealthStatusCritical = "CRITICAL"
)

// TemplateListOption is a functional option for listing templates.
type TemplateListOption func(*TemplateListParams)

// WithTemplateFields sets the fields to retrieve for templates.
func WithTemplateFields(fields ...string) TemplateListOption {
	return func(p *TemplateListParams) {
		fieldsStr := ""
		for i, field := range fields {
			if i > 0 {
				fieldsStr += ","
			}
			fieldsStr += field
		}
		p.Fields = fieldsStr
	}
}

// WithTemplateStatus filters templates by status.
func WithTemplateStatus(status TemplateStatus) TemplateListOption {
	return func(p *TemplateListParams) {
		p.Status = string(status)
	}
}

// WithTemplateCategory filters templates by category.
func WithTemplateCategory(category TemplateCategory) TemplateListOption {
	return func(p *TemplateListParams) {
		p.Category = string(category)
	}
}

// WithTemplateLanguage filters templates by language.
func WithTemplateLanguage(language string) TemplateListOption {
	return func(p *TemplateListParams) {
		p.Language = language
	}
}

// WithTemplateLimit sets the maximum number of templates to retrieve.
func WithTemplateLimit(limit int) TemplateListOption {
	return func(p *TemplateListParams) {
		p.Limit = limit
	}
}

// WithTemplatePagination sets pagination cursors.
func WithTemplatePagination(after, before string) TemplateListOption {
	return func(p *TemplateListParams) {
		p.After = after
		p.Before = before
	}
}

// ConversationAnalytics represents conversation analytics data.
type ConversationAnalytics struct {
	Data   []DataPoint `json:"data"`
	Paging *Paging     `json:"paging,omitempty"`
}

// DataPoint represents a single data point in analytics.
type DataPoint struct {
	ConversationType      string  `json:"conversation_type"`
	ConversationDirection string  `json:"conversation_direction"`
	Country               string  `json:"country"`
	PhoneNumber           string  `json:"phone_number"`
	Cost                  float64 `json:"cost"`
	ConversationCount     int     `json:"conversation_count"`
	DataPoint             string  `json:"data_point"`
	StartTime             string  `json:"start"`
	EndTime               string  `json:"end"`
}

// GetStartTime returns the start time as a time.Time.
func (d *DataPoint) GetStartTime() (time.Time, error) {
	return time.Parse(time.RFC3339, d.StartTime)
}

// GetEndTime returns the end time as a time.Time.
func (d *DataPoint) GetEndTime() (time.Time, error) {
	return time.Parse(time.RFC3339, d.EndTime)
}

// Paging represents pagination information.
type Paging struct {
	Cursors  *Cursors `json:"cursors,omitempty"`
	Next     string   `json:"next,omitempty"`
	Previous string   `json:"previous,omitempty"`
}

// Cursors represents pagination cursors.
type Cursors struct {
	Before string `json:"before"`
	After  string `json:"after"`
}

// ConversationType represents the type of conversation.
type ConversationType string

const (
	ConversationTypeUserInitiated      ConversationType = "USER_INITIATED"
	ConversationTypeBusinessInitiated  ConversationType = "BUSINESS_INITIATED"
	ConversationTypeReferralConversion ConversationType = "REFERRAL_CONVERSION"
)

// ConversationDirection represents the direction of conversation.
type ConversationDirection string

const (
	DirectionInbound  ConversationDirection = "INBOUND"
	DirectionOutbound ConversationDirection = "OUTBOUND"
)

// MediaUploadResponse represents the response from media upload.
type MediaUploadResponse struct {
	ID string `json:"id"`
}

// MediaInfo represents information about uploaded media.
type MediaInfo struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	SHA256   string `json:"sha256"`
	FileSize int64  `json:"file_size"`
}

// BusinessProfileUpdateRequest represents a request to update business profile.
type BusinessProfileUpdateRequest struct {
	MessagingProduct  string   `json:"messaging_product"`
	About             string   `json:"about,omitempty"`
	Address           string   `json:"address,omitempty"`
	Description       string   `json:"description,omitempty"`
	Email             string   `json:"email,omitempty"`
	ProfilePictureURL string   `json:"profile_picture_url,omitempty"`
	Websites          []string `json:"websites,omitempty"`
	Vertical          string   `json:"vertical,omitempty"`
}

// TemplateCreateRequest represents a request to create a message template.
type TemplateCreateRequest struct {
	Name       string              `json:"name"`
	Category   TemplateCategory    `json:"category"`
	Language   string              `json:"language"`
	Components []TemplateComponent `json:"components"`
}

// TemplateDeleteResponse represents the response from deleting a template.
type TemplateDeleteResponse struct {
	Success bool `json:"success"`
}

// PhoneNumberInfo represents detailed information about a phone number.
type PhoneNumberInfo struct {
	PhoneNumber
	TwoStepVerificationStatus string `json:"two_step_verification_status,omitempty"`
	AccountMode               string `json:"account_mode,omitempty"`
	CertificateStatus         string `json:"certificate_status,omitempty"`
	NameStatus                string `json:"name_status,omitempty"`
	NewNameStatus             string `json:"new_name_status,omitempty"`
	SearchVisibility          string `json:"search_visibility,omitempty"`
	IsOfficialBusinessAccount bool   `json:"is_official_business_account,omitempty"`
}

// BusinessVertical represents business verticals.
type BusinessVertical string

const (
	VerticalUndefined     BusinessVertical = "UNDEFINED"
	VerticalOther         BusinessVertical = "OTHER"
	VerticalAutoMotive    BusinessVertical = "AUTO"
	VerticalBeauty        BusinessVertical = "BEAUTY"
	VerticalApparel       BusinessVertical = "APPAREL"
	VerticalEducation     BusinessVertical = "EDU"
	VerticalEntertainment BusinessVertical = "ENTERTAIN"
	VerticalEventPlanning BusinessVertical = "EVENT_PLAN"
	VerticalFinance       BusinessVertical = "FINANCE"
	VerticalGrocery       BusinessVertical = "GROCERY"
	VerticalGovernment    BusinessVertical = "GOVT"
	VerticalHotel         BusinessVertical = "HOTEL"
	VerticalHealth        BusinessVertical = "HEALTH"
	VerticalNonProfit     BusinessVertical = "NONPROFIT"
	VerticalProfessional  BusinessVertical = "PROF_SERVICES"
	VerticalShopping      BusinessVertical = "SHOPPING"
	VerticalTravel        BusinessVertical = "TRAVEL"
	VerticalRestaurant    BusinessVertical = "RESTAURANT"
)

// IsValidStatus checks if a phone number status is valid.
func (s PhoneNumberStatus) IsValid() bool {
	switch s {
	case StatusConnected, StatusDisconnected, StatusUnverified, StatusPending, StatusFlagged, StatusRestricted:
		return true
	default:
		return false
	}
}

// IsValidQualityRating checks if a quality rating is valid.
func (q QualityRating) IsValid() bool {
	switch q {
	case QualityGreen, QualityYellow, QualityRed, QualityUnknown:
		return true
	default:
		return false
	}
}

// IsValidTemplateStatus checks if a template status is valid.
func (s TemplateStatus) IsValid() bool {
	switch s {
	case TemplateStatusApproved, TemplateStatusPending, TemplateStatusRejected, TemplateStatusDisabled:
		return true
	default:
		return false
	}
}

// IsValidTemplateCategory checks if a template category is valid.
func (c TemplateCategory) IsValid() bool {
	switch c {
	case CategoryMarketing, CategoryUtility, CategoryAuthentication:
		return true
	default:
		return false
	}
}

// MessageMetrics represents message analytics data.
type MessageMetrics struct {
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Sent      int       `json:"sent"`
	Delivered int       `json:"delivered"`
	Read      int       `json:"read"`
	Failed    int       `json:"failed"`
}

// GetStartTime returns the start time as a time.Time.
func (m *MessageMetrics) GetStartTime() time.Time {
	return m.Start
}

// GetEndTime returns the end time as a time.Time.
func (m *MessageMetrics) GetEndTime() time.Time {
	return m.End
}

// DeliveryRate calculates the delivery rate as a percentage.
func (m *MessageMetrics) DeliveryRate() float64 {
	if m.Sent == 0 {
		return 0
	}
	return float64(m.Delivered) / float64(m.Sent) * 100
}

// ReadRate calculates the read rate as a percentage.
func (m *MessageMetrics) ReadRate() float64 {
	if m.Delivered == 0 {
		return 0
	}
	return float64(m.Read) / float64(m.Delivered) * 100
}

// FailureRate calculates the failure rate as a percentage.
func (m *MessageMetrics) FailureRate() float64 {
	if m.Sent == 0 {
		return 0
	}
	return float64(m.Failed) / float64(m.Sent) * 100
}
