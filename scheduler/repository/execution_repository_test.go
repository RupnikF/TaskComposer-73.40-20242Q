package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gotest.tools/v3/assert"
	"testing"
	"time"
)

func setupTestDB() (func(), *ExecutionRepository, error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForSQL("5432/tcp", "postgres",
			func(host string, port nat.Port) string {
				return fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable", host, port.Port())
			}).WithStartupTimeout(60 * time.Second),
	}

	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, err
	}

	host, _ := postgresC.Host(ctx)
	port, _ := postgresC.MappedPort(ctx, "5432")

	dsn := fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=testdb sslmode=disable", host, port.Port())

	fmt.Println("Connecting to:", dsn)

	var db *gorm.DB
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	// Migrate the schema
	err = db.AutoMigrate(&Execution{}, &State{}, &Step{}, &KeyValueOutput{}, &KeyValueArgument{}, &KeyValueStep{}, &ExecutionParams{}, &Tags{})
	if err != nil {
		return nil, nil, err
	}

	// Clean up database before each test
	db.Exec("DELETE FROM executions CASCADE")

	// Tear down the container after tests
	// defer postgresC.Terminate(ctx)
	cleanup := func() {
		err := postgresC.Terminate(ctx)
		if err != nil {
			fmt.Errorf("failed to terminate container: %v", err)
		}
	}

	return cleanup, NewExecutionRepository(db), nil
}
func GetGenericExecution() Execution {
	return Execution{
		WorkflowID:    1,
		ExecutionUUID: "123e4567-e89b-12d3-a456-426614174000",
		Tags: []*Tags{
			{Tag: "test"},
			{Tag: "automation"},
		},
		State: &State{
			Step:    "Initialize",
			Status:  PENDING,
			Outputs: []*KeyValueOutput{},
			Arguments: []*KeyValueArgument{
				{Key: "param1", Value: "value1"},
				{Key: "param2", Value: "value2"},
			},
		},
		Steps: []*Step{
			{
				Name:    "Step 1",
				Service: "AuthService",
				Task:    "Authenticate",
				Inputs: []*KeyValueStep{
					{Key: "username", Value: "test_user"},
					{Key: "password", Value: "test_pass"},
				},
			},
			{
				Name:    "Step 2",
				Service: "DataService",
				Task:    "FetchData",
				Inputs: []*KeyValueStep{
					{Key: "query", Value: "SELECT * FROM users"},
				},
			},
		},
		Params: &ExecutionParams{
			DelayedSeconds: 10,
			CronDefinition: sql.NullString{String: "0 0 * * *", Valid: true},
		},
	}
}

func TestExecutionRepository_CancelExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	cleanup, repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()
	testExec := GetGenericExecution()
	repo.db.Create(&testExec)

	repo.CancelExecution(context.Background(), &testExec)
	newExec := Execution{}
	repo.db.Preload("State").First(&newExec, testExec.ID)

	assert.Equal(t, newExec.State.Status, CANCELLED)
}

func TestExecutionRepository_CreateExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	cleanup, repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()
	testExec := GetGenericExecution()
	repo.CreateExecution(context.Background(), &testExec)

	assert.Equal(t, testExec.ID, uint(1))
	assert.Equal(t, testExec.ExecutionUUID, "123e4567-e89b-12d3-a456-426614174000")
	assert.Equal(t, testExec.State.ExecutionID, testExec.ID)
	assert.Equal(t, testExec.State.Step, "Initialize")
	assert.Equal(t, testExec.Steps[0].ExecutionID, testExec.ID)
	assert.Equal(t, testExec.Params.DelayedSeconds, uint(10))
	assert.Equal(t, testExec.Params.ExecutionID, testExec.ID)
	assert.Equal(t, testExec.Tags[0].ExecutionID, testExec.ID)
	assert.Equal(t, testExec.State.Arguments[0].StateID, testExec.State.ID)
	assert.Equal(t, testExec.Steps[0].Inputs[0].StepID, testExec.Steps[0].ID)

}

func TestExecutionRepository_GetExecutionById(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	cleanup, repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()
	testExec := GetGenericExecution()
	repo.db.Create(&testExec)
	exec := repo.GetExecutionById(testExec.ID)
	if exec == nil {
		t.Fatalf("failed to get execution by id: %v", err)
	}
	assert.Equal(t, exec.ID, testExec.ID)
}

func TestExecutionRepository_GetExecutionByUUID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	cleanup, repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()
	testExec := GetGenericExecution()
	repo.db.Create(&testExec)
	exec := repo.GetExecutionByUUID(testExec.ExecutionUUID)
	if exec == nil {
		t.Fatalf("failed to get execution by uuid: %v", err)
	}
	assert.Equal(t, exec.ExecutionUUID, testExec.ExecutionUUID)
}

func TestExecutionRepository_GetExecutionsByTags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	cleanup, repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()
	testExec := GetGenericExecution()
	repo.db.Create(&testExec)
	exec := repo.GetExecutionsByTags(context.Background(), []string{"test"})
	if exec == nil {
		t.Fatalf("failed to get execution by tags: %v", err)
	}
	assert.Equal(t, exec[0].ExecutionUUID, testExec.ExecutionUUID)
	assert.Equal(t, len(exec), 1)
	found := false
	for _, tag := range exec[0].Tags {
		if tag.Tag == "test" {
			found = true
		}
	}
	assert.Equal(t, found, true)
}

func TestExecutionRepository_GetStateByExecutionID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	cleanup, repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()
	testExec := GetGenericExecution()
	repo.db.Create(&testExec)
	state := repo.GetStateByExecutionID(testExec.ID)
	if state == nil {
		t.Fatalf("failed to get state by execution id: %v", err)
	}
	assert.Equal(t, state.ExecutionID, testExec.ID)
}

func TestExecutionRepository_UpdateState(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	cleanup, repo, err := setupTestDB()
	if err != nil {
		t.Fatalf("failed to setup test db: %v", err)
	}
	defer cleanup()
	testExec := GetGenericExecution()
	repo.db.Create(&testExec)
	testExec.State.Status = EXECUTING
	repo.UpdateState(context.Background(), testExec.State)
	newExec := Execution{}
	repo.db.Preload("State").First(&newExec, testExec.ID)

	assert.Equal(t, newExec.State.Status, EXECUTING)

}
