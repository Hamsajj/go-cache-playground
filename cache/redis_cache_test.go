package cache

import (
	"cache-api/config"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	redisConteiner "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestRedisCache_Set(t *testing.T) {
	connectionString := setupRedis(t)
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: connectionString,
	})
	redisCfg := &config.RedisConfig{
		Host: connectionString,
	}
	cacheCfg := &config.CacheConfig{
		TTLSec: 0,
	}

	logger := zerolog.Nop()
	t.Run("No expiry", func(t *testing.T) {
		redisCache := NewRedisCache(ctx, cacheCfg, redisCfg, &logger)

		key := "key"
		expected := "expected"

		require.NoError(t, redisCache.Set(key, expected))
		got, ok := rdb.Get(ctx, key).Result()
		require.NoError(t, ok)
		require.Equal(t, expected, got)

		// re-write
		expected = "new"
		require.NoError(t, redisCache.Set(key, expected))
		got, err := rdb.Get(ctx, key).Result()
		require.NoError(t, err)
	})

	t.Run("with expiry", func(t *testing.T) {
		redisCacheWithTTL := NewRedisCache(ctx, &config.CacheConfig{TTLSec: 1}, redisCfg, &logger)
		key := "keyWillExpire"
		value := "toExpire"
		require.NoError(t, redisCacheWithTTL.Set(key, value))
		time.Sleep(1 * time.Second)
		_, err := rdb.Get(ctx, key).Result()
		require.ErrorIs(t, err, redis.Nil)
	})
}

func TestRedisCache_Get(t *testing.T) {
	connectionString := setupRedis(t)
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: connectionString,
	})
	redisCfg := &config.RedisConfig{
		Host: connectionString,
	}
	cacheCfg := &config.CacheConfig{
		TTLSec: 0,
	}

	logger := zerolog.Nop()
	cache := NewRedisCache(ctx, cacheCfg, redisCfg, &logger)

	require.NoError(t, rdb.Set(ctx, "key", "value", 0).Err())
	t.Run("key exists", func(t *testing.T) {
		value, ok := cache.Get("key")
		require.True(t, ok)
		require.Equal(t, "value", value)
	})

	t.Run("key does not exist", func(t *testing.T) {
		value, ok := cache.Get("nonExisting")
		require.False(t, ok)
		require.Empty(t, value)
	})
}

func setupRedis(t *testing.T) string {
	ctx := context.Background()

	redisContainer, err := redisConteiner.Run(ctx,
		"docker.io/redis:7",
	)
	if err != nil {
		t.Fatalf("Could not start redis container: %s", err)
	}

	host, err := redisContainer.Host(ctx)
	port, err := redisContainer.MappedPort(ctx, "6379/tcp")
	if err != nil {
		t.Fatalf("Could not get connection string: %s", err)
	}
	return fmt.Sprintf("%s:%d", host, port.Int())
}
