package broker

import (
	"context"
	"fmt"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"log"
	"regexp"
	"scheduler/repository"
	"strings"

	"github.com/goccy/go-json"
	kafka "github.com/segmentio/kafka-go"
)

var handlerLogger = otelslog.NewLogger("handlers")

type Handler struct {
	executionRepository  *repository.ExecutionRepository
	serviceRepository    *repository.ServiceRepository
	executionStepsWriter *kafka.Writer
	serviceWriter        *kafka.Writer
	tracer               trace.Tracer
}

func NewHandler(
	executionRepository *repository.ExecutionRepository,
	serviceRepository *repository.ServiceRepository,
	executionStepsWriter *kafka.Writer,
	serviceWriter *kafka.Writer,
	tracerProvider trace.TracerProvider,
) *Handler {
	return &Handler{
		executionRepository,
		serviceRepository,
		executionStepsWriter,
		serviceWriter,
		tracerProvider.Tracer("kafka-handlers"),
	}
}

func (h *Handler) HandleExecutionSubmission(message []byte, header []kafka.Header) {
	ctx, span := h.CreateOrGetSpan("HandleExecutionSubmission", header)
	defer span.End()

	handlerLogger.Info("Received message from submission", string(message))
	submission := repository.ExecutionSubmissionDTO{}
	err := json.Unmarshal(message, &submission)
	if err != nil {
		handlerLogger.Error("Failed to unmarshal message: %s\n", err)
		span.RecordError(err)
		return
	}
	log.Printf("Received submission: %v\n", submission)
	execution := submission.ToExecution(repository.PENDING)
	if execution == nil {
		log.Printf("Error parsing execution\n")
		span.RecordError(err)
		return
	}
	h.executionRepository.CreateExecution(context.Background(), execution)
	stepToExecute := execution.Steps[0].ToExecutionStepDTO()

	//Enqueue the step
	bytes, err := json.Marshal(stepToExecute)
	if err != nil {
		log.Printf("Failed to marshal message: %s\n", err)
		span.RecordError(err)
		return
	}
	err = ProduceMessage(h.executionStepsWriter, h.PassHeader(ctx), bytes)
	if err != nil {
		log.Printf("Failed to produce message: %s\n", err)
		span.RecordError(err)
		return
	}
}

type ServiceMessage struct {
	ExecutionId uint                   `json:"executionId"`
	TaskName    string                 `json:"taskName"`
	Inputs      map[string]interface{} `json:"inputs"`
	TraceId     string                 `json:"traceId"`
}

// Matches with all strings that start with args.

func (h *Handler) CreateOrGetSpan(spanName string, header []kafka.Header) (context.Context, trace.Span) {
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

func (h *Handler) PassHeader(ctx context.Context) []kafka.Header {
	propagator := otel.GetTextMapPropagator()
	carrier := make(propagation.MapCarrier)
	propagator.Inject(ctx, carrier)
	if carrier == nil {
		return []kafka.Header{}
	}
	var headers []kafka.Header = make([]kafka.Header, 0)
	for k, v := range carrier {
		headers = append(headers, kafka.Header{
			Key: k, Value: []byte(v),
		})
	}
	return headers
}

func (h *Handler) HandleExecutionStep(message []byte, header []kafka.Header) {
	ctx, span := h.CreateOrGetSpan("HandleExecutionStep", header)
	defer span.End()

	step := repository.ExecutionStepDTO{}
	err := json.Unmarshal(message, &step)
	if err != nil {
		log.Printf("Failed to unmarshal message: %s\n", err)
		span.RecordError(err)
		return
	}

	log.Printf("Received step: %v\n", step)
	span.SetAttributes(attribute.Int("ExecutionId", int(step.ExecutionID)))
	span.SetAttributes(attribute.String("ServiceName", step.Service))
	span.SetAttributes(attribute.String("Task", step.Task))
	span.SetAttributes(attribute.String("Step", step.Name))

	config := h.serviceRepository.GetService(step.Service)
	if config == nil {
		log.Printf("Service not found: %s\n", step.Service)
		span.RecordError(fmt.Errorf("service not found: %s", step.Service))
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
			span.RecordError(err)
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
			h.executionRepository.UpdateState(context.Background(), state)
			log.Printf("Required output key not found: %s\n", key)
			span.RecordError(err)
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
		span.RecordError(err)
		return
	}

	log.Printf("Sending message: %s\n", message)

	err = ProduceTopicMessage(h.serviceWriter, message, h.PassHeader(ctx), config.Topic)
	if err != nil {
		state.Status = repository.FAILED
		h.executionRepository.UpdateState(context.Background(), state)
		log.Printf("Failed to produce message: %s\n", err)
		span.RecordError(err)
		return
	}
	state.Status = repository.EXECUTING
	h.executionRepository.UpdateState(context.Background(), state)
}

type ServiceResponse struct {
	ExecutionID uint                   `json:"executionId"`
	Outputs     map[string]interface{} `json:"outputs"`
	TraceId     string                 `json:"traceId"`
}

func (h *Handler) HandleServiceResponse(message []byte, header []kafka.Header) {
	ctx, span := h.CreateOrGetSpan("HandleExecutionSubmission", header)
	defer span.End()

	response := ServiceResponse{}
	err := json.Unmarshal(message, &response)
	if err != nil {
		log.Printf("Failed to unmarshal message: %s\n", err)
		span.RecordError(err)
		return
	}

	execution := h.executionRepository.GetExecutionById(response.ExecutionID)
	if execution == nil {
		log.Printf("Execution not found: %d\n", response.ExecutionID)
		span.RecordError(err)
		return
	}
	state := execution.State
	if response.Outputs["error"] != nil {
		execution.State.Status = repository.FAILED
		h.executionRepository.UpdateState(context.Background(), execution.State)
		log.Printf("Service failed: %s\n", response.Outputs["error"])
		span.RecordError(err)
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
		span.SetAttributes(attribute.Bool("Finished", true))
	} else {
		state.Step = execution.Steps[nextStepIndex].Name
		state.Status = repository.PENDING
		stepToExecute := execution.Steps[nextStepIndex].ToExecutionStepDTO()

		//Enqueue the step
		bytes, err := json.Marshal(stepToExecute)
		if err != nil {
			log.Printf("Failed to marshal message: %s\n", err)
			// TODO manejar este caso
		}
		err = ProduceMessage(h.executionStepsWriter, h.PassHeader(ctx), bytes)
		if err != nil {
			log.Printf("Failed to produce message: %s\n", err)
			// TODO manejar este caso
		}
	}
	h.executionRepository.UpdateState(context.Background(), state)
}
