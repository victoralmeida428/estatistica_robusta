package utils

import (
	"math"
	"slices"
)

// Quantile calcula o quantil amostral por interpolação linear (tipo R-7,
// Hyndman & Fan, 1996 — o método padrão do R e do numpy).
//
// Dado um vetor ordenado x[0..n-1] e uma probabilidade q em [0,1]:
//   pos = q × (n-1)
//   j = floor(pos), g = pos - j
//   retorna (1 - g) × x[j] + g × x[j+1]
//
// Casos de borda: q = 0 retorna o mínimo, q = 1 retorna o máximo.
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
