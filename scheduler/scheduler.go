package main

import (
	"log"
	"os"
	"scheduler/broker"
	"scheduler/repository"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// Load the .env file
	ENV_MODE := os.Getenv("ENV_MODE")
	if ENV_MODE == "development" || ENV_MODE == "" {
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

	kafkaHost := []string{os.Getenv("KAFKA_HOST") + ":9092"}
	executionStepsWriter := broker.GetWriter(kafkaHost, broker.GetStepKafkaTopic())
	serviceWriter := broker.GetGenericWriter(kafkaHost)

	handler := broker.NewHandler(executionRepository, serviceRepository, executionStepsWriter, serviceWriter)

	executionReader := broker.GetExecutionReader()

	go broker.ConsumeMessageWithHandler(
		executionReader,
		-1,
		handler.HandleExecutionSubmission,
	)

	r.Run() // listen and serve on 0.0.0.0:8080
}
