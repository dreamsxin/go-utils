package stats

import (
	"encoding/json"
	"sort"
	"strconv"
)

type HistogramStats struct {
	stats map[string]uint64
}

func NewHistogramStats() *HistogramStats {
	return &HistogramStats{
		stats: make(map[string]uint64),
	}
}

func (hs *HistogramStats) Reset() {
	hs.stats = make(map[string]uint64)
}

func (hs *HistogramStats) Update(delta map[string]uint64) {
	for key, val := range delta {
		hs.stats[key] += val
	}
}

func (hs *HistogramStats) Get() map[string]uint64 {
	stats := make(map[string]uint64)
	for key, val := range hs.stats {
		stats[key] = val
	}
	return stats
}

func (hs *HistogramStats) Copy() *HistogramStats {
	copyHS := &HistogramStats{}
	copyHS.stats = hs.Get()
	return copyHS
}

func (hs *HistogramStats) UpdateWithHistogram(hs1 *HistogramStats) {
	hs.Update(hs1.stats)
}

func (hs *HistogramStats) MarshalJSON() ([]byte, error) {
	return json.Marshal(hs.stats)
}

func (hs *HistogramStats) PercentileN(p int) int {
	latencyStats := hs.stats

	var samples sort.IntSlice
	var numSamples uint64
	for bin, binCount := range latencyStats {
		sample, err := strconv.Atoi(bin)
		if err == nil {
			samples = append(samples, sample)
			numSamples += binCount
		}
	}
	sort.Sort(samples)
	i := numSamples*uint64(p)/100 - 1

	var counter uint64
	var prevSample int
	for _, sample := range samples {
		if counter > i {
			return prevSample
		}
		counter += latencyStats[strconv.Itoa(sample)]
		prevSample = sample
	}

	if len(samples) > 0 {
		return samples[len(samples)-1]
	}
	return 0
}

func (hs *HistogramStats) LatencyPercentile() map[string]int {
	ls := make(map[string]int)
	ls["50"] = hs.PercentileN(50)
	ls["80"] = hs.PercentileN(80)
	ls["90"] = hs.PercentileN(90)
	ls["95"] = hs.PercentileN(95)
	ls["99"] = hs.PercentileN(99)
	ls["100"] = hs.PercentileN(100)
	return ls
}
