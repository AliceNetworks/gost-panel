package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 基于内存的限流器
type RateLimiter struct {
	mu       sync.RWMutex
	attempts map[string]*attemptInfo
	// 配置
	maxAttempts int           // 最大尝试次数
	window      time.Duration // 时间窗口
	blockTime   time.Duration // 封锁时间
}

type attemptInfo struct {
	count     int
	firstTime time.Time
	blocked   bool
	blockTime time.Time
}

// NewRateLimiter 创建限流器
func NewRateLimiter(maxAttempts int, window, blockTime time.Duration) *RateLimiter {
	rl := &RateLimiter{
		attempts:    make(map[string]*attemptInfo),
		maxAttempts: maxAttempts,
		window:      window,
		blockTime:   blockTime,
	}
	// 启动清理协程
	go rl.cleanup()
	return rl
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	info, exists := rl.attempts[key]

	if !exists {
		rl.attempts[key] = &attemptInfo{
			count:     1,
			firstTime: now,
		}
		return true
	}

	// 检查是否被封锁
	if info.blocked {
		if now.Sub(info.blockTime) > rl.blockTime {
			// 解除封锁
			info.blocked = false
			info.count = 1
			info.firstTime = now
			return true
		}
		return false
	}

	// 检查时间窗口
	if now.Sub(info.firstTime) > rl.window {
		// 重置计数
		info.count = 1
		info.firstTime = now
		return true
	}

	// 增加计数
	info.count++
	if info.count > rl.maxAttempts {
		info.blocked = true
		info.blockTime = now
		return false
	}

	return true
}

// Reset 重置某个 key 的限流状态（登录成功后调用）
func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, key)
}

// GetBlockTimeRemaining 获取剩余封锁时间
func (rl *RateLimiter) GetBlockTimeRemaining(key string) time.Duration {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	info, exists := rl.attempts[key]
	if !exists || !info.blocked {
		return 0
	}

	remaining := rl.blockTime - time.Since(info.blockTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// cleanup 定期清理过期记录
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, info := range rl.attempts {
			// 删除超过封锁时间的记录
			if info.blocked && now.Sub(info.blockTime) > rl.blockTime {
				delete(rl.attempts, key)
			}
			// 删除超过时间窗口且未被封锁的记录
			if !info.blocked && now.Sub(info.firstTime) > rl.window*2 {
				delete(rl.attempts, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用客户端 IP 作为限流 key
		key := c.ClientIP()

		if !rl.Allow(key) {
			remaining := rl.GetBlockTimeRemaining(key)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests, please try again later",
				"retry_after": int(remaining.Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
