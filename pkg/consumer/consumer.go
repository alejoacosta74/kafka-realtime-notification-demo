package consumer

import (
	"context"
	"sync"
	"time"

	"kafka-notify/pkg/models"
	"kafka-notify/pkg/server"

	"github.com/alejoacosta74/go-logger"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	ConsumerGroup = "notifications-group"
	ConsumerTopic = "notifications"
	ConsumerPort  = ":8081"
)

var KafkaServerAddress string

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

func Run() {
	KafkaServerAddress = viper.GetString("kafka-broker-address")

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

	// create and start http server to expose the consumer endpoint
	httpServer := server.NewServer(ConsumerPort)
	httpServer.Get("/notifications/:userID", func(ctx *gin.Context) {
		handleNotifications(ctx, store)
	})
	httpServer.ListenAndServe()

	logger.Infof("Kafka CONSUMER (Group: %s) 👥📥 started at http://localhost:%v", ConsumerGroup, ConsumerPort)

	interruptCh := server.NewInterruptSignalChannel()
	<-interruptCh

	// cancel the context to stop the consumer
	cancel()

	ctxWithTimeout, _ := context.WithTimeout(ctx, 5*time.Second)
	httpServer.Shutdown(ctxWithTimeout)

	<-ctx.Done()
	logger.Info("Kafka consumer finished")
}
