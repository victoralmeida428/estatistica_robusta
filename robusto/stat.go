package robusto

import (
	"math"
	"robusto/utils"
	"slices"
)

type Statistics struct {
	data []float64
}

func (s Statistics) getDiff(p int) []float64 {
	diff := make([]float64, 0)
	for i := range s.data {
		for j := i + 1; j < p; j++ {
			diff = append(diff, math.Abs(s.data[j]-s.data[i]))
		}
	}
	slices.Sort(diff)
	return diff
}

func (s Statistics) Qn() (float64, float64) {
	p := len(s.data)
	diff := s.getDiff(p)
	if len(diff) == 1 {
		return diff[0], 0
	}
	
	var h, k int
	
	if p%2 == 0 {
		h = (p / 2) + 1
	} else {
		h = ((p - 1) / 2) + 1
	}
	
	k = h * (h - 1) / 2
	dk := diff[k-1]
	
	var bp float64
	
	if p >= 2 && p <= 12 {
		list := []float64{0.399356, 0.99365, 0.51321, 0.84401, 0.61220, 0.85877, 0.66993, 0.87344, 0.72014, 0.88906, 0.75743}
		bp = list[p-2]
	} else {
		pf := float64(p)
		var rp float64
		if p%2 == 0 {
			rp = (1 / pf) * (3.67561 + (1/pf)*(1.9654+(1/pf)*(6.987-(77/pf))))
		} else {
			rp = (1 / pf) * (1.60188 + (1/pf)*(-2.1284-(5.172/pf)))
		}
		bp = 1 / (rp + 1)
	}
	
	std := 2.21914 * dk * bp
	return s.hampelMean(std), std
}

func (s Statistics) QMethod() (float64, float64) {
	p := len(s.data)
	pf := float64(p)
	diff := s.getDiff(p)
	if len(diff) == 1 {
		return diff[0], 0
	}
	
	frequence := make(map[float64]int)
	
	for _, d := range diff {
		frequence[d]++
	}
	
	//contagem
	diffCompact := slices.Compact(diff)
	
	diffStruct := make([]statDiff, len(diffCompact))
	
	for i := range diffCompact {
		var ds statDiff
		ds.diff = diffCompact[i]
		ds.freq = frequence[diffCompact[i]]
		ds.fac = frequence[diffCompact[i]]
		if i > 0 {
			ds.fac += diffStruct[i-1].fac
		}
		diffStruct[i] = ds
	}
	
	diff0 := make([]statDiff, 0)
	if !slices.Contains(diffCompact, 0) {
		var ds statDiff
		diff0 = append(diff0, ds)
	}
	
	finalDiff := make([]statDiff, 0)
	finalDiff = append(finalDiff, diff0...)
	finalDiff = append(finalDiff, diffStruct...)
	
	var h0 float64
	
	gx := make([]float64, len(finalDiff))
	for i, ds := range finalDiff {
		hx := 2 / (pf * (pf - 1)) * float64(ds.fac)
		finalDiff[i].hx = hx
		if i == 0 {
			h0 = hx
			
			finalDiff[i].gx = hx * 0.5
		} else {
			finalDiff[i].gx = (hx + finalDiff[i-1].hx) * 0.5
		}
		gx[i] = finalDiff[i].gx
	}
	//Encontrando a inversa (Gx^-1) em y
	y := 0.25 + 0.75*h0
	gx = append(gx, y)
	slices.Sort(gx)
	var yIndex int
	for i := range gx {
		if gx[i] == y {
			yIndex = i
		}
	}
	numerador := finalDiff[yIndex-1].diff + (finalDiff[yIndex].diff-finalDiff[yIndex-1].diff)/(finalDiff[yIndex].gx-finalDiff[yIndex-1].gx)*(y-finalDiff[yIndex-1].gx)
	denominador := math.Sqrt(2) * utils.NormPPF(0.625+0.375*h0)
	std := numerador / denominador
	return s.hampelMean(std), std
}

func (s Statistics) recurrenceAlgorithmA(median, dam float64) (float64, float64) {
	sigma := 1.5 * dam
	finalData := make([]float64, len(s.data))
	
	for i, value := range s.data {
		if value > (median + sigma) {
			finalData[i] = median + sigma
		} else if value < (median - sigma) {
			finalData[i] = median - sigma
		} else {
			finalData[i] = value
		}
	}
	
	mean := utils.Mean(finalData)
	return mean, utils.Std(finalData, &mean) * 1.134
}

func (s Statistics) AlgorithmA(iter bool) (float64, float64) {
	
	median := utils.Quantile(s.data, 0.5)
	distMedian := make([]float64, len(s.data))
	for i, value := range s.data {
		distMedian[i] = math.Abs(value - median)
	}
	dam := 1.4826 * utils.Quantile(distMedian, 0.5)
	
	mean, std := s.recurrenceAlgorithmA(median, dam)
	
	if iter {
		for i := 0; i < 100_000; i++ {
			newMean, newStd := s.recurrenceAlgorithmA(mean, std)
			if (math.Abs(mean-newMean) < 0.0001) && (math.Abs(std-newStd) < 0.0001) {
				return newMean, newStd
			}
			mean = newMean
			std = newStd
			
		}
	}
	
	return mean, std
}

func (s Statistics) Traditional() (float64, float64) {
	q1 := utils.Quantile(s.data, 0.25)
	q3 := utils.Quantile(s.data, 0.75)
	iqr := q3 - q1
	cerca_inf := q1 - 1.5*iqr
	cerca_sup := q3 + 1.5*iqr
	
	filtered := slices.DeleteFunc(slices.Clone(s.data), func(value float64) bool {
		return value < cerca_inf || cerca_sup < value
	})
	
	mean := utils.Mean(filtered)
	std := utils.Std(filtered, &mean)
	filtered = nil
	return mean, std
}

func (s Statistics) DamN() (float64, float64) {
	median := utils.Quantile(s.data, 0.5)
	aux := make([]float64, len(s.data))
	for i := range aux {
		aux[i] = math.Abs(s.data[i] - median)
	}
	mad := utils.Quantile(aux, 0.5)
	dam := 1.4826 * mad
	return median, dam
}

func (s Statistics) NiQr() (float64, float64) {
	median := utils.Quantile(s.data, 0.5)
	q1 := utils.Quantile(s.data, 0.25)
	q3 := utils.Quantile(s.data, 0.75)
	iqr := q3 - q1
	n_iqr := iqr / 1.349
	return median, n_iqr
}

func (s *Statistics) SetData(data []float64) {
	s.data = data
}

func (s *Statistics) findMean(std float64, ref *float64) float64 {
	if ref == nil {
		refV := utils.Quantile(s.data, 0.5)
		ref = &refV
	}
	
	//find q
	q := make([]float64, len(s.data))
	for i := range q {
		q[i] = math.Abs((s.data[i] - *ref) / std)
	}
	
	//find w
	w := make([]float64, len(s.data))
	for i := range w {
		if q[i] > 3 && q[i] <= 4.5 {
			w[i] = (4.5 - q[i]) / q[i]
		}
		if q[i] > 1.5 && q[i] <= 3 {
			w[i] = 1.5 / q[i]
		}
		if q[i] <= 1.5 {
			w[i] = 1
		}
	}
	
	//get mean
	numerator := 0.0
	denominator := 0.0
	for i := range s.data {
		numerator += s.data[i] * w[i]
		denominator += w[i]
	}
	return numerator / denominator
	
}

func (s Statistics) hampelMean(std float64) float64 {
	erro := 0.01 * std / math.Sqrt(float64(len(s.data)))
	mean := s.findMean(std, nil)
	
	for i := 0; i < 100_000; i++ {
		newMean := s.findMean(std, &mean)
		if math.Abs(mean-newMean) <= erro {
			return newMean
		}
		mean = newMean
	}
	
	return mean
}

func New(data []float64) *Statistics {
	return &Statistics{data: data}
}
