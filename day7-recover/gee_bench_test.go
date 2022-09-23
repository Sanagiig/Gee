package main_test

import (
	"example7/gee"
	"io"
	"net/http"
	"strings"
	"testing"
)

type httpRequest = func(url string) (resp *http.Response, err error)

func runServer(t *testing.B, baseUrl string) {
	addr := ":8080"
	portIdx := strings.LastIndex(baseUrl, ":")

	if portIdx != -1 {
		addr = baseUrl[portIdx:]
	}

	g := gee.New()
	g.Use(gee.Recovery())

	group1 := g.Group("/group1")
	group2 := g.Group("/group2")

	g.GET("/param/:param", func(ctx *gee.Context) {
		param := ctx.Param("param")
		ctx.String(http.StatusOK, param)
	})
	group1.GET("/1", func(ctx *gee.Context) {
		ctx.String(http.StatusOK, "group1/1")
	})
	group11 := group1.Group("/1")
	group11.GET("/1", func(ctx *gee.Context) {
		ctx.String(http.StatusOK, "group1/1/1")
	})

	group1.GET("/2", func(ctx *gee.Context) {
		ctx.String(http.StatusOK, "group1/2")
	})
	group12 := group1.Group("/2")
	group12.GET("/2", func(ctx *gee.Context) {
		ctx.String(http.StatusOK, "group1/2/2")
	})

	group2.GET("/1", func(ctx *gee.Context) {
		ctx.String(http.StatusOK, "group2/1")
	})

	group2.GET("/2", func(ctx *gee.Context) {
		ctx.String(http.StatusOK, "group2/2")
	})

	t.Fatal(g.Run(addr))
}

func request(t *testing.B, reqFunc httpRequest, url string, result string) {
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

var ok bool = false

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

	if !ok {
		ok = true
		go runServer(t, baseUrl)
	}

	for n := 0; n < t.N; n++ {
		for path, result := range getReq {
			url := baseUrl + path
			request(t, http.Get, url, result)
		}
	}

}
