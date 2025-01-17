package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zhiyin2021/zycli/tools/logger"
)

type item struct {
	value      any
	sliding    time.Duration
	expiration int64
	f          func(any)
}

// NewMemory memory模式
func NewMemory(ctx context.Context) *Memory {
	cc := &Memory{
		items: map[any]*item{},
	}
	go cc.runing(ctx)
	return cc
}

type Memory struct {
	sync.Mutex
	items  map[any]*item
	expire time.Duration
}

func (*Memory) String() string {
	return "memory"
}
func (m *Memory) runing(ctx context.Context) {
	tmr := time.NewTicker(time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tmr.C:
			func() {
				m.Lock()
				defer m.Unlock()
				now := NowMicr().UnixMicro()
				for key, item := range m.items {
					if item.expiration > 0 && item.expiration < now {
						if item.f != nil {
							go func() {
								defer func() {
									if e := recover(); e != nil {
										logger.Errorln("memory.cache.del.err", e)
									}
								}()
								item.f(item.value)
							}()
						}
						delete(m.items, key)
					}
				}
			}()
		}
	}
}
func (m *Memory) setItem(key, val any, sliding time.Duration, expire time.Duration) *item {
	item := &item{
		value:      val,
		sliding:    sliding,
		expiration: 0,
	}
	if expire > 0 {
		item.expiration = NowMicr().Add(expire).UnixMicro()
	}
	m.items[key] = item
	return item
}
func (m *Memory) del(key any) any {
	item, ok := m.items[key]
	if ok {
		delete(m.items, key)
		return item.value
	}
	return nil
}
func (m *Memory) Get(key any) any {
	m.Lock()
	defer m.Unlock()
	if item, flag := m.items[key]; flag {
		if item.sliding > 0 {
			item.expiration = NowMicr().Add(item.sliding).UnixMicro()
		}
		return item.value
	}
	return nil
}

func (m *Memory) GetBy(check func(any) bool) any {
	if check != nil {
		m.Lock()
		defer m.Unlock()
		for _, v := range m.items {
			if check(v.value) {
				return v.value
			}
		}
	}
	return nil
}

func (m *Memory) GetOrStoreBySliding(key any, val any, expire time.Duration) (any, bool) {
	m.Lock()
	defer m.Unlock()
	if t, flag := m.items[key]; flag {
		t.expiration = NowMicr().Add(expire).UnixMicro()
		return t.value, true
	} else {
		m.setItem(key, val, expire, expire)
		return val, false
	}
}

func (m *Memory) GetOrStore(key any, val any) (any, bool) {
	m.Lock()
	defer m.Unlock()

	if t, flag := m.items[key]; flag {
		return t.value, true
	} else {
		m.setItem(key, val, m.expire, m.expire)
		return t.value, false
	}
}

func (m *Memory) GetAndDel(key any) any {
	m.Lock()
	defer m.Unlock()
	if t, ok := m.items[key]; ok {
		delete(m.items, key)
		return t.value
	}
	return nil
}

// 获取缓存时自动延时
func (m *Memory) SetBySliding(key, val any, expire time.Duration) {
	m.Lock()
	defer m.Unlock()
	m.setItem(key, val, expire, expire)
}

func (m *Memory) Set(key, val any) {
	m.Lock()
	defer m.Unlock()
	m.setItem(key, val, 0, 0)
}

func (m *Memory) SetByEmpty(key, val any) bool {
	m.Lock()
	defer m.Unlock()
	_, ok := m.items[key]
	m.setItem(key, val, 0, m.expire)
	return ok
}
func (m *Memory) SetByExpire(key, val any, expire time.Duration) bool {
	m.Lock()
	defer m.Unlock()
	_, ok := m.items[key]
	m.setItem(key, val, 0, expire)
	return ok
}
func (m *Memory) SetByExpireCallback(key, val any, expire time.Duration, f func(any)) bool {
	m.Lock()
	defer m.Unlock()
	item, ok := m.items[key]
	if ok {
		item.expiration = NowMicr().Add(expire).UnixMicro()
		item.value = val
		item.f = f
	} else {
		item = m.setItem(key, val, 0, expire)
		item.f = f
	}
	return ok
}
func (m *Memory) Del(key any) any {
	m.Lock()
	defer m.Unlock()
	return m.del(key)
}

func (m *Memory) Increase(key any) error {
	return m.calculate(key, 1)
}

func (m *Memory) Decrease(key any) error {
	return m.calculate(key, -1)
}

func (m *Memory) calculate(key any, num int) error {
	m.Lock()
	defer m.Unlock()
	if item, ok := m.items[key]; ok {
		switch n := item.value.(type) {
		case int:
			item.value = n + num
		case int8:
			item.value = n + int8(num)
		case int16:
			item.value = n + int16(num)
		case int32:
			item.value = n + int32(num)
		case int64:
			item.value = n + int64(num)
		default:
			return fmt.Errorf("value of %s type not int", key)
		}
		if item.sliding > 0 {
			item.expiration = NowMicr().Add(item.sliding).UnixMicro()
		}
		return nil
	}
	return fmt.Errorf("key not found")
}

func (m *Memory) Count(callback func(any, any) bool) int {
	m.Lock()
	defer m.Unlock()
	count := 0
	for k, v := range m.items {
		if callback != nil && callback(k, v.value) {
			count++
		}
	}
	return count
}

func (m *Memory) Keys() []any {
	m.Lock()
	defer m.Unlock()
	if count := len(m.items); count > 0 {
		keys := make([]any, count)
		i := 0
		for k := range m.items {
			keys[i] = k
			i++
		}
		return keys
	}
	return []any{}
}

func (m *Memory) List() []any {
	m.Lock()
	defer m.Unlock()
	var val []any
	fmt.Printf("memory.list:%d,%p", len(m.items), m)
	if count := len(m.items); count > 0 {
		val = make([]any, count)
		i := 0
		for _, v := range m.items {
			val[i] = v.value
			i++
		}
	}
	return val
}

func (m *Memory) Range(callback func(any, any) bool) {
	m.Lock()
	defer m.Unlock()
	for k, v := range m.items {
		if callback != nil && !callback(k, v.value) {
			return
		}
	}
}
