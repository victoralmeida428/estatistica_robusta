package robusto

import "slices"

// kthPairwiseDiff retorna a k-ésima menor (k a partir de 1) das n(n-1)/2
// diferenças y[j] - y[i], j > i, de um vetor ordenado y, em O(n log² n) e
// sem gerar todas as diferenças.
//
// Segue a ideia de Croux & Rousseeuw (1992): cada linha i da matriz
// implícita de diferenças é crescente em j; a cada passo, o pivô é a
// mediana ponderada das medianas dos candidatos restantes de cada linha, e
// a contagem de diferenças menores que o pivô é feita em O(n) com dois
// ponteiros, descartando pelo menos um quarto dos candidatos. Quando restam
// no máximo n candidatos, eles são ordenados diretamente.
//
// Referência: Croux, C. & Rousseeuw, P.J. (1992). Time-efficient algorithms
// for two highly robust estimators of scale. Computational Statistics, 1, 411-428.
func kthPairwiseDiff(y []float64, k int) float64 {
	n := len(y)
	// Candidatos da linha i: colunas j em [left[i], right[i]].
	left := make([]int, n)
	right := make([]int, n)
	candidates := 0
	for i := range y {
		left[i], right[i] = i+1, n-1
		candidates += right[i] - left[i] + 1
	}
	// below conta as diferenças já descartadas por serem menores que a resposta.
	below := 0

	type weighted struct {
		value  float64
		weight int
	}
	medians := make([]weighted, 0, n)
	less := make([]int, n)
	lessEq := make([]int, n)

	for candidates > n {
		medians = medians[:0]
		for i := range y {
			if left[i] <= right[i] {
				mid := (left[i] + right[i]) / 2
				medians = append(medians, weighted{y[mid] - y[i], right[i] - left[i] + 1})
			}
		}
		slices.SortFunc(medians, func(a, b weighted) int {
			switch {
			case a.value < b.value:
				return -1
			case a.value > b.value:
				return 1
			}
			return 0
		})
		var pivot float64
		cumulative := 0
		for _, m := range medians {
			cumulative += m.weight
			if 2*cumulative >= candidates {
				pivot = m.value
				break
			}
		}

		// less[i]: última coluna j com y[j]-y[i] < pivot (ou i se nenhuma);
		// lessEq[i]: idem para ≤. Ambas são não decrescentes em i.
		countLess, countLessEq := 0, 0
		jl, je := 0, 0
		for i := range y {
			jl, je = max(jl, i), max(je, i)
			for jl+1 < n && y[jl+1]-y[i] < pivot {
				jl++
			}
			for je+1 < n && y[je+1]-y[i] <= pivot {
				je++
			}
			less[i], lessEq[i] = jl, je
			countLess += jl - i
			countLessEq += je - i
		}

		switch {
		case k <= countLess:
			for i := range y {
				right[i] = min(right[i], less[i])
			}
		case k > countLessEq:
			for i := range y {
				left[i] = max(left[i], lessEq[i]+1)
			}
		default:
			return pivot
		}

		candidates, below = 0, 0
		for i := range y {
			below += left[i] - (i + 1)
			if left[i] <= right[i] {
				candidates += right[i] - left[i] + 1
			}
		}
	}

	rest := make([]float64, 0, candidates)
	for i := range y {
		for j := left[i]; j <= right[i]; j++ {
			rest = append(rest, y[j]-y[i])
		}
	}
	slices.Sort(rest)
	return rest[k-below-1]
}
