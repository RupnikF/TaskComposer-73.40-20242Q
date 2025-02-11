package broker

import (
	"context"
	"fmt"
	"log"
	"scheduler/repository"
	"strings"

	"github.com/goccy/go-json"
	"go.opentelemetry.io/otel/trace"
)

type nativeFn func(
	step repository.ExecutionStepDTO,
	state *repository.State,
	inputs map[string]interface{},
	span trace.Span,
	ctx context.Context,
)

// TODO: Unit Test? | Integration Test?
func (h *Handler) ConditionalHandler(
	step repository.ExecutionStepDTO,
	state *repository.State,
	inputs map[string]interface{},
	span trace.Span,
	ctx context.Context,
) {
	execution := h.executionRepository.GetExecutionById(ctx, state.ExecutionID)
	leftValue, leftOk := inputs["leftValue"]
	rightValue, rightOk := inputs["rightValue"]
	operator, opOk := inputs["operator"]
	onTruePath, onTrueOk := inputs["onTrue"]
	onFalsePath, onFalseOk := inputs["onFalse"]

	values := [...]bool{leftOk, rightOk, opOk, onTrueOk, onFalseOk}
	valuesName := [...]string{"leftValue", "rightValue", "operator", "onTrue", "onFalse"}

	failed := make([]string, 0)
	for i, value := range values {
		if !value {
			failed = append(failed, valuesName[i])
		}
	}
	if len(failed) > 0 {
		state.Status = repository.FAILED
		state.Outputs = append(state.Outputs, &repository.KeyValueOutput{
			Key:   "error",
			Value: fmt.Sprintf("Following required properties are not specified %s", strings.Join(failed, ",")),
		})
		h.executionRepository.UpdateState(context.Background(), state)
		return
	}

	// Evaluate condition
	var evalPath string
	var evalResult bool
	if operator == "==" {
		evalResult = leftValue == rightValue
		if evalResult {
			evalPath = onTruePath.(string)
		} else {
			evalPath = onFalsePath.(string)
		}
	} else if operator == "!=" {
		evalResult = leftValue != rightValue
		if evalResult {
			evalPath = onFalsePath.(string)
		} else {
			evalPath = onTruePath.(string)
		}
	} else {
		state.Status = repository.FAILED
		state.Outputs = append(state.Outputs, &repository.KeyValueOutput{
			Key:   "error",
			Value: fmt.Sprintf("%s is not a valid operator", operator),
		})
		h.executionRepository.UpdateState(context.Background(), state)
		return
	}

	steps := execution.Steps
	var currentStep *repository.Step
	var nextStep *repository.Step
	var inmediateNextStep *repository.Step
	for _, value := range steps {
		if evalPath == value.Name {
			nextStep = value
		}
		if currentStep != nil {
			inmediateNextStep = value
		}
		if value.Name == step.Name {
			currentStep = value
		}
	}

	if evalPath == "continue" {
		nextStep = inmediateNextStep
	} else if nextStep == nil {
		// No found path
		state.Status = repository.FAILED
		state.Outputs = append(state.Outputs, &repository.KeyValueOutput{
			Key:   "error",
			Value: fmt.Sprintf("Path %s is not found", evalPath),
		})
		h.executionRepository.UpdateState(context.Background(), state)
		return
	}

	// Got to end of steps
	if nextStep == nil {
		state.Outputs = append(state.Outputs, &repository.KeyValueOutput{
			Key:   step.Name + ".result",
			Value: fmt.Sprintf("%t", evalResult),
		})
		state.Status = repository.SUCCESS
	} else {
		stepToExecute := nextStep.ToExecutionStepDTO()

		//Enqueue the step
		bytes, err := json.Marshal(stepToExecute)
		if err != nil {
			log.Printf("Failed to marshal message: %s\n", err)
			span.RecordError(err)
			state.Status = repository.FAILED
			h.executionRepository.UpdateState(context.Background(), state)
			return
		}
		err = ProduceMessage(h.executionStepsWriter, h.PassHeader(ctx), bytes)
		if err != nil {
			log.Printf("Failed to produce message: %s\n", err)
			state.Status = repository.FAILED
			h.executionRepository.UpdateState(context.Background(), state)
			return
		}
		state.Step = nextStep.Name
		state.Outputs = append(state.Outputs, &repository.KeyValueOutput{
			Key:   step.Name + ".result",
			Value: fmt.Sprintf("%t", evalResult),
		})
	}
	h.executionRepository.UpdateState(context.Background(), state)
}

func (h *Handler) AbortHandler(
	step repository.ExecutionStepDTO,
	state *repository.State,
	inputs map[string]interface{},
	span trace.Span,
	ctx context.Context,
) {
	state.Status = repository.SUCCESS
	h.executionRepository.UpdateState(context.Background(), state)
}

func (h *Handler) ErrorHandler(
	step repository.ExecutionStepDTO,
	state *repository.State,
	inputs map[string]interface{},
	span trace.Span,
	ctx context.Context,
) {
	state.Status = repository.FAILED
	state.Outputs = append(state.Outputs, &repository.KeyValueOutput{
		Key:   "error",
		Value: fmt.Sprintf("%s is not a valid native taskname", step.Task),
	})
	h.executionRepository.UpdateState(context.Background(), state)
}
