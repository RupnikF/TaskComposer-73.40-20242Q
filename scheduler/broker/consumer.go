package broker

import (
	"context"
	"fmt"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

// Tendria que llamarse con una go routine
// func ConsumeMessages(config *kafka.ReaderConfig, topic string, timeout time.Duration) {
// 	c := kafka.NewReader(*config)
// 	if err != nil {
// 		log.Fatalf("Failed to create consumer: %s\n", err)
// 	}

// 	// A signal handler or similar could be used to set this to false to break the loop.
// 	run := true

// 	for run {
// 		msg, err := c.ReadMessage(context.Background())

// 		if err == nil {
// 			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
// 			//Aca se podria poner un handler
// 			//Tambien se tendria que mandar algo que pueda parsear el mensaje si es que esta serializado
// 		} else if !err {
// 			// The client will automatically try to recover from all errors.
// 			// Timeout is not considered an error because it is raised by
// 			// ReadMessage in absence of messages.
// 			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
// 		}
// 	}
// 	err = c.Close()
// 	if err != nil {
// 		log.Fatalf("Failed to close consumer: %s\n", err)
// 	}
// }

func ConsumeMessageWithHandler(c *kafka.Reader, timeout time.Duration, handler func([]byte)) {

	// A signal handler or similar could be used to set this to false to break the loop.
	run := true

	for run {
		msg, err := c.ReadMessage(context.Background())
		if err == nil {
			go handler(msg.Value)
		} else {
			// The client will automatically try to recover from all errors.
			// Timeout is not considered an error because it is raised by
			// ReadMessage in absence of messages.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
	err := c.Close()
	if err != nil {
		log.Fatalf("Failed to close consumer: %s\n", err)
	}
}
