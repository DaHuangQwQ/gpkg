package ratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// fixWindowLimiter 固定窗口算法
type fixWindowLimiter struct {
	// 起始窗口时间
	timestamp int64
	// 窗口大小
	interval time.Duration
	// 最大请求数
	rate int64

	cnt int64

	mutex sync.Mutex
}

func NewFixWindowLimiter(interval time.Duration, rate int64) Limiter {
	return &fixWindowLimiter{
		timestamp: time.Now().UnixNano(),
		interval:  interval,
		rate:      rate,
		cnt:       0,
		mutex:     sync.Mutex{},
	}
}

func (f *fixWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	timestamp := atomic.LoadInt64(&f.timestamp)
	cnt := atomic.LoadInt64(&f.cnt)
	cur := time.Now().UnixNano()

	if timestamp+int64(f.interval) < cur {
		if atomic.CompareAndSwapInt64(&f.timestamp, timestamp, cur) {
			atomic.CompareAndSwapInt64(&f.cnt, cnt, 0)
		}
	}
	cnt = atomic.AddInt64(&f.cnt, 1)
	if cnt > f.rate {
		return true, ErrLimitExceeded
	}
	return false, nil
}
