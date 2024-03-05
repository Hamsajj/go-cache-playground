package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	ttl := 10 * time.Second
	cache := NewCache[string](context.Background(), ttl)
	if cache == nil {
		t.Errorf("Expected a cache but got nil")
	}
	if cache.ttl != ttl {
		t.Errorf("Expected %d but got %d", ttl, cache.ttl)
	}
}

func TestCache_Set(t *testing.T) {
	cache := NewCache[string](context.Background(), 10*time.Second)
	cache.Set("key", "value")
	assertValueExists(t, cache, "key", "value")

	// re-write
	cache.Set("key", "newValue")
	assertValueExists(t, cache, "key", "newValue")

}

func TestCache_Get(t *testing.T) {
	cache := NewCache[string](context.Background(), 10*time.Second)
	cache.items["key"] = item[string]{
		value:     "value",
		expiresAt: time.Now().Add(10 * time.Second),
	}
	val, ok := cache.Get("key")
	if !ok {
		t.Errorf("Expected 'key' to be present in the cache")
	}
	if val != "value" {
		t.Errorf("Expected 'key' to have value 'value'")
	}

	val, ok = cache.Get("nonExistentKey")
	if ok {
		t.Errorf("Expected 'nonExistentKey' to be absent from the cache")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache[string](context.Background(), 10*time.Second)
	cache.items["key"] = item[string]{
		value:     "value",
		expiresAt: time.Now().Add(10 * time.Second),
	}
	cache.Delete("key")
	if _, ok := cache.items["key"]; ok {
		t.Errorf("Expected 'key' to be deleted from the cache")
	}
}

func TestCache_DeleteExpired(t *testing.T) {
	items := map[string]item[string]{
		"expiredKey1": {
			value:     "value1",
			expiresAt: time.Now().Add(-10 * time.Second),
		},
		"expiredKey2": {
			value:     "value2",
			expiresAt: time.Now().Add(-10 * time.Second),
		},
		"key2": {
			value:     "value2",
			expiresAt: time.Now().Add(10 * time.Second),
		},
	}

	cache := &Cache[string]{
		ctx:          context.Background(),
		items:        items,
		ttl:          10 * time.Second,
		mutex:        &sync.RWMutex{},
		stopEviction: make(chan bool),
	}

	cache.DeleteExpired()
	if _, ok := cache.items["expiredKey1"]; ok {
		t.Errorf("Expected 'expiredKey1' to be deleted from the cache")
	}
	if _, ok := cache.items["expiredKey2"]; ok {
		t.Errorf("Expected 'expiredKey2' to be deleted from the cache")
	}
	assertValueExists(t, cache, "key2", "value2")
}

func TestCache_Eviction(t *testing.T) {
	items := map[string]item[string]{
		"expiredKey1": {
			value:     "value1",
			expiresAt: time.Now().Add(-10 * time.Second),
		},
		"expiredKey2": {
			value:     "value2",
			expiresAt: time.Now().Add(-10 * time.Second),
		},
		"key2": {
			value:     "value2",
			expiresAt: time.Now().Add(10 * time.Second),
		},
	}

	cache := &Cache[string]{
		ctx:              context.Background(),
		items:            items,
		ttl:              10 * time.Second,
		mutex:            &sync.RWMutex{},
		stopEviction:     make(chan bool),
		evictionInterval: time.Millisecond * 500,
	}
	cache.startEviction()
	time.Sleep(1 * time.Second)

	if _, ok := cache.items["expiredKey1"]; ok {
		t.Errorf("Expected 'expiredKey1' to be deleted from the cache")
	}
	if _, ok := cache.items["expiredKey2"]; ok {
		t.Errorf("Expected 'expiredKey2' to be deleted from the cache")
	}
	assertValueExists(t, cache, "key2", "value2")
}

func TestCache_StopEviction(t *testing.T) {
	cache := NewCache[string](context.Background(), 10*time.Second)
	if cache.isEvictionRunning == false {
		t.Errorf("Expected eviction to be running")
	}
	cache.StopEviction()
	cache.items["expiredKey1"] = item[string]{
		value:     "value1",
		expiresAt: time.Now().Add(-10 * time.Second),
	}
	time.Sleep(1 * time.Second)
	if _, ok := cache.items["expiredKey1"]; !ok {
		t.Errorf("Expected 'expiredKey1' to be present in the cache, as eviction is stopped")
	}
	if cache.isEvictionRunning == true {
		t.Errorf("Expected eviction to be stopped")
	}
}

func assertValueExists(t *testing.T, cache *Cache[string], key string, expectedValue string) {
	val, ok := cache.items[key]
	if !ok {
		t.Errorf("Expected '%s' to be present in the cache", key)
	}
	if val.value != expectedValue {
		t.Errorf("Expected '%s' to have value '%s'", key, expectedValue)
	}
	if val.expiresAt == (time.Time{}) {
		t.Errorf("Expected '%s' to have a non-zero expiration time", key)
	}
}
