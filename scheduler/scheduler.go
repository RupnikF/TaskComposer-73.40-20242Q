package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"scheduler/broker"
	"scheduler/repository"
)

func main() {

	// Load the .env file
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize the repository and broker
	executionRepository := repository.NewExecutionRepository(repository.Initialize())

	broker.Initialize()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
