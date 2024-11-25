package broker

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log"
	"time"
)

// Tendria que llamarse con una go routine
func ConsumeMessages(config *kafka.ConfigMap, topic string, timeout time.Duration) {
	c, err := kafka.NewConsumer(config)
	if err != nil {
		log.Fatalf("Failed to create consumer: %s\n", err)
	}

	err = c.SubscribeTopics([]string{topic}, nil)

	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %s\n", err)
	}

	// A signal handler or similar could be used to set this to false to break the loop.
	run := true

	for run {
		msg, err := c.ReadMessage(timeout)
		if err == nil {
			//Aca se podria poner un handler
			//Tambien se tendria que mandar algo que pueda parsear el mensaje si es que esta serializado
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
		} else if !err.(kafka.Error).IsTimeout() {
			// The client will automatically try to recover from all errors.
			// Timeout is not considered an error because it is raised by
			// ReadMessage in absence of messages.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
	err = c.Close()
	if err != nil {
		log.Fatalf("Failed to close consumer: %s\n", err)
	}
}
