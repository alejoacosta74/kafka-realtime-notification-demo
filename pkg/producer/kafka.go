package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"kafka-notify/pkg/models"
	"strconv"

	"github.com/IBM/sarama"
	"github.com/alejoacosta74/go-logger"
	"github.com/gin-gonic/gin"
)

// setupProducer initializes and configures a Kafka producer for synchronous message sending
func setupProducer(ctx context.Context) (sarama.SyncProducer, error) {
	// Create a new Kafka configuration with default settings
	config := sarama.NewConfig()
	// Enable producer acknowledgments so we can confirm messages were sent successfully
	config.Producer.Return.Successes = true
	// Create a new synchronous producer connected to our Kafka broker
	producer, err := sarama.NewSyncProducer([]string{KafkaServerAddress}, config)
	// If producer creation fails, wrap the error with additional context
	if err != nil {
		logger.Error("Failed to setup producer", "error", err)
		return nil, fmt.Errorf("failed to setup producer: %w", err)
	}
	logger.Info("New kafka producer created")
	// Return the successfully created producer

	// Monitor context for shutting down the producer
	go func() {
		<-ctx.Done()
		producer.Close()
		logger.Warn("Kafka producer closed")
	}()

	return producer, nil
}

// sendKafkaProducerMessage sends a notification message to Kafka from one user to another
func sendKafkaProducerMessage(producer sarama.SyncProducer,
	users []models.User, ctx *gin.Context, fromID, toID int) error {
	// Get the message content from the HTTP form data
	message := ctx.PostForm("message")

	// Find the sender user by their ID
	fromUser, err := findUserByID(fromID, users)
	if err != nil {
		logger.Error("Failed to find sender user", "error", err)
		return err
	}

	// Find the recipient user by their ID
	toUser, err := findUserByID(toID, users)
	if err != nil {
		logger.Error("Failed to find recipient user", "error", err)
		return err
	}

	// Create a notification object with the sender, recipient and message
	notification := models.Notification{
		From:    fromUser,
		To:      toUser,
		Message: message,
	}

	// Convert the notification struct to JSON bytes
	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		logger.Error("Failed to marshal notification", "error", err)
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Create a Kafka producer message with the topic, recipient ID as key, and JSON as value
	msg := &sarama.ProducerMessage{
		Topic: KafkaTopic,
		Key:   sarama.StringEncoder(strconv.Itoa(toUser.ID)),
		Value: sarama.StringEncoder(notificationJSON),
	}

	// Send the message to Kafka and return any error
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		logger.Error("Failed to send message to Kafka", "error", err)
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	logger.Info("Message sent to Kafka", "partition: ", partition, "offset: ", offset)
	return nil
}
