package main

import (
	"context"
	"log"
	"os"
	"scheduler/broker"
	"scheduler/repository"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
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

var (
	serviceName  = os.Getenv("SERVICE_NAME")
	collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
)

func initTracer() func(context.Context) error {
	secureOption := otlptracegrpc.WithInsecure()

	log.Printf("COLLECTOR_ENDPOINT" + collectorURL)
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(collectorURL),
		),
	)

	if err != nil {
		log.Fatal(err)
	}
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Printf("Could not set resources: ", err)
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
		panic("failed to initialize exporter")
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
	defer lp.Shutdown(ctx)

	cleanup := initTracer()
	defer func() {
		err := cleanup(context.Background())
		if err != nil {
			log.Printf("Error cleaning up tracer: %v", err)
		}
	}()
	// Load the .env file
	EnvMode := os.Getenv("ENV_MODE")
	if EnvMode == "development" || EnvMode == "" {
		err := godotenv.Load(".env.local")
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}

	// Initialize the repository and broker
	executionRepository := repository.NewExecutionRepository(repository.Initialize())
	serviceRepository := repository.NewServiceRepository()
	broker.Initialize()

	r := gin.Default()
	r.Use(otelgin.Middleware(serviceName))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/executions/:uuid", func(c *gin.Context) {
		stringUUID := c.Param("uuid")
		state := executionRepository.GetExecutionByUUID(stringUUID).State
		if state == nil {
			c.JSON(404, gin.H{
				"error": "execution not found",
			})
			return
		}
		c.JSON(200, state.ToResponseStateDTO())
	})

	kafkaHost := []string{os.Getenv("KAFKA_HOST") + ":" + os.Getenv("KAFKA_PORT")}
	executionStepsWriter := broker.GetWriter(kafkaHost, broker.GetStepKafkaTopic())
	serviceWriter := broker.GetGenericWriter(kafkaHost)
	tp := otel.GetTracerProvider()
	handler := broker.NewHandler(executionRepository, serviceRepository, executionStepsWriter, serviceWriter, tp)

	executionReader := broker.GetExecutionReader()

	go broker.ConsumeMessageWithHandler(
		executionReader,
		-1,
		handler.HandleExecutionSubmission,
	)

	stepReader := broker.GetStepReader()

	go broker.ConsumeMessageWithHandler(
		stepReader,
		-1,
		handler.HandleExecutionStep,
	)
	serviceHost := os.Getenv("SERVICE_HOST") + ":" + os.Getenv("NATIVE_PORT")
	nativeTopic := os.Getenv("NATIVE_OUTPUT_TOPIC")
	nativeServiceReader := broker.GetReader([]string{serviceHost}, nativeTopic, "native-service")
	go broker.ConsumeMessageWithHandler(
		nativeServiceReader,
		-1,
		handler.HandleServiceResponse,
	)
	err := r.Run()
	if err != nil {
		return
	} // listen and serve on 0.0.0.0:8080
}
