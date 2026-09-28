package utils

import "math"

// Mean calcula a média aritmética simples do vetor.
func Mean(data []float64) float64 {
	sum := 0.0
	count := len(data)
	for _, v := range data {
		sum += v
	}
	return sum / float64(count)
}

// Std calcula o desvio padrão amostral (divisor n-1).
// Se mean não for nil, usa o valor fornecido; caso contrário calcula internamente.
func Std(data []float64, mean *float64) float64 {
	diffSquare := 0.0
	count := len(data)

	var newMean float64

	if mean == nil {
		newMean = Mean(data)
	} else {
		newMean = *mean
	}

	for _, v := range data {
		d := v - newMean
		diffSquare += d * d
	}
	return math.Sqrt(diffSquare / float64(count-1))
}

// NormPPF é a função quantil (inversa da CDF) da distribuição normal padrão.
//
// Usa a identidade Φ⁻¹(p) = -√2 · erfc⁻¹(2p), com precisão de máquina em
// todo o domínio (inclusive nas caudas, onde 2p-1 perderia dígitos).
// Retorna NaN se p ≤ 0 ou p ≥ 1.
func NormPPF(p float64) float64 {
	if p <= 0 || p >= 1 {
		return math.NaN()
	}
	return -math.Sqrt2 * math.Erfcinv(2*p)
}
