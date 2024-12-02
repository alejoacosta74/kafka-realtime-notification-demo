package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"kafka-notify/pkg/models"

	"github.com/gin-gonic/gin"
)

const (
	ConsumerGroup      = "notifications-group"
	ConsumerTopic      = "notifications"
	ConsumerPort       = ":8081"
	KafkaServerAddress = "192.168.5.142:9092"
)

// UserNotifications is a custom type that maps user IDs to their slice of notifications
// This allows efficient storage and retrieval of notifications per user
type UserNotifications map[string][]models.Notification

// NotificationStore provides thread-safe storage of user notifications
// Uses a mutex to safely handle concurrent access to the data
type NotificationStore struct {
	data UserNotifications // Holds the actual notification data
	mu   sync.RWMutex      // RWMutex allows multiple readers but only one writer
}

// Add safely adds a new notification to a user's notification list
// Uses a write lock to ensure thread-safe updates to the data
func (ns *NotificationStore) Add(userID string,
	notification models.Notification) {
	ns.mu.Lock()                                            // Acquire exclusive write lock
	defer ns.mu.Unlock()                                    // Release lock when function returns
	ns.data[userID] = append(ns.data[userID], notification) // Append new notification to user's list
}

// Get safely retrieves all notifications for a given user
// Uses a read lock since it's not modifying data
func (ns *NotificationStore) Get(userID string) []models.Notification {
	ns.mu.RLock()          // Acquire shared read lock
	defer ns.mu.RUnlock()  // Release lock when function returns
	return ns.data[userID] // Return user's notifications
}

func main() {
	// Initialize notification store with empty map
	store := &NotificationStore{
		data: make(UserNotifications),
	}

	// Create cancellable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	// Start Kafka consumer group in separate goroutine
	go setupConsumerGroup(ctx, store)
	// Ensure context is cancelled when main exits
	defer cancel()

	// Configure Gin to run in production mode
	gin.SetMode(gin.ReleaseMode)
	// Create default Gin router with middleware
	router := gin.Default()
	// Set up GET endpoint for retrieving user notifications
	router.GET("/notifications/:userID", func(ctx *gin.Context) {
		handleNotifications(ctx, store)
	})

	// Log server startup with consumer group info
	fmt.Printf("Kafka CONSUMER (Group: %s) 👥📥 "+
		"started at http://localhost%s\n", ConsumerGroup, ConsumerPort)

	// Start HTTP server, log any errors that occur
	if err := router.Run(ConsumerPort); err != nil {
		log.Printf("failed to run the server: %v", err)
	}
}
