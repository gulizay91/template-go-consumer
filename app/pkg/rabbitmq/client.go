package rabbitmq

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"time"
)

type IConsumerHandler interface {
	ConsumeMessages()
}

// MessageBrokerManager manages RabbitMQ connections and channels
type MessageBrokerManager struct {
	Conn     *amqp.Connection
	Channels map[string]*amqp.Channel
	Hosts    []string
	Username string
	Password string
}

type ExchangeType string

const (
	Direct ExchangeType = "direct"
	Fanout ExchangeType = "fanout"
	Topic  ExchangeType = "topic"
)

const DeadLetterQueueSuffix = "_error"
const HeaderRetryCountKey = "x-retry-count"

var (
	ErrConnectionClosed = errors.New("connection closed")
)

// ValidateJSON checks if the given string is a valid JSON
func validateJSON(input string) error {
	var js map[string]interface{}
	return json.Unmarshal([]byte(input), &js)
}

// ConnectRabbitMQ connects to a RabbitMQ server
func ConnectRabbitMQ(hosts []string, username string, password string) (*amqp.Connection, error) {
	var err error
	var conn *amqp.Connection
	for _, host := range hosts {
		rabbitmqURL := fmt.Sprintf("amqp://%s:%s@%s/",
			username,
			password,
			host,
		)
		conn, err = amqp.Dial(rabbitmqURL)
		if err == nil {
			log.Printf("Connected to RabbitMQ at %s", host)
			return conn, nil
		}
		log.Printf("Failed to connect to RabbitMQ at %s: %v", host, err)
	}
	return nil, err
}

// CreateChannel creates a new RabbitMQ channel
func CreateChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}
	return ch, nil
}

// ReconnectRabbitMQ tries to reconnect to RabbitMQ and create a new channel
func ReconnectRabbitMQ(hosts []string, username string, password string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := ConnectRabbitMQ(hosts, username, password)
	if err != nil {
		return nil, nil, err
	}
	ch, err := CreateChannel(conn)
	if err != nil {
		return nil, nil, err
	}
	return conn, ch, nil
}

// IsConnectionOpen checks if the RabbitMQ connection is open
func IsConnectionOpen(conn *amqp.Connection) bool {
	return conn != nil && conn.IsClosed() == false
}

// IsChannelOpen checks if a RabbitMQ channel is open
func IsChannelOpen(ch *amqp.Channel) bool {
	if ch == nil {
		return false
	}
	// Attempt a no-op publish to check if the channel is open
	err := ch.Publish("", "", false, false, amqp.Publishing{})
	return err == nil
}

// NewMessageBrokerManager creates a new MessageBrokerManager instance
func NewMessageBrokerManager(hosts []string, username string, password string) (*MessageBrokerManager, error) {
	conn, err := ConnectRabbitMQ(hosts, username, password)
	if err != nil {
		return nil, err
	}
	return &MessageBrokerManager{
		Conn:     conn,
		Channels: make(map[string]*amqp.Channel),
		Hosts:    hosts,
		Username: username,
		Password: password,
	}, nil
}

// GetChannel returns an existing channel or creates a new one if it does not exist
func (mbm *MessageBrokerManager) GetChannel(queueName string) (*amqp.Channel, error) {
	ch, exists := mbm.Channels[queueName]
	if exists && IsChannelOpen(ch) {
		return ch, nil
	}
	// Recreate channel if it does not exist or is not open
	newCh, err := CreateChannel(mbm.Conn)
	if err != nil {
		return nil, err
	}
	mbm.Channels[queueName] = newCh
	return newCh, nil
}

// MonitorConnection continuously monitors the RabbitMQ connection and reconnects if needed
func (mbm *MessageBrokerManager) MonitorConnection() {
	for {
		if !IsConnectionOpen(mbm.Conn) {
			log.Println("RabbitMQ connection lost, reconnecting...")
			var err error
			mbm.Conn, err = ConnectRabbitMQ(mbm.Hosts, mbm.Username, mbm.Password)
			if err != nil {
				log.Printf("Failed to reconnect RabbitMQ: %v", err)
				time.Sleep(5 * time.Second) // Wait before retrying
				continue
			}

			// Recreate all channels
			newChannels := make(map[string]*amqp.Channel)
			for queueName := range mbm.Channels {
				newCh, err := CreateChannel(mbm.Conn)
				if err != nil {
					log.Printf("Failed to create channel for %s: %v", queueName, err)
					continue
				}
				newChannels[queueName] = newCh
			}
			mbm.Channels = newChannels
		}
		time.Sleep(10 * time.Second) // Check connection every 10 seconds
	}
}

// DeclareExchange declares a RabbitMQ exchange
func DeclareExchange(ch *amqp.Channel, name, kind string) {
	err := ch.ExchangeDeclare(
		name,  // name of the exchange
		kind,  // type
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Errorf("Failed to declare exchange %s: %v", name, err)
	}
}

// DeclareQueue declares a RabbitMQ queue
func DeclareQueue(ch *amqp.Channel, name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table, hasDlq bool) {
	if hasDlq {
		if args == nil {
			args = amqp.Table{}
		}
		args["x-dead-letter-exchange"] = ""                              // default exchange
		args["x-dead-letter-routing-key"] = name + DeadLetterQueueSuffix // dead-letter queue name
	}
	_, err := ch.QueueDeclare(
		name,       // name
		durable,    // durable
		autoDelete, // delete when unused
		exclusive,  // exclusive
		noWait,     // no-wait
		args,       // arguments
	)
	if err != nil {
		log.Errorf("Failed to declare queue %s: %v", name, err)
	}
}

// BindQueue binds a RabbitMQ queue to an exchange
func BindQueue(ch *amqp.Channel, queueName, routingKey, exchangeName string) {
	err := ch.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		log.Errorf("Failed to bind queue %s to exchange %s: %v", queueName, exchangeName, err)
	}
}

// DeleteQueueIfExistsAndEmpty deletes a RabbitMQ queue if it exists and is empty
func DeleteQueueIfExistsAndEmpty(ch *amqp.Channel, queueName string) {
	queue, err := ch.QueueInspect(queueName)
	if err != nil {
		log.Warnf("Queue %s does not exist or cannot be inspected: %v", queueName, err)
		return
	}

	if queue.Messages == 0 {
		_, err := ch.QueueDelete(queueName, false, false, false)
		if err != nil {
			log.Errorf("Failed to delete queue %s: %v", queueName, err)
		} else {
			log.Infof("Queue %s deleted", queueName)
		}
	} else {
		log.Warnf("Queue %s is not empty and cannot be deleted", queueName)
	}
}

// SendMessageToDLQ sends the message to the Dead Letter Queue (DLQ) and acknowledges the original message
func SendMessageToDLQ(ch *amqp.Channel, queueName string, d amqp.Delivery) {
	dlqName := queueName + DeadLetterQueueSuffix
	messageBody := d.Body

	// Ensure the DLQ is declared
	DeclareQueue(ch, dlqName, true, false, false, false, nil, false)

	// Publish to the DLQ
	message := amqp.Publishing{
		ContentType: "application/json",
		Body:        messageBody,
	}

	err := ch.Publish(
		"",      // exchange
		dlqName, // routing key
		false,   // mandatory
		false,   // immediate
		message,
	)
	if err != nil {
		log.Errorf("Failed to send message to DLQ %s: %v", dlqName, err)
	} else {
		log.Infof("Message sent to DLQ %s", dlqName)
	}

	// Acknowledge the original message after sending it to DLQ
	if err := d.Ack(false); err != nil {
		log.Errorf("Failed to ack message after sending to DLQ: %v", err)
	}
}

// RetryMessage retries the failed message
func RetryMessage(ch *amqp.Channel, d amqp.Delivery, retryCount int32) error {
	// Retrieve existing headers or initialize a new map
	headers := d.Headers
	if headers == nil {
		headers = make(map[string]interface{})
	}

	// Update the retry count
	headers[HeaderRetryCountKey] = retryCount

	// Publish the message with updated headers
	err := ch.Publish(
		d.Exchange,   // exchange
		d.RoutingKey, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        d.Body,
			Headers:     headers,
		},
	)
	if err != nil {
		log.Errorf("Failed to retry message: %v", err)
	}
	return err
}

// RetryOrSendToDLQ retries the message if needed or sends it to the Dead Letter Queue (DLQ)
func RetryOrSendToDLQ(ch *amqp.Channel, d amqp.Delivery, queueName string, retryMaxCount int, retryDelaySeconds int) {
	// Retrieve retry count from message headers
	retryCountHeader, ok := d.Headers[HeaderRetryCountKey]
	if !ok {
		retryCountHeader = int32(0) // Default to 0 if not present
	}

	// Convert retryCountHeader to integer
	retryCount, ok := retryCountHeader.(int32)
	if !ok {
		log.Errorf("Retry count header is not of type int32")
		retryCount = 0 // Fallback to 0 if conversion fails
	}

	if retryCount < int32(retryMaxCount) {
		// Wait before retrying
		time.Sleep(time.Second * time.Duration(retryDelaySeconds))
		if err := RetryMessage(ch, d, retryCount+1); err != nil {
			log.Errorf("Failed to retry message: %v", err)
		}
		// Acknowledge the original message to remove it from the queue
		if err := d.Ack(false); err != nil {
			log.Errorf("Failed to ack message after retry: %v", err)
		}
	} else {
		// If retries exhausted, send to DLQ
		SendMessageToDLQ(ch, queueName, d)
	}
}

func PublishEvent(ch *amqp.Channel, messageJSON []byte, exchangeName string, routingKey string) error {

	// Validate JSON format
	if err := validateJSON(string(messageJSON)); err != nil {
		return err
	}

	err := ch.Publish(
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageJSON,
		})
	if err != nil {
		log.Errorf("Failed to publish a message: %s", err)
		return err
	}

	return nil
}
