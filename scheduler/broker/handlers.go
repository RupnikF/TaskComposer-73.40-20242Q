package broker

import (
	"log"
	"regexp"
	"scheduler/repository"
	"strings"

	"github.com/goccy/go-json"
	kafka "github.com/segmentio/kafka-go"
)

type Handler struct {
	executionRepository  *repository.ExecutionRepository
	serviceRepository    *repository.ServiceRepository
	executionStepsWriter *kafka.Writer
	serviceWriter        *kafka.Writer
}

func NewHandler(
	executionRepository *repository.ExecutionRepository,
	serviceRepository *repository.ServiceRepository,
	executionStepsWriter *kafka.Writer,
	serviceWriter *kafka.Writer,
) *Handler {
	return &Handler{executionRepository, serviceRepository, executionStepsWriter, serviceWriter}
}

func (h *Handler) HandleExecutionSubmission(message []byte) {
	log.Printf(string(message))
	submission := repository.ExecutionSubmissionDTO{}
	err := json.Unmarshal(message, &submission)
	if err != nil {
		log.Printf("Failed to unmarshal message: %s\n", err)
		return
	}
	log.Printf("Received submission: %v\n", submission)
	execution := submission.ToExecution(repository.PENDING)
	if execution == nil {
		log.Printf("Error parsing execution\n")
		return
	}
	h.executionRepository.CreateExecution(execution)
	stepToExecute := execution.Steps[0].ToExecutionStepDTO()

	//Enqueue the step
	bytes, err := json.Marshal(stepToExecute)
	if err != nil {
		log.Printf("Failed to marshal message: %s\n", err)
	}
	err = ProduceMessage(h.executionStepsWriter, bytes)
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

func (h *Handler) HandleExecutionStep(message []byte) {
	step := repository.ExecutionStepDTO{}
	err := json.Unmarshal(message, &step)
	if err != nil {
		log.Printf("Failed to unmarshal message: %s\n", err)
		return
	}

	log.Printf("Received step: %v\n", step)

	config := h.serviceRepository.GetService(step.Service)
	if config == nil {
		log.Printf("Service not found: %s\n", step.Service)
		return
	}

	state := h.executionRepository.GetStateByExecutionID(step.ExecutionID)

	// TODO: move argsMatcher so it only compiles once.
	argsMatcher := regexp.MustCompile(`\$args\.(.+)`)
	// Build corresponding inputs
	argsMap := make(map[string]interface{})
	outputMap := make(map[string]interface{})

	for _, arg := range state.Arguments {
		var value any
		err = json.Unmarshal([]byte(arg.Value), &value)
		if err != nil {
			log.Printf("Failed to unmarshal value: %s - %s\n", arg.Value, err.Error())
			argsMap[arg.Key] = arg.Value
		} else {
			argsMap[arg.Key] = value
		}
	}

	for _, output := range state.Outputs {
		outputMap[output.Key] = output.Value
	}

	inputs := make(map[string]interface{})

	for arg, key := range step.Input {
		isArgs := argsMatcher.MatchString(key)
		existMapping := true
		if isArgs {
			argKey := strings.Replace(key, "$args.", "", 1)
			result, ok := argsMap[argKey]
			if !ok {
				existMapping = false
			} else {
				inputs[arg] = result
			}
		} else {
			result, ok := outputMap[key]
			if !ok {
				existMapping = false
			} else {
				inputs[arg] = result
			}
		}
		if !existMapping {
			state.Status = repository.FAILED
			h.executionRepository.UpdateState(state)
			log.Printf("Required output key not found: %s\n", key)
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

	log.Printf("Sending message: %s\n", message)
	err = ProduceTopicMessage(h.serviceWriter, message, config.Topic)
	if err != nil {
		state.Status = repository.FAILED
		h.executionRepository.UpdateState(state)
		log.Printf("Failed to produce message: %s\n", err)
		return
	}
	state.Status = repository.EXECUTING
	h.executionRepository.UpdateState(state)
}

type ServiceResponse struct {
	ExecutionID uint                   `json:"executionId"`
	Outputs     map[string]interface{} `json:"outputs"`
}

func (h *Handler) HandleServiceResponse(message []byte) {
	response := ServiceResponse{}
	err := json.Unmarshal(message, &response)
	if err != nil {
		log.Printf("Failed to unmarshal message: %s\n", err)
		return
	}

	execution := h.executionRepository.GetExecutionById(response.ExecutionID)
	if execution == nil {
		log.Printf("Execution not found: %d\n", response.ExecutionID)
		return
	}
	state := execution.State
	if response.Outputs["error"] != nil {
		execution.State.Status = repository.FAILED
		h.executionRepository.UpdateState(execution.State)
		log.Printf("Service failed: %s\n", response.Outputs["error"])
		return
	}
	for k, v := range response.Outputs {
		state.Outputs = append(state.Outputs, &repository.KeyValueOutput{Key: k, Value: v.(string)})
	}
	//Check if all steps are done
	var nextStepIndex int
	for _, step := range execution.Steps {
		if step.Name == state.Step {
			nextStepIndex = step.StepOrder + 1
			break
		}
	}
	if nextStepIndex >= len(execution.Steps) {
		state.Status = repository.SUCCESS
	} else {
		state.Step = execution.Steps[nextStepIndex].Name
		state.Status = repository.PENDING
		stepToExecute := execution.Steps[nextStepIndex].ToExecutionStepDTO()

		//Enqueue the step
		bytes, err := json.Marshal(stepToExecute)
		if err != nil {
			log.Printf("Failed to marshal message: %s\n", err)
		}
		err = ProduceMessage(h.executionStepsWriter, bytes)
		if err != nil {
			log.Printf("Failed to produce message: %s\n", err)
		}
	}
	h.executionRepository.UpdateState(state)
}
