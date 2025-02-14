package mathex

import "time"

// Types of numbers that may be used in routines.
type Number interface {
	int | int32 | int64 | float32 | float64 | time.Duration
}

// Min returns the minimum number of two numbers.
func Min[N Number](v1 N, v2 N) N {
	if v1 < v2 {
		return v1
	}
	return v2
}

// Max returns the maximum value of two numbers.
func Max[N Number](v1 N, v2 N) N {
	if v1 > v2 {
		return v1
	}
	return v2
}
