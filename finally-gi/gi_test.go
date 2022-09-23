package main_test

import (
	"finally-gi/gi"
	"finally-gi/gi/middlewares"
	"io"
	"net/http"
	"strings"
	"testing"
)

type httpRequest = func(url string) (resp *http.Response, err error)

func runServer(t testing.TB, baseUrl string) {
	addr := ":8080"
	portIdx := strings.LastIndex(baseUrl, ":")

	if portIdx != -1 {
		addr = baseUrl[portIdx:]
	}

	g := gi.New()
	g.Use(middlewares.Recover())

	group1 := g.Group("/group1")
	group2 := g.Group("/group2")

	g.Get("/panic", func(ctx *gi.Context) {
		panic("hehe")
	})

	g.Get("/param/:param", func(ctx *gi.Context) {
		param := ctx.Param("param")
		ctx.String(param)
	})

	group1.Get("/1", func(ctx *gi.Context) {
		ctx.String("group1/1")
	})
	group11 := group1.Group("/1")
	group11.Get("/1", func(ctx *gi.Context) {
		ctx.String("group1/1/1")
	})

	group1.Get("/2", func(ctx *gi.Context) {
		ctx.String("group1/2")
	})
	group12 := group1.Group("/2")
	group12.Get("/2", func(ctx *gi.Context) {
		ctx.String("group1/2/2")
	})

	group2.Get("/1", func(ctx *gi.Context) {
		ctx.String("group2/1")
	})

	group2.Get("/2", func(ctx *gi.Context) {
		ctx.String("group2/2")
	})
	t.Fatal(g.Run(addr))
}

func request(t testing.TB, reqFunc httpRequest, url string, result string) {
	t.Helper()
	res, err := reqFunc(url)
	if err != nil {
		t.Fatalf("[%30s] err \n %v\n", url, err.Error())
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !strings.EqualFold(string(data), result) {

		t.Fatalf("test1 err\n expected:  \"%v\" \n get wrong data: \"%v\" \n", result, string(data))
	}
}

func TestTire(t *testing.T) {
	baseUrl := "http://localhost:8081"

	getReq := map[string]string{
		"/group1/1":   "group1/1",
		"/group1/1/1": "group1/1/1",
		"/group1/2":   "group1/2",
		"/group1/2/2": "group1/2/2",
		"/group2/1":   "group2/1",
		"/group2/2":   "group2/2",
	}

	go runServer(t, baseUrl)

	for path, result := range getReq {
		url := baseUrl + path
		request(t, http.Get, url, result)
	}
}

func Benchmark(t *testing.B) {
	baseUrl := "http://localhost:8081"

	getReq := map[string]string{
		//"/param/1111": "1111",
		"/group1/1": "group1/1",
		//"/group1/1/1": "group1/1/1",
		//"/group1/2":   "group1/2",
		//"/group1/2/2": "group1/2/2",
		//"/group2/1":   "group2/1",
		//"/group2/2":   "group2/2",
	}

	for n := 0; n < t.N; n++ {
		for path, result := range getReq {
			url := baseUrl + path
			request(t, http.Get, url, result)
		}
	}
}
