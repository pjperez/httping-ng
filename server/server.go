package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/pjperez/httping-ng/logging"
)

type Server struct {
	Addr  string
	Delay time.Duration
}

func NewServer(addr string, delay time.Duration) *Server {
	return &Server{
		Addr:  addr,
		Delay: delay,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handlePing)

	server := &http.Server{
		Addr:    s.Addr,
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		logging.Info("server", "Listening on %s", s.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error("server", "Server error: %v", err)
		}
	}()

	<-stop
	logging.Info("server", "Received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	if s.Delay > 0 {
		time.Sleep(s.Delay)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, "pong")
}
