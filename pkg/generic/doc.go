// Package generic provides type-safe generic utility functions.
//
// # Overview
//
// This package offers a collection of generic utility functions and types
// for common programming tasks:
//   - Result type for error handling
//   - Option type for optional values
//   - Slice operations (Map, Filter, Reduce, etc.)
//   - Numeric utilities (Min, Max, Sum)
//
// # Result Type
//
// The Result type represents the outcome of an operation that may fail:
//
//	func divide(a, b float64) generic.Result[float64] {
//	    if b == 0 {
//	        return generic.Err[float64](errors.New("division by zero"))
//	    }
//	    return generic.Ok(a / b)
//	}
//
//	result := divide(10, 2)
//	if val, err := result.Value(); err == nil {
//	    fmt.Println(val) // 5
//	}
//
// # Option Type
//
// The Option type represents an optional value:
//
//	func findUser(id int) generic.Option[User] {
//	    if user, found := cache[id]; found {
//	        return generic.Some(user)
//	    }
//	    return generic.None[User]()
//	}
//
//	opt := findUser(123)
//	if val, ok := opt.Value(); ok {
//	    fmt.Println(val.Name)
//	}
//
// # Slice Operations
//
// Transform and query slices with type-safe functions:
//
//	// Filter even numbers
//	numbers := []int{1, 2, 3, 4, 5, 6}
//	evens := generic.Filter(numbers, func(n int) bool { return n%2 == 0 })
//
//	// Map to strings
//	strs := generic.MapSlice(numbers, strconv.Itoa)
//
//	// Reduce to sum
//	sum := generic.Reduce(numbers, 0, func(acc, n int) int { return acc + n })
//
// # Grouping and Partitioning
//
//	// Group by first letter
//	words := []string{"apple", "banana", "apricot"}
//	groups := generic.GroupBy(words, func(s string) string { return string(s[0]) })
//
//	// Partition by predicate
//	evens, odds := generic.Partition(numbers, func(n int) bool { return n%2 == 0 })
//
// # Numeric Utilities
//
//	min := generic.Min(1, 2, 3)  // 1
//	max := generic.Max(1, 2, 3)  // 3
//	sum := generic.Sum(1, 2, 3)  // 6
package generic
