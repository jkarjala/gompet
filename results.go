package gompet

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// Results collects overall statistics
type Results struct {
	Start         time.Time
	LastProgress  time.Time
	LastStats     time.Time
	PeriodicStats int
	Progress      bool
	Count         int64
	LastCount     int64
	Times         []float64
	Results       map[string]int64
	Errs          map[string]int64
}

var percentiles = []float64{50, 90, 95, 98, 100}

var statsChan chan Results
var statsDone chan bool

// NewResults returns a newly initialized Results
func NewResults(progress bool, periodicStats int) *Results {
	var results Results
	results.Start = time.Now()
	results.LastProgress = time.Now()
	results.LastStats = time.Now()
	results.PeriodicStats = periodicStats
	results.Progress = progress
	results.Times = make([]float64, 0)
	results.Results = make(map[string]int64)
	results.Errs = make(map[string]int64)
	if periodicStats > 0 {
		statsChan = make(chan Results, 2) // buffer to reduce blocking the Update
		statsDone = make(chan bool)
		go statsReporter()
	}
	return &results
}

// Update results from one Run
func (results *Results) Update(res *RunResult) {
	results.Count++
	now := time.Now()
	if now.Sub(results.LastProgress) > 1*time.Second {
		if results.Progress {
			elapsed := now.Sub(results.Start).Seconds()
			cps := FormatDecimals(float64(results.Count) / elapsed)
			fmt.Printf("%.0fs %d commands in %0.1f seconds, %s cmds/sec\r",
				elapsed, results.Count, elapsed, cps)
		} else if results.PeriodicStats > 0 &&
			now.Sub(results.LastStats) > time.Duration(results.PeriodicStats)*time.Second {

			statsChan <- *results
			// must allocate a new buffer to preserve old for reporter
			results.Times = make([]float64, len(results.Times)) // old size is good estimate
			results.Times = results.Times[:0]                   // clear the slice to append results later
			results.LastCount = results.Count
			results.LastStats = now
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
func (results *Results) Report() {
	elapsed := time.Since(results.Start).Seconds()
	if results.PeriodicStats > 0 {
		statsChan <- *results
		close(statsChan)
		<-statsDone
	}
	if results.Progress {
		fmt.Println("")
	}
	PrintMap("Result counts:", results.Results)
	if len(results.Errs) > 0 {
		PrintMap("Error counts:", results.Errs)
	}
	cps := FormatDecimals(float64(results.Count) / elapsed)
	fmt.Printf("Total %d commands in %0.1f seconds, %s cmds/sec\n", results.Count, elapsed, cps)
	if results.PeriodicStats == 0 {
		fmt.Println("Latency percentiles:")
		sort.Float64s(results.Times)
		// fmt.Printf("Times:%v\n", results.Times)
		for _, p := range percentiles {
			PrintPercentile(results.Times, p)
		}
	}
}

// PercentileRowHeader returns the header for the stats rows
func (results *Results) PercentileRowHeader() string {
	var res = "Secs\t"
	for _, p := range percentiles {
		res += fmt.Sprintf("%.0f%% ms\t", p)
	}
	res += "Cmds\tCmds/sec"
	return res
}

// PercentileRow formats a row of percentiles from results
func (results *Results) PercentileRow() string {
	// NOTE: must not modify "results"
	elapsed := time.Since(results.Start).Seconds()
	lastElapsed := time.Since(results.LastStats).Seconds()
	var res strings.Builder
	res.WriteString(fmt.Sprintf("%.0f\t", elapsed))
	sort.Float64s(results.Times)
	for _, p := range percentiles {
		v := Percentile(results.Times, p) * 1000
		res.WriteString(FormatDecimals(v))
		res.WriteString("\t")
	}
	c := results.Count - results.LastCount
	cps := FormatDecimals(float64(c) / lastElapsed)
	res.WriteString(fmt.Sprintf("%d\t%s", c, cps))
	return res.String()
}

// FormatDecimals returns the value as a string with "nice" number of decimals
func FormatDecimals(v float64) string {
	switch true {
	case v < 0.999:
		return (fmt.Sprintf("%.3f", v))
	case v < 9.99:
		return (fmt.Sprintf("%.2f", v))
	case v < 99.9:
		return (fmt.Sprintf("%.1f", v))
	default:
		return (fmt.Sprintf("%.0f", v))
	}
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
	fmt.Printf("%3.0f%%\t%s ms\n", percent, FormatDecimals(Percentile(input, percent)*1000))
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

// Go routine to report progress during test without blocking the test
func statsReporter() {
	for results := range statsChan {
		if results.LastCount == 0 {
			fmt.Println(results.PercentileRowHeader())
		}
		fmt.Println(results.PercentileRow())
	}
	statsDone <- true
}
