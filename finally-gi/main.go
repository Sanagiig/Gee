package main

import (
	"finally-gi/giCache"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type ByteView struct {
	b []byte
}

func (b *ByteView) Len() int {
	return len(b.b)
}

func (b *ByteView) String() string {
	return string(b.b)
}

func TestLru() {
	lru := giCache.NewLruCache(10, nil)
	lru.Set("1", &ByteView{[]byte("1")})
	lru.Set("22", &ByteView{[]byte("22")})

	res1, _ := lru.Get("1")
	res2, _ := lru.Get("22")

	fmt.Printf("res1 :%s \n", res1)
	if res1.(*ByteView).String() != "1" || res2.(*ByteView).String() != "22" {
		//t.Fatalf("res1 :%s != 1 || \n res2 : %s != 22\n", res1, res2)
	}
	lru.Set("333", &ByteView{[]byte("333")})
	res3, _ := lru.Get("333")

	if res3.(*ByteView).String() != "333" {
		//t.Fatalf("res3 :%s != 333 \n", res3)
	}

	res1, _ = lru.Get("1")
	if res1.(*ByteView) != nil {
		//t.Fatalf("res1 not nil (%s) \n", res1)
	}
}

func TestGiCacheServer() {
	var data = map[string]string{
		"1": "111",
		"2": "222",
		"3": "333",
	}

	giCache.New("test", 2<<10,
		giCache.GetterFunc(func(key string) ([]byte, error) {
			v := data[key]
			if v == "" {
				return nil, fmt.Errorf("data[%s] not found\n", key)
			}
			fmt.Printf("get data[%s] => %s\n", key, v)
			return []byte(v), nil
		}),
		nil,
	)

	cs := giCache.NewCacheServer("localhost:8888")
	cs.Run()
}

var data1 = map[string]string{
	"1": "111",
	"2": "222",
	"3": "333",
}

var data2 = map[string]string{
	"4": "444",
	"5": "555",
	"6": "666",
}

type httpRequest = func(url string) (resp *http.Response, err error)

func request(t testing.TB, reqFunc httpRequest, url string, result string) {
	res, err := reqFunc(url)
	if err != nil {
		log.Fatalf("[%30s] err \n %v\n", url, err.Error())
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err.Error())
	}

	if !strings.EqualFold(string(data), result) {

		log.Fatalf("[%s] err\n expected:  \"%v\" \n get wrong data: \"%v\" \n", url, result, string(data))
	}
}

func runCacheServer(addrs ...string) {
	for _, addr := range addrs {
		c1 := giCache.New("cache1", 5, giCache.GetterFunc(func(key string) ([]byte, error) {
			log.Printf("%s - get key %s", addr, key)
			if _, ok := data1[key]; !ok {
				return nil, fmt.Errorf("%s - get key (%s) not exist", addr, key)
			}

			return []byte(data1[key]), nil

		}), nil)

		c2 := giCache.New("cache2", 5, giCache.GetterFunc(func(key string) ([]byte, error) {
			log.Printf("%s - get key %s", addr, key)
			if _, ok := data2[key]; !ok {
				return nil, fmt.Errorf("%s - get key (%s) not exist", addr, key)
			}

			return []byte(data2[key]), nil

		}), nil)
		cs := giCache.NewCacheServer(addr)
		cs.AddCache(c1, c2)

		go cs.Run()
	}
}

func createCachePool() *giCache.CachePool {
	addrs := []string{
		"localhost:8081",
		"localhost:8082",
		"localhost:8083",
	}

	cp := giCache.NewCachePool()

	runCacheServer(addrs...)
	cp.AddPeer(addrs...)
	return cp
}

func runCachePoolServe(addr string) {
	prePath := "/cacheApi/"
	cp := createCachePool()
	http.HandleFunc(prePath, func(w http.ResponseWriter, r *http.Request) {
		keyPath := r.URL.Path[len(prePath):]
		parts := strings.SplitN(keyPath, "/", 2)
		data, err := cp.Get(parts[0], parts[1])
		if err != nil {
			log.Println(err.Error())
		}
		log.Printf("cacheApi - GET - [key: %s/%s] [val: \"%s\"]", parts[0], parts[1], data)
		w.Write(data.ByteSlice())
	})

	log.Fatal(http.ListenAndServe(addr, nil))
}

func TestSingleFlight() {
	var data = map[string]string{
		"1": "111",
		"2": "222",
		"3": "333",
	}

	getReq := map[string]string{
		"/_giCache/test/1": "111",
		"/_giCache/test/2": "222",
		"/_giCache/test/3": "333",
	}
	addr := "localhost:9998"
	baseUrl := "http://" + addr
	count := 0
	okCount := 0
	cs := giCache.NewCacheServer(addr)
	c1 := giCache.New("test", 2<<10,
		giCache.GetterFunc(func(key string) ([]byte, error) {
			v := data[key]
			if v == "" {
				return nil, fmt.Errorf("data[%s] not found\n", key)
			}
			time.Sleep(time.Second * 3)
			count++
			fmt.Printf("get data[%s] => %s\n", key, v)
			return []byte(v), nil
		}),
		nil,
	)

	wg := sync.WaitGroup{}
	c := &http.Client{Timeout: time.Second * 10000}
	for i := 0; i < 50; i++ {
		for path, result := range getReq {
			url := baseUrl + path
			wg.Add(1)
			go func(url string, result string) {
				request(nil, c.Get, url, result)
				okCount++
				wg.Done()
			}(url, result)
		}
	}

	cs.AddCache(c1)
	go cs.Run()

	wg.Wait()
	log.Println("ok count is :", okCount)
	if l := len(data); count > l {
		log.Fatalf("get remote data times is too much . expected [%d] but got [%d]", l, count)
	}
}

func main() {
	TestSingleFlight()
	//runCachePoolServe("localhost:8888")
	//TestGiCacheServer()
	//g := gi.New()
	//g.Use(middlewares.Recover())
	//g.Use(middlewares.TimeLog())
	//group1 := g.Group("/group1")
	//group2 := g.Group("/group2")
	//
	//g.Static("/static", "../static")
	//
	//g.Get("/panic", func(ctx *gi.Context) {
	//	panic("hehe")
	//})
	//
	//group1.Get("/1", func(ctx *gi.Context) {
	//	ctx.String("group1/1")
	//})
	//group11 := group1.Group("/1")
	//group11.Get("/1", func(ctx *gi.Context) {
	//	ctx.String("group1/1/1")
	//})
	//
	//group1.Get("/2", func(ctx *gi.Context) {
	//	ctx.String("group1/2")
	//})
	//group12 := group1.Group("/2")
	//group12.Get("/2", func(ctx *gi.Context) {
	//	time.Sleep(time.Second)
	//	ctx.String("group1/2/2")
	//})
	//
	//group2.Get("/1", func(ctx *gi.Context) {
	//	ctx.String("group2/1")
	//})
	//
	//group2.Get("/2", func(ctx *gi.Context) {
	//	ctx.String("group2/2")
	//})
	//
	//log.Fatal(g.Run(":8080"))
}
