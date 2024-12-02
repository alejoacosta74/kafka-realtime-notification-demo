package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"kafka-notify/pkg/models"

	"github.com/IBM/sarama"
	"github.com/alejoacosta74/go-logger"
)

// Consumer struct holds a reference to the notification store for persisting messages
type Consumer struct {
	store *NotificationStore
}

// Setup is called when the consumer group session starts
// Returns nil as no setup is needed
func (*Consumer) Setup(sarama.ConsumerGroupSession) error { return nil }

// Cleanup is called when the consumer group session ends
// Returns nil as no cleanup is needed
func (*Consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim handles the consumption of messages from a Kafka partition
// Implements the sarama.ConsumerGroupHandler interface
func (consumer *Consumer) ConsumeClaim(
	sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Continuously read messages from the claim's message channel
	for msg := range claim.Messages() {
		// Extract the userID from the message key
		userID := string(msg.Key)
		// Create a notification object to store the message data
		var notification models.Notification
		// Deserialize the JSON message value into the notification struct
		err := json.Unmarshal(msg.Value, &notification)
		if err != nil {
			// Log any JSON parsing errors and continue to next message
			logger.Errorf("failed to unmarshal notification: %v", err)
			continue
		}
		// Store the notification in the notification store for the user
		consumer.store.Add(userID, notification)
		// Mark the message as processed
		sess.MarkMessage(msg, "")
	}
	return nil
}

// initializeConsumerGroup creates and configures a new Kafka consumer group
// Returns the consumer group instance and any error that occurred
func initializeConsumerGroup() (sarama.ConsumerGroup, error) {
	// Create default Sarama configuration
	config := sarama.NewConfig()

	logger.Infof("Attempting to connect to Kafka server at %s", KafkaServerAddress)

	// Enable error reporting for the consumer group
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Disable IPv6 resolution
	config.Net.SASL.Enable = false
	config.Net.DialTimeout = 10 * time.Second

	// Create a new consumer group with the specified address, group name and config
	consumerGroup, err := sarama.NewConsumerGroup(
		[]string{KafkaServerAddress}, ConsumerGroup, config)
	if err != nil {
		// Return error if consumer group creation fails
		logger.Errorf("failed to initialize consumer group: %v", err)
		return nil, fmt.Errorf("failed to initialize consumer group: %w", err)
	}
	logger.Infof("Successfully connected to Kafka server at %s", KafkaServerAddress)
	return consumerGroup, nil
}

// setupConsumerGroup initializes and runs the consumer group processing loop
// Takes a context for cancellation and notification store for persistence
func setupConsumerGroup(ctx context.Context, store *NotificationStore) {
	// Initialize the consumer group
	consumerGroup, err := initializeConsumerGroup()
	if err != nil {
		// Log any initialization errors
		logger.Fatalf("initialization error: %v", err)
	}
	// Ensure consumer group is closed when function returns
	defer consumerGroup.Close()

	// Create consumer instance with reference to notification store
	consumer := &Consumer{
		store: store,
	}
	logger.Infof("Starting to consume from topic: %s", ConsumerTopic)
	logger.Infof("Consumer group: %s", ConsumerGroup)

	// Run continuous processing loop
	for {
		select {
		case <-ctx.Done():
			logger.Warn("Context cancelled, stopping consumer")
			return
		default:
			// Consume messages from the topic
			if err := consumerGroup.Consume(ctx, []string{ConsumerTopic}, consumer); err != nil {
				if ctx.Err() != nil {
					// Context was cancelled while consuming
					logger.Warn("Context cancelled while consuming, stopping consumer")
					return
				}
				// Real consumption error occurred
				logger.Errorf("Error consuming topic: %v", err)
				time.Sleep(time.Second)
			}
		}
	}
}
