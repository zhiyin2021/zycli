package tools

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/zhiyin2021/zycli/tools/cache"
	"github.com/zhiyin2021/zycli/tools/logger"
)

type subitem struct {
	Accept int
	Reject int
	Tm     int64
}
type SideRateLimiter struct {
	mutex     sync.Mutex
	window    time.Duration
	maxTokens int
	tokens    []time.Time
	record    map[int64]*subitem
	chToken   chan struct{}
	last      *subitem
}

func NewSideRateLimiter(ctx context.Context, window time.Duration, maxTokens int) *SideRateLimiter {
	rl := &SideRateLimiter{
		window:    window,
		maxTokens: maxTokens,
		tokens:    []time.Time{},
		record:    map[int64]*subitem{},
		chToken:   make(chan struct{}, maxTokens),
	}
	go func() {
		// delay := window / time.Duration(maxTokens)
		tm := time.NewTicker(time.Microsecond * 10)
		for {
			select {
			case <-ctx.Done():
				logger.Println("limiter2.quit")
				return
			case <-tm.C:
				now := cache.NowMicr()
				for len(rl.tokens) > 0 && now.Sub(rl.tokens[0]) > rl.window {
					rl.tokens = rl.tokens[1:]
					<-rl.chToken
				}
				if len(rl.tokens) > 0 {
					sec := time.Second - now.Sub(rl.tokens[0])
					if sec <= 0 {
						sec = time.Microsecond
					}
					tm.Reset(sec)
				} else {
					tm.Reset(time.Second)
				}
			}
		}
	}()
	return rl
}

func (rl *SideRateLimiter) Allow() (flag bool) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	select {
	case rl.chToken <- struct{}{}:
		now := cache.NowSec()
		rl.tokens = append(rl.tokens, now)
		rl.last = rl.calcCount(now)
		rl.last.Accept++
		flag = true
	default:
		now := cache.NowSec()
		rl.last = rl.calcCount(now)
		rl.last.Reject++
		flag = false
	}
	return flag
}

func (rl *SideRateLimiter) Wait() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	rl.wait()
}
func (rl *SideRateLimiter) wait() {
	rl.chToken <- struct{}{}
	now := cache.NowSec()
	rl.tokens = append(rl.tokens, now)

	rl.last = rl.calcCount(now)
	rl.last.Accept++
}

func (rl *SideRateLimiter) CheckWait() time.Duration {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	now := cache.NowSec()

	// 检查是否可以添加新的 token
	if len(rl.tokens) >= rl.maxTokens && len(rl.tokens) > 0 {
		if sec := now.Sub(rl.tokens[0]); sec > 0 {
			return sec
		}
	}
	return 0
}

func (rl *SideRateLimiter) calcCount(now time.Time) *subitem {
	unixTm := now.Unix()
	// 移除过期的 超期的记录
	for k := range rl.record {
		if unixTm-k > 30 {
			delete(rl.record, k)
		}
	}
	if v, ok := rl.record[unixTm]; ok {
		return v
	} else {
		si := &subitem{Tm: unixTm, Accept: 0, Reject: 0}
		rl.record[unixTm] = si
		return si
	}
}
func (rl *SideRateLimiter) IsFull() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	return len(rl.tokens) >= rl.maxTokens
}

func (rl *SideRateLimiter) Record() []subitem {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	now := cache.NowSec()
	rl.calcCount(now)
	var data []subitem
	for _, v := range rl.record {
		data = append(data, *v)
	}
	return data
}
func (rl *SideRateLimiter) Total() int {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	return len(rl.tokens)
}
func (m *subitem) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

func (rl *SideRateLimiter) Last() subitem {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	if rl.last == nil {
		return subitem{Tm: cache.NowSec().Unix(), Accept: 0, Reject: 0}
	}
	return subitem{Tm: rl.last.Tm, Accept: rl.last.Accept, Reject: rl.last.Reject}
}
