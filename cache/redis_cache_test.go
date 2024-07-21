package cache

import (
	"cache-api/config"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	redisConteiner "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestNewRedisCache(t *testing.T) {
	t.Run("Invalid redis config", func(t *testing.T) {
		_, err := NewRedisCache(context.Background(), &config.CacheConfig{}, &config.RedisConfig{
			Host: "invalid:1234",
		}, &zerolog.Logger{})
		if err == nil {
			t.Errorf("Expected an error but got nil")
		}
	})

	t.Run("Valid redis config", func(t *testing.T) {
		connectionString := setupRedis(t)
		redisCfg := &config.RedisConfig{
			Host: connectionString,
		}
		cacheCfg := &config.CacheConfig{
			TTLSec: 0,
		}
		logger := zerolog.Nop()
		_, err := NewRedisCache(context.Background(), cacheCfg, redisCfg, &logger)
		if err != nil {
			t.Error(err)
		}
	})
}

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
		redisCache, err := NewRedisCache(ctx, cacheCfg, redisCfg, &logger)
		if err != nil {
			t.Fatal(err)
		}
		if err != nil {
			t.Fatal(err)
		}
		key := "key"
		expected := "expected"

		err = redisCache.Set(key, expected)
		if err != nil {
			t.Fatal(err)
		}
		got, err := rdb.Get(ctx, key).Result()
		if err != nil {
			t.Fatal(err)
		}
		if got != expected {
			t.Errorf("Expected %s but got %s", expected, got)
		}

		// re-write
		expected = "new"
		err = redisCache.Set(key, expected)
		if err != nil {
			t.Fatal(err)
		}
		got, err = rdb.Get(ctx, key).Result()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("with expiry", func(t *testing.T) {
		redisCacheWithTTL, err := NewRedisCache(ctx, &config.CacheConfig{TTLSec: 1}, redisCfg, &logger)
		if err != nil {
			t.Fatal(err)
		}
		key := "keyWillExpire"
		value := "toExpire"
		err = redisCacheWithTTL.Set(key, value)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(1 * time.Second)
		_, err = rdb.Get(ctx, key).Result()
		if !errors.Is(err, redis.Nil) {
			t.Errorf("Expected key to be expired but it was not")
		}
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
	cache, err := NewRedisCache(ctx, cacheCfg, redisCfg, &logger)
	if err != nil {
		t.Fatal(err)
	}
	err = rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("key exists", func(t *testing.T) {
		value, ok := cache.Get("key")
		if ok != true {
			t.Errorf("Expected key to exist but it did not")
		}
		if value != "value" {
			t.Errorf("Expected value to be 'value' but got %s", value)
		}
	})

	t.Run("key does not exist", func(t *testing.T) {
		_, ok := cache.Get("nonExisting")
		if ok != false {
			t.Errorf("Expected key to not exist but it did")
		}
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
