package httpserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// EchoContext implements HTTPContext for Echo framework
type EchoContext struct {
	echoCtx interface{} // echo.Context - using interface{} to avoid import dependency
}

func NewEchoContext(echoCtx interface{}) *EchoContext {
	return &EchoContext{echoCtx: echoCtx}
}

// Request methods
func (c *EchoContext) Method() string {
	// In real implementation: return c.echoCtx.(echo.Context).Request().Method
	if ctx, ok := c.echoCtx.(interface{ Request() *http.Request }); ok {
		return ctx.Request().Method
	}
	return "GET" // fallback
}

func (c *EchoContext) Path() string {
	// In real implementation: return c.echoCtx.(echo.Context).Request().URL.Path
	if ctx, ok := c.echoCtx.(interface{ Request() *http.Request }); ok {
		return ctx.Request().URL.Path
	}
	return "/" // fallback
}

func (c *EchoContext) Query(key string) string {
	// In real implementation: return c.echoCtx.(echo.Context).QueryParam(key)
	if ctx, ok := c.echoCtx.(interface{ QueryParam(string) string }); ok {
		return ctx.QueryParam(key)
	}
	return "" // fallback
}

func (c *EchoContext) Header(key string) string {
	// In real implementation: return c.echoCtx.(echo.Context).Request().Header.Get(key)
	if ctx, ok := c.echoCtx.(interface{ Request() *http.Request }); ok {
		return ctx.Request().Header.Get(key)
	}
	return "" // fallback
}

func (c *EchoContext) Body() ([]byte, error) {
	// In real implementation: return io.ReadAll(c.echoCtx.(echo.Context).Request().Body)
	if ctx, ok := c.echoCtx.(interface{ Request() *http.Request }); ok {
		return io.ReadAll(ctx.Request().Body)
	}
	return nil, fmt.Errorf("unable to read body")
}

func (c *EchoContext) Context() context.Context {
	// In real implementation: return c.echoCtx.(echo.Context).Request().Context()
	if ctx, ok := c.echoCtx.(interface{ Request() *http.Request }); ok {
		return ctx.Request().Context()
	}
	return context.Background()
}

func (c *EchoContext) WithContext(ctx context.Context) {
	// In real implementation: c.echoCtx.(echo.Context).SetRequest(c.echoCtx.(echo.Context).Request().WithContext(ctx))
	if echoCtx, ok := c.echoCtx.(interface {
		Request() *http.Request
		SetRequest(*http.Request)
	}); ok {
		echoCtx.SetRequest(echoCtx.Request().WithContext(ctx))
	}
}

// Response methods
func (c *EchoContext) Status(code int) {
	// In real implementation: c.echoCtx.(echo.Context).Response().WriteHeader(code)
	if ctx, ok := c.echoCtx.(interface{ Response() interface{ WriteHeader(int) } }); ok {
		ctx.Response().WriteHeader(code)
	}
}

func (c *EchoContext) SetHeader(key, value string) {
	// In real implementation: c.echoCtx.(echo.Context).Response().Header().Set(key, value)
	if ctx, ok := c.echoCtx.(interface{ Response() interface{ Header() http.Header } }); ok {
		ctx.Response().Header().Set(key, value)
	}
}

func (c *EchoContext) Write(data []byte) error {
	// In real implementation: return c.echoCtx.(echo.Context).Blob(http.StatusOK, "application/octet-stream", data)
	if ctx, ok := c.echoCtx.(interface{ Blob(int, string, []byte) error }); ok {
		return ctx.Blob(http.StatusOK, "application/octet-stream", data)
	}
	return fmt.Errorf("unable to write data")
}

func (c *EchoContext) JSON(code int, obj interface{}) error {
	// In real implementation: return c.echoCtx.(echo.Context).JSON(code, obj)
	if ctx, ok := c.echoCtx.(interface{ JSON(int, interface{}) error }); ok {
		return ctx.JSON(code, obj)
	}
	return fmt.Errorf("unable to write JSON")
}

func (c *EchoContext) String(code int, format string, values ...interface{}) error {
	// In real implementation: return c.echoCtx.(echo.Context).String(code, fmt.Sprintf(format, values...))
	if ctx, ok := c.echoCtx.(interface{ String(int, string) error }); ok {
		return ctx.String(code, fmt.Sprintf(format, values...))
	}
	return fmt.Errorf("unable to write string")
}

func (c *EchoContext) Native() interface{} {
	return c.echoCtx
}

// EchoAdapter implements Adapter for Echo framework
type EchoAdapter struct{}

func NewEchoAdapter() *EchoAdapter {
	return &EchoAdapter{}
}

func (a *EchoAdapter) WrapContext(native interface{}) HTTPContext {
	return NewEchoContext(native)
}

func (a *EchoAdapter) WrapHandler(handler HandlerFunc) interface{} {
	// Returns echo.HandlerFunc
	return func(echoCtx interface{}) error {
		ctx := NewEchoContext(echoCtx)
		return handler(ctx)
	}
}

func (a *EchoAdapter) WrapMiddleware(middleware Middleware) interface{} {
	// Returns echo.MiddlewareFunc
	return func(next interface{}) interface{} {
		return func(echoCtx interface{}) error {
			ctx := NewEchoContext(echoCtx)
			
			nextHandler := func(ctx HTTPContext) error {
				// In real implementation: return next.(echo.HandlerFunc)(echoCtx.(echo.Context))
				if nextFunc, ok := next.(func(interface{}) error); ok {
					return nextFunc(echoCtx)
				}
				return nil
			}
			
			wrappedHandler := middleware(nextHandler)
			return wrappedHandler(ctx)
		}
	}
}

func (a *EchoAdapter) Name() string {
	return "echo"
}

// EchoRouter implements Router for Echo framework
type EchoRouter struct {
	echoEngine interface{} // echo.Echo - using interface{} to avoid import dependency
	adapter    *EchoAdapter
	group      interface{} // echo.Group
}

func NewEchoRouter(echoEngine interface{}) *EchoRouter {
	return &EchoRouter{
		echoEngine: echoEngine,
		adapter:    NewEchoAdapter(),
		group:      echoEngine, // Initially, group is the engine itself
	}
}

func (r *EchoRouter) GET(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("GET", path, handler, middleware...)
}

func (r *EchoRouter) POST(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("POST", path, handler, middleware...)
}

func (r *EchoRouter) PUT(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("PUT", path, handler, middleware...)
}

func (r *EchoRouter) DELETE(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("DELETE", path, handler, middleware...)
}

func (r *EchoRouter) PATCH(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("PATCH", path, handler, middleware...)
}

func (r *EchoRouter) addRoute(method, path string, handler HandlerFunc, middleware ...Middleware) {
	// Convert middleware to echo middleware
	echoMiddleware := make([]interface{}, len(middleware))
	for i, mw := range middleware {
		echoMiddleware[i] = r.adapter.WrapMiddleware(mw)
	}
	
	// Convert handler to echo handler
	echoHandler := r.adapter.WrapHandler(handler)
	
	// In real implementation, we would call the appropriate method on the echo router
	switch method {
	case "GET":
		if router, ok := r.group.(interface {
			GET(string, interface{}, ...interface{}) interface{}
		}); ok {
			router.GET(path, echoHandler, echoMiddleware...)
		}
	case "POST":
		if router, ok := r.group.(interface {
			POST(string, interface{}, ...interface{}) interface{}
		}); ok {
			router.POST(path, echoHandler, echoMiddleware...)
		}
	case "PUT":
		if router, ok := r.group.(interface {
			PUT(string, interface{}, ...interface{}) interface{}
		}); ok {
			router.PUT(path, echoHandler, echoMiddleware...)
		}
	case "DELETE":
		if router, ok := r.group.(interface {
			DELETE(string, interface{}, ...interface{}) interface{}
		}); ok {
			router.DELETE(path, echoHandler, echoMiddleware...)
		}
	case "PATCH":
		if router, ok := r.group.(interface {
			PATCH(string, interface{}, ...interface{}) interface{}
		}); ok {
			router.PATCH(path, echoHandler, echoMiddleware...)
		}
	}
}

func (r *EchoRouter) Group(prefix string, middleware ...Middleware) Router {
	// Convert middleware to echo middleware
	echoMiddleware := make([]interface{}, len(middleware))
	for i, mw := range middleware {
		echoMiddleware[i] = r.adapter.WrapMiddleware(mw)
	}
	
	// In real implementation: newGroup := r.group.(echo.Group).Group(prefix, echoMiddleware...)
	var newGroup interface{}
	if router, ok := r.group.(interface {
		Group(string, ...interface{}) interface{}
	}); ok {
		newGroup = router.Group(prefix, echoMiddleware...)
	} else {
		newGroup = r.group // fallback
	}
	
	return &EchoRouter{
		echoEngine: r.echoEngine,
		adapter:    r.adapter,
		group:      newGroup,
	}
}

func (r *EchoRouter) Use(middleware ...Middleware) {
	echoMiddleware := make([]interface{}, len(middleware))
	for i, mw := range middleware {
		echoMiddleware[i] = r.adapter.WrapMiddleware(mw)
	}
	
	// In real implementation: r.group.(echo.Group).Use(echoMiddleware...)
	if router, ok := r.group.(interface {
		Use(...interface{})
	}); ok {
		router.Use(echoMiddleware...)
	}
}

func (r *EchoRouter) Native() interface{} {
	return r.echoEngine
}

// EchoServer implements Server for Echo framework
type EchoServer struct {
	router     *EchoRouter
	echoEngine interface{} // echo.Echo
}

func NewEchoServer(echoEngine interface{}) *EchoServer {
	return &EchoServer{
		router:     NewEchoRouter(echoEngine),
		echoEngine: echoEngine,
	}
}

func (s *EchoServer) Router() Router {
	return s.router
}

func (s *EchoServer) Start(addr string) error {
	// In real implementation: return s.echoEngine.(*echo.Echo).Start(addr)
	if engine, ok := s.echoEngine.(interface{ Start(string) error }); ok {
		return engine.Start(addr)
	}
	return fmt.Errorf("unable to start echo server")
}

func (s *EchoServer) Shutdown(ctx context.Context) error {
	// In real implementation: return s.echoEngine.(*echo.Echo).Shutdown(ctx)
	if engine, ok := s.echoEngine.(interface{ Shutdown(context.Context) error }); ok {
		return engine.Shutdown(ctx)
	}
	return nil
}

func (s *EchoServer) Native() interface{} {
	return s.echoEngine
}
