package singleFlight

import (
	"finally-gi/giCache"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type httpRequest = func(url string) (resp *http.Response, err error)

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

		t.Fatalf("[%s] err\n expected:  \"%v\" \n get wrong data: \"%v\" \n", url, result, string(data))
	}
}

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
	cs := giCache.NewCacheServer(addr)
	c1 := giCache.New("test", 2<<10,
		giCache.GetterFunc(func(key string) ([]byte, error) {
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

	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		for path, result := range getReq {
			url := baseUrl + path
			wg.Add(1)
			go func(url string, result string) {
				time.Sleep(time.Second)
				request(t, http.Get, url, result)
				wg.Done()
			}(url, result)
		}
	}

	cs.AddCache(c1)
	go cs.Run()

	wg.Wait()
	if l := len(data); count > l {
		t.Fatalf("get remote data times is too much . expected [%d] but got [%d]", l, count)
	}
}
