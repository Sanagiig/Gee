package main

import (
	"example3/gee"
	"log"
	"net/http"
)

func main() {
	g := gee.New()
	g.GET("/hello", func(ctx *gee.Context) {
		ctx.Data(http.StatusOK, []byte("hello"))
	})

	g.GET("/start/*test/test", func(ctx *gee.Context) {
		ctx.JSON(http.StatusOK, ctx.Params)
	})

	println("run in 8081")
	log.Fatal(g.Run(":8081"))
}
