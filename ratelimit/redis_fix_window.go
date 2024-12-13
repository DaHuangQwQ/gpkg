package ratelimit

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed lua/fix_window.lua
var luaFixScript string

type redisFixWindowLimiter struct {
	client   redis.Cmdable
	interval time.Duration
	// 阈值
	rate int
}

func newRedisFixWindowLimiter(client redis.Cmdable, interval time.Duration, rate int) Limiter {
	return &redisFixWindowLimiter{
		client:   client,
		interval: interval,
		rate:     rate,
	}
}

func (r *redisFixWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.client.Eval(ctx, luaFixScript, []string{key},
		r.interval.Milliseconds(), r.rate).Bool()
}
