package broker

import (
	"context"
	"log"

	kafka "github.com/segmentio/kafka-go"
)

// Tendria que llamarse con una go routine
func ProduceMessage(writer *kafka.Writer, headers []kafka.Header, message []byte) error {
	msg := kafka.Message{
		Value:   message,
		Headers: headers,
	}

	err := writer.WriteMessages(context.Background(), msg)
	if err != nil {
		log.Println("Failed to write messages:", err)
		return err
	}

	return nil
}

func ProduceTopicMessage(writer *kafka.Writer, message []byte, headers []kafka.Header, topic string) error {
	msg := kafka.Message{
		Value:   message,
		Topic:   topic,
		Headers: headers,
	}

	err := writer.WriteMessages(context.Background(), msg)
	if err != nil {
		log.Println("Failed to write messages:", err)
		return err
	}

	return nil
}
