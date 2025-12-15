// Package generic provides generic utility functions using Go 1.18+ generics.
// These functions demonstrate modern Go patterns for type-safe operations.
package generic

import (
	"cmp"
	"slices"
)

// Result represents a result type that can hold either a value or an error.
// This is inspired by Rust's Result type and provides explicit error handling.
type Result[T any] struct {
	value T
	err   error
	ok    bool
}

// Ok creates a successful Result with the given value.
func Ok[T any](value T) Result[T] {
	return Result[T]{value: value, ok: true}
}

// Err creates a failed Result with the given error.
func Err[T any](err error) Result[T] {
	return Result[T]{err: err, ok: false}
}

// IsOk returns true if the result is successful.
func (r Result[T]) IsOk() bool {
	return r.ok
}

// IsErr returns true if the result is an error.
func (r Result[T]) IsErr() bool {
	return !r.ok
}

// Unwrap returns the value if ok, panics otherwise.
func (r Result[T]) Unwrap() T {
	if !r.ok {
		panic("called Unwrap on an Err value")
	}
	return r.value
}

// UnwrapOr returns the value if ok, otherwise returns the default.
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.ok {
		return r.value
	}
	return defaultValue
}

// UnwrapErr returns the error.
func (r Result[T]) UnwrapErr() error {
	return r.err
}

// Value returns the value and a boolean indicating success.
func (r Result[T]) Value() (T, bool) {
	return r.value, r.ok
}

// Error returns the error, implementing the error interface.
func (r Result[T]) Error() error {
	return r.err
}

// Option represents an optional value that may or may not be present.
// This is inspired by Rust's Option type.
type Option[T any] struct {
	value   T
	present bool
}

// Some creates an Option with a value.
func Some[T any](value T) Option[T] {
	return Option[T]{value: value, present: true}
}

// None creates an empty Option.
func None[T any]() Option[T] {
	return Option[T]{present: false}
}

// IsSome returns true if the option contains a value.
func (o Option[T]) IsSome() bool {
	return o.present
}

// IsNone returns true if the option is empty.
func (o Option[T]) IsNone() bool {
	return !o.present
}

// Unwrap returns the value if present, panics otherwise.
func (o Option[T]) Unwrap() T {
	if !o.present {
		panic("called Unwrap on a None value")
	}
	return o.value
}

// UnwrapOr returns the value if present, otherwise returns the default.
func (o Option[T]) UnwrapOr(defaultValue T) T {
	if o.present {
		return o.value
	}
	return defaultValue
}

// Value returns the value and a boolean indicating presence.
func (o Option[T]) Value() (T, bool) {
	return o.value, o.present
}

// Map applies a function to the option value if present.
func Map[T, U any](opt Option[T], fn func(T) U) Option[U] {
	if opt.present {
		return Some(fn(opt.value))
	}
	return None[U]()
}

// Filter filters a slice based on a predicate function.
func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// MapSlice applies a function to each element of a slice.
func MapSlice[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, item := range slice {
		result[i] = fn(item)
	}
	return result
}

// Reduce reduces a slice to a single value using an accumulator function.
func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, item := range slice {
		result = fn(result, item)
	}
	return result
}

// Find finds the first element matching the predicate.
func Find[T any](slice []T, predicate func(T) bool) Option[T] {
	for _, item := range slice {
		if predicate(item) {
			return Some(item)
		}
	}
	return None[T]()
}

// FindIndex finds the index of the first element matching the predicate.
func FindIndex[T any](slice []T, predicate func(T) bool) int {
	for i, item := range slice {
		if predicate(item) {
			return i
		}
	}
	return -1
}

// Contains checks if a slice contains an element.
func Contains[T comparable](slice []T, element T) bool {
	return slices.Contains(slice, element)
}

// Unique returns a slice with duplicate elements removed.
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]struct{}, len(slice))
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// GroupBy groups slice elements by a key function.
func GroupBy[T any, K comparable](slice []T, keyFn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range slice {
		key := keyFn(item)
		result[key] = append(result[key], item)
	}
	return result
}

// Partition partitions a slice into two based on a predicate.
func Partition[T any](slice []T, predicate func(T) bool) (matching, notMatching []T) {
	matching = make([]T, 0)
	notMatching = make([]T, 0)
	for _, item := range slice {
		if predicate(item) {
			matching = append(matching, item)
		} else {
			notMatching = append(notMatching, item)
		}
	}
	return
}

// Min returns the minimum value from a slice of ordered values.
func Min[T cmp.Ordered](slice []T) Option[T] {
	if len(slice) == 0 {
		return None[T]()
	}
	return Some(slices.Min(slice))
}

// Max returns the maximum value from a slice of ordered values.
func Max[T cmp.Ordered](slice []T) Option[T] {
	if len(slice) == 0 {
		return None[T]()
	}
	return Some(slices.Max(slice))
}

// Sum calculates the sum of numeric values in a slice.
func Sum[T cmp.Ordered](slice []T) T {
	var sum T
	for _, item := range slice {
		sum += item
	}
	return sum
}

// Chunk splits a slice into chunks of the specified size.
func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return nil
	}
	chunks := make([][]T, 0, (len(slice)+size-1)/size)
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// Zip combines two slices into a slice of pairs.
func Zip[T, U any](a []T, b []U) []struct {
	First  T
	Second U
} {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	result := make([]struct {
		First  T
		Second U
	}, minLen)
	for i := 0; i < minLen; i++ {
		result[i] = struct {
			First  T
			Second U
		}{a[i], b[i]}
	}
	return result
}

// Keys returns the keys of a map.
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns the values of a map.
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// Ptr returns a pointer to the value.
func Ptr[T any](v T) *T {
	return &v
}

// Deref returns the value pointed to, or the zero value if nil.
func Deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// DerefOr returns the value pointed to, or the default if nil.
func DerefOr[T any](p *T, defaultValue T) T {
	if p == nil {
		return defaultValue
	}
	return *p
}
