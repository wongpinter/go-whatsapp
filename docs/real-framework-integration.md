# Real Framework Integration Guide

This guide shows how to integrate the WhatsApp package with real HTTP frameworks like Gin and Echo using build tags.

## Overview

The HTTP server abstraction supports both mock implementations (for testing and demonstration) and real framework implementations. Real implementations are enabled using Go build tags.

## Build Tags

### Available Build Tags

- `gin` - Enables real Gin framework integration
- `echo` - Enables real Echo framework integration

### How Build Tags Work

Build tags allow conditional compilation of Go code. When you build with a specific tag, only files with that tag (or no tag) are compiled.

```bash
# Build with Gin support
go build -tags gin

# Build with Echo support  
go build -tags echo

# Build with both Gin and Echo support
go build -tags "gin echo"

# Build without any framework (uses mock implementations)
go build
```

## Gin Integration

### Prerequisites

```bash
go get github.com/gin-gonic/gin
```

### Basic Integration

```go
//go:build gin
// +build gin

package main

import (
    "github.com/gin-gonic/gin"
    "github.com/wongpinter/go-whatsapp/internal/httpserver"
    "github.com/wongpinter/go-whatsapp/webhook"
)

func main() {
    // Option 1: Use existing Gin engine
    r := gin.Default()
    
    factory := httpserver.NewServerFactory()
    server, err := factory.CreateServer(
        httpserver.FrameworkGin,
        httpserver.WithNativeEngine(r),
    )
    if err != nil {
        panic(err)
    }
    
    // Setup webhook
    webhookAdapter := webhook.NewServerAdapter("secret", "token", logger)
    webhookAdapter.RegisterRoutes(server.Router())
    
    // Start server
    server.Start(":8080")
}
```

### Advanced Gin Integration

```go
func advancedGinExample() {
    // Create Gin with custom configuration
    gin.SetMode(gin.ReleaseMode)
    r := gin.New()
    r.Use(gin.Logger(), gin.Recovery())
    
    // Add existing routes
    r.GET("/existing", existingHandler)
    
    // Create server
    factory := httpserver.NewServerFactory()
    server, err := factory.CreateServer(
        httpserver.FrameworkGin,
        httpserver.WithNativeEngine(r),
        httpserver.WithDebug(false),
        httpserver.WithTrustedProxies([]string{"127.0.0.1"}),
    )
    if err != nil {
        panic(err)
    }
    
    // Setup webhook with custom prefix
    webhookAdapter := webhook.NewServerAdapter("secret", "token", logger)
    webhookAdapter.RegisterRoutesWithPrefix(server.Router(), "/whatsapp")
    
    // Add API routes using abstraction
    router := server.Router()
    apiGroup := router.Group("/api/v1")
    
    apiGroup.GET("/status", func(ctx httpserver.HTTPContext) error {
        return ctx.JSON(200, map[string]string{
            "status": "healthy",
            "framework": "gin",
        })
    })
    
    // Start with graceful shutdown
    startWithGracefulShutdown(server)
}
```

### Building and Running

```bash
# Build with Gin support
go build -tags gin -o webhook-server examples/gin-integration/main.go

# Run the server
./webhook-server
```

## Echo Integration

### Prerequisites

```bash
go get github.com/labstack/echo/v4
go get github.com/labstack/echo/v4/middleware
```

### Basic Integration

```go
//go:build echo
// +build echo

package main

import (
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/wongpinter/go-whatsapp/internal/httpserver"
    "github.com/wongpinter/go-whatsapp/webhook"
)

func main() {
    // Option 1: Use existing Echo instance
    e := echo.New()
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    
    factory := httpserver.NewServerFactory()
    server, err := factory.CreateServer(
        httpserver.FrameworkEcho,
        httpserver.WithNativeEngine(e),
    )
    if err != nil {
        panic(err)
    }
    
    // Setup webhook
    webhookAdapter := webhook.NewServerAdapter("secret", "token", logger)
    webhookAdapter.RegisterRoutes(server.Router())
    
    // Start server
    server.Start(":8080")
}
```

### Advanced Echo Integration

```go
func advancedEchoExample() {
    // Create Echo with custom configuration
    e := echo.New()
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())
    
    // Add existing routes
    e.GET("/existing", existingHandler)
    
    // Create server
    factory := httpserver.NewServerFactory()
    server, err := factory.CreateServer(
        httpserver.FrameworkEcho,
        httpserver.WithNativeEngine(e),
        httpserver.WithDebug(false),
    )
    if err != nil {
        panic(err)
    }
    
    // Setup webhook
    webhookAdapter := webhook.NewServerAdapter("secret", "token", logger)
    router := server.Router()
    
    // Add global middleware
    router.Use(timeoutMiddleware(30 * time.Second))
    
    // Register webhook routes
    webhookAdapter.RegisterRoutes(router)
    
    // Add API routes
    apiGroup := router.Group("/api/v1")
    apiGroup.GET("/health", healthHandler)
    
    // Start with graceful shutdown
    startWithGracefulShutdown(server)
}
```

### Building and Running

```bash
# Build with Echo support
go build -tags echo -o webhook-server examples/echo-integration/main.go

# Run the server
./webhook-server
```

## Framework Detection

The factory automatically detects which framework implementation to use:

```go
func frameworkDetectionExample() {
    factory := httpserver.NewServerFactory()
    
    // When built with gin tag, this will use real Gin implementation
    // When built without tags, this will use mock implementation
    server, err := factory.CreateServer(httpserver.FrameworkGin)
    if err != nil {
        log.Printf("Gin not available: %v", err)
    }
    
    // Check what was actually created
    nativeEngine := server.Native()
    detectedFramework := httpserver.DetectFramework(nativeEngine)
    log.Printf("Detected framework: %s", detectedFramework)
}
```

## Migration Strategies

### From net/http to Gin

**Before:**
```go
http.HandleFunc("/webhook", webhookHandler)
http.ListenAndServe(":8080", nil)
```

**After:**
```go
// Build with: go build -tags gin
r := gin.Default()
factory := httpserver.NewServerFactory()
server, _ := factory.CreateServer(httpserver.FrameworkGin, 
    httpserver.WithNativeEngine(r))
webhookAdapter.RegisterRoutes(server.Router())
server.Start(":8080")
```

### From Echo to Abstraction

**Before:**
```go
e := echo.New()
e.POST("/webhook", echoWebhookHandler)
e.Start(":8080")
```

**After:**
```go
// Build with: go build -tags echo
e := echo.New()
factory := httpserver.NewServerFactory()
server, _ := factory.CreateServer(httpserver.FrameworkEcho,
    httpserver.WithNativeEngine(e))
webhookAdapter.RegisterRoutes(server.Router())
server.Start(":8080")
```

## Best Practices

### 1. Use Build Tags in Your Application

```go
//go:build gin
// +build gin

package main

// Your Gin-specific code here
```

### 2. Provide Fallbacks

```go
func createServer() (httpserver.Server, error) {
    factory := httpserver.NewServerFactory()
    
    // Try Gin first
    if server, err := factory.CreateServer(httpserver.FrameworkGin); err == nil {
        return server, nil
    }
    
    // Fall back to standard HTTP
    return factory.CreateServer(httpserver.FrameworkStandard)
}
```

### 3. Environment-Based Configuration

```go
func createServerFromEnv() (httpserver.Server, error) {
    framework := os.Getenv("HTTP_FRAMEWORK")
    if framework == "" {
        framework = "standard"
    }
    
    factory := httpserver.NewServerFactory()
    return factory.CreateServer(httpserver.Framework(framework))
}
```

### 4. Graceful Shutdown

```go
func startWithGracefulShutdown(server httpserver.Server) {
    go func() {
        if err := server.Start(":8080"); err != nil {
            log.Printf("Server error: %v", err)
        }
    }()
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
}
```

## Testing

### Testing with Different Frameworks

```bash
# Test with standard HTTP
go test ./...

# Test with Gin
go test -tags gin ./...

# Test with Echo
go test -tags echo ./...

# Test with all frameworks
go test -tags "gin echo" ./...
```

### Framework-Specific Tests

```go
//go:build gin
// +build gin

func TestGinIntegration(t *testing.T) {
    r := gin.New()
    factory := httpserver.NewServerFactory()
    server, err := factory.CreateServer(httpserver.FrameworkGin,
        httpserver.WithNativeEngine(r))
    
    assert.NoError(t, err)
    assert.IsType(t, &gin.Engine{}, server.Native())
}
```

## Deployment

### Docker with Build Args

```dockerfile
FROM golang:1.21-alpine AS builder

ARG FRAMEWORK_TAGS=""
WORKDIR /app
COPY . .
RUN go build -tags "${FRAMEWORK_TAGS}" -o webhook-server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/webhook-server .
CMD ["./webhook-server"]
```

Build with different frameworks:

```bash
# Build with Gin
docker build --build-arg FRAMEWORK_TAGS=gin -t webhook-server:gin .

# Build with Echo
docker build --build-arg FRAMEWORK_TAGS=echo -t webhook-server:echo .

# Build with both
docker build --build-arg FRAMEWORK_TAGS="gin echo" -t webhook-server:full .
```

## Troubleshooting

### Common Issues

1. **Build tag not working**: Ensure the build tag comment is the first line of the file
2. **Framework not detected**: Check that dependencies are installed and build tags are used
3. **Import errors**: Make sure framework packages are available when building with tags

### Debug Framework Detection

```go
func debugFrameworkDetection() {
    factory := httpserver.NewServerFactory()
    
    frameworks := []httpserver.Framework{
        httpserver.FrameworkStandard,
        httpserver.FrameworkGin,
        httpserver.FrameworkEcho,
    }
    
    for _, framework := range frameworks {
        server, err := factory.CreateServer(framework)
        if err != nil {
            log.Printf("❌ %s: %v", framework, err)
        } else {
            log.Printf("✅ %s: %T", framework, server.Native())
        }
    }
}
```

This integration guide provides everything needed to use real HTTP frameworks with the WhatsApp package while maintaining the flexibility of the abstraction layer.
