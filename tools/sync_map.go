package tools

import "sync"

type SyncMap[T any] struct {
	pool sync.Map
}

func NewSyncMap[T any]() *SyncMap[T] {
	return &SyncMap[T]{}
}
func (s *SyncMap[T]) Get(key string) *T {

	if c, ok := s.pool.Load(key); ok {
		return c.(*T)
	}
	return nil
}

func (s *SyncMap[T]) Set(key string, val *T) {
	s.pool.Store(key, val)
}

func (s *SyncMap[T]) Del(key string, callback func(*T)) {
	if callback != nil {
		if c, ok := s.pool.Load(key); ok {
			callback(c.(*T))
		}
	}
	s.pool.Delete(key)
}
