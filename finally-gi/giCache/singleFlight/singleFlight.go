package singleFlight

import (
	"reflect"
	"runtime"
	"strings"
	"sync"
)

type CallerFunc func() (any, error)
type caller struct {
	wg     sync.WaitGroup
	symbol string
	fn     CallerFunc
	args   []string
	value  any
	err    error
}

func NewCaller(synbol string, fn CallerFunc, args ...string) *caller {
	return &caller{
		wg:     sync.WaitGroup{},
		symbol: synbol,
		fn:     fn,
		args:   args,
	}
}

func (c *caller) Exc() {
	c.wg.Add(1)
	c.value, c.err = c.fn()
	c.wg.Done()
}

// 单飞处理 ， 防止缓存击穿
type SingleFlight struct {
	mu          sync.Mutex
	callerGroup map[string]*caller
}

func NewSingleFlight() *SingleFlight {
	return &SingleFlight{
		mu:          sync.Mutex{},
		callerGroup: make(map[string]*caller),
	}
}

func (s *SingleFlight) addCaller(fn CallerFunc, callerKey string, args ...string) *caller {
	c := NewCaller(callerKey, fn, args...)
	s.callerGroup[callerKey] = c
	return c
}

func (s *SingleFlight) getCaller(callerKey string) *caller {
	return s.callerGroup[callerKey]
}

func (s *SingleFlight) Do(fn CallerFunc, args ...string) (any, error) {
	s.mu.Lock()

	fnName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	callerKey := fnName + strings.Join(args, "-")
	curCaller := s.getCaller(callerKey)

	if curCaller == nil {
		// 此处如果 Lock 放在 addCaller外 是为了防止 Exc 执行错误，删除组
		// 时在 别的 goroutine 获取到异常数据
		// 但是放在 addCaller 内部影响也不大，就某些 goroutine 获取到错误，会重新请求

		curCaller = s.addCaller(fn, callerKey, args...)
		s.mu.Unlock()
		curCaller.Exc()

		// 清空
		if curCaller.err != nil {
			delete(s.callerGroup, callerKey)
		}

		return curCaller.value, curCaller.err
	}

	s.mu.Unlock()
	curCaller.wg.Wait()
	return curCaller.value, curCaller.err
}
