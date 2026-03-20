package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	*redis.Client
	config RedisConfig
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// NewRedis creates a new Redis client
func NewRedis(cfg RedisConfig, log *zerolog.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		// Pool settings
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info().Str("addr", fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)).Msg("Redis connected successfully")

	return &RedisClient{
		Client: client,
		config: cfg,
	}, nil
}

// CacheSet sets a value in the cache with expiration
func (r *RedisClient) CacheSet(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Set(ctx, r.cacheKey(key), value, expiration).Err()
}

// CacheGet gets a value from the cache
func (r *RedisClient) CacheGet(ctx context.Context, key string, dest interface{}) error {
	val, err := r.Get(ctx, r.cacheKey(key)).Result()
	if err != nil {
		return err
	}

	// For simple string values
	if s, ok := dest.(*string); ok {
		*s = val
		return nil
	}

	return nil
}

// CacheDelete deletes a value from the cache
func (r *RedisClient) CacheDelete(ctx context.Context, key string) error {
	return r.Del(ctx, r.cacheKey(key)).Err()
}

// CacheDeleteByPattern deletes all keys matching a pattern
func (r *RedisClient) CacheDeleteByPattern(ctx context.Context, pattern string) error {
	iter := r.Scan(ctx, 0, r.cacheKey(pattern), 0).Iterator()
	for iter.Next(ctx) {
		if err := r.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// SetTenantCache sets a tenant-scoped cache value
func (r *RedisClient) SetTenantCache(ctx context.Context, tenantID, key string, value interface{}, expiration time.Duration) error {
	cacheKey := r.tenantCacheKey(tenantID, key)
	return r.Set(ctx, cacheKey, value, expiration).Err()
}

// GetTenantCache gets a tenant-scoped cache value
func (r *RedisClient) GetTenantCache(ctx context.Context, tenantID, key string) (string, error) {
	cacheKey := r.tenantCacheKey(tenantID, key)
	return r.Get(ctx, cacheKey).Result()
}

// DeleteTenantCache deletes all cache keys for a tenant
func (r *RedisClient) DeleteTenantCache(ctx context.Context, tenantID string) error {
	pattern := r.tenantCacheKey(tenantID, "*")
	return r.CacheDeleteByPattern(ctx, pattern)
}

// TokenBlacklist adds a token to the blacklist
func (r *RedisClient) TokenBlacklist(ctx context.Context, tokenID string, expiration time.Duration) error {
	key := r.tokenBlacklistKey(tokenID)
	return r.Set(ctx, key, "1", expiration).Err()
}

// IsTokenBlacklisted checks if a token is blacklisted
func (r *RedisClient) IsTokenBlacklisted(ctx context.Context, tokenID string) bool {
	key := r.tokenBlacklistKey(tokenID)
	exists, _ := r.Exists(ctx, key).Result()
	return exists > 0
}

// cacheKey creates a cache key with prefix
func (r *RedisClient) cacheKey(key string) string {
	return fmt.Sprintf("cache:%s", key)
}

// tenantCacheKey creates a tenant-scoped cache key
func (r *RedisClient) tenantCacheKey(tenantID, key string) string {
	return fmt.Sprintf("tenant:%s:%s", tenantID, key)
}

// tokenBlacklistKey creates a token blacklist key
func (r *RedisClient) tokenBlacklistKey(tokenID string) string {
	return fmt.Sprintf("token_blacklist:%s", tokenID)
}

// HealthCheck checks the health of Redis
func (r *RedisClient) HealthCheck(ctx context.Context) error {
	return r.Ping(ctx).Err()
}

// Stats returns Redis statistics
func (r *RedisClient) Stats() map[string]interface{} {
	poolStats := r.PoolStats()
	return map[string]interface{}{
		"hits":       poolStats.Hits,
		"misses":     poolStats.Misses,
		"timeouts":   poolStats.Timeouts,
		"total_conns": poolStats.TotalConns,
		"idle_conns":  poolStats.IdleConns,
		"stale_conns": poolStats.StaleConns,
	}
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.Client.Close()
}
