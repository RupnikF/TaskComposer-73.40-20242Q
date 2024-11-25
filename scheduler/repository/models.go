package repository

import (
	"database/sql"
	"gorm.io/gorm"
	"time"
)

type KeyValueOutput struct {
	gorm.Model
	Key     string
	Value   string
	StateID uint
}
type KeyValueArgument struct {
	gorm.Model
	Key     string
	Value   string
	StateID uint
}
type KeyValueStep struct {
	gorm.Model
	Key    string
	Value  string
	StepID uint
}
type ExecutionParams struct {
	gorm.Model
	ScheduleTime   *time.Time
	CronDefinition sql.NullString
	ExecutionID    uint
}
type Step struct {
	gorm.Model
	ExecutionID uint
	Name        string
	Service     string
	Task        string
	StepOrder   int // Order of the step in the workflow MUST BE DONE MANUALLY
	Inputs      []*KeyValueStep
}

func (s *Step) toExecutionStepDTO() ExecutionStepDTO {
	inputs := make(map[string]string)
	for _, i := range s.Inputs {
		inputs[i.Key] = i.Value
	}
	return ExecutionStepDTO{
		Service:     s.Service,
		Name:        s.Name,
		Task:        s.Task,
		Input:       inputs,
		ExecutionID: s.ExecutionID,
		StepOrder:   s.StepOrder,
	}
}

type State struct {
	gorm.Model
	ExecutionID uint
	Step        string
	Status      string
	Outputs     []*KeyValueOutput
	Arguments   []*KeyValueArgument
}

type Execution struct {
	gorm.Model
	WorkflowID uint
	Tags       []string
	State      *State
	Steps      []*Step
	Params     *ExecutionParams
}
