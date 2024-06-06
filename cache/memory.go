package cache

import (
	"fmt"
	"sync"
	"sync/atomic"
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
		Count:  0,
	}
}

type Memory struct {
	items  sync.Map
	mutex  sync.RWMutex
	expire time.Duration
	Count  int32
}

func (*Memory) String() string {
	return "memory"
}

func (m *Memory) Get(key any) any {
	item := m.getItem(key)
	if item == nil {
		return nil
	}
	if item.sliding > 0 {
		item.timer.Reset(item.sliding)
	}
	return item.value
}
func (m *Memory) GetOrStoreBySliding(key any, val any, expire time.Duration) (any, bool) {
	t := &item{
		value:   val,
		sliding: expire,
	}
	t1, flag := m.items.LoadOrStore(key, t)
	if !flag {
		t.timer = m.afterDel(expire, key)
	} else {
		t1.(*item).timer.Reset(expire)
	}
	return t1.(*item).value, flag
}

func (m *Memory) GetOrStore(key any, val any, expire time.Duration) (any, bool) {
	t := &item{
		value:   val,
		sliding: 0,
	}
	t1, flag := m.items.LoadOrStore(key, t)
	if !flag {
		t.timer = m.afterDel(m.expire, key)
	}
	return t1.(*item).value, flag
}

func (m *Memory) GetAndDel(key any) any {
	if t, ok := m.items.LoadAndDelete(key); ok {
		t.(*item).timer.Stop()
		return t.(*item).value
	}
	return nil
}

func (m *Memory) getItem(key any) *item {
	i, ok := m.items.Load(key)
	if !ok {
		return nil
	}
	item := i.(*item)
	return item
}

// 获取缓存时自动延时
func (m *Memory) SetBySliding(key, val any, expire time.Duration) {
	m.Del(key)
	item := &item{
		value:   val,
		sliding: expire,
		timer:   m.afterDel(expire, key),
	}
	atomic.AddInt32(&m.Count, 1)
	m.items.Store(key, item)
}

func (m *Memory) Set(key, val any) {
	m.Del(key)
	item := &item{
		value:   val,
		sliding: 0,
		timer:   m.afterDel(m.expire, key),
	}
	atomic.AddInt32(&m.Count, 1)
	m.items.Store(key, item)
}

func (m *Memory) SetByEmpty(key, val any) bool {
	t := &item{
		value:   val,
		sliding: 0,
		timer:   m.afterDel(m.expire, key),
	}
	_, flag := m.items.LoadOrStore(key, t)
	if flag {
		t.timer.Stop()
	} else {
		atomic.AddInt32(&m.Count, 1)
	}
	return flag
}
func (m *Memory) SetByEmptyExpire(key, val any, expire time.Duration) bool {
	t := &item{
		value:   val,
		sliding: 0,
		timer:   m.afterDel(expire, key),
	}
	_, flag := m.items.LoadOrStore(key, t)
	if flag {
		t.timer.Stop()
	} else {
		atomic.AddInt32(&m.Count, 1)
	}
	return flag
}
func (m *Memory) SetByExpire(key, val any, expire time.Duration) {
	item := &item{
		value:   val,
		sliding: 0,
		timer:   m.afterDel(expire, key),
	}
	atomic.AddInt32(&m.Count, 1)
	m.items.Store(key, item)
}
func (m *Memory) afterDel(expire time.Duration, key any) *time.Timer {
	return time.AfterFunc(expire, func() {
		logrus.Debugln("delete", key)
		atomic.AddInt32(&m.Count, -1)
		m.items.Delete(key)
	})
}
func (m *Memory) Del(key any) {
	if t, ok := m.items.LoadAndDelete(key); ok {
		t.(*item).timer.Stop()
		atomic.AddInt32(&m.Count, -1)
	}
}

func (m *Memory) Increase(key any) error {
	return m.calculate(key, 1)
}

func (m *Memory) Decrease(key any) error {
	return m.calculate(key, -1)
}

func (m *Memory) calculate(key any, num int) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	item := m.getItem(key)
	if item == nil {
		return nil
	}
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
		item.timer.Reset(item.sliding)
	}
	m.items.Store(key, item)
	return nil
}
