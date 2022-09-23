package giCache

import (
	"fmt"
	"testing"
)

var data = map[string]string{
	"1": "111",
	"2": "222",
	"3": "333",
}

func TestGiCache(t *testing.T) {
	cache := New("c1", 100, GetterFunc(func(key string) ([]byte, error) {
		v := data[key]
		if v == "" {
			return nil, fmt.Errorf("data[%s] not found\n", key)
		}
		fmt.Printf("get data[%s] => %s\n", key, v)
		return []byte(v), nil
	}), nil)

	for i := 0; i < 3; i++ {
		for k, _ := range data {
			v, e := cache.Get(k)
			if e != nil {
				t.Fatalf("get err: %s \n", e)
			}
			fmt.Printf("key:%s , val:%s \n", k, v)
		}
	}
}
