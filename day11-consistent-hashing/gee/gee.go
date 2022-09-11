package gee

import (
	"net/http"
	"strings"
)

type HandlerFunc func(ctx *Context)

type Engine struct {
	*RouterGroup
	router *Router
	groups []*RouterGroup
}

func New() *Engine {
	engine := &Engine{router: NewRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// GET defines the method to add GET request
func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.router.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.router.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc

	for _, group := range e.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	c := newContext(w, req)
	c.handlers = middlewares
	e.router.handle(c)
}
