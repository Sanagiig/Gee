package giCache

import (
	"finally-gi/giCache/singleFlight"
	"fmt"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

// 获取远程数据
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	cacheGroup = make(map[string]*GiCache)
	groupMu    = sync.RWMutex{}
)

type GiCache struct {
	name         string
	getter       Getter
	mainCache    *Cache
	remoteLoader *singleFlight.SingleFlight
}

func New(name string, capacity int, getter Getter, evictedFunc EvictedFunc) *GiCache {
	if getter == nil {
		panic("getter got nil")
	}

	c := &GiCache{
		name:         name,
		getter:       getter,
		mainCache:    NewCache(capacity, evictedFunc),
		remoteLoader: singleFlight.NewSingleFlight(),
	}

	groupMu.Lock()
	cacheGroup[name] = c
	groupMu.Unlock()
	return c
}

func (g *GiCache) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	v, ok := g.mainCache.Get(key)
	if !ok {
		return g.getLocally(key)
	}

	return v.(ByteView), nil
}

func (g *GiCache) getLocally(key string) (ByteView, error) {
	v, e := g.remoteLoader.Do(func() (any, error) {
		return g.getter.Get(key)
	}, key)

	if e != nil {
		return ByteView{}, e
	}

	b := ByteView{b: copyByte(v.([]byte))}
	g.populate2Cache(key, b)
	return b, nil
}

func (g *GiCache) populate2Cache(key string, val ByteView) {
	e := g.mainCache.Set(key, val)
	if e != nil {
		panic(e)
	}
}

func GetCache(name string) *GiCache {
	groupMu.RLock()
	defer groupMu.RUnlock()
	return cacheGroup[name]
}
