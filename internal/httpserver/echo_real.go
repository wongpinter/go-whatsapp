//go:build echo
// +build echo

package httpserver

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

func init() {
	// Register the real Echo server creation function
	createRealEchoServerFunc = func(options ...ServerOption) (Server, error) {
		return CreateRealEchoServer(options...)
	}
}

// RealEchoContext implements HTTPContext for real Echo framework
type RealEchoContext struct {
	echoCtx echo.Context
}

func NewRealEchoContext(echoCtx echo.Context) *RealEchoContext {
	return &RealEchoContext{echoCtx: echoCtx}
}

// Request methods
func (c *RealEchoContext) Method() string {
	return c.echoCtx.Request().Method
}

func (c *RealEchoContext) Path() string {
	return c.echoCtx.Request().URL.Path
}

func (c *RealEchoContext) Query(key string) string {
	return c.echoCtx.QueryParam(key)
}

func (c *RealEchoContext) Header(key string) string {
	return c.echoCtx.Request().Header.Get(key)
}

func (c *RealEchoContext) Body() ([]byte, error) {
	return io.ReadAll(c.echoCtx.Request().Body)
}

func (c *RealEchoContext) Context() context.Context {
	return c.echoCtx.Request().Context()
}

func (c *RealEchoContext) WithContext(ctx context.Context) {
	c.echoCtx.SetRequest(c.echoCtx.Request().WithContext(ctx))
}

// Response methods
func (c *RealEchoContext) Status(code int) {
	c.echoCtx.Response().WriteHeader(code)
}

func (c *RealEchoContext) SetHeader(key, value string) {
	c.echoCtx.Response().Header().Set(key, value)
}

func (c *RealEchoContext) Write(data []byte) error {
	return c.echoCtx.Blob(http.StatusOK, "application/octet-stream", data)
}

func (c *RealEchoContext) JSON(code int, obj interface{}) error {
	return c.echoCtx.JSON(code, obj)
}

func (c *RealEchoContext) String(code int, format string, values ...interface{}) error {
	return c.echoCtx.String(code, fmt.Sprintf(format, values...))
}

func (c *RealEchoContext) Native() interface{} {
	return c.echoCtx
}

// RealEchoAdapter implements Adapter for real Echo framework
type RealEchoAdapter struct{}

func NewRealEchoAdapter() *RealEchoAdapter {
	return &RealEchoAdapter{}
}

func (a *RealEchoAdapter) WrapContext(native interface{}) HTTPContext {
	if echoCtx, ok := native.(echo.Context); ok {
		return NewRealEchoContext(echoCtx)
	}
	panic("invalid context type for RealEchoAdapter")
}

func (a *RealEchoAdapter) WrapHandler(handler HandlerFunc) interface{} {
	return echo.HandlerFunc(func(echoCtx echo.Context) error {
		ctx := NewRealEchoContext(echoCtx)
		return handler(ctx)
	})
}

func (a *RealEchoAdapter) WrapMiddleware(middleware Middleware) interface{} {
	return echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx echo.Context) error {
			ctx := NewRealEchoContext(echoCtx)

			nextHandler := func(ctx HTTPContext) error {
				return next(echoCtx)
			}

			wrappedHandler := middleware(nextHandler)
			return wrappedHandler(ctx)
		}
	})
}

func (a *RealEchoAdapter) Name() string {
	return "echo-real"
}

// RealEchoRouter implements Router for real Echo framework
type RealEchoRouter struct {
	echoEngine *echo.Echo
	adapter    *RealEchoAdapter
	group      *echo.Group
}

func NewRealEchoRouter(echoEngine *echo.Echo) *RealEchoRouter {
	return &RealEchoRouter{
		echoEngine: echoEngine,
		adapter:    NewRealEchoAdapter(),
		group:      nil, // nil means use the engine directly
	}
}

func (r *RealEchoRouter) GET(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("GET", path, handler, middleware...)
}

func (r *RealEchoRouter) POST(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("POST", path, handler, middleware...)
}

func (r *RealEchoRouter) PUT(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("PUT", path, handler, middleware...)
}

func (r *RealEchoRouter) DELETE(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("DELETE", path, handler, middleware...)
}

func (r *RealEchoRouter) PATCH(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("PATCH", path, handler, middleware...)
}

func (r *RealEchoRouter) addRoute(method, path string, handler HandlerFunc, middleware ...Middleware) {
	// Convert middleware to echo middleware
	echoMiddleware := make([]echo.MiddlewareFunc, len(middleware))
	for i, mw := range middleware {
		echoMiddleware[i] = r.adapter.WrapMiddleware(mw).(echo.MiddlewareFunc)
	}

	// Convert handler to echo handler
	echoHandler := r.adapter.WrapHandler(handler).(echo.HandlerFunc)

	// Register route with Echo
	if r.group != nil {
		// Use group
		switch method {
		case "GET":
			r.group.GET(path, echoHandler, echoMiddleware...)
		case "POST":
			r.group.POST(path, echoHandler, echoMiddleware...)
		case "PUT":
			r.group.PUT(path, echoHandler, echoMiddleware...)
		case "DELETE":
			r.group.DELETE(path, echoHandler, echoMiddleware...)
		case "PATCH":
			r.group.PATCH(path, echoHandler, echoMiddleware...)
		}
	} else {
		// Use engine directly
		switch method {
		case "GET":
			r.echoEngine.GET(path, echoHandler, echoMiddleware...)
		case "POST":
			r.echoEngine.POST(path, echoHandler, echoMiddleware...)
		case "PUT":
			r.echoEngine.PUT(path, echoHandler, echoMiddleware...)
		case "DELETE":
			r.echoEngine.DELETE(path, echoHandler, echoMiddleware...)
		case "PATCH":
			r.echoEngine.PATCH(path, echoHandler, echoMiddleware...)
		}
	}
}

func (r *RealEchoRouter) Group(prefix string, middleware ...Middleware) Router {
	// Convert middleware to echo middleware
	echoMiddleware := make([]echo.MiddlewareFunc, len(middleware))
	for i, mw := range middleware {
		echoMiddleware[i] = r.adapter.WrapMiddleware(mw).(echo.MiddlewareFunc)
	}

	// Create new group
	var newGroup *echo.Group
	if r.group != nil {
		newGroup = r.group.Group(prefix, echoMiddleware...)
	} else {
		newGroup = r.echoEngine.Group(prefix, echoMiddleware...)
	}

	return &RealEchoRouter{
		echoEngine: r.echoEngine,
		adapter:    r.adapter,
		group:      newGroup,
	}
}

func (r *RealEchoRouter) Use(middleware ...Middleware) {
	echoMiddleware := make([]echo.MiddlewareFunc, len(middleware))
	for i, mw := range middleware {
		echoMiddleware[i] = r.adapter.WrapMiddleware(mw).(echo.MiddlewareFunc)
	}

	if r.group != nil {
		r.group.Use(echoMiddleware...)
	} else {
		r.echoEngine.Use(echoMiddleware...)
	}
}

func (r *RealEchoRouter) Native() interface{} {
	return r.echoEngine
}

// RealEchoServer implements Server for real Echo framework
type RealEchoServer struct {
	router     *RealEchoRouter
	echoEngine *echo.Echo
}

func NewRealEchoServer(echoEngine *echo.Echo) *RealEchoServer {
	return &RealEchoServer{
		router:     NewRealEchoRouter(echoEngine),
		echoEngine: echoEngine,
	}
}

func (s *RealEchoServer) Router() Router {
	return s.router
}

func (s *RealEchoServer) Start(addr string) error {
	return s.echoEngine.Start(addr)
}

func (s *RealEchoServer) Shutdown(ctx context.Context) error {
	return s.echoEngine.Shutdown(ctx)
}

func (s *RealEchoServer) Native() interface{} {
	return s.echoEngine
}

// Helper function to create a real Echo server
func CreateRealEchoServer(options ...ServerOption) (*RealEchoServer, error) {
	config := &ServerConfig{}
	for _, opt := range options {
		opt(config)
	}

	var echoEngine *echo.Echo

	if config.NativeEngine != nil {
		var ok bool
		echoEngine, ok = config.NativeEngine.(*echo.Echo)
		if !ok {
			return nil, fmt.Errorf("provided engine is not a *echo.Echo")
		}
	} else {
		// Create new Echo instance
		echoEngine = echo.New()

		// Configure Echo
		echoEngine.Debug = config.Debug

		// Set trusted proxies if provided
		if len(config.TrustedProxies) > 0 {
			echoEngine.IPExtractor = echo.ExtractIPFromXFFHeader(
				echo.TrustLoopback(false),
				echo.TrustLinkLocal(false),
				echo.TrustPrivateNet(false),
			)
		}
	}

	return NewRealEchoServer(echoEngine), nil
}
