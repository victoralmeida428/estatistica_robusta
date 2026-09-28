package robusto

import (
	"errors"
	"fmt"
	"math"
)

var (
	// ErrTooFewValues indica que há menos valores (ou laboratórios) que o
	// mínimo exigido pelo estimador.
	ErrTooFewValues = errors.New("robusto: são necessários pelo menos 2 valores")
	// ErrZeroScale indica que a dispersão robusta é zero (por exemplo, mais
	// da metade dos valores são iguais); a média retornada é a mediana.
	ErrZeroScale = errors.New("robusto: dispersão robusta igual a zero")
	// ErrNotConverged indica que o procedimento iterativo atingiu o limite
	// de iterações; o resultado retornado é o da última iteração.
	ErrNotConverged = errors.New("robusto: o procedimento iterativo não convergiu")
	// ErrUnknownMethod indica um Method inválido.
	ErrUnknownMethod = errors.New("robusto: método desconhecido")
)

// Method identifica um estimador de média e dispersão.
type Method int

const (
	// MethodClassical: média aritmética e desvio padrão amostral.
	MethodClassical Method = iota
	// MethodTraditional: média e desvio padrão após remover outliers pelas
	// cercas de Tukey.
	MethodTraditional
	// MethodMADe: mediana e MADe.
	MethodMADe
	// MethodNIQR: mediana e nIQR.
	MethodNIQR
	// MethodAlgorithmA: Algorithm A iterado (ISO 13528, C.3).
	MethodAlgorithmA
	// MethodQnHampel: média de Hampel com a dispersão Qn.
	MethodQnHampel
	// MethodQHampel: média de Hampel com a dispersão do método Q (ISO 13528, C.5).
	MethodQHampel
)

// Methods lista todos os estimadores disponíveis, na ordem em que Estimate
// os aceita.
var Methods = []Method{
	MethodClassical, MethodTraditional, MethodMADe, MethodNIQR,
	MethodAlgorithmA, MethodQnHampel, MethodQHampel,
}

func (m Method) String() string {
	switch m {
	case MethodClassical:
		return "Classical"
	case MethodTraditional:
		return "Traditional"
	case MethodMADe:
		return "MADe"
	case MethodNIQR:
		return "nIQR"
	case MethodAlgorithmA:
		return "AlgorithmA"
	case MethodQnHampel:
		return "Qn/Hampel"
	case MethodQHampel:
		return "Q/Hampel"
	}
	return fmt.Sprintf("Method(%d)", int(m))
}

// Result é o resultado de um estimador.
type Result struct {
	Method Method
	// Mean é a estimativa de média (valor designado x*).
	Mean float64
	// SD é a estimativa de desvio padrão (s*).
	SD float64
	// Uncertainty é a incerteza padrão de Mean: SD/√N para os métodos
	// clássicos e 1.25·SD/√N para os robustos (ISO 13528:2022, 7.7.3).
	Uncertainty float64
	// N é o número de valores usados (em MethodTraditional, após remover
	// os outliers).
	N int
	// Iterations é o número de iterações do procedimento iterativo (Algorithm
	// A ou média de Hampel); zero para os demais métodos.
	Iterations int
	// Converged indica se o procedimento iterativo convergiu; true para os
	// métodos não iterativos.
	Converged bool
}

// RobustUncertainty retorna a incerteza padrão do valor designado calculado
// por um estimador robusto: u(x*) = 1.25 · s* / √p (ISO 13528:2022, 7.7.3).
func RobustUncertainty(std float64, n int) float64 {
	return 1.25 * std / math.Sqrt(float64(n))
}

// Estimate calcula a média, a dispersão e a incerteza pelo método m.
//
// Erros: ErrTooFewValues com menos de 2 valores; ErrZeroScale quando a
// dispersão é zero (Mean é a mediana); ErrNotConverged quando o limite de
// iterações é atingido (o Result da última iteração é retornado mesmo
// assim); ErrUnknownMethod para m inválido.
func (s Statistics) Estimate(m Method) (Result, error) {
	n := len(s.data)
	r := Result{Method: m, N: n, Converged: true}
	if n < 2 {
		r.Mean, r.SD, r.Uncertainty = math.NaN(), math.NaN(), math.NaN()
		return r, ErrTooFewValues
	}

	robust := true
	switch m {
	case MethodClassical:
		robust = false
		r.Mean, r.SD = s.Classical()
	case MethodTraditional:
		robust = false
		r.N = len(s.withoutOutliers())
		if r.N < 2 {
			r.Mean, r.SD, r.Uncertainty = math.NaN(), math.NaN(), math.NaN()
			return r, ErrTooFewValues
		}
		r.Mean, r.SD = s.Traditional()
	case MethodMADe:
		r.Mean, r.SD = s.DamN()
	case MethodNIQR:
		r.Mean, r.SD = s.NiQr()
	case MethodAlgorithmA:
		r.Mean, r.SD, r.Iterations, r.Converged = s.algorithmA(maxIterations)
	case MethodQnHampel:
		r.SD = s.qnScale()
		r.Mean, r.Iterations, r.Converged = s.hampel(r.SD)
	case MethodQHampel:
		r.SD = s.qScale()
		r.Mean, r.Iterations, r.Converged = s.hampel(r.SD)
	default:
		return r, ErrUnknownMethod
	}

	if robust {
		r.Uncertainty = RobustUncertainty(r.SD, r.N)
	} else {
		r.Uncertainty = r.SD / math.Sqrt(float64(r.N))
	}

	switch {
	case r.SD == 0:
		return r, ErrZeroScale
	case !r.Converged:
		return r, ErrNotConverged
	}
	return r, nil
}
