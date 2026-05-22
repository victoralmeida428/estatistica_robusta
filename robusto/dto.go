// Package robusto fornece estimadores estatísticos robustos para média e
// dispersão, com foco em testes de proficiência laboratorial (ISO 13528).
package robusto

// statDiff é um DTO interno usado pelo método Q para representar uma
// diferença pareada com sua frequência (freq), frequência acumulada (fac),
// ponto médio da FDA empírica (hx) e ponto médio da FDA linearizada (gx).
type statDiff struct {
	diff       float64 // valor da diferença absoluta entre pares
	hx, gx     float64 // h(d): freq.acum./total; g(d): ponto médio entre h(d_i) e h(d_{i-1})
	freq, fac  int     // freq: contagem; fac: soma acumulada das frequências
}
