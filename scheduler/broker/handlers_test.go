package broker

import (
	"context"
	"github.com/goccy/go-json"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	trace2 "go.opentelemetry.io/otel/trace"
	"scheduler/repository"
	"testing"
)

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) ProduceMessage(writer *kafka.Writer, headers []kafka.Header, msg []byte) error {
	args := m.Called(writer, headers, msg)
	return args.Error(0)
}

func createTracerProvider() *trace.TracerProvider {
	spanRecorder := tracetest.NewSpanRecorder()
	return trace.NewTracerProvider(trace.WithSpanProcessor(spanRecorder))
}

func createSpan() (context.Context, trace2.Span) {
	tp := createTracerProvider()
	tracer := tp.Tracer("broker-scheduler-test")
	return tracer.Start(context.Background(), "test-span")
}

func TestHandler_EnqueueExecutionStep(t *testing.T) {
	mockProducer := new(MockProducer)

	h := NewHandler(nil, nil, nil, nil, createTracerProvider(), nil, mockProducer.ProduceMessage)

	step := repository.ExecutionStepDTO{}
	ctx, span := createSpan()
	msg, _ := json.Marshal(step)
	mockProducer.On("ProduceMessage", h.executionStepsWriter, mock.Anything, msg).Return(nil)

	h.EnqueueExecutionStep(step, ctx, span)

	mockProducer.AssertCalled(t, "ProduceMessage", h.executionStepsWriter, mock.Anything, msg)
	mockProducer.AssertExpectations(t)
}
