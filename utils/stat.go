package utils

import (
	"errors"
	"math"
)

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
		diffSquare += math.Pow(v-newMean, 2)
	}
	return math.Sqrt(diffSquare / float64(count-1))
}

// Fatorial calcula n! iterativamente. Retorna erro se n < 0.
func Fatorial(n int) (int, error) {
	if n < 0 {
		return 0, errors.New("n must be greater than or equal 0")
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result, nil
}

// NormPPF é a função quantil (inversa da CDF) da distribuição normal padrão.
//
// Implementa a aproximação racional de Peter Acklam, dividida em três regiões:
//   - Cauda inferior  (p < 0.02425): polinômio em √(-2 ln p)
//   - Cauda superior  (p > 0.97575): polinômio em √(-2 ln(1-p)), sinal negado
//   - Região central  (demais): razão de polinômios de 6º/5º grau em (p-0.5)²
//
// Precisão: ~1.15 × 10⁻⁹ nas caudas, ~2.2 × 10⁻¹⁶ no centro.
// Retorna NaN se p ≤ 0 ou p ≥ 1.
//
// Referência: Acklam, P.J. "An algorithm for computing the inverse normal
// cumulative distribution function".
func NormPPF(p float64) float64 {
	if p <= 0 || p >= 1 {
		return math.NaN()
	}

	const (
		a1 = -3.969683028665376e+01
		a2 = 2.209460984245205e+02
		a3 = -2.759285104469687e+02
		a4 = 1.383577518672690e+02
		a5 = -3.066479806614716e+01
		a6 = 2.506628277459239e+00

		b1 = -5.447609879822406e+01
		b2 = 1.615858368580409e+02
		b3 = -1.556989798598866e+02
		b4 = 6.680131188771972e+01
		b5 = -1.328068155288572e+01

		c1 = -7.784894002430293e-03
		c2 = -3.223964580411365e-01
		c3 = -2.400758277161838e+00
		c4 = -2.549732539343734e+00
		c5 = 4.374664141464968e+00
		c6 = 2.938163982698783e+00

		d1 = 7.784695709041462e-03
		d2 = 3.224671290700398e-01
		d3 = 2.445134137142996e+00
		d4 = 3.754408661907416e+00
	)

	switch {
	case p < 0.02425:
		q := math.Sqrt(-2 * math.Log(p))
		return ((((c1*q+c2)*q+c3)*q+c4)*q+c5)*q + c6

	case p > 0.97575:
		q := math.Sqrt(-2 * math.Log(1-p))
		return -(((((c1*q+c2)*q+c3)*q+c4)*q+c5)*q + c6)

	default:
		q := p - 0.5
		r := q * q
		return (((((a1*r+a2)*r+a3)*r+a4)*r+a5)*r + a6) * q /
			(((((b1*r+b2)*r+b3)*r+b4)*r+b5)*r + 1)
	}
}
