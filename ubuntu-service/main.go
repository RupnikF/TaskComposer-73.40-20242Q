package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"ubuntu-service/service"

	"github.com/joho/godotenv"
	kafka "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	setupLog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

type Response struct {
	ExecutionId int                    `json:"executionId"`
	Outputs     map[string]interface{} `json:"outputs"`
}

type TaskRequest struct {
	ExecutionId int                    `json:"executionId"`
	TaskName    string                 `json:"taskName"`
	Inputs      map[string]interface{} `json:"inputs"`
}

func (t *TaskRequest) ToError(msg string) Response {
	return Response{
		ExecutionId: t.ExecutionId,
		Outputs: map[string]interface{}{
			"error": map[string]string{
				"msg": msg,
			},
		},
	}
}

func (t *TaskRequest) ToResponse(output map[string]interface{}) Response {
	return Response{
		ExecutionId: t.ExecutionId,
		Outputs:     output,
	}
}

var (
	serviceName      = os.Getenv("SERVICE_NAME")
	grpcCollectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT_GRPC")
)

func CreateOrGetSpan(spanName string, header []kafka.Header) (context.Context, trace.Span) {
	headerMap := make(map[string]string)
	for _, h := range header {
		headerMap[h.Key] = string(h.Value)
	}

	propagator := otel.GetTextMapPropagator()
	carrier := propagation.MapCarrier(headerMap)
	ctx := propagator.Extract(context.Background(), carrier)
	tracer := otel.Tracer("kafka-handlers")

	return tracer.Start(ctx, spanName, trace.WithAttributes())
}

func PassHeader(ctx context.Context) []kafka.Header {
	propagator := otel.GetTextMapPropagator()
	carrier := make(propagation.MapCarrier)
	propagator.Inject(ctx, carrier)
	var headers = make([]kafka.Header, 0)
	for k, v := range carrier {
		headers = append(headers, kafka.Header{
			Key: k, Value: []byte(v),
		})
	}
	return headers
}

func initTracer() func(context.Context) error {
	secureOption := otlptracegrpc.WithInsecure()

	log.Printf("COLLECTOR_ENDPOINT_GRPC" + grpcCollectorURL)
	log.Printf("SERVICE_NAME" + serviceName)
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
		fmt.Printf("Error initializing logger, %s", err)
	}

	resources, err := resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Printf("Could not set resources: %s", err)
	}

	// Create the logger provider
	lp := setupLog.NewLoggerProvider(
		setupLog.WithProcessor(
			setupLog.NewBatchProcessor(logExporter),
		),
		setupLog.WithResource(resources),
	)

	global.SetLoggerProvider(lp)
	return ctx, lp
}

func main() {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx, lp := initLogger()
	defer func() {
		lp.Shutdown(ctx)
	}()
	logger := otelslog.NewLogger("ubuntu-service")

	err := initTracer()
	if err != nil {
		logger.Error("Error initializing tracer", slog.Any("err", err))
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
				logger.Error("Error reading message:", slog.Any("err", err))
			} else {
				ctx, span := CreateOrGetSpan("ubuntu-service", msg.Headers)
				var request TaskRequest
				err := json.Unmarshal(msg.Value, &request)
				if err != nil {
					logger.Error("Error unmarshaling message:", slog.Any("err", err))
					span.RecordError(err)
					span.End()
					continue
				}
				span.SetAttributes(attribute.String("task.name", request.TaskName))
				logger.Debug("Received message: %s", slog.Any("msg", string(msg.Value)))
				var kafkaResponse Response
				switch request.TaskName {
				case "bash":
					res, err := service.RunShell(request.Inputs, span)
					if err != nil {
						span.RecordError(err)
						kafkaResponse = request.ToError(err.Error())
					} else {
						kafkaResponse = request.ToResponse(map[string]interface{}{
							"stdout": res.Stdout,
							"stderr": res.Stderr,
						})
					}
				case "eval":
					res, err := service.Eval(request.Inputs, span)
					if err != nil {
						span.RecordError(err)
						kafkaResponse = request.ToError(err.Error())
					} else {
						kafkaResponse = request.ToResponse(map[string]interface{}{
							"result": res,
						})
					}
				default:
					span.RecordError(fmt.Errorf("unknown task name: %s", request.TaskName))
					kafkaResponse = request.ToError("Invalid task")
				}

				finalMsg, err := json.Marshal(kafkaResponse)
				if err != nil {
					logger.Error("Error marshaling final response", slog.Any("err", err))
					span.RecordError(fmt.Errorf("error marshaling final response: %v", err))
					span.End()
					continue
				}
				logger.Info("Sending response:", slog.Any("response", string(finalMsg)))
				err = writer.WriteMessages(context.Background(), kafka.Message{Value: finalMsg, Headers: PassHeader(ctx)})
				if err != nil {
					logger.Error("Error writing final response", slog.Any("err", err))
				}
				span.End()
			}
		}
	}()

	log.Printf("Server started on port %s\n", HostPort)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, "Healthy\n")
		if err != nil {
			return
		}
	})

	errServer := http.ListenAndServe(":8080", nil)
	if errServer != nil {
		log.Printf("Error binding to port %s", HostPort)
	}
}
