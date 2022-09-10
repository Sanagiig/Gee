package gee

import (
	"io"
	"log"
	"net/http"
	"sync"
	"testing"
)

//func newTestRouter() *Router {
//	r := NewRouter()
//	r.addRoute("GET", "/", nil)
//	r.addRoute("GET", "/hello/:name", nil)
//	r.addRoute("GET", "/hello/b/c", nil)
//	r.addRoute("GET", "/hi/:name", nil)
//	r.addRoute("GET", "/assets/*filepath", nil)
//	return r
//}
//
//func TestParsePattern(t *testing.T) {
//	if !reflect.DeepEqual(parsePattern("/p/:name"), []string{"p", ":name"}) {
//		t.Fatal("test1 parsePattern failed")
//	}
//
//	if !reflect.DeepEqual(parsePattern("/p/*"), []string{"p", "*"}) {
//		t.Fatal("test2 parsePattern failed")
//	}
//
//	if !reflect.DeepEqual(parsePattern("/p/*name/*"), []string{"p", "*name"}) {
//		t.Fatal("test3 parsePattern failed")
//	}
//}
//
//func TestGetRoute(t *testing.T) {
//	r := newTestRouter()
//	n, ps := r.getRoute("GET", "/hello/geektutu")
//
//	if n == nil {
//		t.Fatal("nil shouldn't be returned")
//	}
//
//	if n.pattern != "/hello/:name" {
//		t.Fatal("should match /hello/:name")
//	}
//
//	if ps["name"] != "geektutu" {
//		t.Fatal("name should be equal to 'geektutu'")
//	}
//
//	fmt.Printf("matched path: %s, params['name']: %s\n", n.pattern, ps["name"])
//}

func initGee() {
	g := New()
	g.GET("/hello", func(ctx *Context) {
		ctx.String(http.StatusOK, "hello")
	})

	g.GET("/start/*test", func(ctx *Context) {
		ctx.String(http.StatusOK, ctx.Param("test"))
	})

	log.Fatal(g.Run(":8888"))
}

func TestRequest(t *testing.T) {
	wg := sync.WaitGroup{}

	go initGee()
	wg.Add(1)
	go func() {
		defer wg.Done()

		resp, err := http.Get("http://localhost:8888/hello")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatal("/hello err", err.Error())
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err.Error())
		}
		if string(body) != "hello" {
			t.Fatal("string(body) != \"hello\"")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		testParam := "123"
		resp, err := http.Get("http://localhost:8888/start/" + testParam)
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Fatal("/start/123 err", err.Error())
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err.Error())
		}
		println("xx")
		if string(body) != testParam {
			t.Fatal("testParam is ", testParam, "get :", string(body))
		}
	}()

	wg.Wait()
}
