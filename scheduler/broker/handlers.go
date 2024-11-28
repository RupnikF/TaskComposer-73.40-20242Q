package broker

import (
	"log"
	"regexp"
	"scheduler/repository"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/goccy/go-json"
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

type ServiceMessage struct {
	ExecutionId uint                   `json:"executionId"`
	TaskName    string                 `json:"taskName"`
	Inputs      map[string]interface{} `json:"inputs"`
}

// Matches with all strings that start with args.

func (h *Handler) handleExecutionStep(message []byte) {

	step := repository.ExecutionStepDTO{}
	err := json.Unmarshal(message, &step)
	if err != nil {
		log.Printf("Failed to unmarshal message: %s\n", err)
	}

	config := h.serviceRepository.GetService(step.Service)
	if config == nil {
		log.Printf("Service not found: %s\n", step.Service)
		return
	}

	state := h.executionRepository.GetStateByExecutionID(step.ExecutionID)

	// TODO: move argsMatcher so it only compiles once.
	argsMatcher := regexp.MustCompile(`args\.(.+)`)
	// Build corresponding inputs
	argsMap := make(map[string]interface{})
	outputMap := make(map[string]interface{})

	for _, arg := range state.Arguments {
		// TODO: type shouldnt just be string
		argsMap[arg.Key] = arg.Value
	}

	for _, output := range state.Outputs {
		outputMap[output.Key] = output.Value
	}

	inputs := make(map[string]interface{})

	// TODO: value is not being useful, should do typechecking
	for k := range step.Input {
		isArgs := argsMatcher.MatchString(k)
		existMapping := true
		if isArgs {
			argKey := strings.Replace(k, "args.", "", 1)
			result, ok := argsMap[k]
			if !ok {
				existMapping = false
			} else {
				inputs[argKey] = result
			}
		} else {
			result, ok := outputMap[k]
			if !ok {
				existMapping = false
			} else {
				inputs[k] = result
			}
		}
		if !existMapping {
			state.Status = repository.FAILED
			h.executionRepository.UpdateState(state)
			log.Printf("Required output key not found: %s\n", k)
			return
		}
	}

	serviceMessage := ServiceMessage{
		ExecutionId: step.ExecutionID,
		TaskName:    step.Task,
		Inputs:      inputs,
	}

	message, err = json.Marshal(serviceMessage)
	if err != nil {
		log.Printf("Failed to marshal message: %s\n", err)
		return
	}

	err = ProduceMessage(&kafka.ConfigMap{
		"bootstrap.servers": config.Server,
	}, config.Topic, message)
	if err != nil {
		log.Printf("Failed to produce message: %s\n", err)
	}
}
