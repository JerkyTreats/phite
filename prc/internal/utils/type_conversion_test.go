package utils

import (
	"math"
	"testing"
)

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil", nil, ""},
		{"string", "test", "test"},
		{"bytes", []byte("test"), "test"},
		{"int", 123, ""},
		{"float", 123.45, ""},
		{"bool", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToString(tt.input)
			if result != tt.expected {
				t.Errorf("ToString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{"nil", nil, 0},
		{"float64", float64(123.45), 123.45},
		{"float32", float32(123.45), float64(float32(123.45))},
		{"int", 123, 123},
		{"int64", int64(123), 123},
		{"int32", int32(123), 123},
		{"uint", uint(123), 123},
		{"uint64", uint64(123), 123},
		{"uint32", uint32(123), 123},
		{"string_valid", "123.45", 123.45},
		{"string_invalid", "not a number", 0},
		{"bool", true, 0},
		{"bytes", []byte("123.45"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToFloat64(tt.input)
			if !floatEquals(result, tt.expected) {
				t.Errorf("ToFloat64(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int64
	}{
		{"nil", nil, 0},
		{"int64", int64(123), 123},
		{"int", 123, 123},
		{"int32", int32(123), 123},
		{"uint", uint(123), 123},
		{"uint64", uint64(123), 123},
		{"uint32", uint32(123), 123},
		{"float64", float64(123.45), 123},
		{"float32", float32(123.45), 123},
		{"string_valid", "123", 123},
		{"string_invalid", "not a number", 0},
		{"bool", true, 0},
		{"bytes", []byte("123"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToInt64(tt.input)
			if result != tt.expected {
				t.Errorf("ToInt64(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// floatEquals compares two float64 values with a small epsilon to account for floating-point precision
func floatEquals(a, b float64) bool {
	epsilon := 1e-10
	return math.Abs(a-b) < epsilon
}
