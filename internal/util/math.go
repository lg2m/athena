package util

// CalcProgress calculates the progress percentage.
func CalcProgress(tot, curr int) int {
	if tot == 0 {
		return 0
	}
	return int(((float64(curr) / float64(tot)) * 100) + 0.5)
}

// Clamp clamps a value within a range.
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
