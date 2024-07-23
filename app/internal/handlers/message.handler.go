package handlers

import (
	"encoding/json"
	"time"

	"github.com/gulizay91/template-go-consumer/config"
	"github.com/gulizay91/template-go-consumer/internal/models"
	"github.com/gulizay91/template-go-consumer/pkg/rabbitmq"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// MessageHandler struct for handling messages
type MessageHandler struct {
	AppConfig   *config.Config
	AmqpChannel *amqp.Channel
	QueueName   string
}

// NewMessageHandler creates a new instance of MessageHandler
func NewMessageHandler(config *config.Config, amqpChannel *amqp.Channel, queueName string) rabbitmq.IConsumerHandler {
	handler := &MessageHandler{
		AppConfig:   config,
		AmqpChannel: amqpChannel,
		QueueName:   queueName,
	}

	return handler
}

// ConsumeMessages function to consume messages from the queue
func (h *MessageHandler) ConsumeMessages() {
	messages, err := h.AmqpChannel.Consume(
		h.QueueName,
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
			h.retryOrSendToDLQ(d)
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

	log.WithFields(log.Fields{
		"payload": message,
	}).Info("Consumed message")
	return nil
}

// retryOrSendToDLQ retries the message if needed or sends it to the Dead Letter Queue (DLQ)
func (h *MessageHandler) retryOrSendToDLQ(d amqp.Delivery) {
	queueConfig := h.AppConfig.GetQueueConfig(h.QueueName)
	if queueConfig == nil {
		log.Errorf("Queue configuration not found for %s", models.TemplateMessageQueue)
		return
	}

	// Retrieve retry count from message headers
	retryCountHeader, ok := d.Headers[rabbitmq.HeaderRetryCountKey]
	if !ok {
		retryCountHeader = int32(0) // Default to 0 if not present
	}

	// Convert retryCountHeader to integer
	retryCount, ok := retryCountHeader.(int32)
	if !ok {
		log.Errorf("Retry count header is not of type int32")
		retryCount = 0 // Fallback to 0 if conversion fails
	}

	if retryCount < int32(queueConfig.Retry.MaxRetries) {
		// Wait before retrying
		time.Sleep(time.Second * time.Duration(queueConfig.Retry.RetryDelaySeconds))
		if err := rabbitmq.RetryMessage(h.AmqpChannel, d, retryCount+1); err != nil {
			log.Errorf("Failed to retry message: %v", err)
		}
		// Acknowledge the original message to remove it from the queue
		if err := d.Ack(false); err != nil {
			log.Errorf("Failed to ack message after retry: %v", err)
		}
	} else {
		// If retries exhausted, send to DLQ
		rabbitmq.SendMessageToDLQ(h.AmqpChannel, h.QueueName, d)
	}
}
