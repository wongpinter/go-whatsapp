//go:build gin
// +build gin

package httpserver

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func init() {
	// Register the real Gin server creation function
	createRealGinServerFunc = func(options ...ServerOption) (Server, error) {
		return CreateRealGinServer(options...)
	}
}

// RealGinContext implements HTTPContext for real Gin framework
type RealGinContext struct {
	ginCtx *gin.Context
}

func NewRealGinContext(ginCtx *gin.Context) *RealGinContext {
	return &RealGinContext{ginCtx: ginCtx}
}

// Request methods
func (c *RealGinContext) Method() string {
	return c.ginCtx.Request.Method
}

func (c *RealGinContext) Path() string {
	return c.ginCtx.Request.URL.Path
}

func (c *RealGinContext) Query(key string) string {
	return c.ginCtx.Query(key)
}

func (c *RealGinContext) Header(key string) string {
	return c.ginCtx.GetHeader(key)
}

func (c *RealGinContext) Body() ([]byte, error) {
	return io.ReadAll(c.ginCtx.Request.Body)
}

func (c *RealGinContext) Context() context.Context {
	return c.ginCtx.Request.Context()
}

func (c *RealGinContext) WithContext(ctx context.Context) {
	c.ginCtx.Request = c.ginCtx.Request.WithContext(ctx)
}

// Response methods
func (c *RealGinContext) Status(code int) {
	c.ginCtx.Status(code)
}

func (c *RealGinContext) SetHeader(key, value string) {
	c.ginCtx.Header(key, value)
}

func (c *RealGinContext) Write(data []byte) error {
	c.ginCtx.Data(http.StatusOK, "application/octet-stream", data)
	return nil
}

func (c *RealGinContext) JSON(code int, obj interface{}) error {
	c.ginCtx.JSON(code, obj)
	return nil
}

func (c *RealGinContext) String(code int, format string, values ...interface{}) error {
	c.ginCtx.String(code, format, values...)
	return nil
}

func (c *RealGinContext) Native() interface{} {
	return c.ginCtx
}

// RealGinAdapter implements Adapter for real Gin framework
type RealGinAdapter struct{}

func NewRealGinAdapter() *RealGinAdapter {
	return &RealGinAdapter{}
}

func (a *RealGinAdapter) WrapContext(native interface{}) HTTPContext {
	if ginCtx, ok := native.(*gin.Context); ok {
		return NewRealGinContext(ginCtx)
	}
	panic("invalid context type for RealGinAdapter")
}

func (a *RealGinAdapter) WrapHandler(handler HandlerFunc) interface{} {
	return gin.HandlerFunc(func(ginCtx *gin.Context) {
		ctx := NewRealGinContext(ginCtx)
		if err := handler(ctx); err != nil {
			ginCtx.JSON(500, gin.H{"error": err.Error()})
		}
	})
}

func (a *RealGinAdapter) WrapMiddleware(middleware Middleware) interface{} {
	return gin.HandlerFunc(func(ginCtx *gin.Context) {
		ctx := NewRealGinContext(ginCtx)

		nextHandler := func(ctx HTTPContext) error {
			ginCtx.Next()
			return nil
		}

		wrappedHandler := middleware(nextHandler)
		if err := wrappedHandler(ctx); err != nil {
			ginCtx.JSON(500, gin.H{"error": err.Error()})
			ginCtx.Abort()
		}
	})
}

func (a *RealGinAdapter) Name() string {
	return "gin-real"
}

// RealGinRouter implements Router for real Gin framework
type RealGinRouter struct {
	ginEngine *gin.Engine
	adapter   *RealGinAdapter
	group     gin.IRouter
}

func NewRealGinRouter(ginEngine *gin.Engine) *RealGinRouter {
	return &RealGinRouter{
		ginEngine: ginEngine,
		adapter:   NewRealGinAdapter(),
		group:     ginEngine,
	}
}

func (r *RealGinRouter) GET(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("GET", path, handler, middleware...)
}

func (r *RealGinRouter) POST(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("POST", path, handler, middleware...)
}

func (r *RealGinRouter) PUT(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("PUT", path, handler, middleware...)
}

func (r *RealGinRouter) DELETE(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("DELETE", path, handler, middleware...)
}

func (r *RealGinRouter) PATCH(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("PATCH", path, handler, middleware...)
}

func (r *RealGinRouter) addRoute(method, path string, handler HandlerFunc, middleware ...Middleware) {
	// Convert middleware to gin middleware
	ginMiddleware := make([]gin.HandlerFunc, len(middleware))
	for i, mw := range middleware {
		ginMiddleware[i] = r.adapter.WrapMiddleware(mw).(gin.HandlerFunc)
	}

	// Convert handler to gin handler
	ginHandler := r.adapter.WrapHandler(handler).(gin.HandlerFunc)

	// Combine middleware and handler
	handlers := append(ginMiddleware, ginHandler)

	// Register route with Gin
	r.group.Handle(method, path, handlers...)
}

func (r *RealGinRouter) Group(prefix string, middleware ...Middleware) Router {
	// Convert middleware to gin middleware
	ginMiddleware := make([]gin.HandlerFunc, len(middleware))
	for i, mw := range middleware {
		ginMiddleware[i] = r.adapter.WrapMiddleware(mw).(gin.HandlerFunc)
	}

	// Create new group
	newGroup := r.group.Group(prefix, ginMiddleware...)

	return &RealGinRouter{
		ginEngine: r.ginEngine,
		adapter:   r.adapter,
		group:     newGroup,
	}
}

func (r *RealGinRouter) Use(middleware ...Middleware) {
	ginMiddleware := make([]gin.HandlerFunc, len(middleware))
	for i, mw := range middleware {
		ginMiddleware[i] = r.adapter.WrapMiddleware(mw).(gin.HandlerFunc)
	}

	r.group.Use(ginMiddleware...)
}

func (r *RealGinRouter) Native() interface{} {
	return r.ginEngine
}

// RealGinServer implements Server for real Gin framework
type RealGinServer struct {
	router    *RealGinRouter
	ginEngine *gin.Engine
	server    *http.Server
}

func NewRealGinServer(ginEngine *gin.Engine) *RealGinServer {
	return &RealGinServer{
		router:    NewRealGinRouter(ginEngine),
		ginEngine: ginEngine,
	}
}

func (s *RealGinServer) Router() Router {
	return s.router
}

func (s *RealGinServer) Start(addr string) error {
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.ginEngine,
	}
	return s.server.ListenAndServe()
}

func (s *RealGinServer) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *RealGinServer) Native() interface{} {
	return s.ginEngine
}

// Helper function to create a real Gin server
func CreateRealGinServer(options ...ServerOption) (*RealGinServer, error) {
	config := &ServerConfig{}
	for _, opt := range options {
		opt(config)
	}

	var ginEngine *gin.Engine

	if config.NativeEngine != nil {
		var ok bool
		ginEngine, ok = config.NativeEngine.(*gin.Engine)
		if !ok {
			return nil, fmt.Errorf("provided engine is not a *gin.Engine")
		}
	} else {
		// Create new Gin engine
		if config.Debug {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}

		ginEngine = gin.New()
		ginEngine.Use(gin.Logger(), gin.Recovery())

		// Set trusted proxies if provided
		if len(config.TrustedProxies) > 0 {
			ginEngine.SetTrustedProxies(config.TrustedProxies)
		}
	}

	return NewRealGinServer(ginEngine), nil
}
