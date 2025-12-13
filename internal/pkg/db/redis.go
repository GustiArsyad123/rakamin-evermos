package db

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

// NewRedis creates a new Redis client with connection pooling
func NewRedis() (*redis.Client, error) {
	host := getenv("REDIS_HOST", "localhost")
	port := getenv("REDIS_PORT", "6379")
	password := os.Getenv("REDIS_PASSWORD") // Optional

	rdb := redis.NewClient(&redis.Options{
		Addr:            host + ":" + port,
		Password:        password,
		DB:              0,                // use default DB
		PoolSize:        10,               // connection pool size
		MinIdleConns:    5,                // minimum idle connections
		ConnMaxLifetime: 30 * time.Minute, // maximum connection age
	})

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}

// RedisCache provides caching operations
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// Set stores a key-value pair with expiration
func (c *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (c *RedisCache) Get(key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Delete removes a key
func (c *RedisCache) Delete(key string) error {
	return c.client.Del(ctx, key).Err()
}

// Exists checks if a key exists
func (c *RedisCache) Exists(key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	return count > 0, err
}

// SetJSON stores a JSON-serializable object
func (c *RedisCache) SetJSON(key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// GetJSON retrieves and unmarshals a JSON object
func (c *RedisCache) GetJSON(key string, dest interface{}) error {
	return c.client.Get(ctx, key).Scan(dest)
}

// FlushAll clears all cache
func (c *RedisCache) FlushAll() error {
	return c.client.FlushAll(ctx).Err()
}
