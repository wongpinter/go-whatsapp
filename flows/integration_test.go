package flows

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp/internal/httpclient"
	"github.com/wongpinter/go-whatsapp/internal/httpserver"
)

// TestFlowsServerIntegration tests the complete Flows server integration
func TestFlowsServerIntegration(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Timestamp().Logger()

	tests := []struct {
		name      string
		framework httpserver.Framework
	}{
		{"Standard HTTP", httpserver.FrameworkStandard},
		{"Gin Framework", httpserver.FrameworkGin},
		{"Echo Framework", httpserver.FrameworkEcho},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFlowsServerWithFramework(t, tt.framework, &logger)
		})
	}
}

func testFlowsServerWithFramework(t *testing.T, framework httpserver.Framework, logger *zerolog.Logger) {
	// Create HTTP client manager
	clientManager := httpclient.NewManager(nil, logger)

	// Create Flows server factory
	factory := NewServerFactory(clientManager, logger)

	// Create test server
	server, adapter, err := factory.CreateFullFlowsServer(
		framework,
		"test-app-secret",
		"test-verify-token",
		WithRoutePrefix("/api/v1"),
		WithRateLimit(true, 1000), // High limit for tests
		WithMetrics(true),
		WithSecurity(false), // Disable for easier testing
		WithServerLogger(logger),
	)
	if err != nil {
		t.Fatalf("Failed to create Flows server: %v", err)
	}

	// Create test server - only for standard HTTP (mock implementations don't support http.Handler)
	if framework != httpserver.FrameworkStandard {
		t.Skip("Skipping HTTP server test for mock framework implementation")
		return
	}

	testServer := httptest.NewServer(server.Router().Native().(http.Handler))
	defer testServer.Close()

	// Test all endpoints
	t.Run("Health Check", func(t *testing.T) {
		testHealthEndpoint(t, testServer.URL)
	})

	t.Run("Metrics", func(t *testing.T) {
		testMetricsEndpoint(t, testServer.URL)
	})

	t.Run("Data Exchange", func(t *testing.T) {
		testDataExchangeEndpoint(t, testServer.URL)
	})

	t.Run("Flow Sending", func(t *testing.T) {
		testFlowSendingEndpoints(t, testServer.URL)
	})

	t.Run("Webhook Integration", func(t *testing.T) {
		testWebhookIntegration(t, testServer.URL)
	})

	// Test adapter functionality
	t.Run("Adapter Metrics", func(t *testing.T) {
		metrics := adapter.GetMetrics()
		if metrics == nil {
			t.Error("Expected metrics to be available")
		}
	})
}

func testHealthEndpoint(t *testing.T, baseURL string) {
	resp, err := http.Get(baseURL + "/api/v1/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	// Verify health response structure
	if status, ok := health["status"]; !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", status)
	}

	if _, ok := health["flows"]; !ok {
		t.Error("Expected 'flows' section in health response")
	}
}

func testMetricsEndpoint(t *testing.T, baseURL string) {
	resp, err := http.Get(baseURL + "/api/v1/metrics")
	if err != nil {
		t.Fatalf("Failed to call metrics endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var metrics map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		t.Fatalf("Failed to decode metrics response: %v", err)
	}

	// Verify metrics response structure
	expectedFields := []string{"total_requests", "successful_flows", "failed_flows", "performance"}
	for _, field := range expectedFields {
		if _, ok := metrics[field]; !ok {
			t.Errorf("Expected field '%s' in metrics response", field)
		}
	}
}

func testDataExchangeEndpoint(t *testing.T, baseURL string) {
	// Test data exchange request
	requestData := map[string]interface{}{
		"version": "3.0",
		"action":  "ping",
		"screen":  "WELCOME",
		"data":    map[string]interface{}{},
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		t.Fatalf("Failed to marshal request data: %v", err)
	}

	resp, err := http.Post(
		baseURL+"/api/v1/data-exchange",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Failed to call data exchange endpoint: %v", err)
	}
	defer resp.Body.Close()

	// For now, we expect the endpoint to be available (even if it returns an error due to missing implementation)
	if resp.StatusCode == http.StatusNotFound {
		t.Error("Data exchange endpoint should be available")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	t.Logf("Data exchange response: %s", string(body))
}

func testFlowSendingEndpoints(t *testing.T, baseURL string) {
	endpoints := []string{
		"/api/v1/send/survey",
		"/api/v1/send/lead",
		"/api/v1/send/custom",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			requestData := map[string]interface{}{
				"recipient": "+1234567890",
				"flow_id":   "test-flow-id",
			}

			jsonData, err := json.Marshal(requestData)
			if err != nil {
				t.Fatalf("Failed to marshal request data: %v", err)
			}

			resp, err := http.Post(
				baseURL+endpoint,
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				t.Fatalf("Failed to call endpoint %s: %v", endpoint, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				t.Errorf("Endpoint %s should be available", endpoint)
			}
		})
	}
}

func testWebhookIntegration(t *testing.T, baseURL string) {
	// Test webhook verification (GET) - note the correct path with /webhook prefix
	resp, err := http.Get(baseURL + "/api/v1/webhook/webhook?hub.mode=subscribe&hub.challenge=test-challenge&hub.verify_token=test-verify-token")
	if err != nil {
		t.Fatalf("Failed to call webhook verification: %v", err)
	}
	defer resp.Body.Close()

	// Test webhook event (POST)
	webhookData := map[string]interface{}{
		"object": "whatsapp_business_account",
		"entry": []map[string]interface{}{
			{
				"id": "test-entry-id",
				"changes": []map[string]interface{}{
					{
						"value": map[string]interface{}{
							"messaging_product": "whatsapp",
							"metadata": map[string]interface{}{
								"display_phone_number": "+1234567890",
								"phone_number_id":      "test-phone-id",
							},
						},
						"field": "messages",
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(webhookData)
	if err != nil {
		t.Fatalf("Failed to marshal webhook data: %v", err)
	}

	resp, err = http.Post(
		baseURL+"/api/v1/webhook/webhook",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Failed to call webhook endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Error("Webhook endpoint should be available")
	}
}

// TestFlowsMiddleware tests the Flows middleware functionality
func TestFlowsMiddleware(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Timestamp().Logger()

	t.Run("Logging Middleware", func(t *testing.T) {
		testLoggingMiddleware(t, &logger)
	})

	t.Run("Metrics Middleware", func(t *testing.T) {
		testMetricsMiddleware(t, &logger)
	})

	t.Run("Rate Limit Middleware", func(t *testing.T) {
		testRateLimitMiddleware(t, &logger)
	})

	t.Run("Security Middleware", func(t *testing.T) {
		testSecurityMiddleware(t, &logger)
	})
}

func testLoggingMiddleware(t *testing.T, logger *zerolog.Logger) {
	middleware := FlowsLoggingMiddleware(logger)

	called := false
	handler := func(ctx httpserver.HTTPContext) error {
		called = true
		return nil
	}

	wrappedHandler := middleware(handler)

	// Create mock context
	mockCtx := &mockHTTPContext{
		method: "POST",
		path:   "/flows/data-exchange",
	}

	err := wrappedHandler(mockCtx)
	if err != nil {
		t.Errorf("Logging middleware should not return error: %v", err)
	}

	if !called {
		t.Error("Handler should have been called")
	}
}

func testMetricsMiddleware(t *testing.T, logger *zerolog.Logger) {
	metrics := &FlowMetrics{
		actionExecutions: make(map[string]int64),
		errorsByType:     make(map[string]int64),
		startTime:        time.Now(),
	}

	middleware := FlowsMetricsMiddleware(metrics, logger)

	handler := func(ctx httpserver.HTTPContext) error {
		return nil
	}

	wrappedHandler := middleware(handler)

	// Create mock context
	mockCtx := &mockHTTPContext{
		method: "POST",
		path:   "/flows/data-exchange",
	}

	initialRequests := metrics.totalRequests
	err := wrappedHandler(mockCtx)
	if err != nil {
		t.Errorf("Metrics middleware should not return error: %v", err)
	}

	if metrics.totalRequests <= initialRequests {
		t.Error("Metrics should have been updated")
	}
}

func testRateLimitMiddleware(t *testing.T, logger *zerolog.Logger) {
	middleware := FlowsRateLimitMiddleware(1, logger) // Very low limit for testing

	handler := func(ctx httpserver.HTTPContext) error {
		return nil
	}

	wrappedHandler := middleware(handler)

	// Create mock context
	mockCtx := &mockHTTPContext{
		method: "POST",
		path:   "/flows/data-exchange",
		headers: map[string]string{
			"X-Forwarded-For": "192.168.1.1",
		},
	}

	// First request should succeed
	err := wrappedHandler(mockCtx)
	if err != nil {
		t.Errorf("First request should succeed: %v", err)
	}

	// Second request should be rate limited (check status code)
	err = wrappedHandler(mockCtx)
	// The middleware returns a JSON response with 429 status, not a Go error
	if mockCtx.status != 429 {
		t.Errorf("Second request should be rate limited with status 429, got %d", mockCtx.status)
	}
}

func testSecurityMiddleware(t *testing.T, logger *zerolog.Logger) {
	middleware := FlowsSecurityMiddleware("test-secret", logger)

	handler := func(ctx httpserver.HTTPContext) error {
		return nil
	}

	wrappedHandler := middleware(handler)

	// Create mock context
	mockCtx := &mockHTTPContext{
		method: "GET",
		path:   "/flows/health",
	}

	err := wrappedHandler(mockCtx)
	if err != nil {
		t.Errorf("Security middleware should not block health checks: %v", err)
	}
}

// Mock HTTP context for testing
type mockHTTPContext struct {
	method  string
	path    string
	headers map[string]string
	body    []byte
	status  int
}

func (m *mockHTTPContext) Method() string                       { return m.method }
func (m *mockHTTPContext) Path() string                         { return m.path }
func (m *mockHTTPContext) Query(key string) string              { return "" }
func (m *mockHTTPContext) Header(key string) string             { return m.headers[key] }
func (m *mockHTTPContext) Body() ([]byte, error)                { return m.body, nil }
func (m *mockHTTPContext) Context() context.Context             { return context.Background() }
func (m *mockHTTPContext) WithContext(ctx context.Context)      {}
func (m *mockHTTPContext) Status(code int)                      { m.status = code }
func (m *mockHTTPContext) SetHeader(key, value string)          {}
func (m *mockHTTPContext) Write(data []byte) error              { return nil }
func (m *mockHTTPContext) JSON(code int, obj interface{}) error { m.status = code; return nil }
func (m *mockHTTPContext) String(code int, format string, values ...interface{}) error {
	m.status = code
	return nil
}
func (m *mockHTTPContext) Native() interface{} { return nil }

// TestFlowsServerFactory tests the server factory functionality
func TestFlowsServerFactory(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Timestamp().Logger()
	clientManager := httpclient.NewManager(nil, &logger)
	factory := NewServerFactory(clientManager, &logger)

	t.Run("Create Data Exchange Server", func(t *testing.T) {
		server, adapter, err := factory.CreateDataExchangeServer(
			httpserver.FrameworkStandard,
			WithRateLimit(true, 100),
			WithMetrics(true),
		)
		if err != nil {
			t.Fatalf("Failed to create data exchange server: %v", err)
		}

		if server == nil {
			t.Error("Server should not be nil")
		}
		if adapter == nil {
			t.Error("Adapter should not be nil")
		}
	})

	t.Run("Create Full Flows Server", func(t *testing.T) {
		server, adapter, err := factory.CreateFullFlowsServer(
			httpserver.FrameworkStandard,
			"test-secret",
			"test-token",
			WithRoutePrefix("/test"),
			WithMetrics(true),
		)
		if err != nil {
			t.Fatalf("Failed to create full Flows server: %v", err)
		}

		if server == nil {
			t.Error("Server should not be nil")
		}
		if adapter == nil {
			t.Error("Adapter should not be nil")
		}
	})

	t.Run("Invalid Configuration", func(t *testing.T) {
		_, _, err := factory.CreateFullFlowsServer(
			httpserver.Framework("invalid"),
			"test-secret",
			"test-token",
		)
		if err == nil {
			t.Error("Should return error for invalid framework")
		}
	})
}

// TestFlowsServerAdapter tests the server adapter functionality
func TestFlowsServerAdapter(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Timestamp().Logger()

	adapter := NewServerAdapter(
		nil, // dataExchangeHandler
		nil, // actionRouter
		nil, // tokenManager
		nil, // stateManager
		&logger,
	)

	t.Run("Metrics Access", func(t *testing.T) {
		metrics := adapter.GetMetrics()
		if metrics == nil {
			t.Error("Metrics should not be nil")
		}

		if metrics.actionExecutions == nil {
			t.Error("Action executions map should be initialized")
		}

		if metrics.errorsByType == nil {
			t.Error("Errors by type map should be initialized")
		}
	})

	t.Run("Data Exchange Handler Access", func(t *testing.T) {
		handler := adapter.GetDataExchangeHandler()
		// Handler can be nil in test setup, just verify method exists
		_ = handler
	})
}

// BenchmarkFlowsServerPerformance benchmarks the Flows server performance
func BenchmarkFlowsServerPerformance(b *testing.B) {
	logger := zerolog.New(zerolog.NewTestWriter(b)).With().Timestamp().Logger()
	clientManager := httpclient.NewManager(nil, &logger)
	factory := NewServerFactory(clientManager, &logger)

	server, _, err := factory.CreateDataExchangeServer(
		httpserver.FrameworkStandard,
		WithRateLimit(false, 0), // Disable rate limiting for benchmark
		WithMetrics(false),      // Disable metrics for benchmark
	)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	testServer := httptest.NewServer(server.Router().Native().(http.Handler))
	defer testServer.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := http.Get(testServer.URL + "/flows/health")
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			resp.Body.Close()
		}
	})
}
