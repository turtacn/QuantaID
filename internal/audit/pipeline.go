package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"os"
)

// Sink is the interface for audit event destinations.
type Sink interface {
	Write(event *AuditEvent) error
	Close() error
}

// Pipeline manages a collection of sinks and distributes audit events to them.
type Pipeline struct {
	logger *zap.Logger
	sinks  []Sink
}

// Emit sends an audit event to all sinks in the pipeline.
func (p *Pipeline) Emit(ctx context.Context, event *AuditEvent) {
	for _, sink := range p.sinks {
		if err := sink.Write(event); err != nil {
			p.logger.Error("failed to write audit event to sink", zap.Error(err))
		}
	}
}

// NewFileSink creates a new sink that writes events to a file.
func NewFileSink(path string) (Sink, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}
	return &fileSink{file: file}, nil
}

// NewStdoutSink creates a new sink that writes events to standard output.
func NewStdoutSink() Sink {
	return &stdoutSink{}
}

// NewPipeline creates a new audit pipeline with the given sinks.
func NewPipeline(logger *zap.Logger, sinks ...Sink) *Pipeline {
	return &Pipeline{
		logger: logger,
		sinks:  sinks,
	}
}

// fileSink implements the Sink interface for writing to a file.
type fileSink struct {
	file *os.File
}

// Write serializes the event to JSON and writes it to the file.
func (s *fileSink) Write(event *AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}
	if _, err := s.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to audit log file: %w", err)
	}
	return nil
}

// Close closes the file handle.
func (s *fileSink) Close() error {
	return s.file.Close()
}

// stdoutSink implements the Sink interface for writing to standard output.
type stdoutSink struct{}

// Write serializes the event to JSON and writes it to standard output.
func (s *stdoutSink) Write(event *AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// Close is a no-op for the stdout sink.
func (s *stdoutSink) Close() error {
	return nil
}
