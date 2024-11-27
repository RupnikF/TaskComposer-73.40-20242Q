package broker

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log"
	"os"
)

func GetExecutionKafkaConfig() kafka.ConfigMap {
	bootstrapServers := os.Getenv("KAFKA_HOST") + ":9092"
	return kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          "submissions",
		"auto.offset.reset": "earliest",
	}
}
func GetExecutionKafkaTopic() string {
	return os.Getenv("SUBMISSIONS_TOPIC")
}

func GetExecutionStepKafkaConfig() kafka.ConfigMap {
	bootstrapServers := os.Getenv("KAFKA_HOST") + ":9092"
	return kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          "steps",
		"auto.offset.reset": "earliest",
	}
}
func GetExecutionStepKafkaTopic() string {
	return os.Getenv("STEPS_TOPIC")
}

func Initialize() {

	bootstrapServers := os.Getenv("KAFKA_HOST") + ":9092"

	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": bootstrapServers})
	if err != nil {
		log.Fatalf("Failed to create AdminClient: %s\n", err)
	}
	defer admin.Close()

	results, err := admin.CreateTopics(
		nil, // Default options
		[]kafka.TopicSpecification{
			{
				Topic: GetExecutionKafkaTopic(),
			},
			{
				Topic: GetExecutionStepKafkaTopic(),
			},
		},
	)
	if err != nil {
		log.Fatalf("Failed to create topic: %s\n", err)
	}

	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError {
			fmt.Printf("Failed to create topic %s: %s\n", result.Topic, result.Error)
		} else {
			fmt.Printf("Successfully created topic: %s\n", result.Topic)
		}
	}
}
