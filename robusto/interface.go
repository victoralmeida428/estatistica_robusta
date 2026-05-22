package robusto

// IStatistics define o contrato público para estimadores robustos.
// Cada método retorna (estimativa de média, estimativa de dispersão).
type IStatistics interface {
	// Qn implementa o estimador Qn de Rousseeuw-Croux (1993) baseado nas
	// diferenças absolutas mediana das diferenças pareadas.
	Qn() (float64, float64)
	// QMethod estima média e desvio padrão a partir da distribuição empírica
	// das diferenças absolutas entre pares (ISO 13528).
	QMethod() (float64, float64)
	// AlgorithmA implementa o Algorithm A da ISO 13528. Se iter=true,
	// repete até convergência (troc 1e-4).
	AlgorithmA(bool) (float64, float64)
	// Traditional filtra outliers com cercas de Tukey (1.5×IQR) e calcula
	// média/desvio padrão clássicos nos dados restantes.
	Traditional() (float64, float64)
	// DamN retorna a mediana e o desvio absoluto mediano escalado (MAD × 1.4826).
	DamN() (float64, float64)
	// NiQr retorna a mediana e o IQR normalizado (IQR / 1.349).
	NiQr() (float64, float64)
	// SetData substitui os dados internos da instância.
	SetData([]float64)
}
