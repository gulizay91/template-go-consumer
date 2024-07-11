package services

import (
	"os"
	"os/signal"
	"syscall"
)

var shutdown = make(chan struct{})

// RegisterGracefulShutdown registers graceful shutdown handler
func RegisterGracefulShutdown(shutdownFunc func(interface{})) {
	// Create a channel to signify a signal being sent
	gracefulShutdownSignal := make(chan os.Signal, 1)
	// SIGINT handles Ctrl+C locally.
	// SIGTERM handles termination signal.
	signal.Notify(gracefulShutdownSignal, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Wait for OS signals or shutdown signal
	go func() {
		sig := <-gracefulShutdownSignal
		shutdownFunc(sig)
	}()
}

// ShutdownChannel returns the shutdown channel
func ShutdownChannel() <-chan struct{} {
	return shutdown
}
