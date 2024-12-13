package ratelimit

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed lua/slide_window.lua
var luaSlideScript string

type redisSlidingWindowLimiter struct {
	client   redis.Cmdable
	interval time.Duration
	// 阈值
	rate int
}

func NewRedisSlidingWindowLimiter(client redis.Cmdable, interval time.Duration, rate int) Limiter {
	return &redisSlidingWindowLimiter{
		client:   client,
		interval: interval,
		rate:     rate,
	}
}

func (b *redisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return b.client.Eval(ctx, luaSlideScript, []string{key},
		b.interval.Milliseconds(), b.rate, time.Now().UnixMilli()).Bool()
}
