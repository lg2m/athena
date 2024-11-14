package util

import "testing"

func TestCalcProgress(t *testing.T) {
	tests := []struct {
		name string
		tot  int
		curr int
		want int
	}{
		{
			name: "zero total",
			tot:  0,
			curr: 5,
			want: 0,
		},
		{
			name: "half progress",
			tot:  100,
			curr: 50,
			want: 50,
		},
		{
			name: "complete",
			tot:  100,
			curr: 100,
			want: 100,
		},
		{
			name: "rounding up",
			tot:  3,
			curr: 2,
			want: 67, // (2/3 * 100) + 0.5 ≈ 67
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcProgress(tt.tot, tt.curr)
			if got != tt.want {
				t.Errorf("CalcProgress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		value, min, max, expected int
	}{
		{value: 5, min: 1, max: 10, expected: 5},   // Within range
		{value: -3, min: 1, max: 10, expected: 1},  // Below min
		{value: 15, min: 1, max: 10, expected: 10}, // Above max
		{value: 1, min: 1, max: 10, expected: 1},   // Exactly at min
		{value: 10, min: 1, max: 10, expected: 10}, // Exactly at max
	}

	for _, tt := range tests {
		result := Clamp(tt.value, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("Clamp(%d, %d, %d) = %d; want %d", tt.value, tt.min, tt.max, result, tt.expected)
		}
	}
}
