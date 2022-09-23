package gi

import (
	"fmt"
	"strings"
)

type Router struct {
	root     map[string]*Node
	handlers map[string]HTTPHandler
}

func NewRouter() *Router {
	return &Router{
		root:     make(map[string]*Node),
		handlers: make(map[string]HTTPHandler),
	}
}

func (r *Router) Add(method string, pattern string, handler HTTPHandler) {
	key := fmt.Sprintf("%v:%v", method, pattern)
	node := r.root[method]

	if node == nil {
		node = &Node{
			part: method,
		}
		r.root[method] = node
	}

	node.Insert(pattern)
	r.handlers[key] = handler
}

func (r *Router) handle(c *Context) {
	if len(c.handlers) == 0 {
		c.Err("not found")
	}
	c.Next()
}

func disposeParams(node *Node, c *Context) {
	paths := parseParts(c.Path)
	parts := parseParts(node.pattern)
	params := make(map[string]string)

	for i, part := range parts {
		if len(part) <= 1 {
			continue
		}

		if part[0] == ':' {
			params[part[1:]] = paths[i]
		} else if part[0] == '*' {
			params[part[1:]] = strings.Join(paths[i:], "/")
			break
		}
	}

	c.Params = params
}
