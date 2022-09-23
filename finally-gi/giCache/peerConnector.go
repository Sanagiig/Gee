package giCache

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	PeerActive = iota
	PeerInactive
)

const MaxDisCoonCount = 3
const DefaultProtocal = "http://"

type trigger func(addr string)
type PeerConnector struct {
	addr         string
	disConnCount int // 连接失效次数
	status       int
	onActive     trigger
	onInactive   trigger
}

func NewPeerConnector(addr string, onActive trigger, onInactive trigger) *PeerConnector {
	p := &PeerConnector{
		addr:         addr,
		disConnCount: 0,
		status:       PeerInactive,
		onActive:     onActive,
		onInactive:   onInactive,
	}

	go p.heartbeat()
	return p
}

func (p *PeerConnector) keepAlive() {
	url := DefaultProtocal + p.addr + CacheKeepAlivePath
	body := bytes.NewBuffer([]byte(CacheKeepAliveSyc))
	data, err := disposeResponse(http.Post(url, "text/plain", body))

	if err != nil {
		p.disConnCount++
		log.Printf("%s keep alive err\n\t%s\n", p.addr, err)
	} else if data.String() != CacheKeepAliveAck {
		p.disConnCount++
		log.Printf("%s keep alive err. got unkonw data:\n\t%s\n", p.addr, data)
	} else {
		p.disConnCount = 0
	}

	switch {
	case p.status == PeerInactive && p.disConnCount == 0:
		p.status = PeerActive
		if p.onActive != nil {
			p.onActive(p.addr)
			log.Printf("Server[%s] has been active", p.addr)
		}
	case p.status == PeerActive && p.disConnCount >= MaxDisCoonCount:
		p.status = PeerInactive
		if p.onInactive != nil {
			p.onInactive(p.addr)
			log.Printf("Server[%s] is been inactive", p.addr)
		}
	}
}

func (p *PeerConnector) heartbeat() {
	for {
		p.keepAlive()
		time.Sleep(time.Second * 3)
	}
}

// 获取 http 接口的数据
func (p *PeerConnector) Get(paths ...string) (ByteView, error) {
	url := DefaultProtocal + p.addr + CacheDefaultPath + strings.Join(paths, "/")
	log.Printf("conn the %s get %s", p.addr, strings.Join(paths, "/"))
	return disposeResponse(http.Get(url))
}

func disposeResponse(resp *http.Response, err error) (ByteView, error) {
	if err != nil {
		return ByteView{}, err
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return ByteView{}, err
	}

	return ByteView{b: copyByte(body)}, nil
}
