package ratelimit

import (
	"context"
	"errors"
	"time"
)

type tokenBucketLimiter struct {
	tokens  chan struct{}
	closeCh chan struct{}
}

// NewTokenBucketLimiter interval 多久产生一个令牌, capacity 令牌数最大限度
func NewTokenBucketLimiter(interval time.Duration, capacity int) Limiter {
	ch := make(chan struct{}, capacity)
	ticker := time.NewTicker(interval)
	closeCh := make(chan struct{})

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// 令牌满了 丢掉
				select {
				case ch <- struct{}{}:
				default:
				}
			case <-closeCh:
				return
			}
		}
	}()

	return &tokenBucketLimiter{
		tokens:  ch,
		closeCh: closeCh,
	}
}

func (t *tokenBucketLimiter) Limit(ctx context.Context, key string) (bool, error) {
	select {
	case <-ctx.Done():
		return true, ctx.Err()
	case <-t.closeCh:
		return true, errors.New("close channel closed")
	case <-t.tokens:
		return false, nil
	default:
		return true, ErrLimitExceeded
	}
}

func (t *tokenBucketLimiter) Close() error {
	close(t.closeCh)
	return nil
}
