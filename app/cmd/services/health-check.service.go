package services

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// RegisterHealthCheckServer starts a simple HTTP server for health checks
func RegisterHealthCheckServer() {
	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/ready", readinessCheckHandler)

	go func() {
		addr := fmt.Sprintf(":%s", config.Service.Port)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Health check server failed: %v", err)
		}
	}()
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Perform any necessary health checks here
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("✅"))
	if err != nil {
		return
	}
}

func readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Perform any necessary readiness checks here
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("✅"))
	if err != nil {
		return
	}
}
