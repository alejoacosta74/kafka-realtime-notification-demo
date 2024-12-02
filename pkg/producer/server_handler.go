package producer

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"kafka-notify/pkg/models"

	"github.com/IBM/sarama"
	"github.com/alejoacosta74/go-logger"
	"github.com/gin-gonic/gin"
)

// sendMessageHandler creates a Gin HTTP handler for sending messages between users
// It takes a Kafka producer and a list of users as parameters
func sendMessageHandler(producer sarama.SyncProducer,
	users []models.User) gin.HandlerFunc {
	// Return a closure that handles the actual HTTP request
	return func(ctx *gin.Context) {
		// Extract and parse the sender's ID from the form data
		fromID, err := getIDFromRequest("fromID", ctx)
		if err != nil {
			logger.Error("Failed to get sender ID from request", "error", err)
			// Return 400 Bad Request if sender ID is invalid
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// Extract and parse the recipient's ID from the form data
		toID, err := getIDFromRequest("toID", ctx)
		if err != nil {
			logger.Error("Failed to get recipient ID from request", "error", err)
			// Return 400 Bad Request if recipient ID is invalid
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// Attempt to send the message to Kafka
		err = sendKafkaProducerMessage(producer, users, ctx, fromID, toID)
		if errors.Is(err, ErrUserNotFoundInProducer) {
			// Return 404 Not Found if either user doesn't exist
			logger.Error("User not found", "error", err)
			ctx.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			return
		}
		if err != nil {
			// Return 500 Internal Server Error for any other errors
			logger.Error("Failed to send message to Kafka", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		// Return 200 OK if message was sent successfully
		logger.Info("Notification sent successfully!")
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Notification sent successfully!",
		})
	}
}

// getIDFromRequest extracts and parses a user ID from the HTTP form data
// It takes the form field name and the Gin context as parameters
func getIDFromRequest(formValue string, ctx *gin.Context) (int, error) {
	// Extract the string value from the form and convert it to an integer
	id, err := strconv.Atoi(ctx.PostForm(formValue))
	if err != nil {
		// If conversion fails, return an error with details about which field failed
		logger.Error("Failed to parse ID from form value", "error", err)
		return 0, fmt.Errorf(
			"failed to parse ID from form value %s: %w", formValue, err)
	}
	// Return the successfully parsed integer ID
	return id, nil
}
