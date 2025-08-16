package prometheus

import (
	"math"
	"sort"
	"time"
)

type Aggregator struct {
	collector *MetricCollector
}

func NewAggregator(collector *MetricCollector) *Aggregator {
	return &Aggregator{
		collector: collector,
	}
}

func (a *Aggregator) Sum(name string, labels map[string]string, duration time.Duration) float64 {
	series, exists := a.collector.GetMetrics(name, labels)
	if !exists {
		return 0
	}

	series.mutex.RLock()
	defer series.mutex.RUnlock()

	cutoff := time.Now().Add(-duration)
	sum := 0.0

	for _, dp := range series.Values {
		if dp.Timestamp.After(cutoff) {
			sum += dp.Value
		}
	}

	return sum
}

func (a *Aggregator) Average(name string, labels map[string]string, duration time.Duration) float64 {
	series, exists := a.collector.GetMetrics(name, labels)
	if !exists {
		return 0
	}

	series.mutex.RLock()
	defer series.mutex.RUnlock()

	cutoff := time.Now().Add(-duration)
	sum := 0.0
	count := 0

	for _, dp := range series.Values {
		if dp.Timestamp.After(cutoff) {
			sum += dp.Value
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return sum / float64(count)
}

func (a *Aggregator) Max(name string, labels map[string]string, duration time.Duration) float64 {
	series, exists := a.collector.GetMetrics(name, labels)
	if !exists {
		return 0
	}

	series.mutex.RLock()
	defer series.mutex.RUnlock()

	cutoff := time.Now().Add(-duration)
	max := math.Inf(-1)

	for _, dp := range series.Values {
		if dp.Timestamp.After(cutoff) {
			if dp.Value > max {
				max = dp.Value
			}
		}
	}

	if math.IsInf(max, -1) {
		return 0
	}

	return max
}

func (a *Aggregator) Percentile(name string, labels map[string]string, duration time.Duration, percentile float64) float64 {
	series, exists := a.collector.GetMetrics(name, labels)
	if !exists {
		return 0
	}

	series.mutex.RLock()
	defer series.mutex.RUnlock()

	cutoff := time.Now().Add(-duration)
	var values []float64

	for _, dp := range series.Values {
		if dp.Timestamp.After(cutoff) {
			values = append(values, dp.Value)
		}
	}

	if len(values) == 0 {
		return 0
	}

	sort.Float64s(values)
	index := int(float64(len(values)) * percentile / 100.0)
	if index >= len(values) {
		index = len(values) - 1
	}

	return values[index]
}

func (a *Aggregator) Rate(name string, labels map[string]string, duration time.Duration) float64 {
	series, exists := a.collector.GetMetrics(name, labels)
	if !exists {
		return 0
	}

	series.mutex.RLock()
	defer series.mutex.RUnlock()

	cutoff := time.Now().Add(-duration)
	count := 0

	for _, dp := range series.Values {
		if dp.Timestamp.After(cutoff) {
			count++
		}
	}

	return float64(count) / duration.Seconds()
}