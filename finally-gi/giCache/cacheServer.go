package giCache

import (
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

type CacheServer struct {
	addr       string
	basePath   string
	rwmu       sync.RWMutex
	cacheGroup map[string]*GiCache
}

func NewCacheServer(addr string) *CacheServer {
	return &CacheServer{
		addr:       addr,
		basePath:   CacheDefaultPath,
		cacheGroup: make(map[string]*GiCache),
	}
}

func (cs *CacheServer) AddCache(caches ...*GiCache) {
	cs.rwmu.Lock()
	defer cs.rwmu.Unlock()
	for _, c := range caches {
		oldC := cs.cacheGroup[c.name]
		cs.cacheGroup[c.name] = c

		if oldC != nil {
			log.Printf("cacheGroup[%s] been over write\n", c.name)
		}
	}
}

func (cs *CacheServer) GetCache(name string) *GiCache {
	cs.rwmu.RLock()
	defer cs.rwmu.RUnlock()
	return cs.cacheGroup[name]
}

func (cs *CacheServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if keepAlive(w, r) {
		return
	}

	if !strings.HasPrefix(r.URL.Path, cs.basePath) {
		dumpEmpty(w)
		return
	}

	absUrl := r.URL.Path[len(cs.basePath):]
	paths := strings.SplitN(absUrl, "/", 2)
	cache := cs.GetCache(paths[0])
	if cache == nil {
		dumpEmpty(w)
		return
	}

	v, err := cache.Get(paths[1])
	if err != nil {
		log.Printf("cache.Get(%s) err:\n%s\n", paths[1], err)
		return
	}

	_, err = w.Write(v.ByteSlice())
	if err != nil {
		log.Printf("w.Write err:\n%s\n", err)
	}
}

func (cs *CacheServer) Run() {
	if len(cs.cacheGroup) == 0 {
		panic("empty cacheGroup !")
	}

	log.Println("Cache Server Running at ", cs.addr)
	log.Fatalln(http.ListenAndServe(cs.addr, cs))
}

func dumpEmpty(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte{})
}

// keepAlive 处理，如果是普通 req 则跳过
func keepAlive(w http.ResponseWriter, r *http.Request) bool {
	isKeepAlive := r.URL.Path == CacheKeepAlivePath
	if isKeepAlive {

		if body, err := io.ReadAll(r.Body); err != nil {
			log.Printf("Cache Server received Err KeepAlive:\n%s\n", err)
			dumpEmpty(w)
		} else if string(body) == CacheKeepAliveSyc {
			w.Write([]byte(CacheKeepAliveAck))
		}
	}

	return isKeepAlive
}
