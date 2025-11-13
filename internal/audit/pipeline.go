package audit

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"os"
	"sync"
)

// Sink is the interface for audit event destinations.
type Sink interface {
	Write(ctx context.Context, event *AuditEvent) error
	Close() error
}

// Pipeline fans out audit events to multiple sinks.
type Pipeline struct {
	sinks  []Sink
	logger *zap.Logger
}

// NewPipeline creates a new audit pipeline.
func NewPipeline(logger *zap.Logger, sinks ...Sink) *Pipeline {
	return &Pipeline{
		sinks:  sinks,
		logger: logger,
	}
}

// Emit sends an audit event to all sinks in the pipeline.
// If a sink returns an error, it is logged, but the pipeline continues to other sinks.
func (p *Pipeline) Emit(ctx context.Context, event *AuditEvent) {
	for _, s := range p.sinks {
		if err := s.Write(ctx, event); err != nil {
			p.logger.Error("Failed to write audit event to sink",
				zap.Error(err),
				zap.Any("event", event),
			)
		}
	}
}

// Close closes all sinks in the pipeline.
func (p *Pipeline) Close() {
	for _, s := range p.sinks {
		if err := s.Close(); err != nil {
			p.logger.Error("Failed to close audit sink", zap.Error(err))
		}
	}
}

// --- FileSink ---

// FileSink writes audit events as JSON lines to a file.
type FileSink struct {
	mu   sync.Mutex
	file *os.File
}

// NewFileSink creates a new FileSink.
// The file is opened in append mode. If it doesn't exist, it's created.
func NewFileSink(path string) (*FileSink, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileSink{file: file}, nil
}

// Write marshals the event to JSON and writes it to the file, followed by a newline.
func (s *FileSink) Write(ctx context.Context, event *AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = s.file.Write(append(data, '\n'))
	return err
}

// Close closes the underlying file.
func (s *FileSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.file.Close()
}

// --- StdoutSink ---

// StdoutSink writes audit events as JSON lines to standard output.
type StdoutSink struct {
	mu sync.Mutex
}

// NewStdoutSink creates a new StdoutSink.
func NewStdoutSink() *StdoutSink {
	return &StdoutSink{}
}

// Write marshals the event to JSON and writes it to standard output.
func (s *StdoutSink) Write(ctx context.Context, event *AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(append(data, '\n'))
	return err
}

// Close is a no-op for StdoutSink.
func (s *StdoutSink) Close() error {
	return nil
}
