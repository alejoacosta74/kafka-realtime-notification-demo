package main

import (
	"kafka-notify/pkg/models"

	"github.com/alejoacosta74/go-logger"
	"github.com/gin-gonic/gin"
)

const (
	ProducerPort       = ":8080"
	KafkaServerAddress = "192.168.5.142:9092"
	KafkaTopic         = "notifications"
)

func main() {
	users := []models.User{
		{ID: 1, Name: "Micho"},
		{ID: 2, Name: "Tito"},
		{ID: 3, Name: "Negro"},
		{ID: 4, Name: "Cabezon"},
	}

	producer, err := setupProducer()
	if err != nil {
		logger.Fatal("Failed to initialize producer", "error", err)
	}
	defer producer.Close()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.POST("/send", sendMessageHandler(producer, users))

	logger.Infof("Kafka PRODUCER 📨 started at http://localhost:%v", ProducerPort)

	if err := router.Run(ProducerPort); err != nil {
		logger.Error("Failed to run the server", "error", err)
	}
}
