package robusto

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// eurachemCase é gerado por testdata/eurachem.R a partir dos 7 datasets do
// material Eurachem (ISO 13528:2022).
type eurachemCase struct {
	Name       string    `json:"name"`
	Input      []float64 `json:"input"`
	Mean       float64   `json:"mean"`
	SD         float64   `json:"sd"`
	Median     float64   `json:"median"`
	MAD        float64   `json:"mad"`
	MADe       float64   `json:"made"`
	NIQR       float64   `json:"niqr"`
	Q1         float64   `json:"q1"`
	Q3         float64   `json:"q3"`
	Mode       float64   `json:"mode"`
	Skewness   float64   `json:"skewness"`
	AlgAMean   float64   `json:"alga_mean"`
	AlgASD     float64   `json:"alga_sd"`
	Qn         float64   `json:"qn"`
	QnHampel   float64   `json:"qn_hampel"`
	Q          float64   `json:"q"`
	QHampel    float64   `json:"q_hampel"`
	MADeHampel float64   `json:"made_hampel"`
}

func loadJSON(t *testing.T, name string, v any) {
	t.Helper()
	b, err := os.ReadFile("../testdata/" + name)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, v))
}

func TestEurachem(t *testing.T) {
	var cases []eurachemCase
	loadJSON(t, "eurachem.json", &cases)
	require.Len(t, cases, 7)

	const delta = 1e-9
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			s := New(c.Input)

			r, err := s.Estimate(MethodAlgorithmA)
			assert.NoError(t, err)
			assert.InDelta(t, c.AlgAMean, r.Mean, delta, "AlgorithmA mean")
			assert.InDelta(t, c.AlgASD, r.SD, delta, "AlgorithmA sd")
			assert.InDelta(t, 1.25*c.AlgASD/math.Sqrt(float64(len(c.Input))), r.Uncertainty, delta)

			r, err = s.Estimate(MethodQnHampel)
			assert.NoError(t, err)
			// robustbase::Qn arredonda a estatística de ordem para float32 em
			// alguns casos (outliers_25: 0.04 vira 0.039999999105930328).
			assert.InEpsilon(t, c.Qn, r.SD, 1e-7, "Qn")
			assert.InEpsilon(t, c.QnHampel, r.Mean, 1e-7, "Qn/Hampel mean")

			r, err = s.Estimate(MethodQHampel)
			assert.NoError(t, err)
			assert.InDelta(t, c.Q, r.SD, delta, "Q")
			assert.InDelta(t, c.QHampel, r.Mean, delta, "Q/Hampel mean")

			assert.InDelta(t, c.MADeHampel, s.HampelMean(c.MADe), delta, "MADe/Hampel mean")

			median, made := s.DamN()
			assert.InDelta(t, c.Median, median, delta)
			assert.InDelta(t, c.MADe, made, delta)
			_, niqr := s.NiQr()
			assert.InDelta(t, c.NIQR, niqr, delta)

			sum, err := s.Describe()
			assert.NoError(t, err)
			assert.InDelta(t, c.Mean, sum.Mean, delta)
			assert.InDelta(t, c.SD, sum.SD, delta)
			assert.InDelta(t, c.MAD, sum.MAD, delta)
			assert.InDelta(t, c.Q1, sum.Q1, delta)
			assert.InDelta(t, c.Q3, sum.Q3, delta)
			assert.Equal(t, c.Mode, sum.Mode)
			assert.InDelta(t, c.Skewness, sum.Skewness, delta)
		})
	}
}

func TestQReplicates(t *testing.T) {
	var cases []struct {
		Name  string    `json:"name"`
		Input []float64 `json:"input"`
		Labs  []string  `json:"labs"`
		SD    float64   `json:"sd"`
	}
	loadJSON(t, "q_replicates.json", &cases)

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			sd, err := QReplicates(c.Input, c.Labs)
			assert.NoError(t, err)
			assert.InDelta(t, c.SD, sd, 1e-12)
		})
	}

	t.Run("um resultado por laboratório igual a QMethod", func(t *testing.T) {
		data := []float64{5, 5, 5, 5, 6, 6, 6, 7, 8, 12}
		labs := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
		sd, err := QReplicates(data, labs)
		assert.NoError(t, err)
		_, want := New(data).QMethod()
		assert.InDelta(t, want, sd, 1e-15)
	})

	t.Run("erros", func(t *testing.T) {
		_, err := QReplicates([]float64{1, 2}, []string{"a"})
		assert.ErrorIs(t, err, ErrLabsMismatch)
		_, err = QReplicates([]float64{1, 2}, []string{"a", "a"})
		assert.ErrorIs(t, err, ErrTooFewValues)
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("poucos valores não causam panic", func(t *testing.T) {
		for _, data := range [][]float64{nil, {}, {1}} {
			s := New(data)
			for _, m := range Methods {
				r, err := s.Estimate(m)
				assert.ErrorIs(t, err, ErrTooFewValues, m.String())
				assert.True(t, math.IsNaN(r.Mean), m.String())
			}
			mean, std := s.Qn()
			assert.True(t, math.IsNaN(mean) && math.IsNaN(std))
			mean, std = s.QMethod()
			assert.True(t, math.IsNaN(mean) && math.IsNaN(std))
			mean, std = s.AlgorithmA(true)
			assert.True(t, math.IsNaN(mean) && math.IsNaN(std))
		}
	})

	t.Run("dois valores: média é o centro, não a diferença", func(t *testing.T) {
		s := New([]float64{10, 12})
		mean, std := s.Qn()
		assert.InDelta(t, 11, mean, 1e-12)
		assert.InDelta(t, 2.21914*2*0.399356, std, 1e-12)
		mean, std = s.QMethod()
		assert.InDelta(t, 11, mean, 1e-12)
		assert.Greater(t, std, 0.0)
	})

	t.Run("dispersão zero retorna a mediana", func(t *testing.T) {
		s := New([]float64{3, 3, 3, 3, 3, 4, 9})
		for _, m := range []Method{MethodMADe, MethodAlgorithmA, MethodQnHampel} {
			r, err := s.Estimate(m)
			assert.ErrorIs(t, err, ErrZeroScale, m.String())
			assert.Equal(t, 3.0, r.Mean, m.String())
			assert.Equal(t, 0.0, r.SD, m.String())
		}
	})

	t.Run("todos iguais", func(t *testing.T) {
		mean, std := New([]float64{7, 7, 7}).QMethod()
		assert.Equal(t, 7.0, mean)
		assert.Equal(t, 0.0, std)
	})

	t.Run("New copia os dados", func(t *testing.T) {
		data := []float64{1, 2, 3, 4, 100}
		s := New(data)
		data[4] = 5
		median, _ := s.DamN()
		assert.Equal(t, 3.0, median)
		assert.Equal(t, []float64{100}, s.Outliers())
	})

	t.Run("Algorithm A independe da escala", func(t *testing.T) {
		base := []float64{0.04, 0.055, 0.178, 0.202, 0.206, 0.227, 0.228, 0.23, 0.23, 0.235, 0.236, 0.237}
		ref, err := New(base).Estimate(MethodAlgorithmA)
		require.NoError(t, err)
		for _, scale := range []float64{1e-6, 1e6} {
			scaled := make([]float64, len(base))
			for i, v := range base {
				scaled[i] = v * scale
			}
			r, err := New(scaled).Estimate(MethodAlgorithmA)
			assert.NoError(t, err)
			assert.InEpsilon(t, ref.Mean*scale, r.Mean, 1e-9)
			assert.InEpsilon(t, ref.SD*scale, r.SD, 1e-9)
			assert.InDelta(t, ref.Iterations, r.Iterations, 2)
		}
	})

	t.Run("método desconhecido", func(t *testing.T) {
		_, err := New([]float64{1, 2, 3}).Estimate(Method(99))
		assert.ErrorIs(t, err, ErrUnknownMethod)
		assert.Equal(t, "Method(99)", Method(99).String())
	})
}

func TestEstimateTraditional(t *testing.T) {
	data := []float64{10, 10.5, 9.8, 10.1, 100}
	s := New(data)
	r, err := s.Estimate(MethodTraditional)
	assert.NoError(t, err)
	assert.Equal(t, 4, r.N)
	mean, std := s.Traditional()
	assert.Equal(t, mean, r.Mean)
	assert.InDelta(t, std/2, r.Uncertainty, 1e-15)
	assert.Equal(t, []float64{100}, s.Outliers())
}

func TestBootstrap(t *testing.T) {
	data := []float64{0.04, 0.055, 0.178, 0.202, 0.206, 0.227, 0.228, 0.23, 0.23, 0.235, 0.236, 0.237}
	median := func(x []float64) float64 { m, _ := New(x).DamN(); return m }

	m1, se1 := Bootstrap(data, median, 500, 339)
	m2, se2 := Bootstrap(data, median, 500, 339)
	assert.Equal(t, m1, m2, "mesma semente, mesmo resultado")
	assert.Equal(t, se1, se2)
	assert.Greater(t, se1, 0.0)

	constant := []float64{2, 2, 2, 2}
	m, se := Bootstrap(constant, median, 100, 1)
	assert.Equal(t, 2.0, m)
	assert.Equal(t, 0.0, se)
}

func TestKernelDensity(t *testing.T) {
	s := New([]float64{0.2, 0.25, 0.26, 0.3, 0.55, 0.57})
	h := Bandwidth(0.05, 6)
	assert.InDelta(t, 0.9*0.05/math.Pow(6, 0.2), h, 1e-15)

	// Integral numérica da densidade ≈ 1.
	const step = 1e-3
	at := make([]float64, 0)
	for x := -1.0; x <= 2; x += step {
		at = append(at, x)
	}
	total := 0.0
	for _, d := range s.KernelDensity(h, at) {
		total += d * step
	}
	assert.InDelta(t, 1, total, 1e-6)
}
