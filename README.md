# Basic proof of concept for a real-time notification system with Go and Kafka

This is a basic proof of concept for a real-time notification system with Go and Kafka.

This project demonstrates a real-time notification system implementation using Go, Apache Kafka, and the Gin web framework. 

The architecture follows a producer-consumer pattern where notifications are published and consumed through Kafka topics. 

The producer service, running on port 8080, exposes a REST endpoint `/send` that accepts `POST` requests with user IDs and message content. 
It uses the Sarama library (`github.com/IBM/sarama`) to handle Kafka message production, with messages being serialized as JSON and containing sender, recipient, and message content. 

The consumer service, running on port 8081, implements a thread-safe notification store using Go's `sync.RWMutex` for concurrent access management. 

It maintains a consumer group (referenced in `cmd/consumer/consumer.go` lines 15-18) that processes messages from the Kafka topic and exposes a `/notifications/:userID` endpoint for retrieving user-specific notifications. 

The system supports basic user-to-user interactions like follows, mentions, and likes, with each notification being properly partitioned in Kafka using the recipient's user ID as the message key (as shown in `cmd/producer/kafka.go` lines 48-52). 

This ensures efficient message routing and maintains message ordering per user while allowing for horizontal scalability through Kafka's partitioning mechanism.

## Usage

### Run the docker compose file to start the kafka broker

- Modify the `docker-compose.yml` file to match your kafka server IP address
```bash
docker compose up
```

## Run the producer API

- This starts a simple API that allows you to send notifications to the kafka topic at the `/send` endpoint

```bash
go run cmd/producer/*.go
```

### Run the consumer API

- This starts a simple API that allows you to retrieve notifications from the kafka topic at the `/notifications/:userID` endpoint

```bash
go run cmd/consumer/*.go
```

### Send notifications (publish messages to kafka topic)

**User 1 (Micho) receives a notification from User 2 (Tito):**

```bash
 curl -X POST http://localhost:8080/send \
-d "fromID=2&toID=1&message=Tito started following you."
```

**User 2 (Tito) receives a notification from User 1 (Micho):**

```go
curl -X POST http://localhost:8080/send \
-d "fromID=1&toID=2&message=Micho mentioned you in a comment: 'Great seeing you yesterday, @Tito!'"
```

**User 1 (Micho) receives a notification from User 4 (Negro):**

```go
curl -X POST http://localhost:8080/send \
-d "fromID=4&toID=1&message=Negro liked your post: 'My weekend getaway!'"
```

### Retrieve notifications (subscribe to kafka topic)

**Retrieving notifications for User 1 (Micho):**

```bash
curl http://localhost:8081/notifications/1
```
