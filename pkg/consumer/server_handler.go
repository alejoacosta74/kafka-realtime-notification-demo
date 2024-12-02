package consumer

import (
	"errors"
	"kafka-notify/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleNotifications processes HTTP requests for retrieving user notifications
// It takes a gin context and notification store as parameters
func handleNotifications(ctx *gin.Context, store *NotificationStore) {
	// Extract the userID from the request parameters and handle any errors
	userID, err := getUserIDFromRequest(ctx)
	if err != nil {
		// If userID is invalid, return a 404 status with error message
		ctx.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	// Retrieve notifications for the user from the store
	notes := store.Get(userID)
	if len(notes) == 0 {
		// If no notifications exist, return 200 OK with empty array
		ctx.JSON(http.StatusOK,
			gin.H{
				"message":       "No notifications found for user",
				"notifications": []models.Notification{},
			})
		return
	}

	// Return 200 OK with the array of notifications
	ctx.JSON(http.StatusOK, gin.H{"notifications": notes})
}

// ErrNoMessagesFound is returned when no messages are found for a user
var ErrNoMessagesFound = errors.New("no messages found")

// getUserIDFromRequest extracts the userID parameter from the gin context
// Returns the userID if present, otherwise returns an error
func getUserIDFromRequest(ctx *gin.Context) (string, error) {
	// Get userID from URL parameters
	userID := ctx.Param("userID")
	if userID == "" {
		// Return error if userID parameter is empty
		return "", ErrNoMessagesFound
	}
	// Return the valid userID
	return userID, nil
}
