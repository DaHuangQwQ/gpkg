package ratelimit

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"testing"
	"time"
)

func ratelimit(t *testing.T, limiter Limiter) {
	var eg errgroup.Group
	for i := 0; i < 100; i++ {
		eg.Go(func() error {
			limit, err := limiter.Limit(context.Background(), "test")
			if err != nil {
				return err
			}
			if limit {
				return errors.New("limited")
			}
			return nil
		})
	}
	err := eg.Wait()
	require.Error(t, err)
}

func TestTokenBucketLimiter(t *testing.T) {
	limiter := NewTokenBucketLimiter(time.Second, 1000)
	ratelimit(t, limiter)
}

func TestFixWindowLimiter(t *testing.T) {
	limiter := NewFixWindowLimiter(time.Second, 10)
	ratelimit(t, limiter)
}

func TestSlidingWindowLimiter(t *testing.T) {
	limiter := NewSlidingWindowLimiter(time.Second, 10)
	ratelimit(t, limiter)
}
