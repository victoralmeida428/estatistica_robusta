package utils

import (
	"math"
	"slices"
)

func Quantile(data []float64, q float64) float64 {
	
	sorted := make([]float64, len(data))
	copy(sorted, data)
	slices.Sort(sorted)
	
	n := len(sorted)
	position := q * float64(n-1)
	j := int(math.Floor(position))
	g := position - float64(j)
	
	if j+1 >= n {
		return sorted[n-1]
	}
	if j < 0 {
		return sorted[0]
	}
	
	return (1-g)*sorted[j] + g*sorted[j+1]
}
