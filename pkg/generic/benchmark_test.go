package generic

import "testing"

// Benchmark tests for generic utility functions

func BenchmarkFilter(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Filter(data, func(n int) bool {
			return n%2 == 0
		})
	}
}

func BenchmarkMapSlice(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MapSlice(data, func(n int) int {
			return n * 2
		})
	}
}

func BenchmarkReduce(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Reduce(data, 0, func(acc, n int) int {
			return acc + n
		})
	}
}

func BenchmarkGroupBy(b *testing.B) {
	type Item struct {
		Category string
		Value    int
	}
	categories := []string{"A", "B", "C", "D", "E"}
	data := make([]Item, 1000)
	for i := range data {
		data[i] = Item{
			Category: categories[i%len(categories)],
			Value:    i,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GroupBy(data, func(item Item) string {
			return item.Category
		})
	}
}

func BenchmarkPartition(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Partition(data, func(n int) bool {
			return n%2 == 0
		})
	}
}

func BenchmarkChunk(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Chunk(data, 10)
	}
}

func BenchmarkZip(b *testing.B) {
	data1 := make([]int, 1000)
	data2 := make([]string, 1000)
	for i := range data1 {
		data1[i] = i
		data2[i] = "item"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Zip(data1, data2)
	}
}

func BenchmarkFind(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Find(data, func(n int) bool {
			return n == 500
		})
	}
}

func BenchmarkSum(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Sum(data)
	}
}

func BenchmarkMin(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Min(data)
	}
}

func BenchmarkMax(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Max(data)
	}
}

func BenchmarkResult_UnwrapOr(b *testing.B) {
	ok := Ok(42)
	err := Err[int](nil)

	b.Run("Ok", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ok.UnwrapOr(0)
		}
	})

	b.Run("Err", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = err.UnwrapOr(0)
		}
	})
}

func BenchmarkOption_UnwrapOr(b *testing.B) {
	some := Some(42)
	none := None[int]()

	b.Run("Some", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = some.UnwrapOr(0)
		}
	})

	b.Run("None", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = none.UnwrapOr(0)
		}
	})
}
