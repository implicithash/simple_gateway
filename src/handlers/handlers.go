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

type Context struct {
	http.ResponseWriter
	*http.Request
	Params []string
}

type Handler func(ctx *Context)

type Route struct {
	Pattern *regexp.Regexp
	Handler Handler
}

type App struct {
	Routes       []Route
	DefaultRoute Handler
}

const (
	maxTaskQueueSize = "max_queue_size"
)

var (
	maxQueueSize = os.Getenv(maxTaskQueueSize)
)

func NewApp() *App {
	app := &App{
		DefaultRoute: func(ctx *Context) {
			ctx.Text(http.StatusNotFound, "Not found")
		},
	}
	return app
}

func (ctx *Context) Text(code int, body string) {
	ctx.ResponseWriter.Header().Set("Content-Type", "text/plain")
	ctx.WriteHeader(code)

	io.WriteString(ctx.ResponseWriter, fmt.Sprintf("%s\n", body))
}

func MapUrls() http.Handler {
	app := NewApp()
	app.Handle("/", func(ctx *Context) {
		if ctx.Request.Method == http.MethodPost {
			controllers.PayloadController.DoRequest(ctx.ResponseWriter, ctx.Request)
		}
	})
	return app
}

func InitPool() {
	maxSize, _ := strconv.Atoi(maxQueueSize)
	services.WorkerPool = services.NewWorker(maxSize)
	services.WorkerPool.Run()
}

func (app *App) Handle(pattern string, handler Handler) {
	re := regexp.MustCompile(pattern)
	route := Route{Pattern: re, Handler: handler}
	app.Routes = append(app.Routes, route)
}

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
