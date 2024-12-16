package repository

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

const (
	PENDING   string = "PENDING"
	EXECUTING string = "EXECUTING"
	SUCCESS   string = "SUCCESS"
	FAILED    string = "FAILED"
	CANCELLED string = "CANCELLED"
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

func (s *Step) ToExecutionStepDTO() ExecutionStepDTO {
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
func (s *State) ToResponseStateDTO() ExecutionStateResponseDTO {
	outputs := make(map[string]interface{})
	for _, o := range s.Outputs {
		outputs[o.Key] = o.Value
	}
	return ExecutionStateResponseDTO{
		Step:    s.Step,
		Status:  s.Status,
		Outputs: outputs,
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
type Tags struct {
	gorm.Model
	ExecutionID uint
	Tag         string
}
type Execution struct {
	gorm.Model
	WorkflowID    uint
	ExecutionUUID string `gorm:"type:varchar(64)"`
	Tags          []*Tags
	State         *State
	Steps         []*Step
	Params        *ExecutionParams
}
