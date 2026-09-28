package robusto

import (
	"math"
	"slices"

	"github.com/victoralmeida428/estatistica_robusta/utils"
)

const (
	// convergenceTolerance é a tolerância relativa (à dispersão) usada como
	// critério de parada do Algorithm A e da média de Hampel.
	convergenceTolerance = 1e-10
	// maxIterations limita as iterações do Algorithm A e da média de Hampel.
	maxIterations = 10_000
	// algorithmAFactor é o fator de correção de consistência do Algorithm A
	// para winsorização em ±1.5 s* (ISO 13528:2022, C.3.1).
	algorithmAFactor = 1.134
)

// Statistics encapsula os dados de entrada e expõe métodos de estimação robusta.
type Statistics struct {
	data   []float64
	sorted []float64
}

// New cria uma nova instância de Statistics com uma cópia dos dados fornecidos.
func New(data []float64) *Statistics {
	s := &Statistics{}
	s.SetData(data)
	return s
}

// SetData substitui os dados internos da instância por uma cópia de data.
func (s *Statistics) SetData(data []float64) {
	s.data = slices.Clone(data)
	s.sorted = slices.Clone(data)
	slices.Sort(s.sorted)
}

func (s Statistics) median() float64 {
	return utils.QuantileSorted(s.sorted, 0.5)
}

// made retorna o desvio absoluto mediano escalado (MAD × 1.4826).
func (s Statistics) made(median float64) float64 {
	dist := make([]float64, len(s.data))
	for i, value := range s.data {
		dist[i] = math.Abs(value - median)
	}
	return 1.4826 * utils.Quantile(dist, 0.5)
}

// Qn implementa o estimador Qn de Rousseeuw-Croux (1993).
// É uma medida de escala robusta com ponto de ruptura de 50%.
//
// O algoritmo:
//  1. Calcula todas as N = n(n-1)/2 diferenças absolutas pareadas.
//  2. Toma a k-ésima diferença ordenada, onde k = C(h,2),
//     com h = floor(n/2) + 1.
//  3. Aplica fator de correção bp (tabelado para n ≤ 12, fórmula
//     assintótica para n > 12).
//  4. Escala por 2.21914 para consistência com o desvio padrão sob normalidade.
//  5. Estima a média via Hampel (HampelMean).
//
// Retorna NaN para menos de 2 valores.
//
// Referência: Rousseeuw, P.J. & Croux, C. (1993). JASA, 88(424), 1273-1283.
func (s Statistics) Qn() (mean float64, std float64) {
	std = s.qnScale()
	return s.HampelMean(std), std
}

func (s Statistics) qnScale() float64 {
	p := len(s.data)
	if p < 2 {
		return math.NaN()
	}

	h := p/2 + 1
	dk := kthPairwiseDiff(s.sorted, h*(h-1)/2)

	var bp float64
	if p <= 12 {
		list := []float64{0.399356, 0.99365, 0.51321, 0.84401, 0.61220, 0.85877, 0.66993, 0.87344, 0.72014, 0.88906, 0.75743}
		bp = list[p-2]
	} else {
		pf := float64(p)
		var rp float64
		if p%2 == 0 {
			rp = (1 / pf) * (3.67561 + (1/pf)*(1.9654+(1/pf)*(6.987-(77/pf))))
		} else {
			rp = (1 / pf) * (1.60188 + (1/pf)*(-2.1284-(5.172/pf)))
		}
		bp = 1 / (rp + 1)
	}

	return 2.21914 * dk * bp
}

// QMethod estima média e desvio padrão robustos conforme a ISO 13528:2022
// (C.5), com um resultado por laboratório. Para réplicas por laboratório,
// use QReplicates.
//
// Algoritmo:
//  1. Gera todas as diferenças absolutas pareadas.
//  2. Constrói a FDA empírica H(x) das diferenças e H(0), a proporção de
//     diferenças nulas.
//  3. Define G(x) nos pontos de descontinuidade positivos x₁ < x₂ < …:
//     G(0) = 0, G(x₁) = [H(x₁) + H(0)]/2, G(xᵢ) = [H(xᵢ) + H(xᵢ₋₁)]/2.
//  4. Obtém G⁻¹(y) por interpolação linear, com y = 0.25 + 0.75·H(0).
//  5. Estima o desvio padrão como G⁻¹(y) / [√2 · Φ⁻¹(0.625 + 0.375·H(0))].
//  6. Estima a média via Hampel (HampelMean).
//
// Retorna NaN para menos de 2 valores.
//
// Referência: ISO 13528:2022, Uhlig (2015), Liu et al. (2019).
func (s Statistics) QMethod() (mean float64, std float64) {
	std = s.qScale()
	return s.HampelMean(std), std
}

func (s Statistics) qScale() float64 {
	if len(s.data) < 2 {
		return math.NaN()
	}
	groups := make([][]float64, len(s.data))
	for i, value := range s.data {
		groups[i] = []float64{value}
	}
	std, _ := qScale(groups)
	return std
}

// algorithmAStep executa uma iteração do Algorithm A da ISO 13528:
// winsoriza os dados ajustando valores além de ±1.5σ para o limite,
// depois recalcula média e desvio padrão (este multiplicado por 1.134).
// A winsorização é refeita em cada passagem para não alocar memória.
func (s Statistics) algorithmAStep(center, scale float64) (mean float64, std float64) {
	lower, upper := center-1.5*scale, center+1.5*scale

	sum := 0.0
	for _, value := range s.data {
		sum += min(max(value, lower), upper)
	}
	mean = sum / float64(len(s.data))

	squares := 0.0
	for _, value := range s.data {
		d := min(max(value, lower), upper) - mean
		squares += d * d
	}
	return mean, math.Sqrt(squares/float64(len(s.data)-1)) * algorithmAFactor
}

// AlgorithmA implementa o Algorithm A da ISO 13528.
//
// A estimativa inicial usa mediana e MADe (1.4826 × MAD). Se iter=true, o
// algoritmo itera até que média e desvio variem menos que 1e-10 × s* entre
// passos (ou 10.000 iterações); se iter=false, retorna apenas a primeira
// iteração, que não é o estimador da norma.
//
// Retorna NaN para menos de 2 valores.
//
// Referência: ISO 13528:2022, Anexo C.3.
func (s Statistics) AlgorithmA(iter bool) (mean float64, std float64) {
	limit := 1
	if iter {
		limit = maxIterations
	}
	mean, std, _, _ = s.algorithmA(limit)
	return
}

func (s Statistics) algorithmA(limit int) (mean, std float64, iterations int, converged bool) {
	if len(s.data) < 2 {
		return math.NaN(), math.NaN(), 0, false
	}

	mean = s.median()
	std = s.made(mean)

	for iterations = 1; iterations <= limit; iterations++ {
		newMean, newStd := s.algorithmAStep(mean, std)
		change := max(math.Abs(newMean-mean), math.Abs(newStd-std))
		mean, std = newMean, newStd
		if change <= convergenceTolerance*std {
			return mean, std, iterations, true
		}
	}

	return mean, std, limit, false
}

// Classical retorna a média aritmética e o desvio padrão amostral de todos
// os dados, sem nenhum tratamento de outliers.
func (s Statistics) Classical() (mean float64, std float64) {
	if len(s.data) < 2 {
		return math.NaN(), math.NaN()
	}
	mean = utils.Mean(s.data)
	return mean, utils.Std(s.data, &mean)
}

// tukeyFences retorna as cercas de Tukey (Q1 - 1.5×IQR, Q3 + 1.5×IQR),
// com quartis tipo R-7.
func (s Statistics) tukeyFences() (lower, upper float64) {
	q1 := utils.QuantileSorted(s.sorted, 0.25)
	q3 := utils.QuantileSorted(s.sorted, 0.75)
	iqr := q3 - q1
	return q1 - 1.5*iqr, q3 + 1.5*iqr
}

// Outliers retorna, na ordem original, os valores fora das cercas de Tukey
// usadas por Traditional.
func (s Statistics) Outliers() []float64 {
	if len(s.data) == 0 {
		return nil
	}
	lower, upper := s.tukeyFences()
	outliers := make([]float64, 0)
	for _, value := range s.data {
		if value < lower || upper < value {
			outliers = append(outliers, value)
		}
	}
	return outliers
}

func (s Statistics) withoutOutliers() []float64 {
	if len(s.data) == 0 {
		return nil
	}
	lower, upper := s.tukeyFences()
	return slices.DeleteFunc(slices.Clone(s.data), func(value float64) bool {
		return value < lower || upper < value
	})
}

// Traditional remove outliers com as cercas de Tukey (Q1 - 1.5×IQR,
// Q3 + 1.5×IQR) e calcula média e desvio padrão clássicos nos dados
// filtrados.
//
// Os quartis são do tipo R-7; o boxplot.stats do R usa as dobradiças de
// Tukey (fivenum), que podem detectar outliers diferentes para alguns n.
func (s Statistics) Traditional() (mean float64, std float64) {
	filtered := s.withoutOutliers()
	if len(filtered) < 2 {
		return math.NaN(), math.NaN()
	}
	mean = utils.Mean(filtered)
	return mean, utils.Std(filtered, &mean)
}

// DamN calcula a mediana e o desvio absoluto mediano escalado (MAD × 1.4826).
// O fator 1.4826 torna o MAD um estimador consistente do desvio padrão
// sob a distribuição normal.
func (s Statistics) DamN() (median float64, dam float64) {
	median = s.median()
	return median, s.made(median)
}

// NiQr retorna a mediana e o IQR normalizado (IQR / 1.349).
// A divisão por 1.349 torna o IQR um estimador consistente do desvio
// padrão sob a distribuição normal padrão.
func (s Statistics) NiQr() (median float64, nIQR float64) {
	median = s.median()
	q1 := utils.QuantileSorted(s.sorted, 0.25)
	q3 := utils.QuantileSorted(s.sorted, 0.75)
	return median, (q3 - q1) / 1.349
}

// hampelStep calcula a média ponderada estilo Hampel com pesos decrescentes
// baseados no z-score absoluto |xᵢ - ref| / std:
//   - |z| ≤ 1.5  → peso = 1
//   - 1.5 < |z| ≤ 3 → peso = 1.5 / |z|
//   - 3 < |z| ≤ 4.5 → peso = (4.5 - |z|) / |z|
//   - |z| > 4.5 → peso = 0 (excluído)
//
// ok é false quando todos os pesos são zero.
func (s Statistics) hampelStep(std, ref float64) (mean float64, ok bool) {
	numerator := 0.0
	denominator := 0.0
	for _, value := range s.data {
		q := math.Abs((value - ref) / std)
		var w float64
		switch {
		case q <= 1.5:
			w = 1
		case q <= 3:
			w = 1.5 / q
		case q <= 4.5:
			w = (4.5 - q) / q
		}
		numerator += value * w
		denominator += w
	}
	if denominator == 0 {
		return ref, false
	}
	return numerator / denominator, true
}

// HampelMean retorna a média robusta de Hampel (ISO 13528:2022, C.5.3) para
// uma dispersão robusta fixa std (por exemplo s* do método Q, Qn ou MADe).
//
// Parte da mediana e itera até que a variação entre passos seja no máximo
// 1e-10 × std (ou 10.000 iterações). Se std for zero, retorna a mediana;
// se std for NaN ou não houver dados, retorna NaN.
func (s Statistics) HampelMean(std float64) float64 {
	mean, _, _ := s.hampel(std)
	return mean
}

func (s Statistics) hampel(std float64) (mean float64, iterations int, converged bool) {
	if len(s.data) == 0 || math.IsNaN(std) {
		return math.NaN(), 0, false
	}
	mean = s.median()
	if std == 0 {
		return mean, 0, true
	}

	for iterations = 1; iterations <= maxIterations; iterations++ {
		newMean, ok := s.hampelStep(std, mean)
		if !ok {
			return mean, iterations, false
		}
		change := math.Abs(mean - newMean)
		mean = newMean
		if change <= convergenceTolerance*std {
			return mean, iterations, true
		}
	}

	return mean, maxIterations, false
}
