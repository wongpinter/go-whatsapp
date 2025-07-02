package httpserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// GinContext implements HTTPContext for Gin framework
type GinContext struct {
	ginCtx interface{} // gin.Context - using interface{} to avoid import dependency
}

func NewGinContext(ginCtx interface{}) *GinContext {
	return &GinContext{ginCtx: ginCtx}
}

// Request methods
func (c *GinContext) Method() string {
	// In real implementation: return c.ginCtx.(*gin.Context).Request.Method
	if ctx, ok := c.ginCtx.(interface {
		Request() *http.Request
	}); ok {
		return ctx.Request().Method
	}
	return "GET" // fallback
}

func (c *GinContext) Path() string {
	// In real implementation: return c.ginCtx.(*gin.Context).Request.URL.Path
	if ctx, ok := c.ginCtx.(interface {
		Request() *http.Request
	}); ok {
		return ctx.Request().URL.Path
	}
	return "/" // fallback
}

func (c *GinContext) Query(key string) string {
	// In real implementation: return c.ginCtx.(*gin.Context).Query(key)
	if ctx, ok := c.ginCtx.(interface{ Query(string) string }); ok {
		return ctx.Query(key)
	}
	return "" // fallback
}

func (c *GinContext) Header(key string) string {
	// In real implementation: return c.ginCtx.(*gin.Context).GetHeader(key)
	if ctx, ok := c.ginCtx.(interface{ GetHeader(string) string }); ok {
		return ctx.GetHeader(key)
	}
	return "" // fallback
}

func (c *GinContext) Body() ([]byte, error) {
	// In real implementation: return io.ReadAll(c.ginCtx.(*gin.Context).Request.Body)
	if ctx, ok := c.ginCtx.(interface {
		Request() *http.Request
	}); ok {
		return io.ReadAll(ctx.Request().Body)
	}
	return nil, fmt.Errorf("unable to read body")
}

func (c *GinContext) Context() context.Context {
	// In real implementation: return c.ginCtx.(*gin.Context).Request.Context()
	if ctx, ok := c.ginCtx.(interface {
		Request() *http.Request
	}); ok {
		return ctx.Request().Context()
	}
	return context.Background()
}

func (c *GinContext) WithContext(ctx context.Context) {
	// In real implementation: c.ginCtx.(*gin.Context).Request = c.ginCtx.(*gin.Context).Request.WithContext(ctx)
	// Gin contexts are typically not modified this way
}

// Response methods
func (c *GinContext) Status(code int) {
	// In real implementation: c.ginCtx.(*gin.Context).Status(code)
	if ctx, ok := c.ginCtx.(interface{ Status(int) }); ok {
		ctx.Status(code)
	}
}

func (c *GinContext) SetHeader(key, value string) {
	// In real implementation: c.ginCtx.(*gin.Context).Header(key, value)
	if ctx, ok := c.ginCtx.(interface{ Header(string, string) }); ok {
		ctx.Header(key, value)
	}
}

func (c *GinContext) Write(data []byte) error {
	// In real implementation: c.ginCtx.(*gin.Context).Data(http.StatusOK, "application/octet-stream", data)
	if ctx, ok := c.ginCtx.(interface{ Data(int, string, []byte) }); ok {
		ctx.Data(http.StatusOK, "application/octet-stream", data)
		return nil
	}
	return fmt.Errorf("unable to write data")
}

func (c *GinContext) JSON(code int, obj interface{}) error {
	// In real implementation: c.ginCtx.(*gin.Context).JSON(code, obj)
	if ctx, ok := c.ginCtx.(interface{ JSON(int, interface{}) }); ok {
		ctx.JSON(code, obj)
		return nil
	}
	return fmt.Errorf("unable to write JSON")
}

func (c *GinContext) String(code int, format string, values ...interface{}) error {
	// In real implementation: c.ginCtx.(*gin.Context).String(code, format, values...)
	if ctx, ok := c.ginCtx.(interface {
		String(int, string, ...interface{})
	}); ok {
		ctx.String(code, format, values...)
		return nil
	}
	return fmt.Errorf("unable to write string")
}

func (c *GinContext) Native() interface{} {
	return c.ginCtx
}

// GinAdapter implements Adapter for Gin framework
type GinAdapter struct{}

func NewGinAdapter() *GinAdapter {
	return &GinAdapter{}
}

func (a *GinAdapter) WrapContext(native interface{}) HTTPContext {
	return NewGinContext(native)
}

func (a *GinAdapter) WrapHandler(handler HandlerFunc) interface{} {
	// Returns gin.HandlerFunc
	return func(ginCtx interface{}) {
		ctx := NewGinContext(ginCtx)
		if err := handler(ctx); err != nil {
			// In real implementation: ginCtx.(*gin.Context).JSON(500, gin.H{"error": err.Error()})
			if c, ok := ginCtx.(interface{ JSON(int, interface{}) }); ok {
				c.JSON(500, map[string]string{"error": err.Error()})
			}
		}
	}
}

func (a *GinAdapter) WrapMiddleware(middleware Middleware) interface{} {
	// Returns gin.HandlerFunc that can be used as middleware
	return func(ginCtx interface{}) {
		ctx := NewGinContext(ginCtx)

		nextHandler := func(ctx HTTPContext) error {
			// In real implementation: ginCtx.(*gin.Context).Next()
			if c, ok := ginCtx.(interface{ Next() }); ok {
				c.Next()
			}
			return nil
		}

		wrappedHandler := middleware(nextHandler)
		if err := wrappedHandler(ctx); err != nil {
			// In real implementation: ginCtx.(*gin.Context).JSON(500, gin.H{"error": err.Error()})
			if c, ok := ginCtx.(interface{ JSON(int, interface{}) }); ok {
				c.JSON(500, map[string]string{"error": err.Error()})
			}
			// In real implementation: ginCtx.(*gin.Context).Abort()
			if c, ok := ginCtx.(interface{ Abort() }); ok {
				c.Abort()
			}
		}
	}
}

func (a *GinAdapter) Name() string {
	return "gin"
}

// GinRouter implements Router for Gin framework
type GinRouter struct {
	ginEngine interface{} // gin.Engine - using interface{} to avoid import dependency
	adapter   *GinAdapter
	group     interface{} // gin.RouterGroup
}

func NewGinRouter(ginEngine interface{}) *GinRouter {
	return &GinRouter{
		ginEngine: ginEngine,
		adapter:   NewGinAdapter(),
		group:     ginEngine, // Initially, group is the engine itself
	}
}

func (r *GinRouter) GET(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("GET", path, handler, middleware...)
}

func (r *GinRouter) POST(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("POST", path, handler, middleware...)
}

func (r *GinRouter) PUT(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("PUT", path, handler, middleware...)
}

func (r *GinRouter) DELETE(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("DELETE", path, handler, middleware...)
}

func (r *GinRouter) PATCH(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("PATCH", path, handler, middleware...)
}

func (r *GinRouter) addRoute(method, path string, handler HandlerFunc, middleware ...Middleware) {
	// Convert middleware to gin middleware
	ginMiddleware := make([]interface{}, len(middleware))
	for i, mw := range middleware {
		ginMiddleware[i] = r.adapter.WrapMiddleware(mw)
	}

	// Convert handler to gin handler
	ginHandler := r.adapter.WrapHandler(handler)

	// Combine middleware and handler
	handlers := append(ginMiddleware, ginHandler)

	// In real implementation, we would call the appropriate method on the gin router
	// r.group.(*gin.RouterGroup).Handle(method, path, handlers...)

	// For demonstration, we'll use a generic interface
	if router, ok := r.group.(interface {
		Handle(string, string, ...interface{})
	}); ok {
		router.Handle(method, path, handlers...)
	}
}

func (r *GinRouter) Group(prefix string, middleware ...Middleware) Router {
	// Convert middleware to gin middleware
	ginMiddleware := make([]interface{}, len(middleware))
	for i, mw := range middleware {
		ginMiddleware[i] = r.adapter.WrapMiddleware(mw)
	}

	// In real implementation: newGroup := r.group.(*gin.RouterGroup).Group(prefix, ginMiddleware...)
	var newGroup interface{}
	if router, ok := r.group.(interface {
		Group(string, ...interface{}) interface{}
	}); ok {
		newGroup = router.Group(prefix, ginMiddleware...)
	} else {
		newGroup = r.group // fallback
	}

	return &GinRouter{
		ginEngine: r.ginEngine,
		adapter:   r.adapter,
		group:     newGroup,
	}
}

func (r *GinRouter) Use(middleware ...Middleware) {
	ginMiddleware := make([]interface{}, len(middleware))
	for i, mw := range middleware {
		ginMiddleware[i] = r.adapter.WrapMiddleware(mw)
	}

	// In real implementation: r.group.(*gin.RouterGroup).Use(ginMiddleware...)
	if router, ok := r.group.(interface {
		Use(...interface{})
	}); ok {
		router.Use(ginMiddleware...)
	}
}

func (r *GinRouter) Native() interface{} {
	return r.ginEngine
}

// GinServer implements Server for Gin framework
type GinServer struct {
	router    *GinRouter
	ginEngine interface{} // gin.Engine
}

func NewGinServer(ginEngine interface{}) *GinServer {
	return &GinServer{
		router:    NewGinRouter(ginEngine),
		ginEngine: ginEngine,
	}
}

func (s *GinServer) Router() Router {
	return s.router
}

func (s *GinServer) Start(addr string) error {
	// In real implementation: return s.ginEngine.(*gin.Engine).Run(addr)
	if engine, ok := s.ginEngine.(interface{ Run(string) error }); ok {
		return engine.Run(addr)
	}
	return fmt.Errorf("unable to start gin server")
}

func (s *GinServer) Shutdown(ctx context.Context) error {
	// Gin doesn't have built-in graceful shutdown, but we can implement it
	// In real implementation, you'd need to wrap the gin engine in an http.Server
	return nil
}

func (s *GinServer) Native() interface{} {
	return s.ginEngine
}
