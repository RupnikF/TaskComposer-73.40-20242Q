package service

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	trace2 "go.opentelemetry.io/otel/trace"
	"testing"
)

func createSpan() (context.Context, trace2.Span) {
	spanRecorder := tracetest.NewSpanRecorder()
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(spanRecorder))
	tracer := tp.Tracer("ubuntu-service-test")
	return tracer.Start(context.Background(), "test-span")
}

func TestUnknownCommand(t *testing.T) {
	_, span := createSpan()
	defer span.End()

	inputs := make(map[string]interface{})
	inputs["cmd"] = "aaaaa"
	response, err := RunShell(inputs, span)
	if err != nil {
		t.Errorf("Unknown command shouldn't have failed")
		return
	}
	assert.Contains(t, response.Stderr, "Unknown command")
}

//func TestListCommands(t *testing.T) {
//	_, span := createSpan()
//	defer span.End()
//	inputs := make(map[string]interface{})
//	inputs["cmd"] = "compgen -c"
//
//	response, err := RunShell(inputs, span)
//	if err != nil {
//		return
//	}
//	fmt.Println(response)
//}

func TestEmpty(t *testing.T) {
	_, span := createSpan()
	defer span.End()

	inputs := make(map[string]interface{})
	_, err := RunShell(inputs, span)
	if err == nil {
		t.Errorf("Empty command should have failed")
	}
}

func TestEcho(t *testing.T) {
	_, span := createSpan()
	defer span.End()
	inputs := make(map[string]interface{})
	inputs["cmd"] = "echo"

	response, err := RunShell(inputs, span)
	if err != nil {
		t.Errorf("Should not have failed")
		return
	}
	assert.Equal(t, "", response.Stdout)
}

func TestReplace(t *testing.T) {
	_, span := createSpan()
	defer span.End()
	inputs := make(map[string]interface{})
	inputs["cmd"] = "echo {{1}}"
	inputs["1"] = "hello world"
	response, err := RunShell(inputs, span)
	if err != nil {
		t.Errorf("Should not have failed")
		return
	}
	assert.Equal(t, "hello world", response.Stdout)
}

func TestMissingReplacement(t *testing.T) {
	_, span := createSpan()
	defer span.End()
	inputs := make(map[string]interface{})
	inputs["cmd"] = "echo {{1}}"
	_, err := RunShell(inputs, span)
	assert.NotNil(t, err)
}

func TestGcEvaluate(t *testing.T) {
	_, span := createSpan()
	defer span.End()
	inputs := make(map[string]interface{})
	inputs["cmd"] = "echo 1+2 | bc"
	res, err := RunShell(inputs, span)
	if err != nil {
		t.Errorf("Should not have failed")
		return
	}
	assert.Equal(t, "3", res.Stdout)
}

func TestPipeUnknownCmd(t *testing.T) {
	_, span := createSpan()
	defer span.End()
	inputs := make(map[string]interface{})
	inputs["cmd"] = "echo 1+2 | aaaa"
	res, err := RunShell(inputs, span)
	if err != nil {
		t.Errorf("Should not have failed")
		return
	}
	assert.Contains(t, res.Stderr, "Unknown command")
}

func TestGeneralExp(t *testing.T) {
	_, span := createSpan()
	defer span.End()
	inputs := make(map[string]interface{})
	inputs["exp"] = "2*1"
	res, err := Eval(inputs, span)
	if err != nil {
		t.Errorf("Should not have failed %s", err)
		return
	}
	assert.Equal(t, "2", res)
}

func TestExpWithReplace(t *testing.T) {
	_, span := createSpan()
	defer span.End()
	inputs := make(map[string]interface{})
	inputs["exp"] = "{{a}}*{{b}}"
	inputs["a"] = "123"
	inputs["b"] = "456"
	res, err := Eval(inputs, span)
	if err != nil {
		t.Errorf("Should not have failed %s", err)
	}
	assert.Equal(t, fmt.Sprintf("%d", 123*456), res)
}

func TestExpWithReplace2(t *testing.T) {
	_, span := createSpan()
	defer span.End()
	inputs := make(map[string]interface{})
	inputs["exp"] = "{{a}} * 100"
	inputs["a"] = "7"
	res, err := Eval(inputs, span)
	if err != nil {
		t.Errorf("Should not have failed %s", err)
	}
	assert.Equal(t, fmt.Sprintf("%d", 7*100), res)
}
