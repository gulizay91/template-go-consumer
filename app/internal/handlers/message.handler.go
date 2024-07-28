package handlers

import (
	"encoding/json"
	"github.com/gulizay91/template-go-consumer/config"
	"github.com/gulizay91/template-go-consumer/pkg/rabbitmq"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// MessageHandler struct for handling messages
type MessageHandler struct {
	AppConfig   *config.Config
	AmqpChannel *amqp.Channel
	QueueConfig *config.QueueConfig
}

// NewMessageHandler creates a new instance of MessageHandler
func NewMessageHandler(config *config.Config, amqpChannel *amqp.Channel, queueName string) rabbitmq.IConsumerHandler {
	queueConfig := config.GetQueueConfig(queueName)
	if queueConfig == nil {
		log.Errorf("Queue configuration not found for %s", queueName)
		return nil
	}
	handler := &MessageHandler{
		AppConfig:   config,
		AmqpChannel: amqpChannel,
		QueueConfig: queueConfig,
	}

	return handler
}

// ConsumeMessages function to consume messages from the queue
func (h *MessageHandler) ConsumeMessages() {
	messages, err := h.AmqpChannel.Consume(
		h.QueueConfig.Name,
		"",
		false, // auto-ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Consume message failed: %v", err)
	}

	go h.consumeMessage(messages)
}

func (h *MessageHandler) consumeMessage(messages <-chan amqp.Delivery) {
	for d := range messages {
		if err := h.processMessage(d); err != nil {
			rabbitmq.RetryOrSendToDLQ(h.AmqpChannel, d, h.QueueConfig.Name, h.QueueConfig.Retry.MaxRetries, h.QueueConfig.Retry.RetryDelaySeconds)
		} else {
			if err := d.Ack(false); err != nil {
				log.Errorf("Failed to ack message: %v", err)
			}
		}
	}
}

// processMessage processes a single message and returns an error if any issue occurs
func (h *MessageHandler) processMessage(d amqp.Delivery) error {
	log.Infof("Received message body: %s", d.Body)

	var message map[string]interface{}
	if err := json.Unmarshal(d.Body, &message); err != nil {
		log.Errorf("JSON parse failed: %s. Message body: %s", err, string(d.Body))
		return err
	}

	// PS: DO WHAT YOU WANT

	log.WithFields(log.Fields{
		"payload": message,
	}).Info("Consumed message")
	return nil
}
