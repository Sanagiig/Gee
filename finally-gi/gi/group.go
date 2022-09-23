package gi

import (
	"net/http"
	"path"
)

type RouterGroup struct {
	Parent      *RouterGroup
	Prefix      string
	engin       *Gi
	middlewares []HTTPHandler
}

func (g *RouterGroup) Group(name string) *RouterGroup {
	group := &RouterGroup{
		Parent:      g,
		Prefix:      g.Prefix + name,
		engin:       g.engin,
		middlewares: g.middlewares[:],
	}

	return group
}

func (g *RouterGroup) Use(middlewares ...HTTPHandler) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (g *RouterGroup) AddRoute(method string, pattern string, handler HTTPHandler) {
	groupPattern := g.Prefix + pattern

	g.engin.routerGroups[groupPattern] = g
	g.engin.Router.Add(method, groupPattern, handler)
}

func (g *RouterGroup) Get(pattern string, handler HTTPHandler) {
	g.AddRoute("GET", pattern, handler)
}

func (g *RouterGroup) Post(pattern string, handler HTTPHandler) {
	g.AddRoute("POST", pattern, handler)
}

func (g *RouterGroup) createStaticHandler(pattern string, fs http.FileSystem) HTTPHandler {
	absPath := path.Join(g.Prefix, pattern)
	fileServer := http.StripPrefix(absPath, http.FileServer(fs))

	return func(ctx *Context) {
		filePath := ctx.Param("filePath")
		if _, err := fs.Open(filePath); err != nil {
			ctx.Fail(http.StatusNotFound, "file not foud")
			return
		}

		fileServer.ServeHTTP(ctx.RespWriter, ctx.Req)
	}
}

func (g *RouterGroup) Static(pattern string, sourceRoot string) {
	handler := g.createStaticHandler(pattern, http.Dir(sourceRoot))
	urlPattern := path.Join(pattern, "/*filePath")
	g.Get(urlPattern, handler)
}
