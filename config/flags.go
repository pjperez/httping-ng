package config

import (
	"flag"
)

type Config struct {
	URL        string
	Workers    int
	IntervalMs int
	Count      int
	Histogram  bool
	Graph      bool
	ServerMode bool
	JSONOutput bool
	UserAgent  string
	Buckets    int
}

func ParseFlags() *Config {
	var cfg Config
	flag.StringVar(&cfg.URL, "url", "", "Target URL to ping")
	flag.IntVar(&cfg.Workers, "n", 1, "Number of concurrent workers (default 1)")
	flag.IntVar(&cfg.IntervalMs, "i", 1000, "Interval between requests (milliseconds, default 1000)")
	flag.IntVar(&cfg.Count, "count", 10, "Number of requests to send before stopping (default 10)")
	flag.BoolVar(&cfg.Histogram, "histogram", false, "Show latency histogram at the end")
	flag.BoolVar(&cfg.Graph, "graph", false, "Show latency graph at the end (not implemented yet)")
	flag.BoolVar(&cfg.ServerMode, "server", false, "Run as HTTP server instead of client")
	flag.BoolVar(&cfg.JSONOutput, "json", false, "Output results as JSON")
	flag.StringVar(&cfg.UserAgent, "user-agent", "httping-ng https://github.com/pjperez/httping-ng", "Custom User-Agent header")
	flag.IntVar(&cfg.Buckets, "buckets", 10, "Number of buckets for histogram display")
	flag.Parse()
	return &cfg
}
