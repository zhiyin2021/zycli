package tools

import (
	"sync"
	"time"
)

type runLast struct {
	sync.Mutex
	next    chan bool
	timeout time.Duration
}

// 解决短时间内多次请求同步刷新数据,只响应指定延迟内最后一次请求
func NewRunLast(timeout time.Duration) *runLast {
	return &runLast{
		next:    make(chan bool),
		timeout: timeout,
	}
}
func (r *runLast) Run(callback func()) {
	r.Lock()
	select {
	case r.next <- true:
	case <-time.After(r.timeout / 2):
	}
	r.Unlock()
	select {
	case <-r.next:
	case <-time.After(r.timeout):
		callback()
	}
}
