package giCache

import (
	"container/list"
	"fmt"
)

type Value interface {
	Len() int
}

type Data struct {
	key string
	val Value
}

func (d *Data) Len() int {
	return len(d.key) + d.val.Len()
}

type EvictedFunc = func(data *Data)
type LruCache struct {
	list      *list.List
	store     map[string]*list.Element
	capacity  int
	size      int
	onEvicted EvictedFunc
}

func NewLruCache(capacity int, onEvicted EvictedFunc) *LruCache {
	return &LruCache{
		list:      &list.List{},
		store:     map[string]*list.Element{},
		capacity:  capacity,
		size:      0,
		onEvicted: onEvicted,
	}
}

func (l *LruCache) Get(key string) (val Value, ok bool) {
	ele := l.store[key]
	if ele != nil {
		return ele.Value.(*Data).val, true
	}
	return
}

func (l *LruCache) Set(key string, val Value) error {
	if val.Len() > l.capacity {
		return fmt.Errorf("this LRU Cache maximum size is %d B", l.capacity)
	}

	oldEle := l.store[key]
	if oldEle != nil {
		l.size += val.Len() - oldEle.Value.(*Data).val.Len()
		l.list.Remove(oldEle)
	} else {
		l.size += len(key) + val.Len()
	}

	ele := l.list.PushFront(&Data{key: key, val: val})
	l.store[key] = ele

	for l.size > l.capacity {
		l.removeOldestEle()
	}
	return nil
}

func (l *LruCache) removeOldestEle() {
	lastEle := l.list.Back()
	lastVal := lastEle.Value.(*Data)

	l.size -= lastVal.Len()
	l.list.Remove(lastEle)
	delete(l.store, lastVal.key)

	if l.onEvicted != nil {
		l.onEvicted(lastVal)
	}
}
