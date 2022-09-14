package geeCache

import (
	pb "example14/geeCache/cachePB"
	"example14/geeCache/singleFlight"
	"fmt"
	"log"
	"sync"
)

var (
	rwmu   sync.RWMutex
	groups = make(map[string]*CacheGroup)
)

type CacheGroup struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
	loader    *singleFlight.Group
}

// NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *CacheGroup {
	if getter == nil {
		panic("nil Getter")
	}

	rwmu.Lock()
	defer rwmu.Unlock()
	g := &CacheGroup{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleFlight.Group{},
	}
	groups[name] = g
	return g
}

func (g *CacheGroup) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// Get value for a key from cache
func (g *CacheGroup) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}

func (g *CacheGroup) load(key string) (value ByteView, err error) {
	viewi, err := g.loader.Do(key, func() (any, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}

	return
}

func (g *CacheGroup) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)

	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

func (g *CacheGroup) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *CacheGroup) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *CacheGroup {
	rwmu.RLock()
	g := groups[name]
	rwmu.RUnlock()
	return g
}
