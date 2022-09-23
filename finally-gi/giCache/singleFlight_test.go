package giCache

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestSingleFlight(t *testing.T) {
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
	addr := "localhost:8888"
	baseUrl := "http://" + addr
	count := 0
	testTimes := 1000

	cs := NewCacheServer(addr)
	c1 := New("test", 2<<10,
		GetterFunc(func(key string) ([]byte, error) {
			v := data[key]
			if v == "" {
				return nil, fmt.Errorf("data[%s] not found\n", key)
			}
			count++
			fmt.Printf("get data[%s] => %s\n", key, v)
			return []byte(v), nil
		}),
		nil,
	)

	oo := 0
	wg := sync.WaitGroup{}
	for i := 0; i < testTimes; i++ {
		for path, result := range getReq {
			url := baseUrl + path
			wg.Add(1)
			oo++
			go func(url string, result string) {
				request(t, http.Get, url, result)
				wg.Done()
			}(url, result)
		}
	}

	cs.AddCache(c1)
	go cs.Run()

	wg.Wait()
	time.Sleep(time.Second)
	if l := len(data); count > l {
		t.Fatalf("get remote data times is too much . expected [%d] but got [%d]", l, count)
	}
}
