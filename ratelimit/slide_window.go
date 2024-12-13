package ratelimit

import (
	"container/list"
	"context"
	_ "embed"
	"sync"
	"time"
)

type slidingWindowLimiter struct {
	// 窗口大小
	interval time.Duration
	rate     int64

	queue *list.List

	mutex sync.Mutex
}

func NewSlidingWindowLimiter(interval time.Duration, rate int64) Limiter {
	return &slidingWindowLimiter{
		interval: interval,
		rate:     rate,
		queue:    list.New(),
		mutex:    sync.Mutex{},
	}
}

func (s *slidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixNano()
	// 快路径
	if int64(s.queue.Len()) < s.rate {
		s.queue.PushBack(now)
		return false, nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	boundary := now - int64(s.interval)
	timestamp := s.queue.Front()
	for timestamp != nil && timestamp.Value.(int64) <= boundary {
		s.queue.Remove(timestamp)
		timestamp = s.queue.Front()
	}

	if int64(s.queue.Len()) > s.rate {
		return true, ErrLimitExceeded
	}

	s.queue.PushBack(now)
	return false, nil
}
