package gi

import (
	"net/http"
)

type HTTPHandler = func(ctx *Context)

type Gi struct {
	*RouterGroup
	routerGroups map[string]*RouterGroup
	Router       *Router
}

func New() *Gi {
	g := &Gi{
		Router:       NewRouter(),
		routerGroups: make(map[string]*RouterGroup),
	}
	g.RouterGroup = &RouterGroup{
		Prefix: "",
		engin:  g,
	}
	return g
}

func (g *Gi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(w, r)

	node := g.Router.root[c.Method].Search(c.Path)

	if node != nil {
		if node.isWild {
			disposeParams(node, c)
		}

		//fmt.Printf("[%v] - [%v]\n", handleKey, ctx.Path)
		if group := g.routerGroups[node.pattern]; group != nil {
			c.handlers = group.middlewares[:]
		}

		if handler := g.Router.handlers[c.Method+":"+node.pattern]; handler != nil {
			c.handlers = append(c.handlers, handler)
		}
	}

	g.Router.handle(c)
}

func (g *Gi) Run(addr string) error {
	return http.ListenAndServe(addr, g)
}
