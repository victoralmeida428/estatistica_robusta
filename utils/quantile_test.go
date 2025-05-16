package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQuantile(t *testing.T) {
	array1 := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	array2 := []float64{1, 2, 3, 4, 5}
	array3 := []float64{1, 2, 3}
	testCases := []struct {
		name     string
		input    []float64
		q        float64
		expected float64
	}{
		{
			name:     "array1 quantile 0,25",
			input:    array1,
			q:        0.25,
			expected: 3.25,
		},
		{
			name:     "array2 quantile 0,25",
			input:    array2,
			q:        0.25,
			expected: 2,
		},
		{
			name:     "array3 quantile 0,25",
			input:    array3,
			q:        0.25,
			expected: 1.5,
		},
		{
			name:     "array1 quantile 0,5",
			input:    array1,
			q:        0.5,
			expected: 5.5,
		},
		{
			name:     "array2 quantile 0,5",
			input:    array2,
			q:        0.5,
			expected: 3,
		},
		{
			name:     "array3 quantile 0,5",
			input:    array3,
			q:        0.5,
			expected: 2,
		},
		{
			name:     "array1 quantile 0,75",
			input:    array1,
			q:        0.75,
			expected: 7.75,
		},
		{
			name:     "array2 quantile 0,75",
			input:    array2,
			q:        0.75,
			expected: 4,
		},
		{
			name:     "array3 quantile 0,75",
			input:    array3,
			q:        0.75,
			expected: 2.5,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := Quantile(tc.input, tc.q)
			assert.InDelta(t, tc.expected, output, 0.0001)
		})
	}
	
}
