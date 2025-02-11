package broker

import (
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"net"
	"os"
	"strconv"

	"github.com/segmentio/kafka-go"
)

var configLogger = otelslog.NewLogger("kafka-consumer")

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

func Initialize(serviceTopics []string) {

	bootstrapServers := os.Getenv("KAFKA_HOST") + ":" + os.Getenv("KAFKA_PORT")

	conn, err := kafka.Dial("tcp", bootstrapServers)
	if err != nil {
		configLogger.Error("Error connecting to Kafka Servers", "error", err)
		panic(err.Error())
	}
	defer func() {
		err := conn.Close()
		if err != nil {
		}
	}()

	controller, err := conn.Controller()
	if err != nil {
		configLogger.Error("Error creating kafka controller", "error", err)
		panic(err.Error())
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		configLogger.Error("Error dialing kafka host port", "error", err)
		panic(err.Error())
	}
	defer func(controllerConn *kafka.Conn) {
		err := controllerConn.Close()
		if err != nil {
			// Explicit skip
		}
	}(controllerConn)

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             GetExecutionKafkaTopic(),
			NumPartitions:     3,
			ReplicationFactor: 2,
		},
		{
			Topic:             GetStepKafkaTopic(),
			NumPartitions:     3,
			ReplicationFactor: 2,
		},
	}
	for _, topic := range serviceTopics {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     3,
			ReplicationFactor: 2,
		})
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		configLogger.Error("Error creating topics", "error", err)
	}
}
