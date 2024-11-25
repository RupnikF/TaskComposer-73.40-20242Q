package broker

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log"
)

// Tendria que llamarse con una go routine
func ProduceMessage(config *kafka.ConfigMap, topic string, message string) error {
	p, err := kafka.NewProducer(config)
	if err != nil {
		log.Fatalf("Failed to create producer: %s\n", err)
	}

	defer p.Close()

	deliveryChan := make(chan kafka.Event)

	err = p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, deliveryChan)
	if err != nil {
		return err
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	} else {
		fmt.Printf("Message %v delivered to topic %v [%v] at offset %v\n", m.Key, *m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	}

	return nil
}
