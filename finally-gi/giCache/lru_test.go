package giCache

import (
	"fmt"
	"testing"
)

func TestLru(t *testing.T) {
	lru := NewLruCache(10, nil)
	lru.Set("1", &ByteView{[]byte("1")})
	lru.Set("22", &ByteView{[]byte("22")})

	res1, _ := lru.Get("1")
	res2, _ := lru.Get("22")

	fmt.Printf("res1 :%s \n", res1)
	if res1.(*ByteView).String() != "1" || res2.(*ByteView).String() != "22" {
		t.Fatalf("res1 :%s != 1 || \n res2 : %s != 22\n", res1, res2)
	}
	lru.Set("333", &ByteView{[]byte("333")})
	res3, _ := lru.Get("333")

	if res3.(*ByteView).String() != "333" {
		t.Fatalf("res3 :%s != 333 \n", res3)
	}

	res1, _ = lru.Get("1")
	if res1 != nil {
		t.Fatalf("res1 not nil (%s) \n", res1)
	}
}
