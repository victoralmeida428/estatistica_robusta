package robusto

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKthPairwiseDiff(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for trial := 0; trial < 300; trial++ {
		n := 2 + rng.IntN(60)
		y := make([]float64, n)
		for i := range y {
			if trial%2 == 0 {
				y[i] = float64(rng.IntN(8)) * 0.1 // muitos empates
			} else {
				y[i] = rng.NormFloat64()
			}
		}
		slices.Sort(y)

		all := make([]float64, 0, n*(n-1)/2)
		for i := range y {
			for j := i + 1; j < n; j++ {
				all = append(all, y[j]-y[i])
			}
		}
		slices.Sort(all)

		for k := 1; k <= len(all); k++ {
			require.Equal(t, all[k-1], kthPairwiseDiff(y, k), "n=%d k=%d y=%v", n, k, y)
		}
	}
}
