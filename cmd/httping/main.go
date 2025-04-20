package main

import (
	"os"
	"time"

	"github.com/pjperez/httping-ng/client"
	"github.com/pjperez/httping-ng/config"
	"github.com/pjperez/httping-ng/logging"
	"github.com/pjperez/httping-ng/server"
)

func main() {
	cfg := config.ParseFlags()

	if cfg.ServerMode {
		startServer()
	} else {
		if err := client.Run(cfg); err != nil {
			logging.Error("main", "Error: %v", err)
			os.Exit(1)
		}
	}
}

func startServer() {
	addr := ":8080"
	delay := 0 * time.Millisecond

	srv := server.NewServer(addr, delay)
	if err := srv.Start(); err != nil {
		logging.Error("main", "Server error: %v", err)
		os.Exit(1)
	}
}
