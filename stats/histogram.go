package stats

import (
	"math"
)

// 定义了一个桶，包含下限、上限和计数。
type Bucket struct {
	LowerBound float64
	UpperBound float64
	Count      int
}

// 包含一个桶数组和一个总数。
type Histogram struct {
	Buckets []Bucket
	Total   int
}

func NewHistogram(bucketSizes []float64) *Histogram {
	histogram := &Histogram{
		Buckets: make([]Bucket, len(bucketSizes)),
	}

	for i, size := range bucketSizes {
		histogram.Buckets[i] = Bucket{
			LowerBound: size,
			UpperBound: math.Inf(1),
			Count:      0,
		}
		if i > 0 {
			histogram.Buckets[i-1].UpperBound = size
		}
	}

	return histogram
}

func (h *Histogram) Add(value float64) {
	for i, bucket := range h.Buckets {
		if value >= bucket.LowerBound && value < bucket.UpperBound {
			h.Buckets[i].Count++
			h.Total++
			break
		}
	}
}

// 计算给定百分位数的值。
func (h *Histogram) Percentile(p float64) float64 {
	if h.Total == 0 {
		return 0
	}

	count := int(float64(h.Total) * p / 100)
	sum := 0
	for _, bucket := range h.Buckets {
		sum += bucket.Count
		if sum >= count {
			return bucket.LowerBound
		}
	}

	return h.Buckets[len(h.Buckets)-1].UpperBound
}

// 计算平均值。
func (h *Histogram) Mean() float64 {
	if h.Total == 0 {
		return 0
	}

	sum := 0.0
	for _, bucket := range h.Buckets {
		sum += bucket.LowerBound * float64(bucket.Count)
	}

	return sum / float64(h.Total)
}

// 计算标准差。
func (h *Histogram) StdDev() float64 {
	if h.Total == 0 {
		return 0
	}

	mean := h.Mean()
	sum := 0.0
	for _, bucket := range h.Buckets {
		diff := bucket.LowerBound - mean
		sum += diff * diff * float64(bucket.Count)
	}

	variance := sum / float64(h.Total)
	return math.Sqrt(variance)
}

func (h *Histogram) Merge(other *Histogram) {
	for i, _ := range h.Buckets {
		h.Buckets[i].Count += other.Buckets[i].Count
	}
	h.Total += other.Total
}

func (h *Histogram) GetBuckets() []Bucket {
	return h.Buckets
}
