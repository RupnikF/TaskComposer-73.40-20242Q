package repository

import "log"

type SubmisssionStepDTO struct {
	Service string            `json:"service"`
	Name    string            `json:"name"`
	Task    string            `json:"task"`
	Input   map[string]string `json:"input"`
}

func (s *SubmisssionStepDTO) ToStep(stepIndex int) Step {
	inputs := make([]*KeyValueStep, len(s.Input))
	for k, v := range s.Input {
		inputs = append(inputs, &KeyValueStep{
			Key:   k,
			Value: v,
		})
	}
	return Step{
		Service:   s.Service,
		Name:      s.Name,
		Task:      s.Task,
		Inputs:    inputs,
		StepOrder: stepIndex,
	}
}

type ExecutionSubmissionDTO struct {
	WorkflowName string               `json:"workflowName"`
	WorkflowID   uint                 `json:"workflowID"`
	Tags         []string             `json:"tags"`
	Parameters   map[string]string    `json:"parameters"`
	Arguments    map[string]string    `json:"args"`
	Steps        []SubmisssionStepDTO `json:"steps"`
}

func (e ExecutionSubmissionDTO) ToExecution(status string) *Execution {

	params := ExecutionParams{}

	outputs := make([]*KeyValueOutput, 0)
	arguments := make([]*KeyValueArgument, len(e.Arguments))
	if e.Arguments != nil {
		for k, v := range e.Arguments {
			arguments = append(arguments, &KeyValueArgument{
				Key:   k,
				Value: v,
			})
		}
	}

	steps := make([]*Step, len(e.Steps))
	if e.Steps == nil {
		e.Steps = []
	}
	for i, s := range e.Steps {
		step := s.ToStep(i)
		steps = append(steps, &step)
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

	return &Execution{
		WorkflowID: e.WorkflowID,
		// Tags:       e.Tags,
		Params: &params,
		Steps:  steps,
		State:  &state,
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
