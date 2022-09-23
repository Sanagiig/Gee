package giCache

import (
	"fmt"
	"log"
)

type CachePool struct {
	cacheAddrHashMap *ConsistentHash
	activeCaches     map[string]*PeerConnector
	inactiveCaches   map[string]*PeerConnector
}

func NewCachePool() *CachePool {
	return &CachePool{
		cacheAddrHashMap: NewConsistentHash(5, nil),
		activeCaches:     make(map[string]*PeerConnector),
		inactiveCaches:   make(map[string]*PeerConnector),
	}
}

func (c CachePool) onCacheActive(addr string) {
	if c.inactiveCaches[addr] == nil {
		log.Printf("cache server active addr [ %s ] is nil", addr)
		return
	}

	exchangePeer(c.activeCaches, c.inactiveCaches, addr)
	c.cacheAddrHashMap.Add(addr)
}

func (c CachePool) onCacheInactive(addr string) {
	if c.activeCaches[addr] == nil {
		log.Printf("cache server inactive addr [ %s ] is nil", addr)
		return
	}

	exchangePeer(c.inactiveCaches, c.activeCaches, addr)
	c.cacheAddrHashMap.Remove(addr)
}

func (c *CachePool) AddPeer(addrs ...string) {
	for _, addr := range addrs {
		c.inactiveCaches[addr] = NewPeerConnector(addr, c.onCacheActive, c.onCacheInactive)
	}
}

func (c *CachePool) Get(name string, key string) (ByteView, error) {
	if len(c.activeCaches) == 0 {
		return ByteView{}, fmt.Errorf("CachePool's active server is empty")
	}

	peerAddr := c.cacheAddrHashMap.Get(name + key)
	peer := c.activeCaches[peerAddr]
	if peer == nil {
		return ByteView{}, fmt.Errorf("CachePool's active server is empty")
	}

	return peer.Get(name, key)
}

func exchangePeer(dist map[string]*PeerConnector, src map[string]*PeerConnector, key string) {
	dist[key] = src[key]
	delete(src, key)
}
