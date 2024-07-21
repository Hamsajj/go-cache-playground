package cache

import (
	"cache-api/config"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func createNewCache() *Cache[string] {
	return NewCache[string](context.Background(), config.CacheConfig{
		TTLSec:                   10,
		EvictionIntervalMilliSec: 0,
	})
}

func TestNewCache(t *testing.T) {
	t.Run("NewCache with default values", func(t *testing.T) {
		cache := NewCache[string](context.Background(), config.CacheConfig{})
		if cache == nil {
			t.Fatalf("Expected a cache but got nil")
		}
		if cache.ttl != defaultTTL {
			t.Errorf("Expected ttl to be %d but got %d", defaultTTL, cache.ttl)
		}
		if cache.evictionInterval != defaultEvictionInterval {
			t.Errorf("Expected interval to be %d but got %d", defaultEvictionInterval, cache.evictionInterval)
		}
	})

	t.Run("NewCache with custom values", func(t *testing.T) {
		cache := NewCache[string](context.Background(), config.CacheConfig{
			TTLSec:                   20,
			EvictionIntervalMilliSec: 1000,
		})
		if cache.ttl != 20*time.Second {
			t.Errorf("Expected ttl to be %d but got %d", 20*time.Second, cache.ttl)
		}
		if cache.evictionInterval != defaultEvictionInterval {
			t.Errorf("Expected interval to be %d but got %d", defaultEvictionInterval, cache.evictionInterval)
		}
	})
}

func TestCache_Set(t *testing.T) {
	cache := createNewCache()
	err := cache.Set("key", "value")
	if err != nil {
		t.Fatalf("Expected no error but got %v", err)
	}
	assertValueExists(t, cache, "key", "value")

	// re-write
	err = cache.Set("key", "newValue")
	if err != nil {
		t.Fatalf("Expected no error for re-write but got %v", err)
	}
	assertValueExists(t, cache, "key", "newValue")

}

func TestCache_Get(t *testing.T) {
	cache := createNewCache()
	cache.items["key"] = cacheItem[string]{
		value:     "value",
		expiresAt: time.Now().Add(10 * time.Second).UnixNano(),
	}
	val, ok := cache.Get("key")
	if !ok {
		t.Errorf("Expected 'key' to be present in the cache")
	}
	if val != "value" {
		t.Errorf("Expected 'key' to have value 'value'")
	}

	cache.items["expiredKey"] = cacheItem[string]{
		value:     "expired",
		expiresAt: time.Now().Add(-10 * time.Second).UnixNano(),
	}
	_, ok = cache.Get("expiredKey")
	if ok {
		t.Errorf("Expected 'expiredKey' to return false indicating value is not ok")
	}

	_, ok = cache.Get("nonExistentKey")
	if ok {
		t.Errorf("Expected 'nonExistentKey' to be absent from the cache")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := createNewCache()
	cache.items["key"] = cacheItem[string]{
		value:     "value",
		expiresAt: time.Now().Add(10 * time.Second).UnixNano(),
	}
	cache.Delete("key")
	if _, ok := cache.items["key"]; ok {
		t.Errorf("Expected 'key' to be deleted from the cache")
	}
}

func TestCache_DeleteExpired(t *testing.T) {
	items := map[string]cacheItem[string]{
		"expiredKey1": {
			value:     "value1",
			expiresAt: time.Now().Add(-10 * time.Second).UnixNano(),
		},
		"expiredKey2": {
			value:     "value2",
			expiresAt: time.Now().Add(-10 * time.Second).UnixNano(),
		},
		"key2": {
			value:     "value2",
			expiresAt: time.Now().Add(10 * time.Second).UnixNano(),
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
	items := map[string]cacheItem[string]{
		"expiredKey1": {
			value:     "value1",
			expiresAt: time.Now().Add(-10 * time.Second).UnixNano(),
		},
		"expiredKey2": {
			value:     "value2",
			expiresAt: time.Now().Add(-10 * time.Second).UnixNano(),
		},
		"key2": {
			value:     "value2",
			expiresAt: time.Now().Add(10 * time.Second).UnixNano(),
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
	cache := createNewCache()
	if cache.isEvictionRunning == false {
		t.Errorf("Expected eviction to be running")
	}
	cache.StopEviction()
	cache.items["expiredKey1"] = cacheItem[string]{
		value:     "value1",
		expiresAt: time.Now().Add(-10 * time.Second).UnixNano(),
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
	if val.expiresAt == 0 {
		t.Errorf("Expected '%s' to have a non-zero expiration time", key)
	}
}

func BenchmarkCache_DeleteExpired(b *testing.B) {
	b.StopTimer()
	cache := NewCache[string](context.Background(), config.CacheConfig{
		TTLSec:                   -1,
		EvictionIntervalMilliSec: 0,
	})
	cache.StopEviction()
	// set half of the items to be expired
	for i := 0; i < b.N; i++ {
		cache.items[fmt.Sprintf("key%d", i)] = cacheItem[string]{
			value:     "value",
			expiresAt: time.Now().Add(time.Duration(1-2*(i%2)) * time.Second).UnixNano(),
		}
	}
	b.StartTimer()
	cache.DeleteExpired()
}

func BenchmarkCache_SetWhileDeleteExpired(b *testing.B) {
	b.StopTimer()
	cache := Cache[string]{
		ctx:              context.Background(),
		items:            make(map[string]cacheItem[string]),
		ttl:              time.Millisecond * 10,
		mutex:            &sync.RWMutex{},
		stopEviction:     make(chan bool),
		evictionInterval: time.Millisecond * 10,
	}
	cache.startEviction()
	b.StartTimer()
	// set half of the items to be expired
	for i := 0; i < b.N; i++ {
		err := cache.Set(fmt.Sprintf("key%d", i), "value")
		if err != nil {
			b.Fatalf("Expected no error but got %v", err)
		}
	}
}
