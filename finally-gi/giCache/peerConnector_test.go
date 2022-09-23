package giCache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

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

func runCacheServer(addrs ...string) {
	for _, addr := range addrs {
		c1 := New("cache1", 5, GetterFunc(func(key string) ([]byte, error) {
			if _, ok := data1[key]; !ok {
				return nil, fmt.Errorf("%s - get key (%s) not exist", addr, key)
			}

			return []byte(data1[key]), nil

		}), nil)

		c2 := New("cache2", 5, GetterFunc(func(key string) ([]byte, error) {
			if _, ok := data2[key]; !ok {
				return nil, fmt.Errorf("%s - get key (%s) not exist", addr, key)
			}

			return []byte(data2[key]), nil

		}), nil)
		cs := NewCacheServer(addr)
		cs.AddCache(c1, c2)

		go cs.Run()
	}
}

func createCachePool() *CachePool {
	addrs := []string{
		"localhost:8081",
		"localhost:8082",
		"localhost:8083",
	}

	cp := NewCachePool()

	runCacheServer(addrs...)
	cp.AddPeer(addrs...)
	return cp
}

func runCachePoolServe(t testing.TB, addr string) {
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

	t.Fatal(http.ListenAndServe(addr, nil))
}

func TestCachePool(t *testing.T) {
	serveAddr := "localhost:8999"
	wg := sync.WaitGroup{}
	getReq := map[string]string{}

	go runCachePoolServe(t, serveAddr)

	for k, v := range data1 {
		key := "http://" + serveAddr + "/cacheApi/cache1/" + k
		getReq[key] = v
	}

	for k, v := range data2 {
		key := "http://" + serveAddr + "/cacheApi/cache2/" + k
		getReq[key] = v
	}

	for req, result := range getReq {
		wg.Add(1)
		go func(req string, result string) {
			time.Sleep(time.Second)
			request(t, http.Get, req, result)
			wg.Done()
		}(req, result)
	}

	wg.Wait()
}
