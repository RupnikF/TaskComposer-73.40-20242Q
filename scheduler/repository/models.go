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
	WorkflowID uint
	Name       string
	Service    string
	Task       string
	StepOrder  int // Order of the step in the workflow MUST BE DONE MANUALLY
	Inputs     []KeyValueStep
}

type State struct {
	gorm.Model
	ExecutionID uint
	Step        string
	Status      string
	Outputs     []KeyValueOutput
	Arguments   []KeyValueArgument
}

type Execution struct {
	gorm.Model
	WorkflowID uint
	Tag        string
	State      State
	Params     *ExecutionParams
}
