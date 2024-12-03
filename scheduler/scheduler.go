package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
	"scheduler/broker"
	"scheduler/repository"
)

func main() {

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

	handler := broker.NewHandler(executionRepository, serviceRepository, executionStepsWriter, serviceWriter)

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
