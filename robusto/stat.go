package robusto

import (
	"math"
	"github.com/victoralmeida428/estatistica_robusta/utils"
	"slices"
)

const tolerancia = 1e-4

// Statistics encapsula os dados de entrada e expõe métodos de estimação robusta.
type Statistics struct {
	data []float64
}

// getDiff calcula todas as diferenças absolutas entre pares de elementos
// (|x_j - x_i| para todo j > i) e retorna o vetor ordenado.
func (s Statistics) getDiff(p int) []float64 {
	diff := make([]float64, 0)
	for i := range s.data {
		for j := i + 1; j < p; j++ {
			diff = append(diff, math.Abs(s.data[j]-s.data[i]))
		}
	}
	slices.Sort(diff)
	return diff
}

// Qn implementa o estimador Qn de Rousseeuw-Croux (1993).
// É uma medida de escala robusta com ponto de ruptura de 50%.
//
// O algoritmo:
//  1. Calcula todas as N = n(n-1)/2 diferenças absolutas pareadas.
//  2. Toma a k-ésima diferença ordenada, onde k = C(h,2) + 1,
//     com h = floor(n/2) + 1.
//  3. Aplica fator de correção bp (tabelado para n ≤ 12, fórmula
//     assintótica para n > 12).
//  4. Escala por 2.21914 para consistência com o desvio padrão sob normalidade.
//  5. Estima a média via Hampel (hampelMean).
//
// Referência: Rousseeuw, P.J. & Croux, C. (1993). JASA, 88(424), 1273-1283.
func (s Statistics) Qn() (mean float64, std float64) {
	p := len(s.data)
	diff := s.getDiff(p)
	if len(diff) == 1 {
		return diff[0], 0
	}

	var h, k int

	if p%2 == 0 {
		h = (p / 2) + 1
	} else {
		h = ((p - 1) / 2) + 1
	}

	k = h * (h - 1) / 2
	dk := diff[k-1]

	var bp float64

	if p >= 2 && p <= 12 {
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

	std = 2.21914 * dk * bp
	return s.hampelMean(std), std
}

// QMethod estima média e desvio padrão robustos conforme a ISO 13528,
// utilizando a distribuição empírica das diferenças absolutas entre pares.
//
// Algoritmo:
//  1. Gera todas as diferenças absolutas pareadas.
//  2. Constrói a FDA empírica H(d) = 2·fac(d) / n(n-1), onde fac é a
//     frequência acumulada de cada diferença.
//  3. Define G(d) como a função escada linearizada (pontos médios entre H(d_i) e H(d_{i-1})).
//  4. Obtém d* = G⁻¹(y) por interpolação linear, com y = 0.25 + 0.75·H(0).
//  5. Estima o desvio padrão como d* / [√2 · Φ⁻¹(0.625 + 0.375·H(0))].
//  6. Estima a média via Hampel (hampelMean).
//
// Referência: ISO 13528:2015, Uhlig (2015), Liu et al. (2019).
func (s Statistics) QMethod() (mean float64, std float64) {
	p := len(s.data)
	pf := float64(p)
	diff := s.getDiff(p)
	if len(diff) == 1 {
		return diff[0], 0
	}

	frequence := make(map[float64]int)

	for _, d := range diff {
		frequence[d]++
	}

	diffCompact := slices.Compact(diff)

	diffStruct := make([]statDiff, len(diffCompact))

	for i := range diffCompact {
		var ds statDiff
		ds.diff = diffCompact[i]
		ds.freq = frequence[diffCompact[i]]
		ds.fac = frequence[diffCompact[i]]
		if i > 0 {
			ds.fac += diffStruct[i-1].fac
		}
		diffStruct[i] = ds
	}

	diff0 := make([]statDiff, 0)
	if !slices.Contains(diffCompact, 0) {
		var ds statDiff
		diff0 = append(diff0, ds)
	}

	finalDiff := make([]statDiff, 0)
	finalDiff = append(finalDiff, diff0...)
	finalDiff = append(finalDiff, diffStruct...)

	var h0 float64

	gx := make([]float64, len(finalDiff))
	for i, ds := range finalDiff {
		hx := 2 / (pf * (pf - 1)) * float64(ds.fac)
		finalDiff[i].hx = hx
		if i == 0 {
			h0 = hx

			finalDiff[i].gx = hx * 0.5
		} else {
			finalDiff[i].gx = (hx + finalDiff[i-1].hx) * 0.5
		}
		gx[i] = finalDiff[i].gx
	}
	// Encontra G⁻¹(y) por interpolação linear
	y := 0.25 + 0.75*h0
	gx = append(gx, y)
	slices.Sort(gx)
	var yIndex int
	for i := range gx {
		if gx[i] == y {
			yIndex = i
		}
	}
	numerador := finalDiff[yIndex-1].diff + (finalDiff[yIndex].diff-finalDiff[yIndex-1].diff)/(finalDiff[yIndex].gx-finalDiff[yIndex-1].gx)*(y-finalDiff[yIndex-1].gx)
	denominador := math.Sqrt(2) * utils.NormPPF(0.625+0.375*h0)
	std = numerador / denominador
	return s.hampelMean(std), std
}

// recurrenceAlgorithmA executa uma iteração do Algorithm A da ISO 13528:
// winsoriza os dados ajustando valores além de ±1.5σ para o limite,
// depois recalcula média e desvio padrão (este multiplicado por 1.134).
func (s Statistics) recurrenceAlgorithmA(median, dam float64) (mean float64, std float64) {
	sigma := 1.5 * dam
	finalData := make([]float64, len(s.data))

	for i, value := range s.data {
		if value > (median + sigma) {
			finalData[i] = median + sigma
		} else if value < (median - sigma) {
			finalData[i] = median - sigma
		} else {
			finalData[i] = value
		}
	}

	mean = utils.Mean(finalData)
	return mean, utils.Std(finalData, &mean) * 1.134
}

// AlgorithmA implementa o Algorithm A da ISO 13528.
//
// A estimativa inicial usa mediana e MAD. Se iter=true, o algoritmo
// itera até convergência (diferença < 0.0001 entre passos) ou 100.000
// iterações, o que ocorrer primeiro.
//
// Referência: ISO 13528:2015, Anexo C.
func (s Statistics) AlgorithmA(iter bool) (mean float64, std float64) {

	median := utils.Quantile(s.data, 0.5)
	distMedian := make([]float64, len(s.data))
	for i, value := range s.data {
		distMedian[i] = math.Abs(value - median)
	}
	dam := 1.4826 * utils.Quantile(distMedian, 0.5)

	mean, std = s.recurrenceAlgorithmA(median, dam)

	if iter {
		for i := 0; i < 100_000; i++ {
			newMean, newStd := s.recurrenceAlgorithmA(mean, std)
			if (math.Abs(mean-newMean) < tolerancia) && (math.Abs(std-newStd) < tolerancia) {
				return newMean, newStd
			}
			mean = newMean
			std = newStd

		}
	}

	return
}

// Traditional remove outliers com as cercas de Tukey (Q1 - 1.5×IQR,
// Q3 + 1.5×IQR) e calcula média e desvio padrão clássicos nos dados
// filtrados.
func (s Statistics) Traditional() (mean float64, std float64) {
	q1 := utils.Quantile(s.data, 0.25)
	q3 := utils.Quantile(s.data, 0.75)
	iqr := q3 - q1
	cerca_inf := q1 - 1.5*iqr
	cerca_sup := q3 + 1.5*iqr

	filtered := slices.DeleteFunc(slices.Clone(s.data), func(value float64) bool {
		return value < cerca_inf || cerca_sup < value
	})

	mean = utils.Mean(filtered)
	std = utils.Std(filtered, &mean)
	filtered = nil
	return mean, std
}

// DamN calcula a mediana e o desvio absoluto mediano escalado (MAD × 1.4826).
// O fator 1.4826 torna o MAD um estimador consistente do desvio padrão
// sob a distribuição normal.
func (s Statistics) DamN() (median float64, dam float64) {
	median = utils.Quantile(s.data, 0.5)
	aux := make([]float64, len(s.data))
	for i := range aux {
		aux[i] = math.Abs(s.data[i] - median)
	}
	mad := utils.Quantile(aux, 0.5)
	dam = 1.4826 * mad
	return
}

// NiQr retorna a mediana e o IQR normalizado (IQR / 1.349).
// A divisão por 1.349 torna o IQR um estimador consistente do desvio
// padrão sob a distribuição normal padrão.
func (s Statistics) NiQr() (median float64, n_iqr float64) {
	median = utils.Quantile(s.data, 0.5)
	q1 := utils.Quantile(s.data, 0.25)
	q3 := utils.Quantile(s.data, 0.75)
	iqr := q3 - q1
	n_iqr = iqr / 1.349
	return
}

// SetData substitui os dados internos da instância.
func (s *Statistics) SetData(data []float64) {
	s.data = data
}

// findMean calcula a média ponderada estilo Hampel com pesos decrescentes
// baseados no z-score absoluto |xᵢ - ref| / std:
//   - |z| ≤ 1.5  → peso = 1
//   - 1.5 < |z| ≤ 3 → peso = 1.5 / |z|
//   - 3 < |z| ≤ 4.5 → peso = (4.5 - |z|) / |z|
//   - |z| > 4.5 → peso = 0 (excluído)
//
// Se ref for nil, usa a mediana como referência inicial.
func (s *Statistics) findMean(std float64, ref *float64) float64 {
	if ref == nil {
		refV := utils.Quantile(s.data, 0.5)
		ref = &refV
	}

	q := make([]float64, len(s.data))
	for i := range q {
		q[i] = math.Abs((s.data[i] - *ref) / std)
	}

	w := make([]float64, len(s.data))
	for i := range w {
		if q[i] > 3 && q[i] <= 4.5 {
			w[i] = (4.5 - q[i]) / q[i]
		}
		if q[i] > 1.5 && q[i] <= 3 {
			w[i] = 1.5 / q[i]
		}
		if q[i] <= 1.5 {
			w[i] = 1
		}
	}

	numerator := 0.0
	denominator := 0.0
	for i := range s.data {
		numerator += s.data[i] * w[i]
		denominator += w[i]
	}
	return numerator / denominator

}

// hampelMean itera o estimador de Hampel até convergência.
// O critério de parada é |mean - newMean| ≤ 0.01 × std / √n.
// Limite de 100.000 iterações como segurança.
func (s Statistics) hampelMean(std float64) float64 {
	erro := 0.01 * std / math.Sqrt(float64(len(s.data)))
	mean := s.findMean(std, nil)

	for i := 0; i < 100_000; i++ {
		newMean := s.findMean(std, &mean)
		if math.Abs(mean-newMean) <= erro {
			return newMean
		}
		mean = newMean
	}

	return mean
}

// New cria uma nova instância de Statistics com os dados fornecidos.
func New(data []float64) *Statistics {
	return &Statistics{data: data}
}
