package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
)

func TestNewZapLogger_JSONFormat(t *testing.T) {
	var buffer bytes.Buffer
	writer := zapcore.AddSync(&buffer)

	encoder := getEncoder("json")
	core := zapcore.NewCore(encoder, writer, zapcore.DebugLevel)
	logger := &zapLogger{logger: zap.New(core)}

	testMessage := "this is a test"
	logger.Info(context.Background(), testMessage, zap.String("key", "value"))

	var logOutput map[string]interface{}
	err := json.Unmarshal(buffer.Bytes(), &logOutput)
	require.NoError(t, err, "Logger output should be valid JSON")

	assert.Equal(t, "INFO", logOutput["level"], "Log level should be 'INFO'")
	assert.Equal(t, testMessage, logOutput["msg"], "Log message should match")
	assert.Equal(t, "value", logOutput["key"], "Log field should match")
}

func TestAddTraceID(t *testing.T) {
	traceID, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := trace.SpanIDFromHex("0102030405060708")
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.TraceFlags(0x01),
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanContext)

	fields := addTraceID(ctx, []zap.Field{})

	require.Len(t, fields, 2, "Two fields should have been added")
	assert.Equal(t, "traceID", fields[0].Key, "Field key should be 'traceID'")
	assert.Equal(t, traceID.String(), fields[0].String, "Field value should match the trace ID")
	assert.Equal(t, "spanID", fields[1].Key, "Field key should be 'spanID'")
	assert.Equal(t, spanID.String(), fields[1].String, "Field value should match the span ID")
}

func TestAddTraceID_NoSpan(t *testing.T) {
	ctx := context.Background()
	fields := addTraceID(ctx, []zap.Field{})
	assert.Empty(t, fields, "No fields should be added when there is no span in context")
}
