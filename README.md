
# Robusto

**Robusto** é uma biblioteca em Go desenvolvida para cálculo de medidas estatísticas robustas, com foco em testes de proficiência e análise de dados laboratoriais, especialmente em contextos onde a presença de outliers pode comprometer estimativas tradicionais.

## ✨ Funcionalidades

A biblioteca oferece múltiplos métodos robustos para estimativa de média e desvio padrão:

- **Qn** – Estatística robusta baseada nas diferenças absolutas entre pares.
- **Q Method** – Estimador robusto baseado em distribuições de diferenças, conforme descrito na norma ISO 13528 e trabalhos correlatos.
- **Algorithm A** – Algoritmo iterativo robusto para exclusão e ajuste de valores extremos.
- **Traditional** – Método tradicional com filtragem via IQR (quartis).
- **DamN** – Estimativa de dispersão baseada em mediana e MAD (Median Absolute Deviation).
- **NiQr** – Estimativa robusta baseada no IQR normalizado.
- **HampelMean** – Média de Hampel (ISO 13528, C.5.3) para qualquer dispersão robusta (Q, Qn, MADe…).
- **QReplicates** – Método Q entre laboratórios com mais de um resultado por laboratório.
- **Estimate** – Média, desvio, incerteza `u(x*) = 1.25·s*/√p`, iterações e convergência, com erros tipados.
- **Describe** – Estatísticas descritivas (quartis, moda, CV, MAD, nIQR, assimetria).
- **KernelDensity / Bandwidth** – Densidade de kernel para inspeção de modas.
- **Bootstrap** – Erro padrão bootstrap de qualquer estimador, reprodutível por semente.


## 📌 Base Teórica

O método **Q Method** implementado está baseado na definição matemática descrita por:

- Uhlig, S. (2015). *Robust estimation of between and within laboratory standard deviation measurement results below the detection limit.*
- Liu et al. (2019). *The Comparison of Three Robust Statistical Methods in Proficiency Testing.*
- ISO 13528:2015.

## 🚀 Como Usar

```go
package main

import (
    "fmt"
    "github.com/victoralmeida428/estatistica_robusta/robusto"
)

func main() {
    data := []float64{10.0, 10.5, 9.8, 10.1, 100.0} // exemplo com outlier
    stats := robusto.New(data)

    mean, std := stats.QMethod()
    fmt.Printf("Média robusta (Q Method): %.3f\nDesvio padrão: %.3f\n", mean, std)
}
```

### Resultado completo e erros

```go
stats := robusto.New(data)
r, err := stats.Estimate(robusto.MethodQHampel)
switch {
case errors.Is(err, robusto.ErrTooFewValues): // menos de 2 valores
case errors.Is(err, robusto.ErrZeroScale):    // s* = 0; r.Mean é a mediana
case errors.Is(err, robusto.ErrNotConverged): // r traz a última iteração
}
fmt.Println(r.Mean, r.SD, r.Uncertainty, r.Iterations)

// Réplicas por laboratório
sd, err := robusto.QReplicates(results, labs)
```

## 🔬 Validação

`robusto/eurachem_test.go` compara os estimadores com o R nos 7 datasets do
material Eurachem "Hands-On Data Analytics" (ISO 13528:2022), com tolerância
de 1e-9. Os valores são gerados por `testdata/eurachem.R`.

Diferenças conhecidas em relação aos scripts R do curso:

- **Algorithm A** usa o fator 1.134 da norma e itera até convergência
  (1e-10 × s*); `metRology::algA` usa o fator γ exato (1.13439) e para em 30
  iterações, o que muda o 4º dígito em alguns datasets.
- **Hampel** parte da mediana (ISO 13528, C.5.3); `rlm` parte da média. Nos 7
  datasets as duas convergem para o mesmo valor.
- **Traditional** usa quartis tipo R-7 nas cercas; `boxplot.stats` usa as
  dobradiças de Tukey, que podem marcar outliers diferentes para alguns n.
- **Bootstrap** usa o gerador PCG do Go: com a mesma semente, os números não
  coincidem com `boot::boot`.

## 📦 Instalação

Como o projeto é em Go, basta incluí-lo como dependência no seu projeto. Certifique-se de importar corretamente o pacote (`robusto`)


