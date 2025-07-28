package templating

import "reflect"

// add returns a + b.
func add(a, b int) int {
	return a + b
}

// sub returns a - b.
func sub(a, b int) int {
	return a - b
}

// div returns a / b (integer division). Returns 0 if b is 0.
func div(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}

// mult returns a * b.
func mult(a, b int) int {
	return a * b
}

// max returns the maximum of a and b.
//
//goland:noinspection GoReservedWordUsedAsName
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of a and b.
//
//goland:noinspection GoReservedWordUsedAsName
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// mod returns a % b.
func mod(a, b int) int {
	if b == 0 {
		return 0
	}
	return a % b
}

// inc returns i + 1.
func inc(i int) int {
	return i + 1
}

// dec returns i - 1.
func dec(i int) int {
	return i - 1
}

// and returns true only if all arguments are true.
func and(args ...bool) bool {
	for _, arg := range args {
		if !arg {
			return false
		}
	}
	return true
}

// or returns true if any argument is true.
func or(args ...bool) bool {
	for _, arg := range args {
		if arg {
			return true
		}
	}
	return false
}

// not returns the boolean opposite of its argument.
func not(arg bool) bool {
	return !arg
}

// isSet returns true if a value is not its zero value.
func isSet(val any) bool {
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return false
	}
	return !v.IsZero()
}
