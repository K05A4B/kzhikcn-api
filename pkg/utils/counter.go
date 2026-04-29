package utils

import (
	"context"
	"runtime"
	"sync"
	"time"
)

var (
	counterRegistry    = map[string]*Counter{}
	counterRegistryMux = sync.Mutex{}

	expireOldCountersSleepDuration = time.Minute
)

func autoZeroedCounter(ctx context.Context, interval time.Duration, counter *Counter) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			counter.Reset()
		}
	}
}

func NewAutoZeroedCounter(interval time.Duration, ctx context.Context) *Counter {
	if ctx == nil {
		ctx = context.Background()
	}

	counter := &Counter{
		maxAge: -1,
	}
	go autoZeroedCounter(ctx, interval, counter)
	return counter
}

type Counter struct {
	updateAt time.Time
	maxAge   time.Duration

	mux   sync.RWMutex
	count int
}

func init() {
	go expireOldCounters()
}

func (c *Counter) update() {
	c.updateAt = time.Now()
}

func (c *Counter) Get() int {
	c.mux.RLock()
	defer c.mux.RUnlock()

	c.update()
	return c.count
}

func (c *Counter) Add() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.update()

	c.count++
}

func (c *Counter) Reset() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.update()

	c.count = 0
}

// 获取计数器
// maxAge表示这个计数器多久不用会被回收（如果使用了则未来的maxAge范围内都不会被回收）
func GetCounter(counterId string, maxAge time.Duration) *Counter {
	_, file, _, _ := runtime.Caller(1)

	counterId = file + ":" + counterId
	counterRegistryMux.Lock()
	defer counterRegistryMux.Unlock()

	_, ok := counterRegistry[counterId]
	if !ok {
		counterRegistry[counterId] = &Counter{
			count:  0,
			maxAge: maxAge,
		}
	}

	return counterRegistry[counterId]
}

func DestroyCounter(counterId string) {
	_, file, _, _ := runtime.Caller(1)
	counterRegistryMux.Lock()
	defer counterRegistryMux.Unlock()

	delete(counterRegistry, file+":"+counterId)
}

func expireOldCounters() {
	for {

		for id, counter := range counterRegistry {
			if counter.maxAge < 0 {
				continue
			}

			if !counter.updateAt.Add(counter.maxAge).After(time.Now()) {
				continue
			}

			DestroyCounter(id)
		}

		time.Sleep(expireOldCountersSleepDuration)
	}
}
