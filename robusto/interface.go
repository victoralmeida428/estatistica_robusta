package robusto

// Estimator calcula média, dispersão, incerteza e diagnóstico de
// convergência por qualquer Method. É a interface recomendada para quem
// só precisa do resultado de um estimador.
type Estimator interface {
	Estimate(Method) (Result, error)
}

// RobustEstimator expõe diretamente os estimadores robustos. Cada método
// retorna (média, dispersão), ou NaN quando há menos de 2 valores.
type RobustEstimator interface {
	// Qn: dispersão Qn de Rousseeuw-Croux (1993) com média de Hampel.
	Qn() (float64, float64)
	// QMethod: dispersão pelo método Q (ISO 13528, C.5) com média de Hampel.
	QMethod() (float64, float64)
	// AlgorithmA: Algorithm A da ISO 13528; iter=true repete até convergência.
	AlgorithmA(iter bool) (float64, float64)
	// DamN: mediana e MADe (MAD × 1.4826).
	DamN() (float64, float64)
	// NiQr: mediana e IQR normalizado (IQR / 1.349).
	NiQr() (float64, float64)
	// HampelMean: média de Hampel para uma dispersão robusta fixa.
	HampelMean(std float64) float64
}

// ClassicalEstimator expõe os estimadores clássicos e a detecção de
// outliers pelas cercas de Tukey.
type ClassicalEstimator interface {
	// Classical: média e desvio padrão amostral sem tratamento.
	Classical() (float64, float64)
	// Traditional: média e desvio padrão após remover os Outliers.
	Traditional() (float64, float64)
	// Outliers: valores fora das cercas de Tukey (1.5 × IQR).
	Outliers() []float64
}

// Describer calcula estatísticas descritivas e a densidade de kernel para
// inspeção da distribuição.
type Describer interface {
	Describe() (Summary, error)
	KernelDensity(h float64, at []float64) []float64
}

// DataSetter substitui os dados de uma instância.
type DataSetter interface {
	SetData([]float64)
}

// IStatistics reúne todas as interfaces implementadas por *Statistics.
// Prefira depender só da interface menor que o seu código usa.
type IStatistics interface {
	Estimator
	RobustEstimator
	ClassicalEstimator
	Describer
	DataSetter
}

var _ IStatistics = (*Statistics)(nil)
