package broker

import (
	"github.com/goccy/go-json"
	"log"
	"scheduler/repository"
)

type Handler struct {
	executionRepository repository.ExecutionRepository
	serviceRepository   repository.ServiceRepository
}

func NewHandler(executionRepository repository.ExecutionRepository, serviceRepository repository.ServiceRepository) *Handler {
	return &Handler{executionRepository, serviceRepository}
}

func (h *Handler) handleExecutionSubmission(message []byte) {
	submission := repository.ExecutionSubmissionDTO{}
	err := json.Unmarshal(message, &submission)
	if err != nil {
		log.Printf("Failed to unmarshal message: %s\n", err)
	}
	execution := submission.ToExecution(repository.PENDING)
	h.executionRepository.CreateExecution(&execution)
	stepToExecute := execution.Steps[0].ToExecutionStepDTO()

	//Enqueue the step
	config := GetExecutionStepKafkaConfig()
	bytes, err := json.Marshal(stepToExecute)
	if err != nil {
		log.Printf("Failed to marshal message: %s\n", err)
	}
	err = ProduceMessage(&config, GetExecutionStepKafkaTopic(), bytes)
	if err != nil {
		log.Printf("Failed to produce message: %s\n", err)
	}

}

func (h *Handler) handleExecutionStep(message []byte) {
	step := repository.ExecutionStepDTO{}
	err := json.Unmarshal(message, &step)
	if err != nil {
		log.Printf("Failed to unmarshal message: %s\n", err)
	}
	// step.Service -> topic
}
