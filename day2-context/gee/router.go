package gee

type Router struct {
	handlers map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{make(map[string]HandlerFunc)}
}

func (r *Router) addRoute(method string, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern
	r.handlers[key] = handler
}
