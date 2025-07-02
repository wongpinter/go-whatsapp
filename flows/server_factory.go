package flows

import (
	"fmt"

	"github.com/rs/zerolog"

	"github.com/wongpinter/go-whatsapp/internal/httpclient"
	"github.com/wongpinter/go-whatsapp/internal/httpserver"
	"github.com/wongpinter/go-whatsapp/webhook"
)

// ServerFactory creates Flows servers with different HTTP frameworks
type ServerFactory struct {
	httpServerFactory *httpserver.ServerFactory
	httpClientManager *httpclient.Manager
	logger            *zerolog.Logger
}

// ServerOption configures the Flows server
type ServerOption func(*FlowsServerConfig)

// FlowsServerConfig holds configuration for Flows server creation
type FlowsServerConfig struct {
	// HTTP Framework
	Framework httpserver.Framework

	// Server options
	ServerOptions []httpserver.ServerOption

	// Flows components
	DataExchangeHandler *DataExchangeHandler
	ActionRouter        *ActionRouter
	TokenManager        *FlowTokenManager
	StateManager        *FlowStateManager

	// Integration options
	EnableWebhookIntegration bool
	WebhookAppSecret         string
	WebhookVerifyToken       string

	// Middleware options
	EnableEncryption bool
	EnableRateLimit  bool
	EnableMetrics    bool
	EnableSecurity   bool
	RateLimitRPM     int

	// Route configuration
	RoutePrefix           string
	EnableHealthChecks    bool
	EnableMetricsEndpoint bool

	// Logging
	Logger *zerolog.Logger
}

// NewServerFactory creates a new Flows server factory
func NewServerFactory(httpClientManager *httpclient.Manager, logger *zerolog.Logger) *ServerFactory {
	if logger == nil {
		nopLogger := zerolog.Nop()
		logger = &nopLogger
	}

	return &ServerFactory{
		httpServerFactory: httpserver.NewServerFactory(),
		httpClientManager: httpClientManager,
		logger:            logger,
	}
}

// CreateFlowsServer creates a complete Flows server with the specified configuration
func (f *ServerFactory) CreateFlowsServer(opts ...ServerOption) (httpserver.Server, *ServerAdapter, error) {
	// Default configuration
	config := &FlowsServerConfig{
		Framework:                httpserver.FrameworkStandard,
		EnableWebhookIntegration: false,
		EnableEncryption:         true,
		EnableRateLimit:          true,
		EnableMetrics:            true,
		EnableSecurity:           true,
		RateLimitRPM:             60,
		EnableHealthChecks:       true,
		EnableMetricsEndpoint:    true,
		Logger:                   f.logger,
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Create HTTP server
	server, err := f.httpServerFactory.CreateServer(config.Framework, config.ServerOptions...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create HTTP server: %w", err)
	}

	// Create Flows components if not provided
	if config.DataExchangeHandler == nil {
		config.DataExchangeHandler = f.createDefaultDataExchangeHandler(config)
	}
	if config.ActionRouter == nil {
		config.ActionRouter = NewActionRouter()
	}
	if config.TokenManager == nil {
		config.TokenManager = DefaultFlowTokenManager
	}
	if config.StateManager == nil {
		config.StateManager = NewFlowStateManager()
	}

	// Create server adapter
	serverAdapter := NewServerAdapter(
		config.DataExchangeHandler,
		config.ActionRouter,
		config.TokenManager,
		config.StateManager,
		config.Logger,
	)

	// Configure routes
	router := server.Router()

	// Add global middleware based on configuration
	f.configureMiddleware(router, config)

	// Register Flows routes
	if config.RoutePrefix != "" {
		serverAdapter.RegisterRoutesWithPrefix(router, config.RoutePrefix)
	} else {
		serverAdapter.RegisterRoutes(router)
	}

	// Add webhook integration if enabled (with different prefix to avoid conflicts)
	if config.EnableWebhookIntegration {
		webhookPrefix := config.RoutePrefix
		if webhookPrefix == "" {
			webhookPrefix = "/webhook"
		} else {
			webhookPrefix = webhookPrefix + "/webhook"
		}
		err := f.addWebhookIntegration(router, config, webhookPrefix)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to add webhook integration: %w", err)
		}
	}

	f.logger.Info().
		Str("framework", string(config.Framework)).
		Str("route_prefix", config.RoutePrefix).
		Bool("webhook_integration", config.EnableWebhookIntegration).
		Msg("Flows server created successfully")

	return server, serverAdapter, nil
}

// CreateDataExchangeServer creates a minimal server focused on data exchange
func (f *ServerFactory) CreateDataExchangeServer(framework httpserver.Framework, opts ...ServerOption) (httpserver.Server, *ServerAdapter, error) {
	config := &FlowsServerConfig{
		Framework:                framework,
		EnableWebhookIntegration: false,
		EnableEncryption:         true,
		EnableRateLimit:          true,
		EnableMetrics:            false,
		EnableSecurity:           true,
		RateLimitRPM:             100,
		EnableHealthChecks:       true,
		EnableMetricsEndpoint:    false,
		Logger:                   f.logger,
	}

	// Apply additional options
	for _, opt := range opts {
		opt(config)
	}

	return f.CreateFlowsServer(func(c *FlowsServerConfig) {
		*c = *config
	})
}

// CreateFullFlowsServer creates a complete Flows server with all features enabled
func (f *ServerFactory) CreateFullFlowsServer(framework httpserver.Framework, appSecret, verifyToken string, opts ...ServerOption) (httpserver.Server, *ServerAdapter, error) {
	config := &FlowsServerConfig{
		Framework:                framework,
		EnableWebhookIntegration: true,
		WebhookAppSecret:         appSecret,
		WebhookVerifyToken:       verifyToken,
		EnableEncryption:         true,
		EnableRateLimit:          true,
		EnableMetrics:            true,
		EnableSecurity:           true,
		RateLimitRPM:             60,
		EnableHealthChecks:       true,
		EnableMetricsEndpoint:    true,
		Logger:                   f.logger,
	}

	// Apply additional options
	for _, opt := range opts {
		opt(config)
	}

	return f.CreateFlowsServer(func(c *FlowsServerConfig) {
		*c = *config
	})
}

// Helper methods

func (f *ServerFactory) createDefaultDataExchangeHandler(config *FlowsServerConfig) *DataExchangeHandler {
	return NewDataExchangeHandler(
		WithActionRouter(config.ActionRouter),
		WithTokenManager(config.TokenManager),
		WithStateManager(config.StateManager),
		WithDataExchangeLogger(config.Logger),
	)
}

func (f *ServerFactory) configureMiddleware(router httpserver.Router, config *FlowsServerConfig) {
	// Always add logging
	router.Use(FlowsLoggingMiddleware(config.Logger))

	// Add metrics if enabled
	if config.EnableMetrics {
		// Create metrics instance for the server adapter
		metrics := &FlowMetrics{}
		router.Use(FlowsMetricsMiddleware(metrics, config.Logger))
	}

	// Add security middleware if enabled
	if config.EnableSecurity && config.WebhookAppSecret != "" {
		router.Use(FlowsSecurityMiddleware(config.WebhookAppSecret, config.Logger))
	}
}

func (f *ServerFactory) addWebhookIntegration(router httpserver.Router, config *FlowsServerConfig, prefix string) error {
	if config.WebhookAppSecret == "" || config.WebhookVerifyToken == "" {
		return fmt.Errorf("webhook app secret and verify token are required for webhook integration")
	}

	webhookAdapter := webhook.NewServerAdapter(
		config.WebhookAppSecret,
		config.WebhookVerifyToken,
		config.Logger,
	)

	// Use the provided prefix to avoid route conflicts
	if prefix != "" {
		webhookAdapter.RegisterRoutesWithPrefix(router, prefix)
	} else {
		webhookAdapter.RegisterRoutes(router)
	}

	return nil
}

// Configuration options

// WithFramework sets the HTTP framework
func WithFramework(framework httpserver.Framework) ServerOption {
	return func(c *FlowsServerConfig) {
		c.Framework = framework
	}
}

// WithServerOptions sets HTTP server options
func WithServerOptions(opts ...httpserver.ServerOption) ServerOption {
	return func(c *FlowsServerConfig) {
		c.ServerOptions = opts
	}
}

// WithDataExchangeHandler sets a custom data exchange handler
func WithDataExchangeHandler(handler *DataExchangeHandler) ServerOption {
	return func(c *FlowsServerConfig) {
		c.DataExchangeHandler = handler
	}
}

// WithServerActionRouter sets a custom action router
func WithServerActionRouter(router *ActionRouter) ServerOption {
	return func(c *FlowsServerConfig) {
		c.ActionRouter = router
	}
}

// WithServerTokenManager sets a custom token manager
func WithServerTokenManager(manager *FlowTokenManager) ServerOption {
	return func(c *FlowsServerConfig) {
		c.TokenManager = manager
	}
}

// WithServerStateManager sets a custom state manager
func WithServerStateManager(manager *FlowStateManager) ServerOption {
	return func(c *FlowsServerConfig) {
		c.StateManager = manager
	}
}

// WithWebhookIntegration enables webhook integration
func WithWebhookIntegration(appSecret, verifyToken string) ServerOption {
	return func(c *FlowsServerConfig) {
		c.EnableWebhookIntegration = true
		c.WebhookAppSecret = appSecret
		c.WebhookVerifyToken = verifyToken
	}
}

// WithRoutePrefix sets a custom route prefix
func WithRoutePrefix(prefix string) ServerOption {
	return func(c *FlowsServerConfig) {
		c.RoutePrefix = prefix
	}
}

// WithRateLimit configures rate limiting
func WithRateLimit(enabled bool, requestsPerMinute int) ServerOption {
	return func(c *FlowsServerConfig) {
		c.EnableRateLimit = enabled
		c.RateLimitRPM = requestsPerMinute
	}
}

// WithSecurity configures security features
func WithSecurity(enabled bool) ServerOption {
	return func(c *FlowsServerConfig) {
		c.EnableSecurity = enabled
	}
}

// WithMetrics configures metrics collection
func WithMetrics(enabled bool) ServerOption {
	return func(c *FlowsServerConfig) {
		c.EnableMetrics = enabled
	}
}

// WithServerLogger sets a custom logger
func WithServerLogger(logger *zerolog.Logger) ServerOption {
	return func(c *FlowsServerConfig) {
		c.Logger = logger
	}
}
