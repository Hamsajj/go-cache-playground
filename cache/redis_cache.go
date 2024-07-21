package cache

import (
	"cache-api/config"
	"cache-api/server"
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var _ server.Cache = &RedisCache{}

type RedisCache struct {
	ctx    context.Context
	rdb    *redis.Client
	logger *zerolog.Logger

	// The time to live for each item in the cache - 0 means no expiration
	ttl time.Duration
}

func NewRedisCache(
	ctx context.Context,
	cacheConfig *config.CacheConfig,
	redisConfig *config.RedisConfig,
	logger *zerolog.Logger,
) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Host,
		Username: redisConfig.Username,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	return &RedisCache{
		ctx:    ctx,
		logger: logger,
		rdb:    client,
		ttl:    time.Duration(cacheConfig.TTLSec) * time.Second}
}

func (r RedisCache) Set(key string, value string) error {
	if err := r.rdb.Set(r.ctx, key, value, r.ttl).Err(); err != nil {
		return err
	}
	return nil
}

func (r RedisCache) Get(key string) (string, bool) {
	val, err := r.rdb.Get(r.ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false
	} else if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get value from redis cache")
		return "", false
	}
	return val, true
}
