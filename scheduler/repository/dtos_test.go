package repository

import (
	"database/sql"
	"reflect"
	"testing"
)

func TestExecutionSubmissionDTO_ToExecution(t *testing.T) {
	type fields struct {
		WorkflowName  string
		WorkflowID    uint
		ExecutionUUID string
		Tags          []string
		Parameters    ExecutionsParamsDTO
		Arguments     map[string]string
		Steps         []SubmissionStepDTO
	}
	type args struct {
		status string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Execution
	}{
		{
			name: "Test CronDefinition",
			fields: fields{
				WorkflowName:  "Test",
				WorkflowID:    5,
				ExecutionUUID: "aaaaa5",
				Tags:          []string{"test", "test2"},
				Parameters: ExecutionsParamsDTO{
					CronDefinition: "* * * * *",
				},
				Arguments: map[string]string{"KeyTest": "ValueTest"},
				Steps: []SubmissionStepDTO{
					{
						Service: "TestService",
						Name:    "TestName",
						Task:    "TestTask",
						Input:   map[string]string{"KeyInputTest": "ValueInputTest"},
					},
					{
						Service: "TestService2",
						Name:    "TestName2",
						Task:    "TestTask2",
						Input:   map[string]string{"KeyInputTest2": "ValueInputTest2"},
					},
				},
			},
			args: args{status: PENDING},
			want: &Execution{
				WorkflowID:    5,
				ExecutionUUID: "aaaaa5",
				Tags:          []*Tags{{Tag: "test"}, {Tag: "test2"}},
				State: &State{
					Step:      "TestName",
					Status:    PENDING,
					Outputs:   make([]*KeyValueOutput, 0),
					Arguments: []*KeyValueArgument{{Key: "KeyTest", Value: "ValueTest"}},
				},
				Steps: []*Step{
					{
						Name:      "TestName",
						Service:   "TestService",
						Task:      "TestTask",
						StepOrder: 0,
						Inputs:    []*KeyValueStep{{Key: "KeyInputTest", Value: "ValueInputTest"}},
					},
					{
						Name:      "TestName2",
						Service:   "TestService2",
						Task:      "TestTask2",
						StepOrder: 1,
						Inputs:    []*KeyValueStep{{Key: "KeyInputTest2", Value: "ValueInputTest2"}},
					},
				},
				Params: &ExecutionParams{
					CronDefinition: sql.NullString{String: "* * * * *", Valid: true},
				},
			},
		},
		{
			name: "Test Delayed",
			fields: fields{
				WorkflowName:  "Test",
				WorkflowID:    5,
				ExecutionUUID: "aaaaa5",
				Tags:          []string{"test", "test2"},
				Parameters: ExecutionsParamsDTO{
					Delayed: "150",
				},
				Arguments: map[string]string{"KeyTest": "ValueTest"},
				Steps: []SubmissionStepDTO{
					{
						Service: "TestService",
						Name:    "TestName",
						Task:    "TestTask",
						Input:   map[string]string{"KeyInputTest": "ValueInputTest"},
					},
					{
						Service: "TestService2",
						Name:    "TestName2",
						Task:    "TestTask2",
						Input:   map[string]string{"KeyInputTest2": "ValueInputTest2"},
					},
				},
			},
			args: args{status: EXECUTING},
			want: &Execution{
				WorkflowID:    5,
				ExecutionUUID: "aaaaa5",
				Tags:          []*Tags{{Tag: "test"}, {Tag: "test2"}},
				State: &State{
					Step:      "TestName",
					Status:    EXECUTING,
					Outputs:   make([]*KeyValueOutput, 0),
					Arguments: []*KeyValueArgument{{Key: "KeyTest", Value: "ValueTest"}},
				},
				Steps: []*Step{
					{
						Name:      "TestName",
						Service:   "TestService",
						Task:      "TestTask",
						StepOrder: 0,
						Inputs:    []*KeyValueStep{{Key: "KeyInputTest", Value: "ValueInputTest"}},
					},
					{
						Name:      "TestName2",
						Service:   "TestService2",
						Task:      "TestTask2",
						StepOrder: 1,
						Inputs:    []*KeyValueStep{{Key: "KeyInputTest2", Value: "ValueInputTest2"}},
					},
				},
				Params: &ExecutionParams{
					CronDefinition: sql.NullString{String: "", Valid: false},
					DelayedSeconds: 150,
				},
			},
		},
		{
			name: "Test Empty Params",
			fields: fields{
				WorkflowName:  "Test",
				WorkflowID:    5,
				ExecutionUUID: "aaaaa5",
				Tags:          []string{"test", "test2"},
				Parameters:    ExecutionsParamsDTO{},
				Arguments:     map[string]string{"KeyTest": "ValueTest"},
				Steps: []SubmissionStepDTO{
					{
						Service: "TestService",
						Name:    "TestName",
						Task:    "TestTask",
						Input:   map[string]string{"KeyInputTest": "ValueInputTest"},
					},
					{
						Service: "TestService2",
						Name:    "TestName2",
						Task:    "TestTask2",
						Input:   map[string]string{"KeyInputTest2": "ValueInputTest2"},
					},
				},
			},
			args: args{status: EXECUTING},
			want: &Execution{
				WorkflowID:    5,
				ExecutionUUID: "aaaaa5",
				Tags:          []*Tags{{Tag: "test"}, {Tag: "test2"}},
				State: &State{
					Step:      "TestName",
					Status:    EXECUTING,
					Outputs:   make([]*KeyValueOutput, 0),
					Arguments: []*KeyValueArgument{{Key: "KeyTest", Value: "ValueTest"}},
				},
				Steps: []*Step{
					{
						Name:      "TestName",
						Service:   "TestService",
						Task:      "TestTask",
						StepOrder: 0,
						Inputs:    []*KeyValueStep{{Key: "KeyInputTest", Value: "ValueInputTest"}},
					},
					{
						Name:      "TestName2",
						Service:   "TestService2",
						Task:      "TestTask2",
						StepOrder: 1,
						Inputs:    []*KeyValueStep{{Key: "KeyInputTest2", Value: "ValueInputTest2"}},
					},
				},
				Params: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ExecutionSubmissionDTO{
				WorkflowName:  tt.fields.WorkflowName,
				WorkflowID:    tt.fields.WorkflowID,
				ExecutionUUID: tt.fields.ExecutionUUID,
				Tags:          tt.fields.Tags,
				Parameters:    tt.fields.Parameters,
				Arguments:     tt.fields.Arguments,
				Steps:         tt.fields.Steps,
			}
			if got := e.ToExecution(tt.args.status); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToExecution() = %v, want %v", got, tt.want)
			}
		})
	}
}
