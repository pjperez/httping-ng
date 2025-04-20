package metrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type MetricBucket struct {
	mu        sync.Mutex
	latencies []time.Duration
	buckets   map[int]int
}

func NewMetricBucket() *MetricBucket {
	return &MetricBucket{
		buckets: make(map[int]int),
	}
}

func (m *MetricBucket) Record(latency time.Duration) {
	ms := int(latency.Milliseconds())
	bucket := (ms / 10) * 10

	m.mu.Lock()
	defer m.mu.Unlock()
	m.latencies = append(m.latencies, latency)
	m.buckets[bucket]++
}

func (m *MetricBucket) Histogram() map[int]int {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(map[int]int)
	for k, v := range m.buckets {
		result[k] = v
	}
	return result
}

func (m *MetricBucket) Latencies() []time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()

	snapshot := make([]time.Duration, len(m.latencies))
	copy(snapshot, m.latencies)
	return snapshot
}

func (m *MetricBucket) PrintPercentiles() {
	latencies := m.Latencies()
	if len(latencies) == 0 {
		fmt.Println("No latency data collected.")
		return
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	p50 := percentile(latencies, 50)
	p75 := percentile(latencies, 75)
	p90 := percentile(latencies, 90)
	p99 := percentile(latencies, 99)

	fmt.Println("\nLatency percentiles:")
	fmt.Printf("p50: %.2f ms\n", p50)
	fmt.Printf("p75: %.2f ms\n", p75)
	fmt.Printf("p90: %.2f ms\n", p90)
	fmt.Printf("p99: %.2f ms\n", p99)
}

func percentile(latencies []time.Duration, p int) float64 {
	if len(latencies) == 0 {
		return 0
	}

	index := int(float64(p) / 100.0 * float64(len(latencies)))
	if index >= len(latencies) {
		index = len(latencies) - 1
	}

	return float64(latencies[index].Microseconds()) / 1000.0
}

func (m *MetricBucket) PrintHistogram(n int) {
	latencies := m.Latencies()
	if len(latencies) == 0 {
		fmt.Println("No latency data collected.")
		return
	}

	// Find min and max in ms
	min := int(latencies[0].Milliseconds())
	max := min
	for _, l := range latencies {
		ms := int(l.Milliseconds())
		if ms < min {
			min = ms
		}
		if ms > max {
			max = ms
		}
	}

	if max == min {
		fmt.Println("All latencies are identical; skipping histogram.")
		return
	}

	rangeMs := max - min
	bucketSize := float64(rangeMs) / float64(n)

	counts := make([]int, n)

	for _, l := range latencies {
		ms := int(l.Milliseconds())
		pos := float64(ms-min) / float64(rangeMs)
		idx := int(pos * float64(n))
		if idx >= n {
			idx = n - 1
		}
		counts[idx]++
	}

	// Get max count for scaling
	maxCount := 0
	for _, v := range counts {
		if v > maxCount {
			maxCount = v
		}
	}

	fmt.Printf("\nLatency histogram (exact %d buckets):\n", n)
	for i := 0; i < n; i++ {
		from := float64(min) + bucketSize*float64(i)
		to := from + bucketSize
		bar := barString(counts[i], maxCount, 40)
		fmt.Printf("%6.0f–%6.0f ms | %-40s %3d\n", from, to, bar, counts[i])
	}
}

func barString(value, max, width int) string {
	if max == 0 {
		return ""
	}
	barLen := (value * width) / max
	return strings.Repeat("█", barLen)
}
