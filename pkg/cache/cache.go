// Package cache provides a generic thread-safe in-memory cache with TTL support.
//
// This package implements production-ready caching with:
//   - Generic type support for keys and values
//   - Time-to-live (TTL) for automatic expiration
//   - Least Recently Used (LRU) eviction policy
//   - Thread-safe operations
//   - Configurable capacity and cleanup intervals
//
// Example usage:
//
//	c := cache.New[string, User](
//	    cache.WithCapacity[string, User](1000),
//	    cache.WithTTL[string, User](5 * time.Minute),
//	)
//	c.Set("user:123", user)
//	if user, ok := c.Get("user:123"); ok {
//	    // use user
//	}
package cache

import (
	"container/list"
	"sync"
	"time"
)

// entry represents a cached item with its metadata.
type entry[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time
	element   *list.Element
}

// Cache is a generic thread-safe cache with TTL and LRU eviction.
type Cache[K comparable, V any] struct {
	mu       sync.RWMutex
	items    map[K]*entry[K, V]
	lru      *list.List
	capacity int
	ttl      time.Duration
	onEvict  func(K, V)
	closed   bool
	stopCh   chan struct{}
}

// Option is a functional option for configuring the cache.
type Option[K comparable, V any] func(*Cache[K, V])

// WithCapacity sets the maximum number of items in the cache.
// When capacity is reached, the least recently used item is evicted.
func WithCapacity[K comparable, V any](capacity int) Option[K, V] {
	return func(c *Cache[K, V]) {
		if capacity > 0 {
			c.capacity = capacity
		}
	}
}

// WithTTL sets the default time-to-live for cache entries.
// Items are automatically expired after this duration.
func WithTTL[K comparable, V any](ttl time.Duration) Option[K, V] {
	return func(c *Cache[K, V]) {
		if ttl > 0 {
			c.ttl = ttl
		}
	}
}

// WithOnEvict sets a callback function that is called when an item is evicted.
func WithOnEvict[K comparable, V any](fn func(K, V)) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.onEvict = fn
	}
}

// New creates a new cache with the given options.
func New[K comparable, V any](opts ...Option[K, V]) *Cache[K, V] {
	c := &Cache[K, V]{
		items:    make(map[K]*entry[K, V]),
		lru:      list.New(),
		capacity: 1000, // default capacity
		ttl:      0,    // no TTL by default
		stopCh:   make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	// Start cleanup goroutine if TTL is set
	if c.ttl > 0 {
		go c.cleanupLoop()
	}

	return c
}

// cleanupLoop periodically removes expired items.
func (c *Cache[K, V]) cleanupLoop() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCh:
			return
		}
	}
}

// cleanup removes all expired items.
func (c *Cache[K, V]) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			c.removeEntry(key, item)
		}
	}
}

// Set adds or updates an item in the cache.
func (c *Cache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL adds or updates an item with a specific TTL.
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	// Update existing entry
	if item, exists := c.items[key]; exists {
		item.value = value
		item.expiresAt = expiresAt
		c.lru.MoveToFront(item.element)
		return
	}

	// Evict if at capacity
	if c.capacity > 0 && len(c.items) >= c.capacity {
		c.evictOldest()
	}

	// Add new entry
	item := &entry[K, V]{
		key:       key,
		value:     value,
		expiresAt: expiresAt,
	}
	item.element = c.lru.PushFront(key)
	c.items[key] = item
}

// Get retrieves an item from the cache.
// Returns the value and true if found and not expired, otherwise zero value and false.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		var zero V
		return zero, false
	}

	// Check expiration
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		c.removeEntry(key, item)
		var zero V
		return zero, false
	}

	// Move to front (LRU)
	c.lru.MoveToFront(item.element)

	return item.value, true
}

// GetOrSet returns the existing value for a key, or sets and returns the given value if not present.
func (c *Cache[K, V]) GetOrSet(key K, value V) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, exists := c.items[key]; exists {
		if item.expiresAt.IsZero() || time.Now().Before(item.expiresAt) {
			c.lru.MoveToFront(item.element)
			return item.value, true
		}
		c.removeEntry(key, item)
	}

	// Add new entry
	var expiresAt time.Time
	if c.ttl > 0 {
		expiresAt = time.Now().Add(c.ttl)
	}

	// Evict if at capacity
	if c.capacity > 0 && len(c.items) >= c.capacity {
		c.evictOldest()
	}

	item := &entry[K, V]{
		key:       key,
		value:     value,
		expiresAt: expiresAt,
	}
	item.element = c.lru.PushFront(key)
	c.items[key] = item

	return value, false
}

// Delete removes an item from the cache.
func (c *Cache[K, V]) Delete(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		return false
	}

	c.removeEntry(key, item)
	return true
}

// Has checks if a key exists in the cache and is not expired.
func (c *Cache[K, V]) Has(key K) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return false
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return false
	}

	return true
}

// Len returns the number of items in the cache.
func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Clear removes all items from the cache.
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		if c.onEvict != nil {
			c.onEvict(key, item.value)
		}
	}

	c.items = make(map[K]*entry[K, V])
	c.lru.Init()
}

// Keys returns all keys in the cache (including expired ones that haven't been cleaned up yet).
func (c *Cache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

// Close stops the cleanup goroutine and releases resources.
func (c *Cache[K, V]) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	c.closed = true
	close(c.stopCh)
}

// evictOldest removes the least recently used item.
func (c *Cache[K, V]) evictOldest() {
	elem := c.lru.Back()
	if elem == nil {
		return
	}

	key, ok := elem.Value.(K)
	if !ok {
		return
	}
	if item, exists := c.items[key]; exists {
		c.removeEntry(key, item)
	}
}

// removeEntry removes an entry from the cache.
func (c *Cache[K, V]) removeEntry(key K, item *entry[K, V]) {
	delete(c.items, key)
	c.lru.Remove(item.element)

	if c.onEvict != nil {
		c.onEvict(key, item.value)
	}
}

// Stats returns cache statistics.
type Stats struct {
	Size     int
	Capacity int
	TTL      time.Duration
}

// Stats returns the current cache statistics.
func (c *Cache[K, V]) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
		Size:     len(c.items),
		Capacity: c.capacity,
		TTL:      c.ttl,
	}
}
