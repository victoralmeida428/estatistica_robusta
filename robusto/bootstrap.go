package robusto

import (
	"math/rand/v2"

	"github.com/victoralmeida428/estatistica_robusta/utils"
)

// Bootstrap reamostra data com reposição replicates vezes, aplica
// estimator a cada amostra e retorna a média e o desvio padrão (erro padrão
// bootstrap) das estimativas.
//
// estimator recebe a mesma fatia reutilizada a cada réplica e não deve
// guardá-la. seed torna o resultado reprodutível. O gerador (PCG) é diferente do R,
// então os valores não coincidem com boot::boot mesmo com a mesma semente.
//
// Exemplo, erro padrão bootstrap da média do Algorithm A:
//
//	mean, se := robusto.Bootstrap(data, func(x []float64) float64 {
//		m, _ := robusto.New(x).AlgorithmA(true)
//		return m
//	}, 1000, 339)
func Bootstrap(data []float64, estimator func([]float64) float64, replicates int, seed uint64) (mean, se float64) {
	rng := rand.New(rand.NewPCG(seed, seed))
	estimates := make([]float64, replicates)
	sample := make([]float64, len(data))
	for r := range estimates {
		for i := range sample {
			sample[i] = data[rng.IntN(len(data))]
		}
		estimates[r] = estimator(sample)
	}
	mean = utils.Mean(estimates)
	return mean, utils.Std(estimates, &mean)
}
