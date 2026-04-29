package utils

import (
	"container/heap"
	"sync"
	"time"
)

type entry[K comparable, V any] struct {
	key    K
	value  V
	expire time.Time
	index  int // 在 heap 中的位置
}

type entryHeap[K comparable, V any] []*entry[K, V]

func (h entryHeap[K, V]) Len() int { return len(h) }

func (h entryHeap[K, V]) Less(i, j int) bool {
	return h[i].expire.Before(h[j].expire)
}

func (h entryHeap[K, V]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *entryHeap[K, V]) Push(x any) {
	e := x.(*entry[K, V])
	e.index = len(*h)
	*h = append(*h, e)
}

func (h *entryHeap[K, V]) Pop() any {
	old := *h
	n := len(old)
	e := old[n-1]
	e.index = -1
	*h = old[:n-1]
	return e
}

type TTLMap[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]*entry[K, V]
	heap  entryHeap[K, V]

	wakeup chan struct{}
	stop   chan struct{}
}

func NewTTLMap[K comparable, V any]() *TTLMap[K, V] {
	m := &TTLMap[K, V]{
		items:  make(map[K]*entry[K, V]),
		wakeup: make(chan struct{}, 1),
		stop:   make(chan struct{}),
	}
	go m.run()
	return m
}

func (m *TTLMap[K, V]) Set(key K, value V, ttl time.Duration) {
	expire := time.Now().Add(ttl)

	m.mu.Lock()
	defer m.mu.Unlock()

	if e, ok := m.items[key]; ok {
		e.value = value
		e.expire = expire
		heap.Fix(&m.heap, e.index)
	} else {
		e := &entry[K, V]{
			key:    key,
			value:  value,
			expire: expire,
		}
		m.items[key] = e
		heap.Push(&m.heap, e)
	}

	m.notify()
}

func (m *TTLMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	e, ok := m.items[key]
	m.mu.RUnlock()

	if !ok {
		var zero V
		return zero, false
	}

	if time.Now().After(e.expire) {
		m.Delete(key)
		var zero V
		return zero, false
	}

	return e.value, true
}

func (m *TTLMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if e, ok := m.items[key]; ok {
		heap.Remove(&m.heap, e.index)
		delete(m.items, key)
	}
}

func (m *TTLMap[K, V]) Range(f func(K, V) bool) {
	now := time.Now()

	m.mu.RLock()
	defer m.mu.RUnlock()

	for k, e := range m.items {
		if now.After(e.expire) {
			continue
		}
		if !f(k, e.value) {
			return
		}
	}
}

func (m *TTLMap[K, V]) Close() {
	close(m.stop)
}

func (m *TTLMap[K, V]) notify() {
	select {
	case m.wakeup <- struct{}{}:
	default:
	}
}

func (m *TTLMap[K, V]) run() {
	for {
		m.mu.Lock()

		if len(m.heap) == 0 {
			m.mu.Unlock()
			select {
			case <-m.wakeup:
				continue
			case <-m.stop:
				return
			}
		}

		next := m.heap[0]
		now := time.Now()
		wait := next.expire.Sub(now)

		if wait <= 0 {
			heap.Pop(&m.heap)
			delete(m.items, next.key)
			m.mu.Unlock()
			continue
		}

		m.mu.Unlock()

		timer := time.NewTimer(wait)

		select {
		case <-timer.C:
		case <-m.wakeup:
			timer.Stop()
		case <-m.stop:
			timer.Stop()
			return
		}
	}
}
