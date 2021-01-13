package handlers

import (
	"fmt"
	"github.com/implicithash/simple_gateway/src/controllers"
	"github.com/implicithash/simple_gateway/src/services"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

// Context is http context
type Context struct {
	http.ResponseWriter
	*http.Request
	Params []string
}

// Handler is a function for route handling
type Handler func(ctx *Context)

// Route is a route contains a handler and patterns
type Route struct {
	Pattern *regexp.Regexp
	Handler Handler
}

// App is an application with a number of routes
type App struct {
	Routes       []Route
	DefaultRoute Handler
}

const (
	maxTaskQueueSize = "max_queue_size"
	incomingReqQty = "incoming_req_qty"
	outgoingReqQty = "outgoing_req_qty"
)

var (
	maxQueueSize = os.Getenv(maxTaskQueueSize)
	incomingRequestQty = os.Getenv(incomingReqQty)
	outgoingRequestQty = os.Getenv(outgoingReqQty)
)

// NewApp constructs a new app
func NewApp() *App {
	app := &App{
		DefaultRoute: func(ctx *Context) {
			ctx.Text(http.StatusNotFound, "Not found")
		},
	}
	return app
}

// Text responds with some plain text
func (ctx *Context) Text(code int, body string) {
	ctx.ResponseWriter.Header().Set("Content-Type", "text/plain")
	ctx.WriteHeader(code)

	io.WriteString(ctx.ResponseWriter, fmt.Sprintf("%s\n", body))
}

// MapUrls is a mapping of existing routes
func MapUrls() http.Handler {
	app := NewApp()
	app.Handle("/", func(ctx *Context) {
		if ctx.Request.Method == http.MethodPost {
			controllers.PayloadController.DoRequest(ctx.ResponseWriter, ctx.Request)
		}
	})
	return app
}

// RunPool inits a worker pool
func RunPool() {
	maxSize, _ := strconv.Atoi(maxQueueSize)
	services.WorkerPool = services.NewWorker(maxSize)
	services.WorkerPool.Run()

	incomingReqQty, _ := strconv.Atoi(incomingRequestQty)
	outgoingReqQty, _ := strconv.Atoi(outgoingRequestQty)
	services.Limiter = services.NewRateLimiter(incomingReqQty, outgoingReqQty)
	services.Limiter.Run()
}

// StopPool stops a worker pool
func StopPool() {
	services.Limiter.Stop()
	services.WorkerPool.Stop()
}

// Handle parses a parameter pattern and append a route
func (app *App) Handle(pattern string, handler Handler) {
	re := regexp.MustCompile(pattern)
	route := Route{Pattern: re, Handler: handler}
	app.Routes = append(app.Routes, route)
}

// ServeHTTP is a main handler to start the server
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{Request: r, ResponseWriter: w}

	for _, rt := range app.Routes {
		if matches := rt.Pattern.FindStringSubmatch(ctx.URL.Path); len(matches) > 0 {
			if len(matches) > 1 {
				ctx.Params = matches[1:]
			}
			rt.Handler(ctx)
			return
		}
	}
	app.DefaultRoute(ctx)
}
