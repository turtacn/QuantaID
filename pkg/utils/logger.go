package utils

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

// Logger defines a standard logging interface for the QuantaID project.
// It abstracts the underlying logging library (Zap) and provides methods
// for different log levels. It also supports structured, contextual logging
// and automatically includes OpenTelemetry trace IDs if available in the context.
type Logger interface {
	// Debug logs a message at the debug level.
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	// Info logs a message at the info level.
	Info(ctx context.Context, msg string, fields ...zap.Field)
	// Warn logs a message at the warning level.
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	// Error logs a message at the error level.
	Error(ctx context.Context, msg string, fields ...zap.Field)
	// With returns a new logger with the given fields added to its context.
	With(fields ...zap.Field) Logger
}

// zapLogger is the Zap implementation of our standard Logger interface.
type ZapLogger struct {
	Logger *zap.Logger
}

// NewZapLogger creates a new logger instance based on the provided configuration.
// It can be configured to output to the console, a file (with rotation), or both,
// and supports both JSON and console-friendly log formats.
//
// Parameters:
//   - config: The configuration for the logger.
//
// Returns:
//   A configured Logger instance, or an error if the configuration is invalid.
func NewZapLogger(config *LoggerConfig) (Logger, error) {
	var core zapcore.Core
	encoder := getEncoder(config.Format)

	var writers []zapcore.WriteSyncer
	if config.File.Enabled {
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.File.Path,
			MaxSize:    config.File.MaxSizeMB,
			MaxBackups: config.File.MaxBackups,
			MaxAge:     config.File.MaxAgeDays,
			Compress:   config.File.Compress,
		})
		writers = append(writers, fileWriter)
	}
	if config.Console.Enabled {
		consoleWriter := zapcore.AddSync(os.Stdout)
		writers = append(writers, consoleWriter)
	}
	combinedWriter := zapcore.NewMultiWriteSyncer(writers...)

	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return nil, err
	}

	core = zapcore.NewCore(encoder, combinedWriter, level)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &ZapLogger{Logger: logger}, nil
}

func getEncoder(format string) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.TimeKey = "timestamp"

	if format == "json" {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func addTraceID(ctx context.Context, fields []zap.Field) []zap.Field {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		fields = append(fields, zap.String("traceID", span.SpanContext().TraceID().String()))
		fields = append(fields, zap.String("spanID", span.SpanContext().SpanID().String()))
	}
	return fields
}

// Debug logs a message at the debug level, including the trace ID from the context if available.
func (l *ZapLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, addTraceID(ctx, fields)...)
}

// Info logs a message at the info level, including the trace ID from the context if available.
func (l *ZapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Info(msg, addTraceID(ctx, fields)...)
}

// Warn logs a message at the warning level, including the trace ID from the context if available.
func (l *ZapLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, addTraceID(ctx, fields)...)
}

// Error logs a message at the error level, including the trace ID from the context if available.
func (l *ZapLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.Logger.Error(msg, addTraceID(ctx, fields)...)
}

// With returns a new logger instance with the specified fields added to its context,
// allowing for structured logging.
func (l *ZapLogger) With(fields ...zap.Field) Logger {
	return &ZapLogger{Logger: l.Logger.With(fields...)}
}

// LoggerConfig defines the settings for creating a logger.
type LoggerConfig struct {
	// Level is the minimum log level to output (e.g., "debug", "info", "warn", "error").
	Level string `yaml:"level"`
	// Format specifies the log output format ("json" or "console").
	Format string `yaml:"format"`
	// Console configures console logging.
	Console ConsoleConfig `yaml:"console"`
	// File configures file-based logging.
	File FileConfig `yaml:"file"`
}

// ConsoleConfig defines settings for console logging.
type ConsoleConfig struct {
	// Enabled determines if logs should be written to the console.
	Enabled bool `yaml:"enabled"`
}

// FileConfig defines settings for file-based logging, including log rotation.
type FileConfig struct {
	// Enabled determines if logs should be written to a file.
	Enabled bool `yaml:"enabled"`
	// Path is the file path for the log file.
	Path string `yaml:"path"`
	// MaxSizeMB is the maximum size in megabytes of the log file before it gets rotated.
	MaxSizeMB int `yaml:"maxSizeMB"`
	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int `yaml:"maxBackups"`
	// MaxAgeDays is the maximum number of days to retain old log files.
	MaxAgeDays int `yaml:"maxAgeDays"`
	// Compress determines if the rotated log files should be compressed.
	Compress bool `yaml:"compress"`
}

// NewNoopLogger creates a logger that discards all logs.
func NewNoopLogger() Logger {
	return &ZapLogger{Logger: zap.NewNop()}
}

// NewZapLoggerWrapper wraps a zap.Logger in a utils.Logger.
func NewZapLoggerWrapper(logger *zap.Logger) Logger {
	return &ZapLogger{Logger: logger}
}
