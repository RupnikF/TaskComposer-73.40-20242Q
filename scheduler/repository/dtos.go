package repository

import (
	"database/sql"
	"log"
	"strconv"
)

type SubmissionStepDTO struct {
	Service string            `json:"service"`
	Name    string            `json:"name"`
	Task    string            `json:"task"`
	Input   map[string]string `json:"input"`
}

func (s *SubmissionStepDTO) ToStep(stepIndex int) Step {
	inputs := make([]*KeyValueStep, len(s.Input))
	i := 0
	for k, v := range s.Input {
		inputs[i] = &KeyValueStep{
			Key:   k,
			Value: v,
		}
		i++
	}
	return Step{
		Service:   s.Service,
		Name:      s.Name,
		Task:      s.Task,
		Inputs:    inputs,
		StepOrder: stepIndex,
	}
}

type ExecutionsParamsDTO struct {
	CronDefinition string `json:"cronDefinition"`
	Delayed        string `json:"delayed"`
}

type ExecutionSubmissionDTO struct {
	WorkflowName  string              `json:"workflowName"`
	WorkflowID    uint                `json:"workflowID"`
	ExecutionUUID string              `json:"ExecutionUUID"`
	Tags          []string            `json:"tags"`
	Parameters    ExecutionsParamsDTO `json:"parameters"`
	Arguments     map[string]string   `json:"args"`
	Steps         []SubmissionStepDTO `json:"steps"`
}

func (e ExecutionSubmissionDTO) ToExecution(status string) *Execution {

	params := &ExecutionParams{}
	paramsEmpty := true
	if e.Parameters.CronDefinition != "" {
		params.CronDefinition = sql.NullString{String: e.Parameters.CronDefinition, Valid: true}
		paramsEmpty = false
	}
	if e.Parameters.Delayed != "" {
		seconds, err := strconv.ParseUint(e.Parameters.Delayed, 10, 32)
		if err != nil {
			log.Printf("Failed to parse delayed seconds: %s\n", err)
			params.DelayedSeconds = 0
		} else {
			params.DelayedSeconds = uint(seconds)
			paramsEmpty = false
		}

	}

	outputs := make([]*KeyValueOutput, 0)
	arguments := make([]*KeyValueArgument, len(e.Arguments))
	if e.Arguments != nil {
		i := 0
		for k, v := range e.Arguments {
			arguments[i] = &KeyValueArgument{
				Key:   k,
				Value: v,
			}
			i++
		}
	}

	steps := make([]*Step, len(e.Steps))
	if e.Steps == nil {
		e.Steps = make([]SubmissionStepDTO, 0)
	}
	for i, s := range e.Steps {
		step := s.ToStep(i)
		steps[i] = &step
	}
	if len(steps) <= 0 {
		log.Printf("No steps provided, skipped\n")
		return nil
	}
	state := State{
		Step:      steps[0].Name,
		Status:    status,
		Outputs:   outputs,
		Arguments: arguments,
	}
	tags := make([]*Tags, len(e.Tags))
	for i, t := range e.Tags {
		tags[i] = &Tags{
			Tag: t,
		}
	}
	if paramsEmpty {
		params = nil
	}
	return &Execution{
		WorkflowID:    e.WorkflowID,
		Tags:          tags,
		Params:        params,
		Steps:         steps,
		State:         &state,
		ExecutionUUID: e.ExecutionUUID,
	}
}

type ExecutionStepDTO struct {
	Service     string            `json:"service"`
	Name        string            `json:"name"`
	Task        string            `json:"task"`
	Input       map[string]string `json:"input"`
	ExecutionID uint              `json:"execution_id"`
	StepOrder   int               `json:"step_order"`
}

type ExecutionStateResponseDTO struct {
	Step    string                 `json:"step"`
	Status  string                 `json:"status"`
	Outputs map[string]interface{} `json:"outputs"`
}

type CancelTagsDTO struct {
	Tags []string `json:"tags"`
}
