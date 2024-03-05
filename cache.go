package main

import (
	"context"
	"sync"
	"time"
)

type Cache[T any] struct {
	ctx context.Context
	// The cache is a map of strings to strings
	items map[string]item[T]
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

type item[T any] struct {
	value     T
	expiresAt time.Time
}

const defaultEvictionInterval = time.Second

// NewCache creates a new cache with the given time to live
func NewCache[T any](ctx context.Context, ttl time.Duration) *Cache[T] {
	stopChan := make(chan bool)
	c := &Cache[T]{
		ctx:              ctx,
		items:            make(map[string]item[T]),
		ttl:              ttl,
		mutex:            &sync.RWMutex{},
		stopEviction:     stopChan,
		evictionInterval: defaultEvictionInterval,
	}
	c.startEviction()
	return c
}

// Set adds a new key-value pair to the cache
func (c *Cache[T]) Set(key string, value T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items[key] = item[T]{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Get returns the value for the given key and a boolean indicating whether the key was found
func (c *Cache[T]) Get(key string) (T, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	item, ok := c.items[key]
	return item.value, ok
}

// Delete removes the key-value pair from the cache
func (c *Cache[T]) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.items, key)
}

func (c *Cache[T]) DeleteExpired() {
	c.mutex.RLock()
	toDelete := make([]string, 0)
	for key, item := range c.items {
		if time.Now().After(item.expiresAt) {
			toDelete = append(toDelete, key)
		}
	}
	c.mutex.RUnlock()
	for _, key := range toDelete {
		c.Delete(key)
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
