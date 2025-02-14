package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/propagation"
	"log"
	"os"
	"scheduler/broker"
	"scheduler/jobs"
	"scheduler/repository"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/contrib/bridges/otelslog"

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
	serviceName      = os.Getenv("SERVICE_NAME")
	grpcCollectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT_GRPC")
)

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
		panic("failed to initialize exporter")
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

	init := otelslog.NewLogger("init")

	// Initialize the repository and broker
	executionRepository := repository.NewExecutionRepository(repository.Initialize())
	serviceRepository := repository.NewServiceRepository()
	serviceTopics := make([]string, 0)
	for _, service := range serviceRepository.GetServices() {
		if service.Server == "" {
			continue
		}
		serviceTopics = append(serviceTopics, service.OutputTopic, service.InputTopic)
	}
	jobsRepository := jobs.Initialize()
	broker.Initialize(serviceTopics)

	r := gin.Default()
	r.Use(otelgin.Middleware(serviceName))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/executions/:uuid", func(c *gin.Context) {
		stringUUID := c.Param("uuid")
		addCron := c.DefaultQuery("addCron", "false")
		if addCron == "true" {
			executions := executionRepository.GetExecutionsByJobID(c.Request.Context(), stringUUID)
			if executions == nil || len(executions) == 0 {
				c.JSON(204, gin.H{
					"message": "no executions found",
				})
			}
			output := make([]repository.ExecutionStateResponseDTO, len(executions))
			for i, execution := range executions {
				output[i] = execution.State.ToResponseStateDTO()
			}
			c.JSON(200, output)
		} else {
			state := executionRepository.GetExecutionByUUID(c.Request.Context(), stringUUID).State
			if state == nil {
				c.JSON(404, gin.H{
					"error": "execution not found",
				})
				return
			}
			c.JSON(200, state.ToResponseStateDTO())
		}

	})
	r.POST("/cancel-execution/:uuid", func(c *gin.Context) {
		stringUUID := c.Param("uuid")
		execution := executionRepository.GetExecutionByUUID(c.Request.Context(), stringUUID)
		if execution == nil {
			c.JSON(404, gin.H{
				"error": "execution not found",
			})
			return
		}
		err := jobsRepository.CancelJob(stringUUID)
		JobMessage := "Job not found"
		if err == nil {
			JobMessage = "Job cancelled"
		}
		log.Printf("%s, %s", stringUUID, JobMessage)
		if execution.State.Status != repository.PENDING && execution.State.Status != repository.EXECUTING {
			c.JSON(200, gin.H{
				"error": "execution already finished",
			})
			return
		}
		executionRepository.CancelExecution(c, execution)
		c.JSON(200, gin.H{
			"message": "execution cancelled",
		})
	})
	r.POST("/cancel-execution", func(c *gin.Context) {
		var cancelRequest repository.CancelTagsDTO
		err := c.BindJSON(&cancelRequest)
		if err != nil {
			c.JSON(400, gin.H{
				"error": "Invalid request, error parsing tags",
			})
			return
		}
		executions := executionRepository.GetExecutionsByTags(c, cancelRequest.Tags)
		if executions == nil || len(executions) == 0 {
			c.JSON(404, gin.H{
				"error": "No executions to cancel found",
			})
			return
		}
		for _, execution := range executions {
			executionRepository.CancelExecution(c, execution)
		}
		c.JSON(200, gin.H{
			"message": fmt.Sprintf("cancelled %d executions", len(executions)),
		})
	})

	kafkaHost := []string{os.Getenv("KAFKA_HOST") + ":" + os.Getenv("KAFKA_PORT")}
	executionStepsWriter := broker.GetWriter(kafkaHost, broker.GetStepKafkaTopic())

	serviceWriters := make(map[string]*kafka.Writer)
	for _, service := range serviceRepository.GetServices() {
		fmt.Printf("Service: %v\n", service.Name)
		if service.Server == "" {
			log.Printf("Service %s has no server", service.Name)
			continue
		}
		serviceWriters[service.Name] = broker.GetWriter([]string{service.Server}, service.InputTopic)
	}
	tp := otel.GetTracerProvider()
	handler := broker.NewHandler(executionRepository, serviceRepository, executionStepsWriter, serviceWriters, tp, jobsRepository, broker.ProduceMessage)

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
	for _, service := range serviceRepository.GetServices() {
		if service.Server == "" {
			continue
		}
		serviceReader := broker.GetReader([]string{service.Server}, service.OutputTopic, service.Name)

		fmt.Printf("Listening for topic %s\n", service.OutputTopic)
		go broker.ConsumeMessageWithHandler(
			serviceReader,
			-1,
			handler.HandleServiceResponse,
		)
	}
	init.Info("Starting scheduler")
	err := r.Run()
	if err != nil {
		return
	}
}
