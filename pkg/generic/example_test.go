package generic_test

import (
	"fmt"
	"sort"

	"github.com/zareh/go-api-starter/pkg/generic"
)

// ExampleResult demonstrates the Result monad for error handling.
func ExampleResult() {
	// Successful result
	ok := generic.Ok(42)
	value, isOk := ok.Value()
	fmt.Printf("IsOk: %v, Value: %v\n", isOk, value)

	// Error result
	err := generic.Err[int](fmt.Errorf("something went wrong"))
	fmt.Printf("IsOk: %v, IsErr: %v\n", err.IsOk(), err.IsErr())

	// UnwrapOr provides a default value
	errValue := err.UnwrapOr(100)
	fmt.Printf("UnwrapOr on error: %v\n", errValue)

	// Output:
	// IsOk: true, Value: 42
	// IsOk: false, IsErr: true
	// UnwrapOr on error: 100
}

// ExampleOption demonstrates the Option type for nullable values.
func ExampleOption() {
	// Some value
	some := generic.Some(42)
	value, isSome := some.Value()
	fmt.Printf("IsSome: %v, Value: %v\n", isSome, value)

	// None value
	none := generic.None[int]()
	fmt.Printf("IsNone: %v\n", none.IsNone())

	// UnwrapOr provides a default value
	defaulted := none.UnwrapOr(100)
	fmt.Printf("UnwrapOr: %v\n", defaulted)

	// Output:
	// IsSome: true, Value: 42
	// IsNone: true
	// UnwrapOr: 100
}

// ExampleFilter demonstrates filtering slices with a predicate.
func ExampleFilter() {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Filter even numbers
	evens := generic.Filter(numbers, func(n int) bool {
		return n%2 == 0
	})
	fmt.Printf("Evens: %v\n", evens)

	// Filter numbers greater than 5
	greaterThan5 := generic.Filter(numbers, func(n int) bool {
		return n > 5
	})
	fmt.Printf("Greater than 5: %v\n", greaterThan5)

	// Output:
	// Evens: [2 4 6 8 10]
	// Greater than 5: [6 7 8 9 10]
}

// ExampleMapSlice demonstrates transforming slices.
func ExampleMapSlice() {
	numbers := []int{1, 2, 3, 4, 5}

	// Double each number
	doubled := generic.MapSlice(numbers, func(n int) int {
		return n * 2
	})
	fmt.Printf("Doubled: %v\n", doubled)

	// Convert to strings
	strings := generic.MapSlice(numbers, func(n int) string {
		return fmt.Sprintf("num-%d", n)
	})
	fmt.Printf("Strings: %v\n", strings)

	// Output:
	// Doubled: [2 4 6 8 10]
	// Strings: [num-1 num-2 num-3 num-4 num-5]
}

// ExampleReduce demonstrates aggregating values.
func ExampleReduce() {
	numbers := []int{1, 2, 3, 4, 5}

	// Sum all numbers
	sum := generic.Reduce(numbers, 0, func(acc, n int) int {
		return acc + n
	})
	fmt.Printf("Sum: %v\n", sum)

	// Product of all numbers
	product := generic.Reduce(numbers, 1, func(acc, n int) int {
		return acc * n
	})
	fmt.Printf("Product: %v\n", product)

	// Concatenate strings
	words := []string{"Hello", " ", "World"}
	joined := generic.Reduce(words, "", func(acc, s string) string {
		return acc + s
	})
	fmt.Printf("Joined: %q\n", joined)

	// Output:
	// Sum: 15
	// Product: 120
	// Joined: "Hello World"
}

// ExampleGroupBy demonstrates grouping elements.
func ExampleGroupBy() {
	type Person struct {
		Name string
		Age  int
	}

	people := []Person{
		{"Alice", 25},
		{"Bob", 30},
		{"Charlie", 25},
		{"David", 30},
	}

	// Group by age
	byAge := generic.GroupBy(people, func(p Person) int {
		return p.Age
	})

	// Sort ages for consistent output
	ages := make([]int, 0, len(byAge))
	for age := range byAge {
		ages = append(ages, age)
	}
	sort.Ints(ages)

	for _, age := range ages {
		names := make([]string, 0, len(byAge[age]))
		for _, p := range byAge[age] {
			names = append(names, p.Name)
		}
		fmt.Printf("Age %d: %v\n", age, names)
	}

	// Output:
	// Age 25: [Alice Charlie]
	// Age 30: [Bob David]
}

// ExamplePartition demonstrates splitting slices.
func ExamplePartition() {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Partition into even and odd
	evens, odds := generic.Partition(numbers, func(n int) bool {
		return n%2 == 0
	})
	fmt.Printf("Evens: %v\n", evens)
	fmt.Printf("Odds: %v\n", odds)

	// Output:
	// Evens: [2 4 6 8 10]
	// Odds: [1 3 5 7 9]
}

// ExampleChunk demonstrates splitting slices into chunks.
func ExampleChunk() {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Split into chunks of 3
	chunks := generic.Chunk(numbers, 3)
	for i, chunk := range chunks {
		fmt.Printf("Chunk %d: %v\n", i, chunk)
	}

	// Output:
	// Chunk 0: [1 2 3]
	// Chunk 1: [4 5 6]
	// Chunk 2: [7 8 9]
	// Chunk 3: [10]
}

// ExampleZip demonstrates combining two slices.
func ExampleZip() {
	names := []string{"Alice", "Bob", "Charlie"}
	ages := []int{25, 30, 35}

	// Zip names and ages
	pairs := generic.Zip(names, ages)
	for _, pair := range pairs {
		fmt.Printf("%s is %d years old\n", pair.First, pair.Second)
	}

	// Output:
	// Alice is 25 years old
	// Bob is 30 years old
	// Charlie is 35 years old
}

// ExampleFind demonstrates finding elements.
func ExampleFind() {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Find first even number greater than 5
	result := generic.Find(numbers, func(n int) bool {
		return n%2 == 0 && n > 5
	})
	value, found := result.Value()
	fmt.Printf("Found: %v, Value: %v\n", found, value)

	// Find element that doesn't exist
	notFound := generic.Find(numbers, func(n int) bool {
		return n > 100
	})
	fmt.Printf("Not found: %v\n", notFound.IsNone())

	// Output:
	// Found: true, Value: 6
	// Not found: true
}

// ExampleMin demonstrates finding minimum values.
func ExampleMin() {
	numbers := []int{5, 2, 8, 1, 9, 3}
	result := generic.Min(numbers)
	value, _ := result.Value()
	fmt.Printf("Min: %v\n", value)

	// With empty slice
	empty := generic.Min([]int{})
	fmt.Printf("Empty IsNone: %v\n", empty.IsNone())

	// Output:
	// Min: 1
	// Empty IsNone: true
}

// ExampleMax demonstrates finding maximum values.
func ExampleMax() {
	numbers := []int{5, 2, 8, 1, 9, 3}
	result := generic.Max(numbers)
	value, _ := result.Value()
	fmt.Printf("Max: %v\n", value)

	// Output:
	// Max: 9
}

// ExampleSum demonstrates summing numeric slices.
func ExampleSum() {
	integers := []int{1, 2, 3, 4, 5}
	floats := []float64{1.1, 2.2, 3.3}

	fmt.Printf("Sum of integers: %v\n", generic.Sum(integers))
	fmt.Printf("Sum of floats: %.1f\n", generic.Sum(floats))

	// Output:
	// Sum of integers: 15
	// Sum of floats: 6.6
}
