package producer

import (
	"context"
	"kafka-notify/pkg/models"
	"kafka-notify/pkg/server"
	"time"

	"github.com/alejoacosta74/go-logger"
	"github.com/spf13/viper"
)

const (
	ProducerPort = ":8080"
	KafkaTopic   = "notifications"
)

var KafkaServerAddress string

func Run() {
	KafkaServerAddress = viper.GetString("kafka-broker-address")
	// Create cancellable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	users := []models.User{
		{ID: 1, Name: "Micho"},
		{ID: 2, Name: "Tito"},
		{ID: 3, Name: "Negro"},
		{ID: 4, Name: "Cabezon"},
	}

	producer, err := setupProducer(ctx)
	if err != nil {
		logger.Fatal("Failed to initialize producer", "error", err)
	}
	defer producer.Close()

	// gin.SetMode(gin.ReleaseMode)
	// router := gin.Default()
	// router.POST("/send", sendMessageHandler(producer, users))

	// create and start http server to expose the consumer endpoint
	httpServer := server.NewServer(ProducerPort)
	httpServer.Post("/send", sendMessageHandler(producer, users))
	httpServer.ListenAndServe()

	logger.Infof("Kafka PRODUCER 📨 started at http://localhost:%v", ProducerPort)

	interruptCh := server.NewInterruptSignalChannel()
	<-interruptCh
	cancel()

	ctxWithTimeout, _ := context.WithTimeout(ctx, 5*time.Second)
	httpServer.Shutdown(ctxWithTimeout)

	<-ctx.Done()
	logger.Info("Kafka producer finished")

}
