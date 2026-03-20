package middleware

import (
	"fmt"
	"strconv"
	"time"

	"Desktop/multitenant/saas/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

// RateLimiter is a Redis-based rate limiter
type RateLimiter struct {
	redis  *redis.Client
	config RateLimitConfig
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *redis.Client, config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		redis:  redisClient,
		config: config,
	}
}

// RateLimit creates a rate limiting middleware
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get identifier (user ID or IP)
		identifier := rl.getIdentifier(c)

		// Check rate limit
		allowed, remaining, resetAt, err := rl.checkLimit(c, identifier)
		if err != nil {
			c.JSON(500, types.NewErrorResponse(
				types.ErrCodeInternal,
				"Failed to check rate limit",
				nil,
			))
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))

		if !allowed {
			c.JSON(429, types.NewErrorResponse(
				"RATE_LIMIT_EXCEEDED",
				"Too many requests",
				map[string]interface{}{
					"retry_after": time.Until(time.Unix(resetAt, 0)).Seconds(),
				},
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// getIdentifier gets a unique identifier for rate limiting
func (rl *RateLimiter) getIdentifier(c *gin.Context) string {
	// Prefer user ID from context
	if userID, exists := GetUserID(c); exists {
		return fmt.Sprintf("user:%s", userID)
	}

	// Fallback to IP address
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// checkLimit checks if the request is within rate limits
func (rl *RateLimiter) checkLimit(c *gin.Context, identifier string) (bool, int, int64, error) {
	ctx := c.Request.Context()
	key := fmt.Sprintf("ratelimit:%s", identifier)
	now := time.Now().Unix()
	windowStart := now - int64(rl.config.Window.Seconds())

	pipe := rl.redis.Pipeline()

	// Remove old entries outside the current window
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

	// Count current requests
	incrCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

	// Set expiration
	pipe.Expire(ctx, key, rl.config.Window)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return false, 0, 0, err
	}

	currentCount := incrCmd.Val()
	remaining := rl.config.Requests - int(currentCount)
	resetAt := now + int64(rl.config.Window.Seconds())

	if currentCount > int64(rl.config.Requests) {
		return false, 0, resetAt, nil
	}

	return true, remaining, resetAt, nil
}

// SlidingWindowRateLimit creates a sliding window rate limiter middleware
func SlidingWindowRateLimit(redisClient *redis.Client, requests int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(redisClient, RateLimitConfig{
		Requests: requests,
		Window:   window,
	})
	return limiter.RateLimit()
}
