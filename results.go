package gompet

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// Results collects overall statistics
type Results struct {
	Start        time.Time
	LastProgress time.Time
	Count        int64
	Times        []float64
	Results      map[string]int64
	Errs         map[string]int64
}

// NewResults returns a newly initialized Results
func NewResults() *Results {
	var results Results
	results.Start = time.Now()
	results.LastProgress = time.Now()
	results.Times = make([]float64, 0)
	results.Results = make(map[string]int64)
	results.Errs = make(map[string]int64)
	return &results
}

// Update results from one Run
func (results *Results) Update(res *RunResult, progress bool) {
	results.Count++
	now := time.Now()
	if now.Sub(results.LastProgress) > 1*time.Second {
		if progress {
			elapsed := time.Since(results.Start).Seconds()
			cps := float64(results.Count) / elapsed
			fmt.Printf("%.0fs %d commands in %0.1f seconds, %.2f cmds/sec\r",
				elapsed, results.Count, elapsed, cps)
		}
		results.LastProgress = now
	}
	results.Times = append(results.Times, res.Time)
	if res.Res != "" {
		results.Results[res.Res]++
	}
	if res.Err != nil {
		results.Errs[fmt.Sprintf("%s", res.Err)]++
	}
}

// Report results to stdout
func (results *Results) Report(progress bool) {
	elapsed := time.Since(results.Start).Seconds()
	if progress {
		fmt.Println("")
	}
	PrintMap("Result counts:", results.Results)
	if len(results.Errs) > 0 {
		PrintMap("Error counts:", results.Errs)
	}
	cps := float64(results.Count) / elapsed
	sort.Float64s(results.Times)
	// fmt.Printf("Times:%v\n", results.Times)
	fmt.Printf("%d commands in %0.1f seconds, %.2f cmds/sec\n", results.Count, elapsed, cps)
	fmt.Println("Latency percentiles:")
	PrintPercentile(results.Times, 50)
	PrintPercentile(results.Times, 90)
	PrintPercentile(results.Times, 95)
	PrintPercentile(results.Times, 98)
	PrintPercentile(results.Times, 100)
}

// PrintMap outputs heading and map key values in increasing alphabetical order of keys
func PrintMap(heading string, m map[string]int64) {
	fmt.Println(heading)

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := m[k]
		fmt.Printf("%d\t%s\n", v, k)
	}
}

// PrintPercentile outputs the percent's percentile from input
func PrintPercentile(input []float64, percent float64) {
	fmt.Printf("%3.0f%%\t%.2f ms\n", percent, Percentile(input, percent)*1000)
}

// Percentile finds the value at given percent in the input which must be pre-sorted
// Adapted from https://github.com/montanaflynn/stats/blob/master/percentile.go
func Percentile(input []float64, percent float64) (percentile float64) {
	length := len(input)
	if length == 0 {
		return math.NaN()
	}

	if length == 1 {
		return input[0]
	}

	if percent <= 0 || percent > 100 {
		return math.NaN()
	}

	index := (percent / 100) * float64(len(input))
	if index == float64(int64(index)) {
		i := int(index)
		percentile = input[i-1]
	} else if index > 1 {
		i := int(index)
		percentile = (input[i-1] + input[i]) / 2

	} else {
		return math.NaN()
	}
	return percentile
}
