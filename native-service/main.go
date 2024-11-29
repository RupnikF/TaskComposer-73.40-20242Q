package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	kafka "github.com/segmentio/kafka-go"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

type EchoResponse struct {
	ExecutionId int                    `json:"executionId"`
	Outputs     map[string]interface{} `json:"outputs"`
}

type EchoRequest struct {
	ExecutionId int                    `json:"executionId"`
	TaskName    string                 `json:"taskName"`
	Inputs      map[string]interface{} `json:"inputs"`
}

func main() {
	//TIP <p>Press <shortcut actionId="ShowIntentionActions"/> when your caret is at the underlined text
	// to see how GoLand suggests fixing the warning.</p><p>Alternatively, if available, click the lightbulb to view possible fixes.</p>

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
				log.Printf("Error reading message: %v", err)
			} else {
				var request EchoRequest
				err := json.Unmarshal(msg.Value, &request)
				if err != nil {
					log.Printf("Error unmarshaling message: %v", err)
					continue
				}

				if request.TaskName == "echo" {
					log.Printf("Received message: %s", string(msg.Value))
					requestMessage, ok := request.Inputs["msg"]
					if !ok {
						continue
					}
					echoResponse := EchoResponse{
						ExecutionId: request.ExecutionId,
						Outputs: map[string]interface{}{
							"msg": requestMessage,
						},
					}

					finalMsg, err := json.Marshal(echoResponse)
					if err != nil {
						log.Printf("Error marshaling final response: %v", err)
						continue
					}
					log.Printf("Sending response: %s", string(finalMsg))
					err = writer.WriteMessages(context.Background(), kafka.Message{Value: finalMsg})
					if err != nil {
						log.Printf("Error writing final response: %v", err)
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
						log.Printf("Error marshaling final response: %v", err)
						continue
					}
					err = writer.WriteMessages(context.Background(), kafka.Message{Value: finalMsg})
					if err != nil {
						log.Printf("Error writing final response: %v", err)
					}
				}
			}
		}
	}()

	log.Print("Server started on port 8080")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	err := http.ListenAndServe(":8080", nil)

	log.Printf("Error starting server: %v", err.Error())
}
