package services

import (
	configs "github.com/gulizay91/template-go-consumer/config"
	"github.com/gulizay91/template-go-consumer/pkg/rabbitmq"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// BrokerManager Global channel manager instance
var BrokerManager *rabbitmq.MessageBrokerManager

// HandlerFactory is a function type that creates an IConsumerHandler instance
type HandlerFactory func(configs *configs.Config, amqpChannel *amqp.Channel, queueName string) rabbitmq.IConsumerHandler

// InitRabbitMQ Initialize RabbitMQ connection and setup exchanges and queues
func InitRabbitMQ() {
	var err error
	BrokerManager, err = rabbitmq.NewMessageBrokerManager(config.RabbitMQ.Hosts, config.RabbitMQ.Username, config.RabbitMQ.Password)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ connection: %v", err)
	}
	log.Println("RabbitMQ initialized.")

	// Start monitoring the RabbitMQ connection
	go BrokerManager.MonitorConnection()
}

// RegisterConsumer registers a consumer for a specific queue
func RegisterConsumer(queueName string, factory HandlerFactory) rabbitmq.IConsumerHandler {

	amqpChannel, err := BrokerManager.GetChannel(queueName)
	if err != nil {
		log.Fatalf("Failed to get RabbitMQ channel for queue %s: %v", queueName, err)
	}

	// Set prefetch count from config
	queueConfig := config.GetQueueConfig(queueName)
	prefetchCount := 10 // Fallback to default if queueConfig is nil or prefetchCount is not set
	if queueConfig != nil {
		prefetchCount = queueConfig.PrefetchCount
	}
	if err := amqpChannel.Qos(prefetchCount, 0, false); err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	handler := factory(config, amqpChannel, queueName)

	log.Printf("Registered Consumer for %s queue.", queueName)

	return handler
}
