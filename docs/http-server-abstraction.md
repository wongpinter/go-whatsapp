# HTTP Server Framework Abstraction

## Overview

The HTTP Server Framework Abstraction provides a unified interface for building HTTP servers that can work with different Go web frameworks like Gin, Echo, Fiber, Chi, and the standard `net/http` library. This abstraction allows your WhatsApp package to be easily integrated into existing projects regardless of their chosen HTTP framework.

## Problem Solved

Previously, the WhatsApp package was tightly coupled to `net/http`, making it difficult to integrate with projects using other popular frameworks. This abstraction solves:

1. **Framework Lock-in**: No longer tied to a specific HTTP framework
2. **Integration Complexity**: Easy integration with existing projects
3. **Code Duplication**: Single codebase works with multiple frameworks
4. **Migration Barriers**: Smooth migration between frameworks

## Architecture

### Core Components

#### 1. HTTPContext Interface
```go
type HTTPContext interface {
    // Request methods
    Method() string
    Path() string
    Query(key string) string
    Header(key string) string
    Body() ([]byte, error)
    Context() context.Context
    WithContext(ctx context.Context)

    // Response methods
    Status(code int)
    SetHeader(key, value string)
    Write(data []byte) error
    JSON(code int, obj interface{}) error
    String(code int, format string, values ...interface{}) error

    // Framework-specific context
    Native() interface{}
}
```

#### 2. Framework Adapters
- **StandardAdapter**: For `net/http`
- **GinAdapter**: For Gin framework
- **EchoAdapter**: For Echo framework
- **FiberAdapter**: For Fiber framework (planned)
- **ChiAdapter**: For Chi framework (planned)

#### 3. Router Interface
```go
type Router interface {
    GET(path string, handler HandlerFunc, middleware ...Middleware)
    POST(path string, handler HandlerFunc, middleware ...Middleware)
    PUT(path string, handler HandlerFunc, middleware ...Middleware)
    DELETE(path string, handler HandlerFunc, middleware ...Middleware)
    PATCH(path string, handler HandlerFunc, middleware ...Middleware)
    
    Group(prefix string, middleware ...Middleware) Router
    Use(middleware ...Middleware)
    
    Native() interface{}
}
```

#### 4. Server Interface
```go
type Server interface {
    Router() Router
    Start(addr string) error
    Shutdown(ctx context.Context) error
    Native() interface{}
}
```

## Supported Frameworks

### Currently Implemented
- ✅ **net/http** (Standard library)
- ✅ **Gin** (Mock implementation for demonstration)
- ✅ **Echo** (Mock implementation for demonstration)

### Planned
- 🔄 **Fiber** (High-performance framework)
- 🔄 **Chi** (Lightweight router)
- 🔄 **Gorilla Mux** (Popular router)

## Usage Examples

### 1. Standard net/http Server

```go
package main

import (
    "github.com/wongpinter/go-whatsapp/internal/httpserver"
    "github.com/wongpinter/go-whatsapp/webhook"
)

func main() {
    // Create server factory
    factory := httpserver.NewServerFactory()
    
    // Create standard HTTP server
    server, err := factory.CreateServer(httpserver.FrameworkStandard)
    if err != nil {
        panic(err)
    }
    
    // Setup webhook
    webhookAdapter := webhook.NewServerAdapter("app-secret", "verify-token", logger)
    webhookAdapter.RegisterRoutes(server.Router())
    
    // Start server
    server.Start(":8080")
}
```

### 2. Gin Framework Integration

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/wongpinter/go-whatsapp/internal/httpserver"
    "github.com/wongpinter/go-whatsapp/webhook"
)

func main() {
    // Create Gin engine
    ginEngine := gin.New()
    
    // Create server with existing Gin engine
    factory := httpserver.NewServerFactory()
    server, err := factory.CreateServer(
        httpserver.FrameworkGin,
        httpserver.WithNativeEngine(ginEngine),
    )
    if err != nil {
        panic(err)
    }
    
    // Setup webhook routes
    webhookAdapter := webhook.NewServerAdapter("app-secret", "verify-token", logger)
    webhookAdapter.RegisterRoutes(server.Router())
    
    // Add custom routes
    router := server.Router()
    router.GET("/custom", func(ctx httpserver.HTTPContext) error {
        return ctx.JSON(200, gin.H{"message": "Custom route"})
    })
    
    // Start server
    server.Start(":8080")
}
```

### 3. Echo Framework Integration

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/wongpinter/go-whatsapp/internal/httpserver"
    "github.com/wongpinter/go-whatsapp/webhook"
)

func main() {
    // Create Echo instance
    echoEngine := echo.New()
    
    // Create server with existing Echo instance
    factory := httpserver.NewServerFactory()
    server, err := factory.CreateServer(
        httpserver.FrameworkEcho,
        httpserver.WithNativeEngine(echoEngine),
    )
    if err != nil {
        panic(err)
    }
    
    // Setup webhook routes
    webhookAdapter := webhook.NewServerAdapter("app-secret", "verify-token", logger)
    webhookAdapter.RegisterRoutes(server.Router())
    
    // Start server
    server.Start(":8080")
}
```

### 4. Framework-Agnostic Webhook Setup

```go
func setupWebhookServer(framework httpserver.Framework) error {
    factory := httpserver.NewServerFactory()
    server, err := factory.CreateServer(framework)
    if err != nil {
        return err
    }
    
    // Create webhook adapter
    webhookAdapter := webhook.NewServerAdapter(
        os.Getenv("WHATSAPP_APP_SECRET"),
        os.Getenv("WHATSAPP_VERIFY_TOKEN"),
        logger,
    )
    
    // Register routes
    router := server.Router()
    
    // Add middleware
    router.Use(
        webhook.LoggingMiddleware(logger),
        webhook.CORSMiddleware(),
        webhook.TimeoutMiddleware(30*time.Second),
    )
    
    // Register webhook routes
    webhookAdapter.RegisterRoutes(router)
    
    // Add health check
    router.GET("/health", func(ctx httpserver.HTTPContext) error {
        return ctx.JSON(200, map[string]string{
            "status": "healthy",
            "framework": string(framework),
        })
    })
    
    return server.Start(":8080")
}
```

## Middleware System

The abstraction includes a framework-agnostic middleware system:

### Built-in Middleware

```go
// Logging middleware
router.Use(webhook.LoggingMiddleware(logger))

// CORS middleware
router.Use(webhook.CORSMiddleware())

// Timeout middleware
router.Use(webhook.TimeoutMiddleware(30 * time.Second))

// Signature validation middleware (for webhooks)
router.POST("/webhook", handler, webhookAdapter.SignatureValidationMiddleware())
```

### Custom Middleware

```go
func customMiddleware() httpserver.Middleware {
    return func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
        return func(ctx httpserver.HTTPContext) error {
            // Pre-processing
            start := time.Now()
            
            // Call next handler
            err := next(ctx)
            
            // Post-processing
            duration := time.Since(start)
            log.Printf("Request took %v", duration)
            
            return err
        }
    }
}
```

## Migration Guide

### From net/http to Framework Abstraction

**Before:**
```go
http.HandleFunc("/webhook", webhookHandler.ServeHTTP)
http.ListenAndServe(":8080", nil)
```

**After:**
```go
factory := httpserver.NewServerFactory()
server, _ := factory.CreateServer(httpserver.FrameworkStandard)
webhookAdapter.RegisterRoutes(server.Router())
server.Start(":8080")
```

### Adding to Existing Gin Project

**Before:**
```go
r := gin.Default()
r.POST("/webhook", ginWebhookHandler)
r.Run(":8080")
```

**After:**
```go
r := gin.Default()
server, _ := factory.CreateServer(httpserver.FrameworkGin, 
    httpserver.WithNativeEngine(r))
webhookAdapter.RegisterRoutes(server.Router())
r.Run(":8080")
```

## Benefits

### 1. **Framework Flexibility**
- Switch between frameworks without changing webhook code
- Support multiple frameworks in the same codebase
- Easy migration path between frameworks

### 2. **Consistent API**
- Same interface regardless of underlying framework
- Unified middleware system
- Consistent error handling

### 3. **Easy Integration**
- Drop-in replacement for existing HTTP handlers
- Works with existing middleware
- Minimal changes to existing code

### 4. **Performance**
- No performance overhead for framework-specific optimizations
- Direct access to native framework features when needed
- Efficient adapter implementations

## Best Practices

### 1. **Use Factory Pattern**
```go
factory := httpserver.NewServerFactory()
server, err := factory.CreateServer(framework, options...)
```

### 2. **Leverage Middleware**
```go
router.Use(
    LoggingMiddleware(logger),
    CORSMiddleware(),
    AuthMiddleware(),
)
```

### 3. **Group Related Routes**
```go
apiGroup := router.Group("/api/v1")
apiGroup.GET("/users", getUsersHandler)
apiGroup.POST("/users", createUserHandler)
```

### 4. **Handle Errors Gracefully**
```go
func handler(ctx httpserver.HTTPContext) error {
    if err := someOperation(); err != nil {
        return ctx.JSON(500, map[string]string{
            "error": err.Error(),
        })
    }
    return ctx.JSON(200, result)
}
```

## Testing

The abstraction makes testing easier by providing mock implementations:

```go
func TestWebhookHandler(t *testing.T) {
    // Create test server
    factory := httpserver.NewServerFactory()
    server, _ := factory.CreateServer(httpserver.FrameworkStandard)
    
    // Setup webhook
    webhookAdapter := webhook.NewServerAdapter("secret", "token", logger)
    webhookAdapter.RegisterRoutes(server.Router())
    
    // Test webhook endpoint
    // ... test implementation
}
```

## Future Enhancements

1. **Additional Framework Support**: Fiber, Chi, Gorilla Mux
2. **Advanced Middleware**: Rate limiting, authentication, caching
3. **Metrics Integration**: Prometheus, custom metrics
4. **WebSocket Support**: Real-time communication
5. **gRPC Integration**: Protocol buffer support

This abstraction provides a solid foundation for building framework-agnostic HTTP services while maintaining the flexibility to use framework-specific features when needed.
