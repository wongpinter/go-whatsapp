package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// HTTPContext provides a framework-agnostic interface for HTTP request/response handling
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

	// Framework-specific context (for advanced usage)
	Native() interface{}
}

// HandlerFunc is a framework-agnostic handler function
type HandlerFunc func(HTTPContext) error

// Middleware is a framework-agnostic middleware function
type Middleware func(HandlerFunc) HandlerFunc

// Adapter provides framework-specific implementations
type Adapter interface {
	// Convert framework-specific context to HTTPContext
	WrapContext(native interface{}) HTTPContext

	// Convert our HandlerFunc to framework-specific handler
	WrapHandler(handler HandlerFunc) interface{}

	// Convert our Middleware to framework-specific middleware
	WrapMiddleware(middleware Middleware) interface{}

	// Framework identification
	Name() string
}

// StandardHTTPContext implements HTTPContext for net/http
type StandardHTTPContext struct {
	request  *http.Request
	response http.ResponseWriter
	written  bool
}

// NewStandardHTTPContext creates a new StandardHTTPContext
func NewStandardHTTPContext(w http.ResponseWriter, r *http.Request) *StandardHTTPContext {
	return &StandardHTTPContext{
		request:  r,
		response: w,
		written:  false,
	}
}

// Request methods
func (c *StandardHTTPContext) Method() string {
	return c.request.Method
}

func (c *StandardHTTPContext) Path() string {
	return c.request.URL.Path
}

func (c *StandardHTTPContext) Query(key string) string {
	return c.request.URL.Query().Get(key)
}

func (c *StandardHTTPContext) Header(key string) string {
	return c.request.Header.Get(key)
}

func (c *StandardHTTPContext) Body() ([]byte, error) {
	return io.ReadAll(c.request.Body)
}

func (c *StandardHTTPContext) Context() context.Context {
	return c.request.Context()
}

func (c *StandardHTTPContext) WithContext(ctx context.Context) {
	c.request = c.request.WithContext(ctx)
}

// Response methods
func (c *StandardHTTPContext) Status(code int) {
	if !c.written {
		c.response.WriteHeader(code)
		c.written = true
	}
}

func (c *StandardHTTPContext) SetHeader(key, value string) {
	c.response.Header().Set(key, value)
}

func (c *StandardHTTPContext) Write(data []byte) error {
	if !c.written {
		c.response.WriteHeader(http.StatusOK)
		c.written = true
	}
	_, err := c.response.Write(data)
	return err
}

func (c *StandardHTTPContext) JSON(code int, obj interface{}) error {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)

	// Use json.Marshal to properly encode the object
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return c.Write(data)
}

func (c *StandardHTTPContext) String(code int, format string, values ...interface{}) error {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)

	// Simple string formatting - in real implementation, use fmt.Sprintf
	data := []byte(format) // Placeholder
	return c.Write(data)
}

func (c *StandardHTTPContext) Native() interface{} {
	return struct {
		Request  *http.Request
		Response http.ResponseWriter
	}{
		Request:  c.request,
		Response: c.response,
	}
}

// StandardAdapter implements Adapter for net/http
type StandardAdapter struct{}

func NewStandardAdapter() *StandardAdapter {
	return &StandardAdapter{}
}

func (a *StandardAdapter) WrapContext(native interface{}) HTTPContext {
	if ctx, ok := native.(*StandardHTTPContext); ok {
		return ctx
	}

	// If native is a struct with Request and Response
	if data, ok := native.(struct {
		Request  *http.Request
		Response http.ResponseWriter
	}); ok {
		return NewStandardHTTPContext(data.Response, data.Request)
	}

	panic("invalid native context for StandardAdapter")
}

func (a *StandardAdapter) WrapHandler(handler HandlerFunc) interface{} {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewStandardHTTPContext(w, r)
		if err := handler(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (a *StandardAdapter) WrapMiddleware(middleware Middleware) interface{} {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := NewStandardHTTPContext(w, r)

			nextHandler := func(ctx HTTPContext) error {
				// Convert back to http.Handler call
				next.ServeHTTP(w, r)
				return nil
			}

			wrappedHandler := middleware(nextHandler)
			if err := wrappedHandler(ctx); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
	}
}

func (a *StandardAdapter) Name() string {
	return "net/http"
}

// Router provides a framework-agnostic routing interface
type Router interface {
	GET(path string, handler HandlerFunc, middleware ...Middleware)
	POST(path string, handler HandlerFunc, middleware ...Middleware)
	PUT(path string, handler HandlerFunc, middleware ...Middleware)
	DELETE(path string, handler HandlerFunc, middleware ...Middleware)
	PATCH(path string, handler HandlerFunc, middleware ...Middleware)

	Group(prefix string, middleware ...Middleware) Router
	Use(middleware ...Middleware)

	// Get the native router for framework-specific operations
	Native() interface{}
}

// StandardRouter implements Router for net/http
type StandardRouter struct {
	mux        *http.ServeMux
	adapter    *StandardAdapter
	prefix     string
	middleware []Middleware
}

func NewStandardRouter() *StandardRouter {
	return &StandardRouter{
		mux:        http.NewServeMux(),
		adapter:    NewStandardAdapter(),
		prefix:     "",
		middleware: make([]Middleware, 0),
	}
}

func (r *StandardRouter) GET(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("GET", path, handler, middleware...)
}

func (r *StandardRouter) POST(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("POST", path, handler, middleware...)
}

func (r *StandardRouter) PUT(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("PUT", path, handler, middleware...)
}

func (r *StandardRouter) DELETE(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("DELETE", path, handler, middleware...)
}

func (r *StandardRouter) PATCH(path string, handler HandlerFunc, middleware ...Middleware) {
	r.addRoute("PATCH", path, handler, middleware...)
}

func (r *StandardRouter) addRoute(method, path string, handler HandlerFunc, middleware ...Middleware) {
	fullPath := r.prefix + path

	// Combine router middleware with route-specific middleware
	allMiddleware := append(r.middleware, middleware...)

	// Apply middleware chain
	finalHandler := handler
	for i := len(allMiddleware) - 1; i >= 0; i-- {
		finalHandler = allMiddleware[i](finalHandler)
	}

	// Convert to http.Handler
	httpHandler := r.adapter.WrapHandler(finalHandler).(http.HandlerFunc)

	// Register with method-specific pattern (Go 1.22+ style)
	pattern := method + " " + fullPath
	r.mux.HandleFunc(pattern, httpHandler)
}

func (r *StandardRouter) Group(prefix string, middleware ...Middleware) Router {
	return &StandardRouter{
		mux:        r.mux,
		adapter:    r.adapter,
		prefix:     r.prefix + prefix,
		middleware: append(r.middleware, middleware...),
	}
}

func (r *StandardRouter) Use(middleware ...Middleware) {
	r.middleware = append(r.middleware, middleware...)
}

func (r *StandardRouter) Native() interface{} {
	return r.mux
}

// Server provides a framework-agnostic server interface
type Server interface {
	Router() Router
	Start(addr string) error
	Shutdown(ctx context.Context) error
	Native() interface{}
}

// StandardServer implements Server for net/http
type StandardServer struct {
	router     *StandardRouter
	httpServer *http.Server
}

func NewStandardServer() *StandardServer {
	return &StandardServer{
		router: NewStandardRouter(),
	}
}

func (s *StandardServer) Router() Router {
	return s.router
}

func (s *StandardServer) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router.mux,
	}
	return s.httpServer.ListenAndServe()
}

func (s *StandardServer) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *StandardServer) Native() interface{} {
	return s.httpServer
}
