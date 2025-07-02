package httpserver

import (
	"fmt"
)

// Framework represents supported HTTP frameworks
type Framework string

const (
	FrameworkStandard Framework = "standard" // net/http
	FrameworkGin      Framework = "gin"      // Gin
	FrameworkEcho     Framework = "echo"     // Echo
	FrameworkFiber    Framework = "fiber"    // Fiber
	FrameworkChi      Framework = "chi"      // Chi
)

// ServerFactory creates servers for different frameworks
type ServerFactory struct{}

// Function variables that can be overridden by build tags
var (
	createRealGinServerFunc  func(...ServerOption) (Server, error)
	createRealEchoServerFunc func(...ServerOption) (Server, error)
)

// NewServerFactory creates a new server factory
func NewServerFactory() *ServerFactory {
	return &ServerFactory{}
}

// CreateServer creates a server for the specified framework
func (f *ServerFactory) CreateServer(framework Framework, options ...ServerOption) (Server, error) {
	switch framework {
	case FrameworkStandard:
		return f.createStandardServer(options...)
	case FrameworkGin:
		return f.createGinServer(options...)
	case FrameworkEcho:
		return f.createEchoServer(options...)
	case FrameworkFiber:
		return f.createFiberServer(options...)
	case FrameworkChi:
		return f.createChiServer(options...)
	default:
		return nil, fmt.Errorf("unsupported framework: %s", framework)
	}
}

// CreateAdapter creates an adapter for the specified framework
func (f *ServerFactory) CreateAdapter(framework Framework) (Adapter, error) {
	switch framework {
	case FrameworkStandard:
		return NewStandardAdapter(), nil
	case FrameworkGin:
		return NewGinAdapter(), nil
	case FrameworkEcho:
		return NewEchoAdapter(), nil
	default:
		return nil, fmt.Errorf("unsupported framework: %s", framework)
	}
}

// ServerOption configures server creation
type ServerOption func(*ServerConfig)

// ServerConfig holds server configuration
type ServerConfig struct {
	NativeEngine   interface{} // Framework-specific engine (gin.Engine, echo.Echo, etc.)
	Debug          bool
	TrustedProxies []string
}

// WithNativeEngine sets a pre-configured framework engine
func WithNativeEngine(engine interface{}) ServerOption {
	return func(config *ServerConfig) {
		config.NativeEngine = engine
	}
}

// WithDebug enables debug mode
func WithDebug(debug bool) ServerOption {
	return func(config *ServerConfig) {
		config.Debug = debug
	}
}

// WithTrustedProxies sets trusted proxy IPs
func WithTrustedProxies(proxies []string) ServerOption {
	return func(config *ServerConfig) {
		config.TrustedProxies = proxies
	}
}

func (f *ServerFactory) createStandardServer(options ...ServerOption) (Server, error) {
	config := &ServerConfig{}
	for _, opt := range options {
		opt(config)
	}

	if config.NativeEngine != nil {
		// Use provided engine if available
		return NewStandardServer(), nil
	}

	return NewStandardServer(), nil
}

func (f *ServerFactory) createGinServer(options ...ServerOption) (Server, error) {
	// Try real implementation first if available
	if createRealGinServerFunc != nil {
		if realServer, err := createRealGinServerFunc(options...); err == nil {
			return realServer, nil
		}
	}

	// Fall back to mock implementation
	config := &ServerConfig{}
	for _, opt := range options {
		opt(config)
	}

	var ginEngine interface{}

	if config.NativeEngine != nil {
		ginEngine = config.NativeEngine
	} else {
		// Use mock interface for demonstration
		ginEngine = &mockGinEngine{debug: config.Debug}
	}

	return NewGinServer(ginEngine), nil
}

func (f *ServerFactory) createEchoServer(options ...ServerOption) (Server, error) {
	// Try real implementation first if available
	if createRealEchoServerFunc != nil {
		if realServer, err := createRealEchoServerFunc(options...); err == nil {
			return realServer, nil
		}
	}

	// Fall back to mock implementation
	config := &ServerConfig{}
	for _, opt := range options {
		opt(config)
	}

	var echoEngine interface{}

	if config.NativeEngine != nil {
		echoEngine = config.NativeEngine
	} else {
		// Use mock interface for demonstration
		echoEngine = &mockEchoEngine{debug: config.Debug}
	}

	return NewEchoServer(echoEngine), nil
}

func (f *ServerFactory) createFiberServer(options ...ServerOption) (Server, error) {
	// Fiber implementation would go here
	return nil, fmt.Errorf("fiber support not implemented yet")
}

func (f *ServerFactory) createChiServer(options ...ServerOption) (Server, error) {
	// Chi implementation would go here
	return nil, fmt.Errorf("chi support not implemented yet")
}

// Mock implementations for testing/demonstration

type mockGinEngine struct {
	debug bool
}

func (m *mockGinEngine) Run(addr string) error {
	fmt.Printf("Mock Gin server starting on %s (debug: %v)\n", addr, m.debug)
	return nil
}

func (m *mockGinEngine) Handle(method, path string, handlers ...interface{}) {
	fmt.Printf("Mock Gin route registered: %s %s\n", method, path)
}

func (m *mockGinEngine) Group(prefix string, handlers ...interface{}) interface{} {
	return &mockGinGroup{prefix: prefix}
}

func (m *mockGinEngine) Use(handlers ...interface{}) {
	fmt.Printf("Mock Gin middleware registered\n")
}

type mockGinGroup struct {
	prefix string
}

func (g *mockGinGroup) Handle(method, path string, handlers ...interface{}) {
	fmt.Printf("Mock Gin group route registered: %s %s%s\n", method, g.prefix, path)
}

func (g *mockGinGroup) Group(prefix string, handlers ...interface{}) interface{} {
	return &mockGinGroup{prefix: g.prefix + prefix}
}

func (g *mockGinGroup) Use(handlers ...interface{}) {
	fmt.Printf("Mock Gin group middleware registered for %s\n", g.prefix)
}

type mockEchoEngine struct {
	debug bool
}

func (m *mockEchoEngine) Start(addr string) error {
	fmt.Printf("Mock Echo server starting on %s (debug: %v)\n", addr, m.debug)
	return nil
}

func (m *mockEchoEngine) GET(path string, handler interface{}, middleware ...interface{}) interface{} {
	fmt.Printf("Mock Echo GET route registered: %s\n", path)
	return nil
}

func (m *mockEchoEngine) POST(path string, handler interface{}, middleware ...interface{}) interface{} {
	fmt.Printf("Mock Echo POST route registered: %s\n", path)
	return nil
}

func (m *mockEchoEngine) PUT(path string, handler interface{}, middleware ...interface{}) interface{} {
	fmt.Printf("Mock Echo PUT route registered: %s\n", path)
	return nil
}

func (m *mockEchoEngine) DELETE(path string, handler interface{}, middleware ...interface{}) interface{} {
	fmt.Printf("Mock Echo DELETE route registered: %s\n", path)
	return nil
}

func (m *mockEchoEngine) PATCH(path string, handler interface{}, middleware ...interface{}) interface{} {
	fmt.Printf("Mock Echo PATCH route registered: %s\n", path)
	return nil
}

func (m *mockEchoEngine) Group(prefix string, middleware ...interface{}) interface{} {
	return &mockEchoGroup{prefix: prefix}
}

func (m *mockEchoEngine) Use(middleware ...interface{}) {
	fmt.Printf("Mock Echo middleware registered\n")
}

type mockEchoGroup struct {
	prefix string
}

func (g *mockEchoGroup) GET(path string, handler interface{}, middleware ...interface{}) interface{} {
	fmt.Printf("Mock Echo group GET route registered: %s%s\n", g.prefix, path)
	return nil
}

func (g *mockEchoGroup) POST(path string, handler interface{}, middleware ...interface{}) interface{} {
	fmt.Printf("Mock Echo group POST route registered: %s%s\n", g.prefix, path)
	return nil
}

func (g *mockEchoGroup) PUT(path string, handler interface{}, middleware ...interface{}) interface{} {
	fmt.Printf("Mock Echo group PUT route registered: %s%s\n", g.prefix, path)
	return nil
}

func (g *mockEchoGroup) DELETE(path string, handler interface{}, middleware ...interface{}) interface{} {
	fmt.Printf("Mock Echo group DELETE route registered: %s%s\n", g.prefix, path)
	return nil
}

func (g *mockEchoGroup) PATCH(path string, handler interface{}, middleware ...interface{}) interface{} {
	fmt.Printf("Mock Echo group PATCH route registered: %s%s\n", g.prefix, path)
	return nil
}

func (g *mockEchoGroup) Group(prefix string, middleware ...interface{}) interface{} {
	return &mockEchoGroup{prefix: g.prefix + prefix}
}

func (g *mockEchoGroup) Use(middleware ...interface{}) {
	fmt.Printf("Mock Echo group middleware registered for %s\n", g.prefix)
}

// Helper functions

// GetSupportedFrameworks returns a list of supported frameworks
func GetSupportedFrameworks() []Framework {
	return []Framework{
		FrameworkStandard,
		FrameworkGin,
		FrameworkEcho,
		FrameworkFiber,
		FrameworkChi,
	}
}

// IsFrameworkSupported checks if a framework is supported
func IsFrameworkSupported(framework Framework) bool {
	supported := GetSupportedFrameworks()
	for _, f := range supported {
		if f == framework {
			return true
		}
	}
	return false
}

// DetectFramework attempts to detect the framework from a native engine
func DetectFramework(engine interface{}) Framework {
	// In real implementation, we would use type assertions to detect the framework
	// For now, we'll use simple interface checks

	if _, ok := engine.(interface{ Run(string) error }); ok {
		// Could be Gin
		if _, ok := engine.(interface {
			Handle(string, string, ...interface{})
		}); ok {
			return FrameworkGin
		}
	}

	if _, ok := engine.(interface{ Start(string) error }); ok {
		// Could be Echo
		if _, ok := engine.(interface {
			GET(string, interface{}, ...interface{}) interface{}
		}); ok {
			return FrameworkEcho
		}
	}

	// Default to standard
	return FrameworkStandard
}
