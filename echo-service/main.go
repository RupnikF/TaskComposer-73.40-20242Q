package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	kafka "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	setupLog "go.opentelemetry.io/otel/sdk/log"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

type EchoResponse struct {
	ExecutionId int                    `json:"executionId"`
	Outputs     map[string]interface{} `json:"outputs"`
}

type TaskRequest struct {
	ExecutionId int                    `json:"executionId"`
	TaskName    string                 `json:"taskName"`
	Inputs      map[string]interface{} `json:"inputs"`
}

var (
	serviceName      = os.Getenv("SERVICE_NAME")
	grpcCollectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT_GRPC")
)

func initTracer() func(context.Context) error {
	secureOption := otlptracegrpc.WithInsecure()

	log.Printf("COLLECTOR_ENDPOINT_GRPC" + grpcCollectorURL)
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(grpcCollectorURL),
		),
	)

	if err != nil {
		log.Println(err)
	}
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Printf("Could not set resources: %s", err)
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)

	return exporter.Shutdown
}

func initLogger() (context.Context, *setupLog.LoggerProvider) {
	ctx := context.Background()

	// Create the OTLP log exporter that sends logs to configured destination
	logExporter, err := otlploghttp.New(ctx)
	if err != nil {
		fmt.Printf("Error initializing logger", err)
	}

	// Create the logger provider
	lp := setupLog.NewLoggerProvider(
		setupLog.WithProcessor(
			setupLog.NewBatchProcessor(logExporter),
		),
	)

	global.SetLoggerProvider(lp)
	return ctx, lp
}

func main() {
	ctx, lp := initLogger()
	defer func() {
		lp.Shutdown(ctx)
	}()
	logger := otelslog.NewLogger("scheduler-init")

	err := initTracer()
	if err != nil {
		logger.Info("Error initiating tracer", err)
	}

	HostPort := os.Getenv("HOST_PORT")

	EnvMode := os.Getenv("ENV_MODE")
	if EnvMode == "development" || EnvMode == "" {
		err := godotenv.Load(".env.local")
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	brokers := []string{os.Getenv("KAFKA_HOST") + ":" + os.Getenv("KAFKA_PORT")}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "native",
		Topic:   os.Getenv("INPUT_TOPIC"),
	})

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Balancer: &kafka.LeastBytes{},
		Topic:    os.Getenv("OUTPUT_TOPIC"),
	})

	log.Print("Listening on topic: " + os.Getenv("INPUT_TOPIC"))
	go func() {
		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				logger.Error("Error reading message:", err)
			} else {
				var request TaskRequest
				err := json.Unmarshal(msg.Value, &request)
				if err != nil {
					logger.Error("Error unmarshaling message:", err)
					continue
				}

				if request.TaskName == "echo" {
					logger.Debug("Received message: %s", string(msg.Value))
					requestMessage, ok := request.Inputs["msg"]
					var echoResponse EchoResponse
					if !ok {
						echoResponse = EchoResponse{
							ExecutionId: request.ExecutionId,
							Outputs: map[string]interface{}{
								"error": map[string]string{
									"msg": "No msg property in inputs",
								},
							},
						}
					} else {
						echoResponse = EchoResponse{
							ExecutionId: request.ExecutionId,
							Outputs: map[string]interface{}{
								"msg": requestMessage,
							},
						}
					}

					finalMsg, err := json.Marshal(echoResponse)
					if err != nil {
						logger.Error("Error marshaling final response: %v", err)
						continue
					}
					logger.Info("Sending response: %s", string(finalMsg))
					err = writer.WriteMessages(context.Background(), kafka.Message{Value: finalMsg})
					if err != nil {
						logger.Error("Error writing final response: %v", err)
					}
				} else {
					finalMsg, err := json.Marshal(EchoResponse{
						ExecutionId: request.ExecutionId,
						Outputs: map[string]interface{}{
							"error": map[string]interface{}{
								"msg": "Invalid task",
							},
						},
					})
					if err != nil {
						logger.Error("Error marshaling final response: %v", err)
						continue
					}
					err = writer.WriteMessages(context.Background(), kafka.Message{Value: finalMsg})
					if err != nil {
						logger.Error("Error writing final response: %v", err)
					}
				}
			}
		}
	}()

	log.Printf("Server started on port %s\n", HostPort)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "This is my website!\n")
	})

	errServer := http.ListenAndServe(":8080", nil)
	if errServer != nil {
		log.Printf("Error binding to port %s", HostPort)
	}
}
