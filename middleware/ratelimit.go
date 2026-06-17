package middleware

import (
	"sync"
	"time"

	"personal-assistant-server/model/common/response"

	"github.com/gin-gonic/gin"
)

type rateLimitEntry struct {
	count       int
	windowStart time.Time
}

// RateLimiter 简单的内存滑动窗口速率限制
type RateLimiter struct {
	mu       sync.Mutex
	entries  map[string]*rateLimitEntry
	maxReqs  int
	interval time.Duration
}

// NewRateLimiter creates a new rate limiter. maxReqs is the maximum number
// of requests allowed within the interval duration (per client IP).
func NewRateLimiter(maxReqs int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries:  make(map[string]*rateLimitEntry),
		maxReqs:  maxReqs,
		interval: interval,
	}
	// 定期清理过期条目
	go func() {
		for {
			time.Sleep(interval)
			rl.mu.Lock()
			now := time.Now()
			for k, v := range rl.entries {
				if now.Sub(v.windowStart) > interval {
					delete(rl.entries, k)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

// RateLimit 返回限速中间件
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()

		rl.mu.Lock()
		now := time.Now()
		entry, exists := rl.entries[key]
		if !exists || now.Sub(entry.windowStart) > rl.interval {
			rl.entries[key] = &rateLimitEntry{count: 1, windowStart: now}
			rl.mu.Unlock()
			c.Next()
			return
		}

		entry.count++
		if entry.count > rl.maxReqs {
			rl.mu.Unlock()
			response.Result(7, nil, "请求过于频繁,请稍后重试", c)
			c.Abort()
			return
		}
		rl.mu.Unlock()
		c.Next()
	}
}
