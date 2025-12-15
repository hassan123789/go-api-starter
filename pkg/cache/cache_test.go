package cache_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zareh/go-api-starter/pkg/cache"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("creates cache with default options", func(t *testing.T) {
		c := cache.New[string, int]()
		defer c.Close()

		stats := c.Stats()
		if stats.Capacity != 1000 {
			t.Errorf("expected default capacity 1000, got %d", stats.Capacity)
		}
		if stats.TTL != 0 {
			t.Errorf("expected default TTL 0, got %v", stats.TTL)
		}
	})

	t.Run("creates cache with custom capacity", func(t *testing.T) {
		c := cache.New[string, int](cache.WithCapacity[string, int](100))
		defer c.Close()

		stats := c.Stats()
		if stats.Capacity != 100 {
			t.Errorf("expected capacity 100, got %d", stats.Capacity)
		}
	})

	t.Run("creates cache with TTL", func(t *testing.T) {
		ttl := 5 * time.Minute
		c := cache.New[string, int](cache.WithTTL[string, int](ttl))
		defer c.Close()

		stats := c.Stats()
		if stats.TTL != ttl {
			t.Errorf("expected TTL %v, got %v", ttl, stats.TTL)
		}
	})
}

func TestSetAndGet(t *testing.T) {
	t.Parallel()

	t.Run("set and get basic values", func(t *testing.T) {
		c := cache.New[string, int]()
		defer c.Close()

		c.Set("key1", 100)
		c.Set("key2", 200)

		if val, ok := c.Get("key1"); !ok || val != 100 {
			t.Errorf("expected 100, got %v (ok=%v)", val, ok)
		}
		if val, ok := c.Get("key2"); !ok || val != 200 {
			t.Errorf("expected 200, got %v (ok=%v)", val, ok)
		}
	})

	t.Run("get returns false for missing key", func(t *testing.T) {
		c := cache.New[string, int]()
		defer c.Close()

		if val, ok := c.Get("missing"); ok {
			t.Errorf("expected ok=false for missing key, got val=%v", val)
		}
	})

	t.Run("set updates existing value", func(t *testing.T) {
		c := cache.New[string, int]()
		defer c.Close()

		c.Set("key", 100)
		c.Set("key", 200)

		if val, ok := c.Get("key"); !ok || val != 200 {
			t.Errorf("expected 200, got %v (ok=%v)", val, ok)
		}
		if c.Len() != 1 {
			t.Errorf("expected 1 item, got %d", c.Len())
		}
	})
}

func TestTTLExpiration(t *testing.T) {
	t.Parallel()

	t.Run("item expires after TTL", func(t *testing.T) {
		ttl := 50 * time.Millisecond
		c := cache.New[string, int](cache.WithTTL[string, int](ttl))
		defer c.Close()

		c.Set("key", 100)

		// Should exist immediately
		if _, ok := c.Get("key"); !ok {
			t.Error("expected key to exist immediately")
		}

		// Wait for expiration
		time.Sleep(ttl + 10*time.Millisecond)

		// Should be expired
		if _, ok := c.Get("key"); ok {
			t.Error("expected key to be expired")
		}
	})

	t.Run("SetWithTTL overrides default TTL", func(t *testing.T) {
		c := cache.New[string, int](cache.WithTTL[string, int](1 * time.Hour))
		defer c.Close()

		customTTL := 50 * time.Millisecond
		c.SetWithTTL("key", 100, customTTL)

		// Should exist immediately
		if _, ok := c.Get("key"); !ok {
			t.Error("expected key to exist immediately")
		}

		// Wait for custom TTL expiration
		time.Sleep(customTTL + 10*time.Millisecond)

		// Should be expired
		if _, ok := c.Get("key"); ok {
			t.Error("expected key to be expired")
		}
	})
}

func TestLRUEviction(t *testing.T) {
	t.Parallel()

	t.Run("evicts least recently used item", func(t *testing.T) {
		c := cache.New[string, int](cache.WithCapacity[string, int](3))
		defer c.Close()

		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)

		// Access 'a' to make it recently used
		c.Get("a")

		// Add new item - should evict 'b' (least recently used)
		c.Set("d", 4)

		if c.Len() != 3 {
			t.Errorf("expected 3 items, got %d", c.Len())
		}

		// 'b' should be evicted
		if _, ok := c.Get("b"); ok {
			t.Error("expected 'b' to be evicted")
		}

		// Others should exist
		if _, ok := c.Get("a"); !ok {
			t.Error("expected 'a' to exist")
		}
		if _, ok := c.Get("c"); !ok {
			t.Error("expected 'c' to exist")
		}
		if _, ok := c.Get("d"); !ok {
			t.Error("expected 'd' to exist")
		}
	})
}

func TestOnEvict(t *testing.T) {
	t.Parallel()

	t.Run("calls onEvict when item is evicted", func(t *testing.T) {
		var evictedKey string
		var evictedVal int

		c := cache.New[string, int](
			cache.WithCapacity[string, int](2),
			cache.WithOnEvict[string, int](func(k string, v int) {
				evictedKey = k
				evictedVal = v
			}),
		)
		defer c.Close()

		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3) // Should evict 'a'

		if evictedKey != "a" || evictedVal != 1 {
			t.Errorf("expected evicted key='a', val=1, got key='%s', val=%d", evictedKey, evictedVal)
		}
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	t.Run("deletes existing key", func(t *testing.T) {
		c := cache.New[string, int]()
		defer c.Close()

		c.Set("key", 100)
		if !c.Delete("key") {
			t.Error("expected Delete to return true")
		}
		if _, ok := c.Get("key"); ok {
			t.Error("expected key to be deleted")
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		c := cache.New[string, int]()
		defer c.Close()

		if c.Delete("missing") {
			t.Error("expected Delete to return false for missing key")
		}
	})
}

func TestHas(t *testing.T) {
	t.Parallel()

	t.Run("returns true for existing key", func(t *testing.T) {
		c := cache.New[string, int]()
		defer c.Close()

		c.Set("key", 100)
		if !c.Has("key") {
			t.Error("expected Has to return true")
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		c := cache.New[string, int]()
		defer c.Close()

		if c.Has("missing") {
			t.Error("expected Has to return false")
		}
	})

	t.Run("returns false for expired key", func(t *testing.T) {
		ttl := 50 * time.Millisecond
		c := cache.New[string, int](cache.WithTTL[string, int](ttl))
		defer c.Close()

		c.Set("key", 100)
		time.Sleep(ttl + 10*time.Millisecond)

		if c.Has("key") {
			t.Error("expected Has to return false for expired key")
		}
	})
}

func TestGetOrSet(t *testing.T) {
	t.Parallel()

	t.Run("returns existing value", func(t *testing.T) {
		c := cache.New[string, int]()
		defer c.Close()

		c.Set("key", 100)
		val, loaded := c.GetOrSet("key", 200)

		if val != 100 {
			t.Errorf("expected 100, got %d", val)
		}
		if !loaded {
			t.Error("expected loaded=true for existing key")
		}
	})

	t.Run("sets and returns new value", func(t *testing.T) {
		c := cache.New[string, int]()
		defer c.Close()

		val, loaded := c.GetOrSet("key", 100)

		if val != 100 {
			t.Errorf("expected 100, got %d", val)
		}
		if loaded {
			t.Error("expected loaded=false for new key")
		}
	})
}

func TestClear(t *testing.T) {
	t.Parallel()

	t.Run("removes all items", func(t *testing.T) {
		c := cache.New[string, int]()
		defer c.Close()

		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)

		c.Clear()

		if c.Len() != 0 {
			t.Errorf("expected 0 items after clear, got %d", c.Len())
		}
	})

	t.Run("calls onEvict for all items", func(t *testing.T) {
		var count int
		c := cache.New[string, int](
			cache.WithOnEvict[string, int](func(k string, v int) {
				count++
			}),
		)
		defer c.Close()

		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)

		c.Clear()

		if count != 3 {
			t.Errorf("expected onEvict called 3 times, got %d", count)
		}
	})
}

func TestKeys(t *testing.T) {
	t.Parallel()

	t.Run("returns all keys", func(t *testing.T) {
		c := cache.New[string, int]()
		defer c.Close()

		c.Set("a", 1)
		c.Set("b", 2)
		c.Set("c", 3)

		keys := c.Keys()
		if len(keys) != 3 {
			t.Errorf("expected 3 keys, got %d", len(keys))
		}

		keySet := make(map[string]bool)
		for _, k := range keys {
			keySet[k] = true
		}
		if !keySet["a"] || !keySet["b"] || !keySet["c"] {
			t.Error("expected keys to contain a, b, c")
		}
	})
}

func TestConcurrency(t *testing.T) {
	t.Parallel()

	t.Run("concurrent read/write", func(t *testing.T) {
		c := cache.New[int, int](cache.WithCapacity[int, int](1000))
		defer c.Close()

		var wg sync.WaitGroup
		numGoroutines := 100
		numOps := 1000

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOps; j++ {
					key := (id*numOps + j) % 500
					c.Set(key, j)
					c.Get(key)
					c.Has(key)
				}
			}(i)
		}

		wg.Wait()
	})

	t.Run("concurrent GetOrSet", func(t *testing.T) {
		c := cache.New[string, *int64]()
		defer c.Close()

		var wg sync.WaitGroup
		numGoroutines := 100
		counter := new(int64)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				val, loaded := c.GetOrSet("counter", counter)
				if !loaded {
					atomic.AddInt64(val, 1)
				}
			}()
		}

		wg.Wait()

		// Only one goroutine should have set the value
		if *counter != 1 {
			t.Errorf("expected counter=1, got %d", *counter)
		}
	})
}

func TestStats(t *testing.T) {
	t.Parallel()

	t.Run("returns correct stats", func(t *testing.T) {
		c := cache.New[string, int](
			cache.WithCapacity[string, int](100),
			cache.WithTTL[string, int](5*time.Minute),
		)
		defer c.Close()

		c.Set("a", 1)
		c.Set("b", 2)

		stats := c.Stats()
		if stats.Size != 2 {
			t.Errorf("expected size 2, got %d", stats.Size)
		}
		if stats.Capacity != 100 {
			t.Errorf("expected capacity 100, got %d", stats.Capacity)
		}
		if stats.TTL != 5*time.Minute {
			t.Errorf("expected TTL 5m, got %v", stats.TTL)
		}
	})
}

// Example tests for documentation
func ExampleNew() {
	// Create a simple string->int cache
	c := cache.New[string, int]()
	defer c.Close()

	c.Set("answer", 42)
	if val, ok := c.Get("answer"); ok {
		fmt.Println(val)
	}
	// Output: 42
}

func ExampleWithCapacity() {
	// Create cache with max 100 items (LRU eviction)
	c := cache.New[string, int](cache.WithCapacity[string, int](100))
	defer c.Close()

	fmt.Println(c.Stats().Capacity)
	// Output: 100
}

func ExampleWithTTL() {
	// Create cache with 5 minute TTL
	c := cache.New[string, int](cache.WithTTL[string, int](5 * time.Minute))
	defer c.Close()

	fmt.Println(c.Stats().TTL)
	// Output: 5m0s
}

func ExampleCache_GetOrSet() {
	c := cache.New[string, int]()
	defer c.Close()

	// First call sets the value
	val, existed := c.GetOrSet("key", 100)
	fmt.Printf("val=%d, existed=%v\n", val, existed)

	// Second call returns existing value
	val, existed = c.GetOrSet("key", 200)
	fmt.Printf("val=%d, existed=%v\n", val, existed)

	// Output:
	// val=100, existed=false
	// val=100, existed=true
}

// Benchmarks
func BenchmarkSet(b *testing.B) {
	c := cache.New[int, int]()
	defer c.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(i, i)
	}
}

func BenchmarkGet(b *testing.B) {
	c := cache.New[int, int]()
	defer c.Close()

	for i := 0; i < 1000; i++ {
		c.Set(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(i % 1000)
	}
}

func BenchmarkSetWithEviction(b *testing.B) {
	c := cache.New[int, int](cache.WithCapacity[int, int](100))
	defer c.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(i, i)
	}
}

func BenchmarkConcurrentAccess(b *testing.B) {
	c := cache.New[int, int]()
	defer c.Close()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				c.Set(i, i)
			} else {
				c.Get(i)
			}
			i++
		}
	})
}
