package fsrs

import "math"

// curveFactor keeps the forgetting curve and its inverse in nextInterval
// consistent: R(S, S) == 0.9 for any decay.
func curveFactor(decay float64) float64 {
	return math.Pow(0.9, 1/decay) - 1
}

// clamp restricts value to be within the range [min, max].
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
