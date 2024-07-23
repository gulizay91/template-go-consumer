package services

import (
	"github.com/gulizay91/template-go-consumer/internal/handlers"
	"github.com/gulizay91/template-go-consumer/internal/models"
	log "github.com/sirupsen/logrus"
)

// RegisterGoRoutines registers and starts all necessary goroutines
func RegisterGoRoutines() {
	// Write your goroutines
	heartbeatHandler := handlers.NewHeartbeatHandler(config)

	// Start the handler as a goroutine
	go heartbeatHandler.Heartbeat(config.Service.Name + " heartbeat")

	// Register consumer with Handler
	messageHandler := RegisterConsumer(string(models.TemplateMessageQueue), handlers.NewMessageHandler)
	if messageHandler != nil {
		go messageHandler.ConsumeMessages()
	}

	log.Println("Consumer started. Press Ctrl+C to exit.")

	// Register graceful shutdown handler
	RegisterGracefulShutdown(func(reason interface{}) {
		log.Printf("Application is shutting down: %v", reason)
		// Perform cleanup actions if needed
	})

	// Wait for shutdown signal
	<-ShutdownChannel()

	log.Println("Shutdown complete.")
}
