package robusto

import (
	"errors"
	"math"
	"slices"

	"github.com/victoralmeida428/estatistica_robusta/utils"
)

// weightedDiff é uma diferença absoluta entre resultados de dois
// laboratórios com o peso que ela tem na FDA H₁ (ISO 13528:2022, C.23).
type weightedDiff struct {
	diff, weight float64
}

// qScale implementa o método Q da ISO 13528:2022 (C.5.2, eq. C.23–C.25) para
// grupos de resultados, um grupo por laboratório. Retorna s* e H₁(0).
//
// H₁(x) é a média, sobre os pares de laboratórios, da proporção de diferenças
// entre os dois laboratórios ≤ x. Com um resultado por laboratório, todas as
// diferenças têm o mesmo peso.
func qScale(groups [][]float64) (std, h0 float64) {
	p := len(groups)
	pairs := float64(p * (p - 1) / 2)

	diffs := make([]weightedDiff, 0)
	for i := 0; i < p-1; i++ {
		for j := i + 1; j < p; j++ {
			weight := 1 / (pairs * float64(len(groups[i])*len(groups[j])))
			for _, a := range groups[i] {
				for _, b := range groups[j] {
					diffs = append(diffs, weightedDiff{math.Abs(a - b), weight})
				}
			}
		}
	}
	slices.SortFunc(diffs, func(a, b weightedDiff) int {
		switch {
		case a.diff < b.diff:
			return -1
		case a.diff > b.diff:
			return 1
		}
		return 0
	})

	// Pontos de descontinuidade positivos xᵢ e H₁(xᵢ) acumulado.
	xs := make([]float64, 0)
	hs := make([]float64, 0)
	cumulative := 0.0
	for i := 0; i < len(diffs); {
		d := diffs[i].diff
		for ; i < len(diffs) && diffs[i].diff == d; i++ {
			cumulative += diffs[i].weight
		}
		if d == 0 {
			h0 = cumulative
			continue
		}
		xs = append(xs, d)
		hs = append(hs, cumulative)
	}
	if len(xs) == 0 {
		return 0, h0
	}

	// G₁(0) = 0, G₁(x₁) = [H₁(x₁)+H₁(0)]/2, G₁(xᵢ) = [H₁(xᵢ)+H₁(xᵢ₋₁)]/2;
	// G₁⁻¹(y) por interpolação linear entre os pontos (x, G₁(x)).
	y := 0.25 + 0.75*h0
	prevX, prevG, prevH := 0.0, 0.0, h0
	for i := range xs {
		g := (hs[i] + prevH) / 2
		if y <= g {
			inverse := prevX + (xs[i]-prevX)*(y-prevG)/(g-prevG)
			return inverse / (math.Sqrt2 * utils.NormPPF(0.625+0.375*h0)), h0
		}
		prevX, prevG, prevH = xs[i], g, hs[i]
	}
	return xs[len(xs)-1] / (math.Sqrt2 * utils.NormPPF(0.625+0.375*h0)), h0
}

// ErrLabsMismatch indica que results e labs têm tamanhos diferentes.
var ErrLabsMismatch = errors.New("robusto: results e labs devem ter o mesmo tamanho")

// QReplicates estima o desvio padrão robusto entre laboratórios s* pelo
// método Q da ISO 13528:2022 (C.5.2) quando cada laboratório pode ter mais
// de um resultado. labs[i] identifica o laboratório de results[i].
//
// Retorna ErrLabsMismatch se os tamanhos diferirem e ErrTooFewValues se
// houver menos de dois laboratórios.
func QReplicates(results []float64, labs []string) (std float64, err error) {
	if len(results) != len(labs) {
		return math.NaN(), ErrLabsMismatch
	}

	index := make(map[string]int)
	groups := make([][]float64, 0)
	for i, lab := range labs {
		g, ok := index[lab]
		if !ok {
			g = len(groups)
			index[lab] = g
			groups = append(groups, nil)
		}
		groups[g] = append(groups[g], results[i])
	}
	if len(groups) < 2 {
		return math.NaN(), ErrTooFewValues
	}

	std, _ = qScale(groups)
	return std, nil
}
