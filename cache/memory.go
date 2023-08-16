package cache

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type item struct {
	value   any
	sliding time.Duration
	timer   *time.Timer
}

// NewMemory memory模式
func NewMemory(expire time.Duration) *Memory {
	if expire == 0 {
		expire = time.Hour
	}
	return &Memory{
		items:  sync.Map{},
		expire: expire,
	}
}

type Memory struct {
	items  sync.Map
	mutex  sync.RWMutex
	expire time.Duration
}

func (*Memory) String() string {
	return "memory"
}

func (m *Memory) Get(key string) any {
	item := m.getItem(key)
	if item == nil {
		return nil
	}
	if item.sliding > 0 {
		item.timer.Reset(item.sliding)
		m.items.Store(key, item)
	}
	return item.value
}

func (m *Memory) getItem(key string) *item {
	i, ok := m.items.Load(key)
	if !ok {
		return nil
	}
	item := i.(*item)
	return item
}

// 获取缓存时自动延时
func (m *Memory) SetBySliding(key string, val any, expire time.Duration) {
	item := &item{
		value:   val,
		sliding: expire,
		timer:   m.afterDel(expire, key),
	}
	m.items.Store(key, item)
}

func (m *Memory) Set(key string, val any) {
	item := &item{
		value:   val,
		sliding: 0,
		timer:   m.afterDel(m.expire, key),
	}
	m.items.Store(key, item)
}
func (m *Memory) SetByExpire(key string, val any, expire time.Duration) {
	item := &item{
		value:   val,
		sliding: 0,
		timer:   m.afterDel(expire, key),
	}
	m.items.Store(key, item)
}
func (m *Memory) afterDel(expire time.Duration, key string) *time.Timer {
	return time.AfterFunc(expire, func() {
		logrus.Println("delete", key)
		m.items.Delete(key)
	})
}
func (m *Memory) Del(key string) {
	if t, ok := m.items.LoadAndDelete(key); ok {
		t.(*item).timer.Stop()
	}
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
	switch item.value.(type) {
	case int:
		n = item.value.(int)
	default:
		return fmt.Errorf("value of %s type not int", key)
	}
	n += num
	item.value = strconv.Itoa(n)
	if item.sliding > 0 {
		item.timer.Reset(item.sliding)
	}
	m.items.Store(key, item)
	return nil
}
