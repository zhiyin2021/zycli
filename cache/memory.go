package cache

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type item struct {
	Value   any
	Expired int64
	Sliding time.Duration
}

// NewMemory memory模式
func NewMemory() *Memory {
	return &Memory{
		items: new(sync.Map),
	}
}

type Memory struct {
	items *sync.Map
	mutex sync.RWMutex
}

func (*Memory) String() string {
	return "memory"
}

func (m *Memory) Get(key string) any {
	item := m.getItem(key)
	if item == nil {
		return nil
	}
	if item.Sliding > 0 {
		now := time.Now().UnixMilli()
		if (item.Expired - now) > int64(item.Sliding/2) {
			m.mutex.RLock()
			defer m.mutex.RUnlock()
			item.Expired = now + int64(item.Sliding)
			m.items.Store(key, item)
		}
	}
	return item.Value
}

func (m *Memory) getItem(key string) *item {
	i, ok := m.items.Load(key)
	if !ok {
		return nil
	}
	item := i.(*item)
	if item.Expired < time.Now().UnixMilli() {
		//过期
		m.items.Delete(key)
		//过期后删除
		return nil
	}
	return item
}

// 获取缓存时自动延时
func (m *Memory) SetBySliding(key string, val any, expire time.Duration) {
	item := &item{
		Value:   val,
		Expired: time.Now().UnixMilli() + expire.Milliseconds(),
		Sliding: expire,
	}
	m.items.Store(key, item)
}

func (m *Memory) Set(key string, val any, expire time.Duration) {
	item := &item{
		Value:   val,
		Expired: time.Now().UnixMilli() + expire.Milliseconds(),
		Sliding: 0,
	}
	m.items.Store(key, item)
}

func (m *Memory) Del(key string) {
	m.items.Delete(key)
}

func (m *Memory) Increase(key string) error {
	return m.calculate(key, 1)
}

func (m *Memory) Decrease(key string) error {
	return m.calculate(key, -1)
}

func (m *Memory) calculate(key string, num int) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	item := m.getItem(key)
	if item == nil {
		return nil
	}
	var n int
	switch item.Value.(type) {
	case int:
		n = item.Value.(int)
	default:
		return fmt.Errorf("value of %s type not int", key)
	}
	n += num
	item.Value = strconv.Itoa(n)
	m.items.Store(key, item)
	return nil
}
