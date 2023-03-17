// Package intutil provides integer utilities.
package intutil

// Min returns the minimum of a and b.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of a and b.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
