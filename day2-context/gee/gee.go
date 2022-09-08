package gee

import (
	"fmt"
	"net/http"
)

type HandlerFunc func(ctx *Context)

type Engine struct {
	router *Router
}

func New() *Engine {
	return &Engine{
		router: NewRouter(),
	}
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
	c := &Context{
		Writer: w,
		Req:    req,
		Method: req.Method,
		Path:   req.URL.Path,
	}

	key := c.Method + "-" + c.Path
	if handler, ok := e.router.handlers[key]; ok {
		handler(c)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}
