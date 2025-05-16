package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMeanStd(t *testing.T) {
	array1 := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	array2 := []float64{1, 2, 3, 4, 5}
	array3 := []float64{1, 2, 3}
	testCases := []struct {
		name      string
		input     []float64
		mean, std float64
	}{
		{
			name:  "array1",
			input: array1,
			mean:  5.5,
			std:   3.0276503540974917,
		},
		{
			name:  "array2",
			input: array2,
			mean:  3,
			std:   1.5811388300841898,
		},
		{
			name:  "array3",
			input: array3,
			mean:  2,
			std:   1,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mean := Mean(tc.input)
			std := Std(tc.input, nil)
			assert.Equal(t, tc.mean, mean)
			assert.InDelta(t, tc.std, std, 0.00000001)
			
			mean2 := Mean(tc.input)
			std2 := Std(tc.input, &mean)
			assert.Equal(t, tc.mean, mean2)
			assert.InDelta(t, tc.std, std2, 0.00000001)
		})
	}
	
}

func TestFatorial(t *testing.T) {
	testCases := []struct {
		name   string
		input  int
		expect int
	}{
		{
			name:   "fatorial de 5",
			input:  5,
			expect: 120,
		},
		{
			name:   "fatorial de 6",
			input:  6,
			expect: 720,
		},
		{
			name:   "fatorial de 0",
			input:  0,
			expect: 1,
		},
		{
			name:   "fatorial de 1",
			input:  1,
			expect: 1,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Fatorial(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, result)
		})
	}
	
	t.Run("fatorial -1", func(t *testing.T) {
		_, err := Fatorial(-1)
		assert.Error(t, err)
	})
	
}

func BenchmarkFatoriais(b *testing.B) {
	n := 1000
	b.Run("Iterativo", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Fatorial(n)
		}
	})
	
	b.Run("Iterativo-Inline", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := 1
			for j := 2; j <= n; j++ {
				result *= j
			}
			_ = result
		}
	})
}

func TestPPF(t *testing.T) {
	
	t.Run("h0 0", func(t *testing.T) {
		
		h0 := 0.0
		assert.InDelta(t, 0.31863936396437514, NormPPF(0.625+0.375*h0), 1e-8)
	})
	t.Run("h0 0.013071895424836602", func(t *testing.T) {
		
		h0 := 0.013071895424836602
		assert.InDelta(t, 0.33159369857326043, NormPPF(0.625+0.375*h0), 1e-8)
	})
	
}
