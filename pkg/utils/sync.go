package utils

import (
	"slices"
	"sync"
)

type SyncMap[K comparable, V any] struct {
	m sync.Map
}

// Has 检查映射是否包含指定的键。
func (m *SyncMap[K, V]) Has(key K) bool {
	_, ok := m.m.Load(key)
	return ok
}

// Get 获取与指定键关联的值。
func (m *SyncMap[K, V]) Get(key K) V {
	val, _ := m.m.Load(key)
	return val.(V)
}

// Set 设置指定键的值。
func (m *SyncMap[K, V]) Set(key K, value V) {
	m.m.Store(key, value)
}

// Delete 从映射中删除指定的键。
func (m *SyncMap[K, V]) Delete(key K) {
	m.m.Delete(key)
}

// Iterator 遍历映射并为每个键值对调用提供的函数。
// 如果函数返回 false，则停止迭代。
func (m *SyncMap[K, V]) Iterator(f func(K, V) bool) {
	m.m.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

func (m *SyncMap[K, V]) Map() map[K]V {
	ret := make(map[K]V)

	m.Iterator(func(k K, v V) bool {
		ret[k] = v
		return true
	})

	return ret
}

type SyncSlice[T any] struct {
	val []T
	rwm sync.RWMutex
}

func (s *SyncSlice[T]) Append(val ...T) {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	s.val = append(s.val, val...)
}

func (s *SyncSlice[T]) Len() int {
	s.rwm.RLock()
	defer s.rwm.RUnlock()

	return len(s.val)
}

func (s *SyncSlice[T]) Get(idx int) T {
	s.rwm.RLock()
	defer s.rwm.RUnlock()

	return s.val[idx]
}

func (s *SyncSlice[T]) Set(idx int, val T) {
	s.rwm.Lock()
	defer s.rwm.Unlock()

	s.val[idx] = val
}

func (s *SyncSlice[T]) Raw() []T {
	return slices.Clone(s.val)
}

func NewSyncSlice[T any](s []T) SyncSlice[T] {
	return SyncSlice[T]{
		val: slices.Clone(s),
	}
}
