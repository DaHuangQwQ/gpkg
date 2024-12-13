package ratelimit

import (
	"context"
	"time"
)

// leakyBucketLimiter 漏桶
type leakyBucketLimiter struct {
	ticker *time.Ticker
}

func NewLeakyBucketLimiter(ticker *time.Ticker) Limiter {
	return &leakyBucketLimiter{
		ticker: ticker,
	}
}

func (l *leakyBucketLimiter) Limit(ctx context.Context, key string) (bool, error) {
	select {
	case <-ctx.Done():
		return true, ctx.Err()
	case <-l.ticker.C:
		return false, nil
	}
}

func (l *leakyBucketLimiter) Close() error {
	l.ticker.Stop()
	return nil
}
