package robusto

type IStatistics interface {
	Qn() (float64, float64)
	QMethod() (float64, float64)
	AlgorithmA(bool) (float64, float64)
	Traditional() (float64, float64)
	DamN() (float64, float64)
	NiQr() (float64, float64)
	SetData([]float64)
}
