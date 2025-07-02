package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/wongpinter/go-whatsapp/bm"
)

func main() {
	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Get configuration from environment
	wabaID := os.Getenv("WABA_ID")
	accessToken := os.Getenv("ACCESS_TOKEN")
	phoneNumberID := os.Getenv("PHONE_NUMBER_ID")

	if wabaID == "" || accessToken == "" {
		log.Fatal("Please set WABA_ID and ACCESS_TOKEN environment variables")
	}

	// Initialize Business Management client
	bmClient := bm.NewClient(accessToken, bm.WithWABAID(wabaID), bm.WithLogger(logger))

	ctx := context.Background()

	// Example 1: Get cost analytics
	fmt.Println("=== Cost Analytics ===")
	if err := getCostAnalytics(ctx, bmClient); err != nil {
		logger.Error().Err(err).Msg("Failed to get cost analytics")
	}

	// Example 2: Get template usage cost analytics
	fmt.Println("\n=== Template Usage Cost Analytics ===")
	if err := getTemplateUsageCostAnalytics(ctx, bmClient); err != nil {
		logger.Error().Err(err).Msg("Failed to get template usage cost analytics")
	}

	// Example 3: Get compliance monitoring
	fmt.Println("\n=== Compliance Monitoring ===")
	if err := getComplianceMonitoring(ctx, bmClient); err != nil {
		logger.Error().Err(err).Msg("Failed to get compliance monitoring")
	}

	// Example 4: Get account quality metrics
	fmt.Println("\n=== Account Quality Metrics ===")
	if err := getAccountQualityMetrics(ctx, bmClient); err != nil {
		logger.Error().Err(err).Msg("Failed to get account quality metrics")
	}

	// Example 3: Get phone number analytics (if phone number ID is provided)
	if phoneNumberID != "" {
		fmt.Println("\n=== Phone Number Analytics ===")
		if err := getPhoneNumberAnalytics(ctx, bmClient, phoneNumberID); err != nil {
			logger.Error().Err(err).Msg("Failed to get phone number analytics")
		}

		fmt.Println("\n=== Phone Number Status ===")
		if err := getPhoneNumberStatus(ctx, bmClient, phoneNumberID); err != nil {
			logger.Error().Err(err).Msg("Failed to get phone number status")
		}
	}

	// Example 4: Advanced quality monitoring
	fmt.Println("\n=== Advanced Quality Monitoring ===")
	if err := getAdvancedQualityMonitoring(ctx, bmClient, phoneNumberID); err != nil {
		logger.Error().Err(err).Msg("Failed to get advanced quality monitoring")
	}

	// Example 5: Advanced delivery analytics (if phone number ID is provided)
	if phoneNumberID != "" {
		fmt.Println("\n=== Advanced Delivery Analytics ===")
		if err := getAdvancedDeliveryAnalytics(ctx, bmClient, phoneNumberID); err != nil {
			logger.Error().Err(err).Msg("Failed to get advanced delivery analytics")
		}

		fmt.Println("\n=== Advanced Phone Number Management ===")
		if err := getAdvancedPhoneNumberManagement(ctx, bmClient, phoneNumberID); err != nil {
			logger.Error().Err(err).Msg("Failed to get advanced phone number management")
		}
	}

	// Example 5: Comprehensive analytics report
	fmt.Println("\n=== Comprehensive Analytics Report ===")
	if err := generateAnalyticsReport(ctx, bmClient, phoneNumberID); err != nil {
		logger.Error().Err(err).Msg("Failed to generate analytics report")
	}
}

// getCostAnalytics demonstrates retrieving cost analytics.
func getCostAnalytics(ctx context.Context, client *bm.Client) error {
	// Get cost analytics for the last 30 days
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

	costAnalytics, err := client.GetCostAnalytics(ctx, startDate, endDate,
		bm.WithAnalyticsGranularity(bm.GranularityDaily),
		bm.WithAnalyticsMetricTypes(bm.MetricTypeCost),
	)
	if err != nil {
		return fmt.Errorf("failed to get cost analytics: %w", err)
	}

	fmt.Printf("Cost Analytics (%s to %s):\n", startDate, endDate)
	fmt.Printf("  Total Cost: %.2f %s\n", costAnalytics.TotalCost.TotalCost, costAnalytics.TotalCost.Currency)
	fmt.Printf("  Message Cost: %.2f %s\n", costAnalytics.TotalCost.MessageCost, costAnalytics.TotalCost.Currency)
	fmt.Printf("  Template Cost: %.2f %s\n", costAnalytics.TotalCost.TemplateCost, costAnalytics.TotalCost.Currency)
	fmt.Printf("  Period: %s\n", costAnalytics.TotalCost.Period)

	if len(costAnalytics.MessageCosts) > 0 {
		fmt.Println("  Message Cost Breakdown:")
		for _, messageCost := range costAnalytics.MessageCosts {
			fmt.Printf("    %s - %s: %d messages, %.2f %s\n",
				messageCost.Date, messageCost.MessageType, messageCost.Count, messageCost.Cost, messageCost.Currency)
		}
	}

	if len(costAnalytics.ConversationCosts) > 0 {
		fmt.Println("  Conversation Cost Breakdown:")
		for _, convCost := range costAnalytics.ConversationCosts {
			fmt.Printf("    %s: %d total conversations, %.2f %s\n",
				convCost.Date, convCost.TotalConversations, convCost.TotalCost, convCost.Currency)
			fmt.Printf("      Business-initiated: %d (%.2f %s)\n",
				convCost.BusinessInitiated, convCost.BusinessInitiatedCost, convCost.Currency)
			fmt.Printf("      User-initiated: %d (%.2f %s)\n",
				convCost.UserInitiated, convCost.UserInitiatedCost, convCost.Currency)
		}
	}

	return nil
}

// getAccountQualityMetrics demonstrates retrieving account quality metrics.
func getAccountQualityMetrics(ctx context.Context, client *bm.Client) error {
	// Get quality metrics for the last 7 days
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	qualityMetrics, err := client.GetAccountQualityMetrics(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get account quality metrics: %w", err)
	}

	fmt.Printf("Account Quality Metrics (%s to %s):\n", startDate, endDate)

	// Quality Score
	fmt.Printf("  Quality Score: %s (Previous: %s, Trend: %s)\n",
		qualityMetrics.QualityScore.Current,
		qualityMetrics.QualityScore.Previous,
		qualityMetrics.QualityScore.Trend)
	fmt.Printf("  Last Updated: %s\n", qualityMetrics.QualityScore.LastUpdate.Format("2006-01-02 15:04:05"))

	if len(qualityMetrics.QualityScore.Factors) > 0 {
		fmt.Println("  Quality Factors:")
		for _, factor := range qualityMetrics.QualityScore.Factors {
			fmt.Printf("    %s (%s impact): %s\n", factor.Factor, factor.Impact, factor.Description)
		}
	}

	// Delivery Metrics
	fmt.Printf("  Delivery Performance:\n")
	fmt.Printf("    Total Messages: %d\n", qualityMetrics.DeliveryMetrics.TotalMessages)
	fmt.Printf("    Delivered: %d (%.2f%%)\n", qualityMetrics.DeliveryMetrics.DeliveredCount, qualityMetrics.DeliveryMetrics.DeliveryRate)
	fmt.Printf("    Failed: %d (%.2f%%)\n", qualityMetrics.DeliveryMetrics.FailedCount, qualityMetrics.DeliveryMetrics.FailureRate)
	fmt.Printf("    Average Latency: %.2f ms\n", qualityMetrics.DeliveryMetrics.AverageLatency)

	if len(qualityMetrics.DeliveryMetrics.FailureReasons) > 0 {
		fmt.Println("    Failure Reasons:")
		for _, reason := range qualityMetrics.DeliveryMetrics.FailureReasons {
			fmt.Printf("      %s: %d (%.2f%%) - %s\n",
				reason.Reason, reason.Count, reason.Percentage, reason.Description)
		}
	}

	// Compliance Status
	fmt.Printf("  Compliance Status: %s\n", qualityMetrics.ComplianceStatus.Status)
	fmt.Printf("  Last Review: %s\n", qualityMetrics.ComplianceStatus.LastReview.Format("2006-01-02 15:04:05"))

	if len(qualityMetrics.ComplianceStatus.PolicyViolations) > 0 {
		fmt.Println("  Policy Violations:")
		for _, violation := range qualityMetrics.ComplianceStatus.PolicyViolations {
			fmt.Printf("    %s (%s): %s [%s]\n",
				violation.Type, violation.Severity, violation.Description, violation.Status)
		}
	}

	if len(qualityMetrics.ComplianceStatus.Restrictions) > 0 {
		fmt.Println("  Account Restrictions:")
		for _, restriction := range qualityMetrics.ComplianceStatus.Restrictions {
			status := "Inactive"
			if restriction.Active {
				status = "Active"
			}
			fmt.Printf("    %s: %s [%s]\n", restriction.Type, restriction.Description, status)
		}
	}

	return nil
}

// getPhoneNumberAnalytics demonstrates retrieving phone number analytics.
func getPhoneNumberAnalytics(ctx context.Context, client *bm.Client, phoneNumberID string) error {
	// Get analytics for the last 30 days
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

	analytics, err := client.GetPhoneNumberAnalytics(ctx, phoneNumberID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get phone number analytics: %w", err)
	}

	fmt.Printf("Phone Number Analytics for %s (%s to %s):\n",
		analytics.DisplayPhoneNumber, startDate, endDate)

	// Message Volume
	volume := analytics.PerformanceMetrics.MessageVolume
	fmt.Printf("  Message Volume:\n")
	fmt.Printf("    Total Messages: %d\n", volume.TotalMessages)
	fmt.Printf("    Inbound: %d\n", volume.InboundMessages)
	fmt.Printf("    Outbound: %d\n", volume.OutboundMessages)

	if len(volume.MessagesByType) > 0 {
		fmt.Println("    By Type:")
		for msgType, count := range volume.MessagesByType {
			fmt.Printf("      %s: %d\n", msgType, count)
		}
	}

	// Delivery Performance
	delivery := analytics.PerformanceMetrics.DeliveryPerformance
	fmt.Printf("  Delivery Performance:\n")
	fmt.Printf("    Success Rate: %.2f%%\n", delivery.SuccessRate)
	fmt.Printf("    Failure Rate: %.2f%%\n", delivery.FailureRate)
	fmt.Printf("    Average Latency: %.2f ms\n", delivery.AverageLatency)

	// Cost Metrics
	cost := analytics.PerformanceMetrics.CostMetrics
	fmt.Printf("  Cost Metrics:\n")
	fmt.Printf("    Total Cost: %.2f %s\n", cost.TotalCost, cost.Currency)
	fmt.Printf("    Cost per Message: %.4f %s\n", cost.CostPerMessage, cost.Currency)

	// Usage Patterns
	usage := analytics.PerformanceMetrics.UsagePatterns
	fmt.Printf("  Usage Patterns:\n")
	fmt.Printf("    Active Hours: %v\n", usage.ActiveHours)
	fmt.Printf("    Active Days: %v\n", usage.ActiveDays)
	fmt.Printf("    Avg Conversation Length: %.2f\n", usage.ConversationFlow.AverageConversationLength)
	fmt.Printf("    Avg Response Time: %.2f minutes\n", usage.ConversationFlow.ResponseTime)

	return nil
}

// getPhoneNumberStatus demonstrates retrieving phone number status.
func getPhoneNumberStatus(ctx context.Context, client *bm.Client, phoneNumberID string) error {
	status, err := client.GetPhoneNumberStatus(ctx, phoneNumberID)
	if err != nil {
		return fmt.Errorf("failed to get phone number status: %w", err)
	}

	fmt.Printf("Phone Number Status for %s:\n", phoneNumberID)
	fmt.Printf("  Status: %s\n", status.Status)
	fmt.Printf("  Verification Status: %s\n", status.VerificationStatus)
	fmt.Printf("  Quality Rating: %s\n", status.QualityRating)
	fmt.Printf("  Last Status Update: %s\n", status.LastStatusUpdate.Format("2006-01-02 15:04:05"))

	// Health Status
	health := status.HealthStatus
	fmt.Printf("  Health Status: %s\n", health.Overall)
	fmt.Printf("  Last Health Check: %s\n", health.LastCheck.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Uptime: %.2f%%\n", health.Metrics.UptimePercentage)
	fmt.Printf("  Error Rate: %.2f%%\n", health.Metrics.ErrorRate)
	fmt.Printf("  Response Time: %.2f ms\n", health.Metrics.ResponseTime)
	fmt.Printf("  Throughput: %d/%d\n", health.Metrics.CurrentThroughput, health.Metrics.ThroughputLimit)

	if len(health.Issues) > 0 {
		fmt.Println("  Health Issues:")
		for _, issue := range health.Issues {
			fmt.Printf("    %s (%s): %s [%s]\n",
				issue.Type, issue.Severity, issue.Description, issue.Status)
		}
	}

	// Capabilities
	if len(status.Capabilities) > 0 {
		fmt.Printf("  Capabilities: %v\n", status.Capabilities)
	}

	// Restrictions
	if len(status.Restrictions) > 0 {
		fmt.Println("  Restrictions:")
		for _, restriction := range status.Restrictions {
			activeStatus := "Inactive"
			if restriction.Active {
				activeStatus = "Active"
			}
			fmt.Printf("    %s: %s [%s]\n", restriction.Type, restriction.Description, activeStatus)
		}
	}

	return nil
}

// generateAnalyticsReport demonstrates generating a comprehensive analytics report.
func generateAnalyticsReport(ctx context.Context, client *bm.Client, phoneNumberID string) error {
	fmt.Println("Generating comprehensive analytics report...")

	// Get data for the last 7 days
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	// Get general analytics
	analytics, err := client.GetAnalytics(ctx, startDate, endDate,
		bm.WithAnalyticsGranularity(bm.GranularityDaily),
		bm.WithAnalyticsMetricTypes(bm.MetricTypeCost, bm.MetricTypeMessage, bm.MetricTypeConversation),
	)
	if err != nil {
		return fmt.Errorf("failed to get analytics: %w", err)
	}

	fmt.Printf("\nAnalytics Report (%s to %s):\n", startDate, endDate)
	fmt.Printf("Data Points Retrieved: %d\n", len(analytics.Data))

	for _, dataPoint := range analytics.Data {
		fmt.Printf("\nMetric: %s\n", dataPoint.Name)
		if dataPoint.Title != "" {
			fmt.Printf("Title: %s\n", dataPoint.Title)
		}
		if dataPoint.Description != "" {
			fmt.Printf("Description: %s\n", dataPoint.Description)
		}
		fmt.Printf("Period: %s\n", dataPoint.Period)
		fmt.Printf("Values: %d\n", len(dataPoint.Values))

		// Show first few values as examples
		for i, value := range dataPoint.Values {
			if i >= 3 { // Limit to first 3 values
				fmt.Printf("  ... and %d more values\n", len(dataPoint.Values)-3)
				break
			}
			fmt.Printf("  %s: %v\n", value.EndTime, value.Value)
		}
	}

	fmt.Println("\nReport generation completed successfully!")
	return nil
}

// getAdvancedQualityMonitoring demonstrates advanced quality monitoring features.
func getAdvancedQualityMonitoring(ctx context.Context, client *bm.Client, phoneNumberID string) error {
	// Get quality score history
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	fmt.Printf("Advanced Quality Monitoring (%s to %s):\n", startDate, endDate)

	// Get quality score history
	history, err := client.GetQualityScoreHistory(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get quality score history: %w", err)
	}

	fmt.Printf("\nQuality Score History (%d data points):\n", len(history))
	for _, point := range history {
		fmt.Printf("  %s: %s (%.1f) - %d events\n",
			point.Date, point.Score, point.Value, len(point.Events))

		// Show events for this date
		for _, event := range point.Events {
			fmt.Printf("    - %s: %s (%s impact)\n",
				event.Type, event.Description, event.Impact)
		}
	}

	// Get delivery trends (if phone number ID is provided)
	if phoneNumberID != "" {
		trends, err := client.GetDeliveryTrends(ctx, phoneNumberID, startDate, endDate)
		if err != nil {
			return fmt.Errorf("failed to get delivery trends: %w", err)
		}

		fmt.Printf("\nDelivery Trends (%d data points):\n", len(trends))
		for _, trend := range trends {
			fmt.Printf("  %s: %.2f%% delivery rate, %d messages, %.0fms latency\n",
				trend.Date, trend.DeliveryRate, trend.Volume, trend.Latency)
		}
	}

	// Get personalized recommendations
	recommendations, err := client.GetQualityRecommendations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get quality recommendations: %w", err)
	}

	fmt.Printf("\nQuality Improvement Recommendations (%d recommendations):\n", len(recommendations))
	for i, rec := range recommendations {
		fmt.Printf("  %d. [%s Priority] %s\n", i+1, rec.Priority, rec.Title)
		fmt.Printf("     Category: %s\n", rec.Category)
		fmt.Printf("     Description: %s\n", rec.Description)
		fmt.Printf("     Action: %s\n", rec.Action)
		fmt.Printf("     Expected Impact: %s\n", rec.Impact)
		fmt.Printf("     Effort: %s | Timeline: %s\n", rec.Effort, rec.Timeline)
		fmt.Println()
	}

	return nil
}

// getAdvancedDeliveryAnalytics demonstrates advanced delivery analytics features.
func getAdvancedDeliveryAnalytics(ctx context.Context, client *bm.Client, phoneNumberID string) error {
	// Get comprehensive delivery analytics
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	fmt.Printf("Advanced Delivery Analytics for %s (%s to %s):\n", phoneNumberID, startDate, endDate)

	// Get comprehensive delivery analytics
	analytics, err := client.GetDeliveryAnalytics(ctx, phoneNumberID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get delivery analytics: %w", err)
	}

	// Display summary
	summary := analytics.Summary
	fmt.Printf("\nDelivery Summary:\n")
	fmt.Printf("  Total Messages: %d\n", summary.TotalMessages)
	fmt.Printf("  Successful Deliveries: %d (%.2f%%)\n", summary.SuccessfulDeliveries, summary.SuccessRate)
	fmt.Printf("  Failed Deliveries: %d (%.2f%%)\n", summary.FailedDeliveries, summary.FailureRate)
	fmt.Printf("  Average Latency: %.0fms\n", summary.AverageLatency)
	fmt.Printf("  Median Latency: %.0fms\n", summary.MedianLatency)
	fmt.Printf("  P95 Latency: %.0fms\n", summary.P95Latency)
	fmt.Printf("  P99 Latency: %.0fms\n", summary.P99Latency)

	// Display failure analysis
	failure := analytics.FailureAnalysis
	fmt.Printf("\nFailure Analysis:\n")
	fmt.Printf("  Top Failure Reasons (%d):\n", len(failure.TopFailureReasons))
	for i, reason := range failure.TopFailureReasons {
		fmt.Printf("    %d. %s (%.1f%%) - %s\n", i+1, reason.Reason, reason.Percentage, reason.Category)
		fmt.Printf("       Error Code: %s | Severity: %s | Trend: %s\n",
			reason.ErrorCode, reason.Severity, reason.Trend)
		if len(reason.Recommendations) > 0 {
			fmt.Printf("       Recommendations:\n")
			for _, rec := range reason.Recommendations {
				fmt.Printf("         - %s (%s priority, %s effort)\n",
					rec.Action, rec.Priority, rec.Effort)
			}
		}
		fmt.Println()
	}

	fmt.Printf("  Failures by Category:\n")
	for category, count := range failure.FailuresByCategory {
		fmt.Printf("    %s: %d\n", category, count)
	}

	fmt.Printf("  Failures by Severity:\n")
	for severity, count := range failure.FailuresBySeverity {
		fmt.Printf("    %s: %d\n", severity, count)
	}

	fmt.Printf("  Impact Assessment:\n")
	impact := failure.ImpactAssessment
	fmt.Printf("    Business Impact: %s\n", impact.BusinessImpact)
	fmt.Printf("    User Experience Impact: %s\n", impact.UserExperienceImpact)
	fmt.Printf("    Revenue Impact: $%.2f\n", impact.RevenueImpact)
	fmt.Printf("    Reputation Risk: %s\n", impact.ReputationRisk)
	fmt.Printf("    Compliance Risk: %s\n", impact.ComplianceRisk)

	// Display optimization suggestions
	fmt.Printf("\nOptimization Suggestions (%d):\n", len(analytics.OptimizationSuggestions))
	for i, suggestion := range analytics.OptimizationSuggestions {
		fmt.Printf("  %d. [%s] %s\n", i+1, suggestion.Priority, suggestion.Title)
		fmt.Printf("     Category: %s | Complexity: %s\n", suggestion.Category, suggestion.Complexity)
		fmt.Printf("     Expected Impact: %s\n", suggestion.ExpectedImpact)
		fmt.Printf("     Implementation: %s\n", suggestion.Implementation)
		fmt.Printf("     Timeline: %s | ROI: %.1fx\n", suggestion.Timeline, suggestion.EstimatedROI)
		if len(suggestion.Prerequisites) > 0 {
			fmt.Printf("     Prerequisites: %v\n", suggestion.Prerequisites)
		}
		if len(suggestion.Metrics) > 0 {
			fmt.Printf("     Metrics to Track: %v\n", suggestion.Metrics)
		}
		fmt.Println()
	}

	// Display performance insights
	insights := analytics.PerformanceInsights
	fmt.Printf("Performance Insights:\n")
	fmt.Printf("  Best Performing Hours: %v\n", insights.BestPerformingHours)
	fmt.Printf("  Worst Performing Hours: %v\n", insights.WorstPerformingHours)
	fmt.Printf("  Best Performing Days: %v\n", insights.BestPerformingDays)
	fmt.Printf("  Worst Performing Days: %v\n", insights.WorstPerformingDays)

	fmt.Printf("  Message Type Performance:\n")
	for msgType, stats := range insights.MessageTypePerformance {
		fmt.Printf("    %s: %d sent, %d delivered (%.2f%% success rate)\n",
			msgType, stats.Sent, stats.Delivered, stats.Rate)
	}

	return nil
}

// getAdvancedPhoneNumberManagement demonstrates advanced phone number management features.
func getAdvancedPhoneNumberManagement(ctx context.Context, client *bm.Client, phoneNumberID string) error {
	fmt.Printf("Advanced Phone Number Management for %s:\n", phoneNumberID)

	// Get comprehensive phone number insights
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

	insights, err := client.GetPhoneNumberInsights(ctx, phoneNumberID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get phone number insights: %w", err)
	}

	fmt.Printf("\nPhone Number Insights:\n")
	fmt.Printf("  Display Number: %s\n", insights.DisplayPhoneNumber)

	// Enhanced message volume with daily breakdown
	volume := insights.PerformanceMetrics.MessageVolume
	fmt.Printf("  Message Volume (30 days):\n")
	fmt.Printf("    Total: %d | Inbound: %d | Outbound: %d\n",
		volume.TotalMessages, volume.InboundMessages, volume.OutboundMessages)

	if len(volume.DailyVolume) > 0 {
		fmt.Printf("    Recent Daily Volume:\n")
		for i, daily := range volume.DailyVolume {
			if i >= 5 { // Show only last 5 days
				break
			}
			fmt.Printf("      %s: %d total (%d in, %d out)\n",
				daily.Date, daily.Total, daily.Inbound, daily.Outbound)
		}
	}

	if len(volume.PeakHours) > 0 {
		fmt.Printf("    Peak Hours:\n")
		for _, peak := range volume.PeakHours {
			dayName := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}[peak.DayOfWeek]
			fmt.Printf("      %s %02d:00 - %d messages\n",
				dayName, peak.Hour, peak.MessageCount)
		}
	}

	// Enhanced user engagement metrics
	engagement := insights.PerformanceMetrics.UsagePatterns.UserEngagement
	fmt.Printf("  User Engagement:\n")
	fmt.Printf("    Unique Users: %d\n", engagement.UniqueUsers)
	fmt.Printf("    Returning Users: %d (%.1f%%)\n",
		engagement.ReturningUsers, engagement.RetentionRate)
	fmt.Printf("    New Users: %d\n", engagement.NewUsers)
	fmt.Printf("    Engagement Rate: %.2f%%\n", engagement.EngagementRate)

	// Conversation flow analysis
	flow := insights.PerformanceMetrics.UsagePatterns.ConversationFlow
	fmt.Printf("  Conversation Flow:\n")
	fmt.Printf("    Avg Conversation Length: %.1f messages\n", flow.AverageConversationLength)
	fmt.Printf("    Conversation Starters: %d\n", flow.ConversationStarters)
	fmt.Printf("    Avg Response Time: %.1f minutes\n", flow.ResponseTime)

	// Perform health check
	fmt.Printf("\nHealth Check:\n")
	health, err := client.GetPhoneNumberHealthCheck(ctx, phoneNumberID)
	if err != nil {
		return fmt.Errorf("failed to perform health check: %w", err)
	}

	fmt.Printf("  Overall Health: %s\n", health.Overall)
	fmt.Printf("  Last Check: %s\n", health.LastCheck.Format("2006-01-02 15:04:05"))

	metrics := health.Metrics
	fmt.Printf("  Performance Metrics:\n")
	fmt.Printf("    Uptime: %.2f%%\n", metrics.UptimePercentage)
	fmt.Printf("    Error Rate: %.2f%%\n", metrics.ErrorRate)
	fmt.Printf("    Response Time: %.0fms\n", metrics.ResponseTime)
	fmt.Printf("    Throughput: %d/%d (%.1f%% capacity)\n",
		metrics.CurrentThroughput, metrics.ThroughputLimit,
		float64(metrics.CurrentThroughput)/float64(metrics.ThroughputLimit)*100.0)

	if len(health.Issues) > 0 {
		fmt.Printf("  Health Issues (%d):\n", len(health.Issues))
		for i, issue := range health.Issues {
			fmt.Printf("    %d. %s (%s severity)\n", i+1, issue.Type, issue.Severity)
			fmt.Printf("       %s [%s]\n", issue.Description, issue.Status)
			fmt.Printf("       Detected: %s\n", issue.DetectedAt.Format("2006-01-02 15:04:05"))
		}
	} else {
		fmt.Printf("  No health issues detected\n")
	}

	// Get usage patterns analysis
	fmt.Printf("\nUsage Patterns Analysis:\n")
	patterns, err := client.GetPhoneNumberUsagePatterns(ctx, phoneNumberID, 30)
	if err != nil {
		return fmt.Errorf("failed to get usage patterns: %w", err)
	}

	fmt.Printf("  Active Hours: %v\n", patterns.ActiveHours)
	fmt.Printf("  Active Days: %v (1=Mon, 7=Sun)\n", patterns.ActiveDays)
	fmt.Printf("  Engagement Rate: %.2f%%\n", patterns.UserEngagement.EngagementRate)
	fmt.Printf("  Retention Rate: %.2f%%\n", patterns.UserEngagement.RetentionRate)

	// Example of updating phone number settings
	fmt.Printf("\nPhone Number Configuration:\n")
	settings := map[string]interface{}{
		"display_name": "Customer Support",
		"about_text":   "We're here to help! Business hours: 9 AM - 6 PM",
		"auto_reply": map[string]interface{}{
			"enabled":           true,
			"message":           "Thanks for contacting us! We'll respond during business hours.",
			"out_of_hours_only": true,
		},
	}

	err = client.UpdatePhoneNumberSettings(ctx, phoneNumberID, settings)
	if err != nil {
		fmt.Printf("  Failed to update settings: %v\n", err)
	} else {
		fmt.Printf("  Settings updated successfully\n")
		fmt.Printf("    Display Name: %s\n", settings["display_name"])
		fmt.Printf("    About Text: %s\n", settings["about_text"])
		fmt.Printf("    Auto Reply: Enabled\n")
	}

	return nil
}

// getTemplateUsageCostAnalytics demonstrates template usage cost analytics features.
func getTemplateUsageCostAnalytics(ctx context.Context, client *bm.Client) error {
	// Get template usage cost analytics for the last 30 days
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

	templateAnalytics, err := client.GetTemplateUsageCostAnalytics(ctx, startDate, endDate,
		bm.WithAnalyticsGranularity(bm.GranularityDaily),
	)
	if err != nil {
		return fmt.Errorf("failed to get template usage cost analytics: %w", err)
	}

	fmt.Printf("Template Usage Cost Analytics (%s to %s):\n", startDate, endDate)

	// Total cost summary
	fmt.Printf("  Total Cost: %.2f %s\n", templateAnalytics.TotalCost.TotalCost, templateAnalytics.TotalCost.Currency)
	fmt.Printf("  Message Cost: %.2f %s\n", templateAnalytics.TotalCost.MessageCost, templateAnalytics.TotalCost.Currency)
	fmt.Printf("  Template Cost: %.2f %s\n", templateAnalytics.TotalCost.TemplateCost, templateAnalytics.TotalCost.Currency)

	// Template costs breakdown
	if len(templateAnalytics.TemplateCosts) > 0 {
		fmt.Println("  Template Cost Breakdown:")
		for _, templateCost := range templateAnalytics.TemplateCosts {
			fmt.Printf("    %s (%s):\n", templateCost.TemplateName, templateCost.Category)
			fmt.Printf("      Usage: %d messages | Cost: %.2f %s | Success Rate: %.1f%%\n",
				templateCost.Count, templateCost.Cost, templateCost.Currency, templateCost.SuccessRate)
		}
	}

	// Cost by category
	if len(templateAnalytics.CostByCategory) > 0 {
		fmt.Println("  Cost by Category:")
		for category, cost := range templateAnalytics.CostByCategory {
			fmt.Printf("    %s: %.2f %s\n", category, cost, templateAnalytics.TotalCost.Currency)
		}
	}

	// Usage patterns
	patterns := templateAnalytics.UsagePatterns
	fmt.Println("  Usage Patterns:")

	if len(patterns.MostUsedTemplates) > 0 {
		fmt.Println("    Most Used Templates:")
		for i, template := range patterns.MostUsedTemplates {
			if i >= 3 { // Show top 3
				break
			}
			fmt.Printf("      %d. %s: %d uses, %.1f%% success rate, %.2f %s\n",
				i+1, template.TemplateName, template.UsageCount, template.SuccessRate, template.Cost, template.Currency)
		}
	}

	if len(patterns.LeastUsedTemplates) > 0 {
		fmt.Println("    Least Used Templates:")
		for i, template := range patterns.LeastUsedTemplates {
			if i >= 3 { // Show bottom 3
				break
			}
			fmt.Printf("      %d. %s: %d uses, %.1f%% success rate, %.2f %s\n",
				i+1, template.TemplateName, template.UsageCount, template.SuccessRate, template.Cost, template.Currency)
		}
	}

	if len(patterns.CategoryDistribution) > 0 {
		fmt.Println("    Category Distribution:")
		for category, count := range patterns.CategoryDistribution {
			fmt.Printf("      %s: %d messages\n", category, count)
		}
	}

	// Performance metrics
	performance := templateAnalytics.PerformanceMetrics
	fmt.Println("  Performance Metrics:")
	fmt.Printf("    Average Success Rate: %.2f%%\n", performance.AverageSuccessRate)
	fmt.Printf("    Average Cost per Message: %.4f %s\n", performance.AverageCostPerMessage, performance.Currency)
	fmt.Printf("    Total Messages: %d\n", performance.TotalMessages)

	if len(performance.PerformanceByCategory) > 0 {
		fmt.Println("    Performance by Category:")
		for category, perf := range performance.PerformanceByCategory {
			fmt.Printf("      %s: %d templates, %d messages, %.2f%% success rate, %.4f %s avg cost\n",
				category, perf.TemplateCount, perf.TotalMessages, perf.SuccessRate, perf.AverageCost, perf.Currency)
		}
	}

	// Optimization insights
	if len(templateAnalytics.OptimizationInsights) > 0 {
		fmt.Printf("  Optimization Insights (%d):\n", len(templateAnalytics.OptimizationInsights))
		for i, insight := range templateAnalytics.OptimizationInsights {
			fmt.Printf("    %d. [%s] %s\n", i+1, insight.Priority, insight.Title)
			fmt.Printf("       Type: %s | Impact: %s\n", insight.Type, insight.Impact)
			fmt.Printf("       Description: %s\n", insight.Description)
			fmt.Printf("       Estimated Savings: %.2f %s\n", insight.EstimatedSavings, insight.Currency)
			if len(insight.AffectedTemplates) > 0 {
				fmt.Printf("       Affected Templates: %v\n", insight.AffectedTemplates)
			}
		}
	}

	// Get detailed insights for a specific template
	if len(templateAnalytics.TemplateCosts) > 0 {
		templateName := templateAnalytics.TemplateCosts[0].TemplateName
		fmt.Printf("\nDetailed Template Performance Insights for '%s':\n", templateName)

		insights, err := client.GetTemplatePerformanceInsights(ctx, templateName, startDate, endDate)
		if err != nil {
			return fmt.Errorf("failed to get template performance insights: %w", err)
		}

		fmt.Printf("  Template: %s (%s)\n", insights.TemplateName, insights.Category)

		// Performance data
		perf := insights.Performance
		fmt.Printf("  Performance:\n")
		fmt.Printf("    Total Sent: %d | Delivered: %d | Failed: %d\n",
			perf.TotalSent, perf.TotalDelivered, perf.TotalFailed)
		fmt.Printf("    Success Rate: %.2f%% | Engagement Rate: %.2f%%\n",
			perf.SuccessRate, perf.EngagementRate)
		fmt.Printf("    Response Rate: %.2f%% | Conversion Rate: %.2f%%\n",
			perf.ResponseRate, perf.ConversionRate)

		// Cost efficiency
		cost := insights.CostEfficiency
		fmt.Printf("  Cost Efficiency:\n")
		fmt.Printf("    Total Cost: %.2f %s | Cost per Message: %.4f %s\n",
			cost.TotalCost, cost.Currency, cost.CostPerMessage, cost.Currency)
		fmt.Printf("    Cost per Success: %.4f %s | Cost per Engagement: %.4f %s\n",
			cost.CostPerSuccess, cost.Currency, cost.CostPerEngagement, cost.Currency)
		fmt.Printf("    Efficiency Rating: %s | Benchmark Comparison: %.1f%%\n",
			cost.EfficiencyRating, cost.BenchmarkComparison)

		// Usage analysis
		usage := insights.UsageAnalysis
		fmt.Printf("  Usage Analysis:\n")
		fmt.Printf("    Frequency: %s | Trend: %s\n", usage.UsageFrequency, usage.UsageTrend)
		fmt.Printf("    Peak Hours: %v | Peak Days: %v\n", usage.PeakUsageHours, usage.PeakUsageDays)
		fmt.Printf("    Seasonality Score: %.1f | Consistency: %.1f\n",
			usage.SeasonalityScore, usage.UsageConsistency)

		// Competitor analysis
		if insights.CompetitorAnalysis.PerformanceRanking != "" {
			comp := insights.CompetitorAnalysis
			fmt.Printf("  Competitor Analysis:\n")
			fmt.Printf("    Performance Ranking: %s\n", comp.PerformanceRanking)
			fmt.Printf("    Your Success Rate: %.2f%% vs Industry: %.2f%%\n",
				comp.YourPerformance.SuccessRate, comp.IndustryBenchmark.SuccessRate)
			if len(comp.CompetitiveAdvantage) > 0 {
				fmt.Printf("    Competitive Advantages: %v\n", comp.CompetitiveAdvantage)
			}
			if len(comp.ImprovementAreas) > 0 {
				fmt.Printf("    Improvement Areas: %v\n", comp.ImprovementAreas)
			}
		}

		// Optimization recommendations
		if len(insights.Recommendations) > 0 {
			fmt.Printf("  Optimization Recommendations (%d):\n", len(insights.Recommendations))
			for i, rec := range insights.Recommendations {
				fmt.Printf("    %d. [%s Priority] %s\n", i+1, rec.Priority, rec.Title)
				fmt.Printf("       Category: %s | Complexity: %s\n", rec.Category, rec.Complexity)
				fmt.Printf("       Expected Impact: %s\n", rec.ExpectedImpact)
				fmt.Printf("       Implementation: %s\n", rec.Implementation)
				fmt.Printf("       Timeline: %s | ROI: %.1fx | Savings: %.2f %s\n",
					rec.Timeline, rec.EstimatedROI, rec.EstimatedSavings, rec.Currency)
				fmt.Println()
			}
		}

		// Get optimization recommendations for the template
		fmt.Printf("Personalized Optimization Recommendations for '%s':\n", templateName)
		recommendations, err := client.GetTemplateOptimizationRecommendations(ctx, templateName)
		if err != nil {
			return fmt.Errorf("failed to get template optimization recommendations: %w", err)
		}

		for i, rec := range recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec.Title)
			fmt.Printf("     Priority: %s | Category: %s | Complexity: %s\n",
				rec.Priority, rec.Category, rec.Complexity)
			fmt.Printf("     Expected Impact: %s\n", rec.ExpectedImpact)
			fmt.Printf("     Estimated Savings: %.2f %s | ROI: %.1fx\n",
				rec.EstimatedSavings, rec.Currency, rec.EstimatedROI)
			fmt.Println()
		}
	}

	return nil
}

// getComplianceMonitoring demonstrates compliance monitoring features.
func getComplianceMonitoring(ctx context.Context, client *bm.Client) error {
	// Get compliance monitoring for the last 30 days
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

	monitoring, err := client.GetComplianceMonitoring(ctx, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get compliance monitoring: %w", err)
	}

	fmt.Printf("Compliance Monitoring (%s to %s):\n", startDate, endDate)

	// Overall status
	fmt.Printf("  Overall Status: %s\n", monitoring.OverallStatus.Status)
	fmt.Printf("  Last Review: %s\n", monitoring.OverallStatus.LastReview.Format("2006-01-02 15:04:05"))

	// Policy compliance
	policy := monitoring.PolicyCompliance
	fmt.Printf("  Policy Compliance:\n")
	fmt.Printf("    Overall Score: %.1f/100 (%s)\n", policy.OverallScore, policy.ComplianceLevel)
	fmt.Printf("    Last Assessment: %s\n", policy.LastAssessment.Format("2006-01-02"))
	fmt.Printf("    Next Assessment: %s\n", policy.NextAssessment.Format("2006-01-02"))

	if len(policy.WhatsAppPolicies) > 0 {
		fmt.Printf("    WhatsApp Policies (%d):\n", len(policy.WhatsAppPolicies))
		for i, pol := range policy.WhatsAppPolicies {
			fmt.Printf("      %d. %s (%s)\n", i+1, pol.PolicyName, pol.PolicyCategory)
			fmt.Printf("         Status: %s | Score: %.1f/100\n", pol.ComplianceStatus, pol.ComplianceScore)
			if len(pol.Violations) > 0 {
				fmt.Printf("         Violations: %d\n", len(pol.Violations))
				for j, violation := range pol.Violations {
					fmt.Printf("           %d. %s (%s severity)\n", j+1, violation.ViolationType, violation.Severity)
					fmt.Printf("              %s\n", violation.Description)
					fmt.Printf("              Affected Messages: %d | Status: %s\n",
						violation.AffectedMessages, violation.Status)
				}
			}
			if len(pol.Requirements) > 0 {
				fmt.Printf("         Requirements: %d\n", len(pol.Requirements))
				for j, req := range pol.Requirements {
					status := "✓"
					if req.Status != "MET" {
						status = "✗"
					}
					fmt.Printf("           %s %d. %s (%.1f%%)\n", status, j+1, req.Description, req.ComplianceLevel)
				}
			}
		}
	}

	if len(policy.BusinessPolicies) > 0 {
		fmt.Printf("    Business Policies (%d):\n", len(policy.BusinessPolicies))
		for i, pol := range policy.BusinessPolicies {
			fmt.Printf("      %d. %s: %s (%.1f/100)\n", i+1, pol.PolicyName, pol.ComplianceStatus, pol.ComplianceScore)
		}
	}

	// Regulatory compliance
	regulatory := monitoring.RegulatoryCompliance
	fmt.Printf("  Regulatory Compliance:\n")
	fmt.Printf("    Overall Score: %.1f/100 (%s)\n", regulatory.OverallScore, regulatory.ComplianceLevel)
	fmt.Printf("    Last Audit: %s\n", regulatory.LastAudit.Format("2006-01-02"))
	fmt.Printf("    Next Audit: %s\n", regulatory.NextAudit.Format("2006-01-02"))

	if len(regulatory.Jurisdictions) > 0 {
		fmt.Printf("    Jurisdictions (%d):\n", len(regulatory.Jurisdictions))
		for i, jurisdiction := range regulatory.Jurisdictions {
			fmt.Printf("      %d. %s (%s)\n", i+1, jurisdiction.Country, jurisdiction.Region)
			fmt.Printf("         Score: %.1f/100 | Status: %s\n", jurisdiction.ComplianceScore, jurisdiction.Status)
			if len(jurisdiction.Regulations) > 0 {
				fmt.Printf("         Regulations: %d\n", len(jurisdiction.Regulations))
				for j, reg := range jurisdiction.Regulations {
					fmt.Printf("           %d. %s: %.1f/100 (%s)\n",
						j+1, reg.RegulationName, reg.ComplianceScore, reg.Status)
				}
			}
		}
	}

	// Data protection
	dataProtection := regulatory.DataProtection
	fmt.Printf("    Data Protection:\n")
	fmt.Printf("      GDPR: %.1f/100 (%s)\n",
		dataProtection.GDPRCompliance.ComplianceScore, dataProtection.GDPRCompliance.Status)

	// Quality compliance
	quality := monitoring.QualityCompliance
	fmt.Printf("  Quality Compliance:\n")
	fmt.Printf("    Overall Score: %.1f/100 (%s)\n", quality.OverallScore, quality.ComplianceLevel)

	msgQuality := quality.MessageQuality
	fmt.Printf("    Message Quality:\n")
	fmt.Printf("      Content Quality: %.1f/100\n", msgQuality.ContentQualityScore)
	fmt.Printf("      Spam Score: %.1f (lower is better)\n", msgQuality.SpamScore)
	fmt.Printf("      Violations: %d\n", msgQuality.ViolationCount)
	fmt.Printf("      Template Compliance: %.1f%% (%d approved, %d rejected, %d pending)\n",
		msgQuality.TemplateCompliance.ComplianceRate,
		msgQuality.TemplateCompliance.ApprovedTemplates,
		msgQuality.TemplateCompliance.RejectedTemplates,
		msgQuality.TemplateCompliance.PendingTemplates)

	deliveryQuality := quality.DeliveryQuality
	fmt.Printf("    Delivery Quality:\n")
	fmt.Printf("      Delivery Rate: %.1f%% (threshold: %.1f%%)\n",
		deliveryQuality.DeliveryRate, deliveryQuality.QualityThreshold)
	fmt.Printf("      Status: %s\n", deliveryQuality.ComplianceStatus)

	engagementQuality := quality.EngagementQuality
	fmt.Printf("    Engagement Quality:\n")
	fmt.Printf("      Engagement Rate: %.1f%% | Response Rate: %.1f%%\n",
		engagementQuality.EngagementRate, engagementQuality.ResponseRate)
	fmt.Printf("      Opt-out Rate: %.1f%% | Status: %s\n",
		engagementQuality.OptOutRate, engagementQuality.ComplianceStatus)

	// Compliance alerts
	if len(monitoring.ComplianceAlerts) > 0 {
		fmt.Printf("  Active Compliance Alerts (%d):\n", len(monitoring.ComplianceAlerts))
		for i, alert := range monitoring.ComplianceAlerts {
			fmt.Printf("    %d. [%s] %s\n", i+1, alert.Severity, alert.Title)
			fmt.Printf("       Type: %s | Category: %s | Status: %s\n",
				alert.AlertType, alert.Category, alert.Status)
			fmt.Printf("       Description: %s\n", alert.Description)
			fmt.Printf("       Created: %s | Updated: %s\n",
				alert.CreatedAt.Format("2006-01-02"), alert.UpdatedAt.Format("2006-01-02"))
			if alert.DueDate != nil {
				fmt.Printf("       Due Date: %s\n", alert.DueDate.Format("2006-01-02"))
			}
			if len(alert.Actions) > 0 {
				fmt.Printf("       Actions: %d\n", len(alert.Actions))
				for j, action := range alert.Actions {
					fmt.Printf("         %d. %s (%s priority, %s)\n",
						j+1, action.Title, action.Priority, action.Status)
				}
			}
			fmt.Println()
		}
	}

	return nil
}
