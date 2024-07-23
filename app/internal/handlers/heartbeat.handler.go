package handlers

import (
	configs "github.com/gulizay91/template-go-consumer/config"
	log "github.com/sirupsen/logrus"
	"time"
)

type HeartbeatHandler struct {
	AppConfig *configs.Config
}

func NewHeartbeatHandler(config *configs.Config) HeartbeatHandler {
	return HeartbeatHandler{AppConfig: config}
}

// Heartbeat Function to perform a periodic task every 2 seconds
func (h HeartbeatHandler) Heartbeat(message string) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println(message)
		}
	}
}
