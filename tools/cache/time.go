package cache

import (
	"sync"
	"time"
)

type timeCache struct {
	mu       sync.RWMutex
	cached   time.Time
	duration time.Duration
}

func newTimeCache(duration time.Duration) *timeCache {
	return &timeCache{
		duration: duration,
		cached:   time.Now(),
	}
}

var millCache = newTimeCache(time.Millisecond)
var micrCache = newTimeCache(time.Microsecond)
var secCache = newTimeCache(time.Second)

func (tc *timeCache) Now() time.Time {
	tc.mu.RLock()
	if time.Since(tc.cached) < tc.duration {
		tc.mu.RUnlock()
		return tc.cached
	}
	tc.mu.RUnlock()

	tc.mu.Lock()
	defer tc.mu.Unlock()

	// 再次检查以避免竞争
	if time.Since(tc.cached) < tc.duration {
		return tc.cached
	}

	tc.cached = time.Now()
	return tc.cached
}

func NowMill() time.Time {
	return millCache.Now()
}

func NowMicr() time.Time {
	return micrCache.Now()
}

func NowSec() time.Time {
	return secCache.Now()
}
