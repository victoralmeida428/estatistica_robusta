
# Robusto

**Robusto** é uma biblioteca em Go desenvolvida para cálculo de medidas estatísticas robustas, com foco em testes de proficiência e análise de dados laboratoriais, especialmente em contextos onde a presença de outliers pode comprometer estimativas tradicionais.

## ✨ Funcionalidades

A biblioteca oferece múltiplos métodos robustos para estimativa de média e desvio padrão:

- **Estimate** – Interface recomendada: média, desvio, incerteza `u(x*) = 1.25·s*/√p`, iterações e convergência, com erros tipados, para qualquer um dos métodos abaixo.
- **Qn** – Estatística robusta baseada nas diferenças absolutas entre pares.
- **Q Method** – Estimador robusto baseado em distribuições de diferenças, conforme descrito na norma ISO 13528 e trabalhos correlatos.
- **Algorithm A** – Algoritmo iterativo robusto para exclusão e ajuste de valores extremos.
- **Traditional** – Método tradicional com filtragem via IQR (quartis).
- **DamN** – Estimativa de dispersão baseada em mediana e MAD (Median Absolute Deviation).
- **NiQr** – Estimativa robusta baseada no IQR normalizado.
- **HampelMean** – Média de Hampel (ISO 13528, C.5.3) para qualquer dispersão robusta (Q, Qn, MADe…).
- **QReplicates** – Método Q entre laboratórios com mais de um resultado por laboratório.
- **Describe** – Estatísticas descritivas (quartis, moda, CV, MAD, nIQR, assimetria).
- **KernelDensity / Bandwidth** – Densidade de kernel para inspeção de modas.
- **Bootstrap** – Erro padrão bootstrap de qualquer estimador, reprodutível por semente.


## 📌 Base Teórica

O método **Q Method** implementado está baseado na definição matemática descrita por:

- Uhlig, S. (2015). *Robust estimation of between and within laboratory standard deviation measurement results below the detection limit.*
- Liu et al. (2019). *The Comparison of Three Robust Statistical Methods in Proficiency Testing.*
- ISO 13528:2015.

## 🚀 Como Usar

`Estimate` é o jeito recomendado de calcular um estimador. Ele devolve a média,
o desvio, a incerteza do valor designado e o diagnóstico de convergência, e
avisa por erro tipado quando o resultado é degenerado.

```go
package main

import (
    "errors"
    "fmt"
    "log"

    "github.com/victoralmeida428/estatistica_robusta/robusto"
)

func main() {
    data := []float64{10.0, 10.5, 9.8, 10.1, 100.0} // exemplo com outlier
    stats := robusto.New(data)

    r, err := stats.Estimate(robusto.MethodQHampel)
    switch {
    case errors.Is(err, robusto.ErrTooFewValues): // menos de 2 valores
        log.Fatal(err)
    case errors.Is(err, robusto.ErrZeroScale): // s* = 0; r.Mean é a mediana
    case errors.Is(err, robusto.ErrNotConverged): // r traz a última iteração
    }

    fmt.Printf("x* = %.3f  s* = %.3f  u(x*) = %.3f\n", r.Mean, r.SD, r.Uncertainty)
    // x* = 10.100  s* = 0.666  u(x*) = 0.372
}
```

Para comparar todos os estimadores, percorra `robusto.Methods`:

```go
for _, m := range robusto.Methods {
    r, err := stats.Estimate(m)
    if err != nil {
        fmt.Printf("%-12v %v\n", m, err)
        continue
    }
    fmt.Printf("%-12v x* = %.4f  s* = %.4f  u = %.4f\n", m, r.Mean, r.SD, r.Uncertainty)
}
```

### Atalhos

Os métodos diretos devolvem só `(média, desvio)`, sem incerteza nem erros
(NaN com menos de 2 valores). Use-os quando só precisar dos números, ou para
o que o `Estimate` não cobre:

```go
mean, std := stats.QMethod()          // mesmo resultado de MethodQHampel
mean, std = stats.AlgorithmA(false)   // um único passo, sem iterar
mean = stats.HampelMean(sigmaPT)      // média de Hampel com dispersão externa
outliers := stats.Outliers()          // valores fora das cercas de Tukey

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


