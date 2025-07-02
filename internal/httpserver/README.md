# HTTP Server Framework Abstraction

A framework-agnostic HTTP server abstraction that allows the WhatsApp package to work seamlessly with different Go web frameworks.

## Quick Start

### Standard net/http

```go
package main

import (
    "github.com/wongpinter/go-whatsapp/internal/httpserver"
    "github.com/wongpinter/go-whatsapp/webhook"
)

func main() {
    // Create server
    factory := httpserver.NewServerFactory()
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

### Gin Framework

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/wongpinter/go-whatsapp/internal/httpserver"
    "github.com/wongpinter/go-whatsapp/webhook"
)

func main() {
    // Create or use existing Gin engine
    r := gin.Default()
    
    // Create server with Gin engine
    factory := httpserver.NewServerFactory()
    server, err := factory.CreateServer(
        httpserver.FrameworkGin,
        httpserver.WithNativeEngine(r),
    )
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

### Echo Framework

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/wongpinter/go-whatsapp/internal/httpserver"
    "github.com/wongpinter/go-whatsapp/webhook"
)

func main() {
    // Create or use existing Echo instance
    e := echo.New()
    
    // Create server with Echo instance
    factory := httpserver.NewServerFactory()
    server, err := factory.CreateServer(
        httpserver.FrameworkEcho,
        httpserver.WithNativeEngine(e),
    )
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

## Supported Frameworks

- ✅ **net/http** - Standard library
- ✅ **Gin** - High-performance HTTP web framework
- ✅ **Echo** - High performance, minimalist Go web framework
- 🔄 **Fiber** - Express inspired web framework (planned)
- 🔄 **Chi** - Lightweight, idiomatic router (planned)

## Features

### Framework-Agnostic API
- Unified interface across all frameworks
- Consistent middleware system
- Same codebase works with any supported framework

### Easy Integration
- Drop-in replacement for existing HTTP handlers
- Works with existing middleware
- Minimal changes to existing code

### Middleware Support
```go
// Built-in middleware
router.Use(
    webhook.LoggingMiddleware(logger),
    webhook.CORSMiddleware(),
    webhook.TimeoutMiddleware(30*time.Second),
)

// Custom middleware
router.Use(func(next httpserver.HandlerFunc) httpserver.HandlerFunc {
    return func(ctx httpserver.HTTPContext) error {
        // Your middleware logic
        return next(ctx)
    }
})
```

### Route Groups
```go
// Create API group
apiGroup := router.Group("/api/v1")
apiGroup.GET("/users", getUsersHandler)
apiGroup.POST("/users", createUserHandler)

// Nested groups
adminGroup := apiGroup.Group("/admin", authMiddleware())
adminGroup.DELETE("/users/:id", deleteUserHandler)
```

## Architecture

### Core Interfaces

#### HTTPContext
Provides framework-agnostic request/response handling:
```go
type HTTPContext interface {
    Method() string
    Path() string
    Query(key string) string
    Header(key string) string
    Body() ([]byte, error)
    
    Status(code int)
    SetHeader(key, value string)
    JSON(code int, obj interface{}) error
    String(code int, format string, values ...interface{}) error
}
```

#### Router
Provides framework-agnostic routing:
```go
type Router interface {
    GET(path string, handler HandlerFunc, middleware ...Middleware)
    POST(path string, handler HandlerFunc, middleware ...Middleware)
    Group(prefix string, middleware ...Middleware) Router
    Use(middleware ...Middleware)
}
```

#### Server
Provides framework-agnostic server management:
```go
type Server interface {
    Router() Router
    Start(addr string) error
    Shutdown(ctx context.Context) error
}
```

## Examples

See `examples/httpserver/main.go` for comprehensive examples showing:
- Standard HTTP server setup
- Gin integration
- Echo integration
- Framework detection
- Webhook integration
- Middleware usage
- Route grouping

## Migration Guide

### From net/http
Replace direct HTTP handler registration with the abstraction:

**Before:**
```go
http.HandleFunc("/webhook", handler)
http.ListenAndServe(":8080", nil)
```

**After:**
```go
factory := httpserver.NewServerFactory()
server, _ := factory.CreateServer(httpserver.FrameworkStandard)
server.Router().POST("/webhook", handler)
server.Start(":8080")
```

### Adding to Existing Projects
The abstraction can be added to existing projects without breaking changes:

```go
// Your existing Gin setup
r := gin.Default()
r.GET("/existing", existingHandler)

// Add WhatsApp webhook support
factory := httpserver.NewServerFactory()
server, _ := factory.CreateServer(httpserver.FrameworkGin, 
    httpserver.WithNativeEngine(r))
webhookAdapter.RegisterRoutes(server.Router())

// Continue using Gin as normal
r.Run(":8080")
```

## Testing

The abstraction makes testing easier by providing consistent interfaces:

```go
func TestHandler(t *testing.T) {
    factory := httpserver.NewServerFactory()
    server, _ := factory.CreateServer(httpserver.FrameworkStandard)
    
    // Setup routes
    server.Router().GET("/test", testHandler)
    
    // Test with any framework
}
```

## Performance

The abstraction adds minimal overhead:
- Direct delegation to framework-specific implementations
- No reflection or runtime type checking in hot paths
- Framework-specific optimizations preserved

## Contributing

To add support for a new framework:

1. Implement the `HTTPContext` interface for the framework
2. Implement the `Adapter` interface
3. Implement the `Router` and `Server` interfaces
4. Add factory support in `factory.go`
5. Add tests and examples

See existing implementations (`gin_adapter.go`, `echo_adapter.go`) for reference.

## License

This package is part of the go-whatsapp project and follows the same license terms.
