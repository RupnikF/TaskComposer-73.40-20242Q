package broker

import (
	"net"
	"os"
	"strconv"

	"github.com/segmentio/kafka-go"
)

func GetExecutionReader() *kafka.Reader {
	bootstrapServers := os.Getenv("KAFKA_HOST") + ":" + os.Getenv("KAFKA_PORT")
	topic := GetExecutionKafkaTopic()
	return GetReader([]string{bootstrapServers}, topic, "submissions")
}

func GetStepReader() *kafka.Reader {
	bootstrapServers := os.Getenv("KAFKA_HOST") + ":" + os.Getenv("KAFKA_PORT")
	topic := GetStepKafkaTopic()
	return GetReader([]string{bootstrapServers}, topic, "steps")
}

func GetReader(bootstrapServers []string, topic string, groupid string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  bootstrapServers,
		GroupID:  groupid,
		Topic:    topic,
		MaxBytes: 10e6,
	})
}

func GetWriter(bootstrapServers []string, topic string) *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:  bootstrapServers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
}

func GetGenericWriter(bootstrapServers []string) *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:  bootstrapServers,
		Balancer: &kafka.LeastBytes{},
	})
}

func GetExecutionKafkaTopic() string {
	return os.Getenv("SUBMISSIONS_TOPIC")
}

func GetStepKafkaTopic() string {
	return os.Getenv("STEPS_TOPIC")
}

func Initialize() {

	bootstrapServers := os.Getenv("KAFKA_HOST") + ":" + os.Getenv("KAFKA_PORT")

	conn, err := kafka.Dial("tcp", bootstrapServers)
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		panic(err.Error())
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		panic(err.Error())
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             GetExecutionKafkaTopic(),
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             GetStepKafkaTopic(),
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             os.Getenv("NATIVE_TOPIC"),
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		panic(err.Error())
	}
}
