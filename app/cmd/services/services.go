package services

import (
	log "github.com/sirupsen/logrus"
	stdLog "log"
	"time"
)

func Run() {
	InitConfig()
	stdLog.Printf("Configuration Initialized for %s", config.Service.Name)

	// Create a forever channel to keep the program running
	forever := make(chan bool)

	// Start the task function as a goroutine
	go heartbeatTask(config.Service.Name + " heartbeat")

	log.Printf("Consumer started. Press Ctrl+C to exit.")
	<-forever // Keeps the program running indefinitely
}

// Function to perform a periodic task every 2 seconds
func heartbeatTask(message string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println(message)
		}
	}
}
