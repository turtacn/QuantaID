package utils

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

// Logger defines the standard logging interface for the QuantaID project.
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
}

// zapLogger is the Zap implementation of our Logger interface.
type zapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new logger instance.
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

	return &zapLogger{logger: logger}, nil
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
		return append(fields, zap.String("traceID", span.SpanContext().TraceID().String()))
	}
	return fields
}

func (l *zapLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Debug(msg, addTraceID(ctx, fields)...)
}

func (l *zapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Info(msg, addTraceID(ctx, fields)...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Warn(msg, addTraceID(ctx, fields)...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Error(msg, addTraceID(ctx, fields)...)
}

func (l *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}

type LoggerConfig struct {
	Level   string       `yaml:"level"`
	Format  string       `yaml:"format"`
	Console ConsoleConfig `yaml:"console"`
	File    FileConfig   `yaml:"file"`
}

type ConsoleConfig struct {
	Enabled bool `yaml:"enabled"`
}

type FileConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Path       string `yaml:"path"`
	MaxSizeMB  int    `yaml:"maxSizeMB"`
	MaxBackups int    `yaml:"maxBackups"`
	MaxAgeDays int    `yaml:"maxAgeDays"`
	Compress   bool   `yaml:"compress"`
}

//Personal.AI order the ending
