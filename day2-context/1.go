package main

import (
	"example2/gee"
	"fmt"
	"log"
	"net/http"
)

func main() {
	g := gee.New()
	g.GET("/", func(ctx *gee.Context) {
		fmt.Fprintf(ctx.Writer, "URL.Path = %q\n", ctx.Path)
	})

	g.GET("/query", func(ctx *gee.Context) {
		fmt.Fprintf(ctx.Writer, "a = %v , b = %v", ctx.Query("a"), ctx.Query("b"))
	})

	g.GET("/json", func(ctx *gee.Context) {
		obj := make(map[string]string)
		obj["x"] = "123"
		obj["y"] = "321"
		ctx.JSON(http.StatusOK, obj)
	})

	g.GET("/data", func(ctx *gee.Context) {
		ctx.Data(http.StatusOK, []byte("data ok"))
	})

	g.GET("/html", func(ctx *gee.Context) {
		ctx.HTML(http.StatusOK, ("<!DOCTYPE html><html><body><h1>hello</h1></body></html>"))
	})

	g.GET("/string", func(ctx *gee.Context) {
		ctx.String(http.StatusOK, ("<!DOCTYPE html><html><body><h1>hello</h1></body></html>"))
	})

	g.GET("/header", func(ctx *gee.Context) {
		for k, v := range ctx.Req.Header {
			fmt.Fprintf(ctx.Writer, "Header[%q] = %q\n", k, v)
		}
	})

	log.Fatal(g.Run(":8081"))
}
