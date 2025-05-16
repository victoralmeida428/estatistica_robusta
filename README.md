
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
    "robusto"
)

func main() {
    data := []float64{10.0, 10.5, 9.8, 10.1, 100.0} // exemplo com outlier
    stats := robusto.New(data)

    mean, std := stats.QMethod()
    fmt.Printf("Média robusta (Q Method): %.3f\nDesvio padrão: %.3f\n", mean, std)
}
```

## 📦 Instalação

Como o projeto é em Go, basta incluí-lo como dependência no seu projeto. Certifique-se de importar corretamente o pacote (`robusto`) e disponibilizar o diretório `utils` se não estiver incluso no repositório.


