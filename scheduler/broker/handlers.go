package broker

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"regexp"
	"scheduler/jobs"
	"scheduler/repository"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/goccy/go-json"
	kafka "github.com/segmentio/kafka-go"
)

var handlerLogger = otelslog.NewLogger("handlers")

type Handler struct {
	executionRepository    *repository.ExecutionRepository
	serviceRepository      *repository.ServiceRepository
	executionStepsWriter   *kafka.Writer
	jobsRepository         *jobs.JobsRepository
	servicesWriters        map[string]*kafka.Writer
	produceMessageFunction func(writer *kafka.Writer, headers []kafka.Header, message []byte) error
	tracer                 trace.Tracer
}

func NewHandler(
	executionRepository *repository.ExecutionRepository,
	serviceRepository *repository.ServiceRepository,
	executionStepsWriter *kafka.Writer,
	servicesWriters map[string]*kafka.Writer,
	tracerProvider trace.TracerProvider,
	jobsRepository *jobs.JobsRepository,
	produceMessageFunction func(writer *kafka.Writer, headers []kafka.Header, message []byte) error,
) *Handler {
	return &Handler{
		executionRepository,
		serviceRepository,
		executionStepsWriter,
		jobsRepository,
		servicesWriters,
		produceMessageFunction,
		tracerProvider.Tracer("kafka-handlers"),
	}
}

func (h *Handler) EnqueueExecutionStep(stepToExecute repository.ExecutionStepDTO, ctx context.Context, span trace.Span) {
	//Enqueue the step
	bytes, err := json.Marshal(stepToExecute)
	if err != nil {
		log.Printf("Failed to marshal message: %s\n", err)
		span.RecordError(err)
		return
	}
	err = h.produceMessageFunction(h.executionStepsWriter, h.PassHeader(ctx), bytes)
	if err != nil {
		log.Printf("Failed to produce message: %s\n", err)
		span.RecordError(err)
		return
	}
}

func (h *Handler) HandleExecutionSubmission(message []byte, header []kafka.Header) {
	ctx, span := h.CreateOrGetSpan("HandleExecutionSubmission", header)
	defer span.End()

	handlerLogger.Info("Received message from submission", "message", string(message))
	submission := repository.ExecutionSubmissionDTO{}
	err := json.Unmarshal(message, &submission)
	if err != nil {
		handlerLogger.Error("Failed to unmarshal message", "error", err)
		span.RecordError(err)
		return
	}
	log.Printf("Received submission: %v\n", submission)
	execution := submission.ToExecution(repository.PENDING)
	if execution == nil {
		handlerLogger.Error("Error parsing execution\n")
		span.RecordError(err)
		return
	}
	var useUUID *bool
	useUUID = new(bool)
	*useUUID = true
	if execution.Params != nil {
		if execution.Params.CronDefinition.Valid {
			h.jobsRepository.CreateCronJob(execution.Params.CronDefinition.String, ctx, func() {
				executionCopy := submission.ToExecution(repository.PENDING)
				executionCopy.JobID = executionCopy.ExecutionUUID
				if !*useUUID {
					executionCopy.ExecutionUUID = uuid.New().String()
				} else {
					*useUUID = false
				}
				h.executionRepository.CreateExecution(context.Background(), executionCopy)
				stepToExecute := executionCopy.Steps[0].ToExecutionStepDTO()
				h.EnqueueExecutionStep(stepToExecute, ctx, span)
			}, execution.ExecutionUUID)
		} else if execution.Params.DelayedSeconds > 0 {
			h.jobsRepository.CreateDelayedJob(execution.Params.DelayedSeconds, ctx, func() {
				h.executionRepository.CreateExecution(context.Background(), execution)
				stepToExecute := execution.Steps[0].ToExecutionStepDTO()
				h.EnqueueExecutionStep(stepToExecute, ctx, span)
			}, execution.ExecutionUUID)
		}
	} else {
		h.executionRepository.CreateExecution(context.Background(), execution)
		stepToExecute := execution.Steps[0].ToExecutionStepDTO()
		h.EnqueueExecutionStep(stepToExecute, ctx, span)
	}
}

type ServiceMessage struct {
	ExecutionId uint                   `json:"executionId"`
	TaskName    string                 `json:"taskName"`
	Inputs      map[string]interface{} `json:"inputs"`
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
	var headers = make([]kafka.Header, 0)
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

	config, err := h.serviceRepository.GetService(step.Service)
	if err != nil {
		log.Printf("Service not found: %s\n", step.Service)
		span.RecordError(fmt.Errorf("service not found: %s", step.Service))
		return
	}

	state := h.executionRepository.GetStateByExecutionID(step.ExecutionID)

	if state.Status != repository.PENDING {
		log.Printf("Execution not pending: %s\n", state.Status)
		span.RecordError(fmt.Errorf("execution not pending: %s", state.Status))
		return
	}
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
		fmt.Println("inputs", arg, key, isArgs)
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
				// Use hardcoded value
				inputs[arg] = key
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
	fmt.Println("Found inputs", inputs)
	if step.Service == "native" {
		h.HandleNativeStep(step, state, inputs, span, ctx)
		return
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
	writer := h.servicesWriters[config.Name]
	if writer == nil {
		log.Printf("Writer not found: %s\n", config.Name)
		return
	}

	err = h.produceMessageFunction(writer, h.PassHeader(ctx), message)
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

func (h *Handler) HandleNativeStep(
	step repository.ExecutionStepDTO,
	state *repository.State,
	inputs map[string]interface{},
	span trace.Span,
	ctx context.Context,
) {
	handlerMapper := map[string]nativeFn{
		"abort": h.AbortHandler,
		"if":    h.ConditionalHandler,
	}

	handler, ok := handlerMapper[step.Task]
	if !ok {
		handler = h.ErrorHandler
	}
	handler(step, state, inputs, span, ctx)
}

type ServiceResponse struct {
	ExecutionID uint                   `json:"executionId"`
	Outputs     map[string]interface{} `json:"outputs"`
	TraceId     string                 `json:"traceId"`
}

func (h *Handler) HandleServiceResponse(message []byte, header []kafka.Header) {
	ctx, span := h.CreateOrGetSpan("HandleServiceResponse", header)
	defer span.End()
	fmt.Println("Received message from service", string(message))
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
	if state.Status != repository.EXECUTING {
		log.Printf("Execution not executing: %s\n", state.Status)
		span.RecordError(fmt.Errorf("execution not executing: %s", state.Status))
		return
	}
	outputErr := response.Outputs["error"]
	if outputErr != nil {
		execution.State.Status = repository.FAILED

		errorMsg, ok := outputErr.(map[string]string)
		if !ok {
			errorMsg = map[string]string{
				"msg": "Error message is not an object",
			}
		}

		str, err := json.Marshal(errorMsg)
		if err != nil {
			span.RecordError(err)
		}
		execution.State.Outputs = append(state.Outputs, &repository.KeyValueOutput{Key: "error", Value: string(str)})
		h.executionRepository.UpdateState(context.Background(), execution.State)
		log.Printf("Service failed: %s\n", response.Outputs["error"])
		span.RecordError(err)
		return
	}
	for k, v := range response.Outputs {
		state.Outputs = append(state.Outputs, &repository.KeyValueOutput{Key: fmt.Sprintf("%s.%s", state.Step, k), Value: v.(string)})
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
		h.executionRepository.UpdateState(context.Background(), state)

		//Enqueue the step
		bytes, err := json.Marshal(stepToExecute)
		if err != nil {
			log.Printf("Failed to marshal message: %s\n", err)
			// TODO manejar este caso
		}
		err = h.produceMessageFunction(h.executionStepsWriter, h.PassHeader(ctx), bytes)
		if err != nil {
			log.Printf("Failed to produce message: %s\n", err)
			// TODO manejar este caso
		}
	}
	h.executionRepository.UpdateState(context.Background(), state)
}
