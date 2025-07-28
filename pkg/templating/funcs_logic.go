package templating

import (
	"math/rand/v2"
	"reflect"
)

// repeat returns a slice of integers from 0 to count-1.
func repeat(count int) []int {
	if count < 0 {
		return []int{}
	}
	s := make([]int, count)
	for i := 0; i < count; i++ {
		s[i] = i
	}
	return s
}

// list returns a slice containing all the arguments passed to it.
func list(args ...any) []any {
	return args
}

// randomChoice selects and returns a single random element from a slice.
func randomChoice(slice any) any {
	if slice == nil {
		return nil
	}

	// Use reflection to inspect the type of the provided interface.
	val := reflect.ValueOf(slice)

	// Ensure we were actually given a slice.
	if val.Kind() != reflect.Slice {
		// Fail silently here
		// Otherwise we have to make this dependent on a TemplateManager instance to use the logger
		return nil
	}

	// Ensure the slice is not empty.
	if val.Len() == 0 {
		return nil
	}

	// Select a random index and return the element at that index.
	randomIndex := rand.IntN(val.Len())
	return val.Index(randomIndex).Interface()
}

// randomInt returns a random integer within the range [min, max).
func randomInt(min, max int) int {
	if min >= max {
		return min
	}
	return rand.IntN(max-min) + min
}
