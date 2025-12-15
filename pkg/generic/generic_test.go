package generic_test

import (
	"errors"
	"testing"

	"github.com/zareh/go-api-starter/pkg/generic"
)

func TestResult(t *testing.T) {
	t.Run("Ok result", func(t *testing.T) {
		result := generic.Ok(42)
		if !result.IsOk() {
			t.Error("expected IsOk to be true")
		}
		if result.IsErr() {
			t.Error("expected IsErr to be false")
		}
		if result.Unwrap() != 42 {
			t.Errorf("expected 42, got %d", result.Unwrap())
		}
	})

	t.Run("Err result", func(t *testing.T) {
		err := errors.New("test error")
		result := generic.Err[int](err)
		if result.IsOk() {
			t.Error("expected IsOk to be false")
		}
		if !result.IsErr() {
			t.Error("expected IsErr to be true")
		}
		if result.UnwrapErr() != err {
			t.Error("expected error to match")
		}
	})

	t.Run("UnwrapOr", func(t *testing.T) {
		okResult := generic.Ok(42)
		errResult := generic.Err[int](errors.New("error"))

		if okResult.UnwrapOr(0) != 42 {
			t.Error("expected 42")
		}
		if errResult.UnwrapOr(100) != 100 {
			t.Error("expected default value 100")
		}
	})
}

func TestOption(t *testing.T) {
	t.Run("Some option", func(t *testing.T) {
		opt := generic.Some("hello")
		if !opt.IsSome() {
			t.Error("expected IsSome to be true")
		}
		if opt.IsNone() {
			t.Error("expected IsNone to be false")
		}
		if opt.Unwrap() != "hello" {
			t.Errorf("expected 'hello', got %s", opt.Unwrap())
		}
	})

	t.Run("None option", func(t *testing.T) {
		opt := generic.None[string]()
		if opt.IsSome() {
			t.Error("expected IsSome to be false")
		}
		if !opt.IsNone() {
			t.Error("expected IsNone to be true")
		}
	})

	t.Run("UnwrapOr", func(t *testing.T) {
		someOpt := generic.Some("value")
		noneOpt := generic.None[string]()

		if someOpt.UnwrapOr("default") != "value" {
			t.Error("expected 'value'")
		}
		if noneOpt.UnwrapOr("default") != "default" {
			t.Error("expected 'default'")
		}
	})
}

func TestFilter(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	evens := generic.Filter(nums, func(n int) bool { return n%2 == 0 })

	if len(evens) != 5 {
		t.Errorf("expected 5 even numbers, got %d", len(evens))
	}
}

func TestMapSlice(t *testing.T) {
	nums := []int{1, 2, 3}
	doubled := generic.MapSlice(nums, func(n int) int { return n * 2 })

	expected := []int{2, 4, 6}
	for i, v := range doubled {
		if v != expected[i] {
			t.Errorf("expected %d, got %d", expected[i], v)
		}
	}
}

func TestReduce(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	sum := generic.Reduce(nums, 0, func(acc, n int) int { return acc + n })

	if sum != 15 {
		t.Errorf("expected 15, got %d", sum)
	}
}

func TestFind(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}

	found := generic.Find(nums, func(n int) bool { return n == 3 })
	if !found.IsSome() || found.Unwrap() != 3 {
		t.Error("expected to find 3")
	}

	notFound := generic.Find(nums, func(n int) bool { return n == 10 })
	if notFound.IsSome() {
		t.Error("expected not to find 10")
	}
}

func TestUnique(t *testing.T) {
	nums := []int{1, 2, 2, 3, 3, 3, 4}
	unique := generic.Unique(nums)

	if len(unique) != 4 {
		t.Errorf("expected 4 unique values, got %d", len(unique))
	}
}

func TestGroupBy(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	people := []Person{
		{"Alice", 30},
		{"Bob", 25},
		{"Charlie", 30},
	}

	grouped := generic.GroupBy(people, func(p Person) int { return p.Age })

	if len(grouped[30]) != 2 {
		t.Errorf("expected 2 people aged 30, got %d", len(grouped[30]))
	}
	if len(grouped[25]) != 1 {
		t.Errorf("expected 1 person aged 25, got %d", len(grouped[25]))
	}
}

func TestPartition(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6}
	evens, odds := generic.Partition(nums, func(n int) bool { return n%2 == 0 })

	if len(evens) != 3 {
		t.Errorf("expected 3 evens, got %d", len(evens))
	}
	if len(odds) != 3 {
		t.Errorf("expected 3 odds, got %d", len(odds))
	}
}

func TestMinMax(t *testing.T) {
	nums := []int{3, 1, 4, 1, 5, 9, 2, 6}

	minVal := generic.Min(nums)
	if !minVal.IsSome() || minVal.Unwrap() != 1 {
		t.Error("expected min to be 1")
	}

	maxVal := generic.Max(nums)
	if !maxVal.IsSome() || maxVal.Unwrap() != 9 {
		t.Error("expected max to be 9")
	}

	emptyMin := generic.Min([]int{})
	if emptyMin.IsSome() {
		t.Error("expected None for empty slice")
	}
}

func TestSum(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	sum := generic.Sum(nums)

	if sum != 15 {
		t.Errorf("expected 15, got %d", sum)
	}
}

func TestChunk(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7}
	chunks := generic.Chunk(nums, 3)

	if len(chunks) != 3 {
		t.Errorf("expected 3 chunks, got %d", len(chunks))
	}
	if len(chunks[0]) != 3 {
		t.Errorf("expected first chunk size 3, got %d", len(chunks[0]))
	}
	if len(chunks[2]) != 1 {
		t.Errorf("expected last chunk size 1, got %d", len(chunks[2]))
	}
}

func TestPtr(t *testing.T) {
	val := 42
	ptr := generic.Ptr(val)

	if *ptr != 42 {
		t.Errorf("expected 42, got %d", *ptr)
	}
}

func TestDeref(t *testing.T) {
	val := 42
	ptr := &val

	if generic.Deref(ptr) != 42 {
		t.Error("expected 42")
	}

	var nilPtr *int
	if generic.Deref(nilPtr) != 0 {
		t.Error("expected zero value for nil pointer")
	}

	if generic.DerefOr(nilPtr, 100) != 100 {
		t.Error("expected default value 100")
	}
}
