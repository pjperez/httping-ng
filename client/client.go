package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pjperez/httping-ng/config"
	"github.com/pjperez/httping-ng/logging"
	"github.com/pjperez/httping-ng/metrics"
)

type result struct {
	latency   time.Duration
	success   bool
	seq       int
	workerID  int
	err       error
	status    int
	respBytes int
}

func Run(cfg *config.Config) error {
	if cfg.URL == "" {
		return errors.New("no URL provided")
	}

	if !strings.HasPrefix(cfg.URL, "http://") && !strings.HasPrefix(cfg.URL, "https://") {
		cfg.URL = "https://" + cfg.URL
		if !cfg.JSONOutput {
			logging.Info("client", "No scheme provided, defaulting to %s", cfg.URL)
		}
	}

	if !cfg.JSONOutput {
		logging.Info("client", "Starting HTTP GET requests to %s", cfg.URL)
	}

	m := metrics.NewMetricBucket()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	var wg sync.WaitGroup
	results := make(chan result, 10000)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(httpClient, cfg, m, id, results)
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var sent, received int
	var latencies []time.Duration

loop:
	for {
		select {
		case r, ok := <-results:
			if !ok {
				break loop
			}
			sent++
			if r.success {
				received++
				latencies = append(latencies, r.latency)
				if !cfg.JSONOutput {
					fmt.Printf("\033[32mHTTP GET %-30s worker=%-2d seq=%-3d status=%-3d size=%-5dB time=%6.2f ms\033[0m\n",
						cfg.URL, r.workerID, r.seq, r.status, r.respBytes, float64(r.latency.Microseconds())/1000.0)
				}
			} else {
				if !cfg.JSONOutput {
					if r.err != nil {
						fmt.Printf("\033[31mRequest to %-30s worker=%-2d seq=%-3d failed: %v\033[0m\n",
							cfg.URL, r.workerID, r.seq, r.err)
					} else {
						fmt.Printf("\033[31mRequest to %-30s worker=%-2d seq=%-3d failed: HTTP %d\033[0m\n",
							cfg.URL, r.workerID, r.seq, r.status)
					}
				}
			}

			if cfg.Count > 0 && sent >= cfg.Count {
				break loop
			}

		case <-signals:
			if !cfg.JSONOutput {
				fmt.Println("\nReceived interrupt signal, shutting down gracefully...")
			}
			break loop
		}
	}

	if cfg.JSONOutput {
		printJSONOutput(cfg.URL, sent, received, latencies)
		return nil
	}

	fmt.Printf("\n--- %s httping statistics ---\n", cfg.URL)
	fmt.Printf("Requests : %d sent, %d received, %.1f%% loss\n",
		sent, received, lossPercent(sent, received))

	if len(latencies) > 0 {
		min, avg, max := calcLatencyStats(latencies)
		fmt.Printf("RTT      : min=%.2f ms, avg=%.2f ms, max=%.2f ms\n",
			min, avg, max)
		if cfg.Histogram {
			m.PrintHistogram(cfg.Buckets)
		} else {
			m.PrintPercentiles()
		}

	}

	return nil
}

func worker(client *http.Client, cfg *config.Config, m *metrics.MetricBucket, workerID int, results chan result) {
	ticker := time.NewTicker(time.Duration(cfg.IntervalMs) * time.Millisecond)
	defer ticker.Stop()

	seq := 0

	for range ticker.C {
		start := time.Now()

		req, err := http.NewRequest("GET", cfg.URL, nil)
		if err != nil {
			results <- result{success: false, seq: seq, workerID: workerID, err: err}
			seq++
			continue
		}
		req.Header.Set("User-Agent", cfg.UserAgent)

		resp, err := client.Do(req)
		if err != nil {
			results <- result{success: false, seq: seq, workerID: workerID, err: err}
			seq++
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			results <- result{success: false, seq: seq, workerID: workerID, err: err}
			seq++
			continue
		}

		latency := time.Since(start)
		success := resp.StatusCode < 400
		if success {
			m.Record(latency)
		}

		results <- result{
			latency:   latency,
			success:   success,
			seq:       seq,
			workerID:  workerID,
			status:    resp.StatusCode,
			respBytes: len(body),
		}
		seq++
	}
}

func lossPercent(sent, received int) float64 {
	if sent == 0 {
		return 100.0
	}
	lost := sent - received
	return (float64(lost) / float64(sent)) * 100
}

func calcLatencyStats(latencies []time.Duration) (min, avg, max float64) {
	if len(latencies) == 0 {
		return 0, 0, 0
	}

	minDur := latencies[0]
	maxDur := latencies[0]
	var sum time.Duration

	for _, l := range latencies {
		if l < minDur {
			minDur = l
		}
		if l > maxDur {
			maxDur = l
		}
		sum += l
	}

	min = float64(minDur.Microseconds()) / 1000.0
	avg = float64(sum.Microseconds()) / 1000.0 / float64(len(latencies))
	max = float64(maxDur.Microseconds()) / 1000.0
	return
}

func calcPercentile(latencies []time.Duration, p int) float64 {
	if len(latencies) == 0 {
		return 0
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	index := int(float64(p) / 100.0 * float64(len(latencies)))
	if index >= len(latencies) {
		index = len(latencies) - 1
	}

	return float64(latencies[index].Microseconds()) / 1000.0
}

type Output struct {
	Target      string  `json:"target"`
	TotalSent   int     `json:"total_sent"`
	TotalRecv   int     `json:"total_received"`
	LossPercent float64 `json:"loss_percent"`
	MinRTT      float64 `json:"rtt_min_ms"`
	AvgRTT      float64 `json:"rtt_avg_ms"`
	MaxRTT      float64 `json:"rtt_max_ms"`
	P50         float64 `json:"rtt_p50_ms"`
	P75         float64 `json:"rtt_p75_ms"`
	P90         float64 `json:"rtt_p90_ms"`
	P99         float64 `json:"rtt_p99_ms"`
}

func printJSONOutput(target string, sent, received int, latencies []time.Duration) {
	if len(latencies) == 0 {
		return
	}

	min, avg, max := calcLatencyStats(latencies)
	p50 := calcPercentile(latencies, 50)
	p75 := calcPercentile(latencies, 75)
	p90 := calcPercentile(latencies, 90)
	p99 := calcPercentile(latencies, 99)

	out := Output{
		Target:      target,
		TotalSent:   sent,
		TotalRecv:   received,
		LossPercent: lossPercent(sent, received),
		MinRTT:      min,
		AvgRTT:      avg,
		MaxRTT:      max,
		P50:         p50,
		P75:         p75,
		P90:         p90,
		P99:         p99,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}
