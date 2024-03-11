package cache

import (
	"context"
	"embarkCache/config"
	"sync"
	"time"
)

type Cache[T any] struct {
	ctx context.Context
	// The cache is a map of strings to strings
	items map[string]cacheItem[T]
	// The time to live for each item in the cache
	ttl time.Duration
	// RW mutex to protect the cache
	mutex *sync.RWMutex
	// StopEviction is a channel to stop the eviction process
	stopEviction chan bool
	// isEvictionRunning is a flag to indicate whether the eviction process is running
	isEvictionRunning bool
	// evictionInterval is the interval at which the cache is checked for expired items
	evictionInterval time.Duration
}

type cacheItem[T any] struct {
	value     T
	expiresAt int64
}

const (
	defaultEvictionInterval = time.Second
	defaultTTL              = 30 * time.Minute
)

// NewCache creates a new cache with the given time to live
func NewCache[T any](ctx context.Context, conf config.CacheConfig) *Cache[T] {
	stopChan := make(chan bool)
	evictionInterval := defaultEvictionInterval
	if conf.EvictionIntervalMilliSec != 0 {
		evictionInterval = time.Duration(conf.EvictionIntervalMilliSec) * time.Millisecond
	}

	ttl := defaultTTL
	if conf.TTLSec != 0 {
		ttl = time.Duration(conf.TTLSec) * time.Second
	}
	c := &Cache[T]{
		ctx:              ctx,
		items:            make(map[string]cacheItem[T]),
		ttl:              ttl,
		mutex:            &sync.RWMutex{},
		stopEviction:     stopChan,
		evictionInterval: evictionInterval,
	}
	c.startEviction()
	return c
}

// Set adds a new key-value pair to the cache
func (c *Cache[T]) Set(key string, value T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items[key] = cacheItem[T]{
		value:     value,
		expiresAt: time.Now().Add(c.ttl).UnixNano(),
	}
}

// Get returns the value for the given key and a boolean indicating whether the key was found
func (c *Cache[T]) Get(key string) (T, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	item, ok := c.items[key]
	if time.Now().UnixNano() > item.expiresAt {
		return item.value, false
	}
	return item.value, ok
}

// Delete removes the key-value pair from the cache
func (c *Cache[T]) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.items, key)
}

// DeleteExpired removes all expired items from the cache
func (c *Cache[T]) DeleteExpired() {
	now := time.Now().UnixNano()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for key, item := range c.items {
		if now > item.expiresAt {
			delete(c.items, key)
		}
	}
}

func (c *Cache[T]) StopEviction() {
	c.stopEviction <- true
}

func (c *Cache[T]) startEviction() {
	if c.isEvictionRunning {
		return
	}
	c.isEvictionRunning = true
	go func() {
		ticker := time.NewTicker(c.evictionInterval)
		for {
			select {
			case <-ticker.C:
				c.DeleteExpired()
			case <-c.stopEviction:
				ticker.Stop()
				c.isEvictionRunning = false
				return
			case <-c.ctx.Done():
				ticker.Stop()
				c.isEvictionRunning = false
				return
			}
		}
	}()
}
