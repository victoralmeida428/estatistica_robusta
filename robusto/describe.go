package robusto

import (
	"math"

	"github.com/victoralmeida428/estatistica_robusta/utils"
)

// Summary reúne as estatísticas descritivas de posição e dispersão.
type Summary struct {
	N                  int
	Min, Max           float64
	Q1, Median, Q3     float64
	Mean, SD, Variance float64
	// CV é o coeficiente de variação SD/Mean.
	CV float64
	// Mode é o valor mais frequente; em empate, o primeiro na ordem original.
	Mode float64
	// MAD é o desvio absoluto mediano sem escala; MADe = 1.4826 × MAD.
	MAD, MADe float64
	// IQR = Q3 - Q1; NIQR = IQR / 1.349.
	IQR, NIQR float64
	// Skewness é o coeficiente de assimetria amostral g₁ = m₃ / m₂^1.5 (o
	// mesmo de moments::skewness no R).
	Skewness float64
}

// Describe calcula as estatísticas descritivas dos dados. Quartis são do
// tipo R-7. Retorna ErrTooFewValues com menos de 2 valores.
func (s Statistics) Describe() (Summary, error) {
	n := len(s.data)
	if n < 2 {
		return Summary{N: n}, ErrTooFewValues
	}

	sum := Summary{
		N:      n,
		Min:    s.sorted[0],
		Max:    s.sorted[n-1],
		Q1:     utils.QuantileSorted(s.sorted, 0.25),
		Median: s.median(),
		Q3:     utils.QuantileSorted(s.sorted, 0.75),
	}
	sum.Mean, sum.SD = s.Classical()
	sum.Variance = sum.SD * sum.SD
	sum.CV = sum.SD / sum.Mean
	sum.MADe = s.made(sum.Median)
	sum.MAD = sum.MADe / 1.4826
	sum.IQR = sum.Q3 - sum.Q1
	sum.NIQR = sum.IQR / 1.349

	counts := make(map[float64]int)
	best := 0
	for _, value := range s.data {
		counts[value]++
		if counts[value] > best {
			best = counts[value]
		}
	}
	for _, value := range s.data {
		if counts[value] == best {
			sum.Mode = value
			break
		}
	}

	var m2, m3 float64
	for _, value := range s.data {
		d := value - sum.Mean
		m2 += d * d
		m3 += d * d * d
	}
	m2 /= float64(n)
	m3 /= float64(n)
	sum.Skewness = m3 / math.Pow(m2, 1.5)

	return sum, nil
}

// Bandwidth retorna a largura de banda 0.9 · scale / n^0.2 usada na
// inspeção de modas por densidade de kernel (ISO 13528:2022, 10.3), com
// scale = nIQR ou MADe. Para a largura baseada em σpt use 0.75 · σpt.
func Bandwidth(scale float64, n int) float64 {
	return 0.9 * scale / math.Pow(float64(n), 0.2)
}

// KernelDensity estima a densidade de kernel gaussiano dos dados, com
// largura de banda h, em cada ponto de at.
func (s Statistics) KernelDensity(h float64, at []float64) []float64 {
	density := make([]float64, len(at))
	norm := 1 / (float64(len(s.data)) * h * math.Sqrt(2*math.Pi))
	for i, x := range at {
		for _, value := range s.data {
			z := (x - value) / h
			density[i] += math.Exp(-0.5 * z * z)
		}
		density[i] *= norm
	}
	return density
}
