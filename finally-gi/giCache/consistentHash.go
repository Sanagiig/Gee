package giCache

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type HashFunc func(bs []byte) uint32

type ConsistentHash struct {
	replicas    int
	rwmu        sync.RWMutex
	hash2keyMap map[int]string
	hashList    []int
	hash        HashFunc
}

func NewConsistentHash(replicas int, hash HashFunc) *ConsistentHash {
	ch := &ConsistentHash{
		replicas:    replicas,
		rwmu:        sync.RWMutex{},
		hash2keyMap: make(map[int]string),
		hashList:    []int{},
		hash:        hash,
	}

	if hash == nil {
		ch.hash = crc32.ChecksumIEEE
	}

	return ch
}

func (c *ConsistentHash) Add(keys ...string) {
	c.rwmu.Lock()
	defer c.rwmu.Unlock()

	for _, k := range keys {
		for i := 0; i < c.replicas; i++ {
			hash := int(c.hash([]byte(strconv.Itoa(i) + k)))
			c.hashList = append(c.hashList, hash)
			c.hash2keyMap[hash] = k
		}
	}

	sort.Ints(c.hashList)
}

func (c *ConsistentHash) Remove(keys ...string) {
	c.rwmu.Lock()
	defer c.rwmu.Unlock()

	listLen := len(c.hashList)
	newHashList := make([]int, listLen, listLen)

	for _, rmk := range keys {
		for i := 0; i < c.replicas; i++ {
			hash := int(c.hash([]byte(strconv.Itoa(i) + rmk)))
			delete(c.hash2keyMap, hash)
		}
	}

	for hash, _ := range c.hash2keyMap {
		newHashList = append(newHashList, hash)
	}

	sort.Ints(newHashList)
	c.hashList = newHashList
}

func (c *ConsistentHash) Get(key string) string {
	c.rwmu.RLock()
	defer c.rwmu.RUnlock()

	hash := int(c.hash([]byte(key)))
	idx := sort.Search(len(c.hashList), func(i int) bool {
		return c.hashList[i] >= hash
	}) % len(c.hashList)

	return c.hash2keyMap[c.hashList[idx]]
}
