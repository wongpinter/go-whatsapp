# WhatsApp Flows HTTP Server Framework Integration

## 🚀 **Overview**

The WhatsApp Flows package now supports framework-agnostic HTTP server integration, allowing you to use Flows with any supported HTTP framework (Gin, Echo, standard HTTP) without changing your code.

## 📋 **Table of Contents**

1. [Quick Start](#quick-start)
2. [Framework Support](#framework-support)
3. [Migration Guide](#migration-guide)
4. [Examples](#examples)
5. [Advanced Configuration](#advanced-configuration)
6. [Security](#security)
7. [Monitoring](#monitoring)
8. [Best Practices](#best-practices)

## 🏃 **Quick Start**

### **Basic Setup**

```go
package main

import (
    "github.com/wongpinter/go-whatsapp/flows"
    "github.com/wongpinter/go-whatsapp/internal/httpclient"
    "github.com/wongpinter/go-whatsapp/internal/httpserver"
)

func main() {
    // Create HTTP client manager
    clientManager := httpclient.NewManager(nil, logger)
    
    // Create Flows server factory
    factory := flows.NewServerFactory(clientManager, logger)
    
    // Create complete Flows server
    server, adapter, err := factory.CreateFullFlowsServer(
        httpserver.FrameworkStandard,
        "your-app-secret",
        "your-verify-token",
    )
    
    // Start server
    server.Start(":8080")
}
```

### **With Custom Framework**

```go
// Using Gin
server, adapter, err := factory.CreateFullFlowsServer(
    httpserver.FrameworkGin,
    appSecret,
    verifyToken,
    flows.WithRoutePrefix("/api/v1"),
    flows.WithRateLimit(true, 100),
)

// Using Echo
server, adapter, err := factory.CreateFullFlowsServer(
    httpserver.FrameworkEcho,
    appSecret,
    verifyToken,
    flows.WithMetrics(true),
    flows.WithSecurity(true),
)
```

## 🔧 **Framework Support**

### **Supported Frameworks**

| Framework | Status | Build Tag | Features |
|-----------|--------|-----------|----------|
| **Standard HTTP** | ✅ Complete | None | Zero dependencies, high reliability |
| **Gin** | ✅ Complete | `gin` | High performance, rich middleware |
| **Echo** | ✅ Complete | `echo` | Lightweight, fast routing |

### **Build Tags**

Use build tags to include real framework implementations:

```bash
# Standard HTTP only
go build .

# With Gin support
go build -tags gin .

# With Echo support  
go build -tags echo .

# With both frameworks
go build -tags "gin echo" .
```

## 📦 **Migration Guide**

### **From Direct net/http Usage**

**Before (Direct net/http):**

```go
// Old approach
http.Handle("/flows/data-exchange", dataExchangeHandler)
http.Handle("/webhook", webhookHandler)
http.ListenAndServe(":8080", nil)
```

**After (Framework Abstraction):**

```go
// New approach
factory := flows.NewServerFactory(clientManager, logger)
server, adapter, err := factory.CreateFullFlowsServer(
    httpserver.FrameworkStandard,
    appSecret,
    verifyToken,
)
server.Start(":8080")
```

### **From Existing Gin/Echo Applications**

**Integrating with Existing Gin App:**

```go
// Create your Gin engine
r := gin.New()
r.Use(gin.Logger(), gin.Recovery())

// Create Flows server with existing engine
server, adapter, err := factory.CreateFullFlowsServer(
    httpserver.FrameworkGin,
    appSecret,
    verifyToken,
    flows.WithServerOptions(httpserver.WithNativeEngine(r)),
    flows.WithRoutePrefix("/flows"),
)

// Your existing routes
r.GET("/", homeHandler)
r.GET("/api/users", getUsersHandler)

// Flows routes are automatically registered under /flows/*
```

**Integrating with Existing Echo App:**

```go
// Create your Echo instance
e := echo.New()
e.Use(middleware.Logger(), middleware.Recover())

// Create Flows server with existing instance
server, adapter, err := factory.CreateFullFlowsServer(
    httpserver.FrameworkEcho,
    appSecret,
    verifyToken,
    flows.WithServerOptions(httpserver.WithNativeEngine(e)),
    flows.WithRoutePrefix("/flows"),
)

// Your existing routes
e.GET("/", homeHandler)
e.GET("/api/users", getUsersHandler)
```

## 📚 **Examples**

### **Data Exchange Only**

For minimal Flows data exchange functionality:

```go
server, adapter, err := factory.CreateDataExchangeServer(
    httpserver.FrameworkStandard,
    flows.WithRateLimit(true, 60),
    flows.WithSecurity(true),
)
```

### **Custom Action Handlers**

```go
// Register custom action handlers
actionRouter := flows.NewActionRouter()
actionRouter.RegisterHandlerFunc("submit_survey", handleSurveySubmission)
actionRouter.RegisterHandlerFunc("book_appointment", handleAppointmentBooking)

server, adapter, err := factory.CreateFullFlowsServer(
    httpserver.FrameworkGin,
    appSecret,
    verifyToken,
    flows.WithActionRouter(actionRouter),
)
```

### **Custom Middleware**

```go
// Add custom middleware
router := server.Router()
router.Use(customAuthMiddleware())
router.Use(customLoggingMiddleware())

// Flows routes with custom middleware
flowsGroup := router.Group("/flows", rateLimitMiddleware())
adapter.RegisterRoutesWithPrefix(flowsGroup, "")
```

## ⚙️ **Advanced Configuration**

### **Server Factory Options**

```go
server, adapter, err := factory.CreateFullFlowsServer(
    httpserver.FrameworkGin,
    appSecret,
    verifyToken,
    
    // Framework options
    flows.WithFramework(httpserver.FrameworkGin),
    flows.WithServerOptions(
        httpserver.WithDebug(true),
        httpserver.WithNativeEngine(customGinEngine),
    ),
    
    // Route configuration
    flows.WithRoutePrefix("/api/v1/flows"),
    
    // Security options
    flows.WithSecurity(true),
    flows.WithRateLimit(true, 100),
    
    // Monitoring options
    flows.WithMetrics(true),
    
    // Custom components
    flows.WithActionRouter(customActionRouter),
    flows.WithTokenManager(customTokenManager),
    flows.WithStateManager(customStateManager),
    
    // Logging
    flows.WithLogger(customLogger),
)
```

### **Custom Data Exchange Handler**

```go
// Create custom handler
handler := flows.NewDataExchangeHandler(
    flows.WithActionRouter(actionRouter),
    flows.WithTokenManager(tokenManager),
    flows.WithStateManager(stateManager),
    flows.WithDataExchangeLogger(logger),
)

server, adapter, err := factory.CreateFullFlowsServer(
    framework,
    appSecret,
    verifyToken,
    flows.WithDataExchangeHandler(handler),
)
```

## 🔒 **Security**

### **Built-in Security Features**

* **Signature Validation**: Automatic webhook signature verification
* **Token Validation**: Flow token security validation
* **Rate Limiting**: Configurable rate limiting per endpoint
* **Encryption**: Request/response encryption support
* **CORS**: Cross-origin request handling
* **Security Headers**: Automatic security header injection

### **Security Configuration**

```go
server, adapter, err := factory.CreateFullFlowsServer(
    framework,
    appSecret,
    verifyToken,
    
    // Enable all security features
    flows.WithSecurity(true),
    
    // Configure rate limiting
    flows.WithRateLimit(true, 60), // 60 requests per minute
    
    // Custom security middleware
    flows.WithServerOptions(
        httpserver.WithMiddleware(customSecurityMiddleware()),
    ),
)
```

### **Custom Security Middleware**

```go
// Add custom security middleware
router := server.Router()
router.Use(flows.FlowsSecurityMiddleware(appSecret, logger))
router.Use(flows.FlowsRateLimitMiddleware(100, logger))
router.Use(flows.FlowsEncryptionMiddleware(privateKey, logger))
```

## 📊 **Monitoring**

### **Health Checks**

```bash
# Check service health
curl http://localhost:8080/flows/health

# Response
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "service": "whatsapp-flows",
  "version": "1.0.0",
  "uptime_seconds": 3600,
  "flows": {
    "active_flows": 5,
    "total_requests": 1000,
    "successful_flows": 950,
    "failed_flows": 50,
    "success_rate": 95.0
  }
}
```

### **Metrics**

```bash
# Get detailed metrics
curl http://localhost:8080/flows/metrics

# Response
{
  "timestamp": "2024-01-01T12:00:00Z",
  "uptime_seconds": 3600,
  "total_requests": 1000,
  "successful_flows": 950,
  "failed_flows": 50,
  "success_rate": 95.0,
  "performance": {
    "average_latency_ms": 150,
    "p95_latency_ms": 300,
    "p99_latency_ms": 500,
    "requests_per_second": 2.5
  },
  "action_executions": {
    "submit_survey": 400,
    "book_appointment": 300,
    "contact_form": 250
  }
}
```

### **Custom Metrics**

```go
// Access metrics from adapter
metrics := adapter.GetMetrics()
fmt.Printf("Total requests: %d\n", metrics.totalRequests)
fmt.Printf("Success rate: %.2f%%\n", metrics.calculateSuccessRate())
```

## 🎯 **Best Practices**

### **1. Framework Selection**

* **Standard HTTP**: Use for maximum compatibility and zero dependencies
* **Gin**: Use for high-performance applications with rich middleware needs
* **Echo**: Use for lightweight applications with fast routing requirements

### **2. Route Organization**

```go
// Organize routes with prefixes
flows.WithRoutePrefix("/api/v1/flows")  // /api/v1/flows/data-exchange
flows.WithRoutePrefix("/whatsapp")      // /whatsapp/data-exchange
```

### **3. Error Handling**

```go
// Implement comprehensive error handling
server, adapter, err := factory.CreateFullFlowsServer(...)
if err != nil {
    log.Fatalf("Failed to create Flows server: %v", err)
}

// Add error middleware
router.Use(errorHandlingMiddleware())
```

### **4. Graceful Shutdown**

```go
// Implement graceful shutdown
go func() {
    if err := server.Start(":8080"); err != nil {
        logger.Error().Err(err).Msg("Server failed")
    }
}()

// Wait for interrupt
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

// Graceful shutdown
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
server.Shutdown(ctx)
```

### **5. Production Configuration**

```go
server, adapter, err := factory.CreateFullFlowsServer(
    httpserver.FrameworkGin,
    appSecret,
    verifyToken,
    
    // Production settings
    flows.WithSecurity(true),
    flows.WithRateLimit(true, 100),
    flows.WithMetrics(true),
    flows.WithRoutePrefix("/api/v1"),
    
    // Monitoring
    flows.WithLogger(productionLogger),
)
```

## 🔗 **API Reference**

### **Server Factory**

* `NewServerFactory(clientManager, logger)` - Create server factory
* `CreateFullFlowsServer(framework, appSecret, verifyToken, options...)` - Complete server
* `CreateDataExchangeServer(framework, options...)` - Data exchange only

### **Configuration Options**

* `WithFramework(framework)` - Set HTTP framework
* `WithRoutePrefix(prefix)` - Set route prefix
* `WithSecurity(enabled)` - Enable security features
* `WithRateLimit(enabled, rpm)` - Configure rate limiting
* `WithMetrics(enabled)` - Enable metrics collection
* `WithLogger(logger)` - Set custom logger

### **Middleware**

* `FlowsLoggingMiddleware(logger)` - Request logging
* `FlowsMetricsMiddleware(metrics, logger)` - Metrics collection
* `FlowsSecurityMiddleware(appSecret, logger)` - Security headers
* `FlowsRateLimitMiddleware(rpm, logger)` - Rate limiting
* `FlowsEncryptionMiddleware(privateKey, logger)` - Encryption

## 🆘 **Troubleshooting**

### **Common Issues**

1. **Build Tag Not Working**: Ensure you're using the correct build tag syntax
2. **Route Conflicts**: Use route prefixes to avoid conflicts
3. **Middleware Order**: Apply middleware in the correct order
4. **Context Values**: Use Go context for passing values between middleware

### **Debug Mode**

```go
// Enable debug mode
flows.WithServerOptions(httpserver.WithDebug(true))
```

This comprehensive integration makes WhatsApp Flows truly framework-agnostic and production-ready! 🚀

## 📖 **Additional Resources**

* [HTTP Server Abstraction Documentation](./http-server-abstraction.md)
* [Flows Examples](../examples/flows-integration/)
* [Framework-Specific Examples](../examples/)
* [API Reference](./api-reference.md)

## 🤝 **Contributing**

To contribute to the Flows HTTP server integration:

1. Follow the existing patterns and interfaces
2. Add comprehensive tests for new features
3. Update documentation for any changes
4. Ensure backward compatibility

## 📝 **Changelog**

### **v1.0.0**

* ✅ Framework-agnostic HTTP server abstraction
* ✅ Support for Gin, Echo, and standard HTTP
* ✅ Comprehensive middleware system
* ✅ Built-in security and monitoring
* ✅ Production-ready examples
* ✅ Complete documentation
