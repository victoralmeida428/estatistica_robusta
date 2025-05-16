package robusto

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type testCase struct {
	Name  string    `json:"name"`
	Input []float64 `json:"input"`
	Mean  float64   `json:"mean"`
	Std   float64   `json:"std"`
}

func TestRobusto(t *testing.T) {
	filenames := []struct {
		name   string
		handle func(stat *Statistics) (float64, float64)
	}{
		{
			name: "dam_n.json",
			handle: func(stat *Statistics) (float64, float64) {
				return stat.DamN()
			},
		},
		{
			name: "niqr.json",
			handle: func(stat *Statistics) (float64, float64) {
				return stat.NiQr()
			},
		},
		{
			name: "traditional.json",
			handle: func(stat *Statistics) (float64, float64) {
				return stat.Traditional()
			},
		},
		{
			name: "algorithm_a.json",
			handle: func(stat *Statistics) (float64, float64) {
				return stat.AlgorithmA(false)
			},
		},
		{
			name: "algorithm_a_iter.json",
			handle: func(stat *Statistics) (float64, float64) {
				return stat.AlgorithmA(true)
			},
		},
		{
			name: "qn.json",
			handle: func(stat *Statistics) (float64, float64) {
				return stat.Qn()
			},
		},
		{
			name: "q_method.json",
			handle: func(stat *Statistics) (float64, float64) {
				return stat.QMethod()
			},
		},
	}
	
	for _, f := range filenames {
		t.Run(f.name, func(t *testing.T) {
			file, err := os.Open(fmt.Sprintf("../testdata/%s", f.name))
			assert.NoError(t, err)
			defer file.Close()
			
			var testData []testCase
			decoder := json.NewDecoder(file)
			err = decoder.Decode(&testData)
			assert.NoError(t, err)
			
			for _, data := range testData {
				t.Run(data.Name, func(t *testing.T) {
					stat := New(data.Input)
					mean, std := f.handle(stat)
					assert.InDelta(t, data.Mean, mean, 1e-6, "error in mean")
					assert.InDelta(t, data.Std, std, 1e-6, "error in std")
				})
			}
		})
		
	}
}
