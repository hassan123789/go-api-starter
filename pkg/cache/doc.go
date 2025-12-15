// Package cache provides a generic thread-safe in-memory cache with TTL and LRU eviction.
//
// # Overview
//
// This package implements a production-ready cache with the following features:
//   - Generic type support for keys (any comparable type) and values (any type)
//   - Time-to-live (TTL) for automatic expiration
//   - Least Recently Used (LRU) eviction policy
//   - Thread-safe operations using sync.RWMutex
//   - Configurable capacity and cleanup intervals
//   - Eviction callbacks
//
// # Basic Usage
//
//	c := cache.New[string, User](
//	    cache.WithCapacity[string, User](1000),
//	    cache.WithTTL[string, User](5 * time.Minute),
//	)
//	defer c.Close()
//
//	c.Set("user:123", user)
//	if user, ok := c.Get("user:123"); ok {
//	    // use user
//	}
//
// # Configuration Options
//
// The cache supports several configuration options via functional options:
//
//   - WithCapacity: Set maximum number of items (default: 1000)
//   - WithTTL: Set default time-to-live for entries
//   - WithOnEvict: Set callback function for evicted items
//
// # LRU Eviction
//
// When the cache reaches capacity, the least recently used item is evicted.
// Items are moved to the front of the LRU list on each Get or Set operation.
//
// # TTL Expiration
//
// Items can expire based on a global TTL or per-item TTL set via SetWithTTL.
// A background goroutine periodically cleans up expired items.
//
// # Thread Safety
//
// All operations are thread-safe. The cache uses sync.RWMutex for locking,
// allowing concurrent reads while serializing writes.
//
// # Resource Management
//
// Always call Close() when done with the cache to stop the cleanup goroutine:
//
//	c := cache.New[string, int](cache.WithTTL[string, int](time.Minute))
//	defer c.Close()
package cache
