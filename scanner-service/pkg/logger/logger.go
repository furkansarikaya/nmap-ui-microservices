package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a wrapper around zap logger
type Logger struct {
	*zap.Logger
}

// Config contains logger configuration
type Config struct {
	Level  string
	Format string
	Output string
}

// NewLogger creates a new Logger instance
func NewLogger(config Config) (*Logger, error) {
	level := getLogLevel(config.Level)

	// Configure encoder based on format
	var encoder zapcore.Encoder
	encConfig := zap.NewProductionEncoderConfig()
	encConfig.TimeKey = "timestamp"
	encConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if config.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encConfig)
	}

	// Configure output
	var output zapcore.WriteSyncer
	if config.Output == "stdout" || config.Output == "" {
		output = zapcore.AddSync(os.Stdout)
	} else {
		file, err := os.OpenFile(config.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		output = zapcore.AddSync(file)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		output,
		level,
	)

	// Create logger
	zapLogger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return &Logger{
		Logger: zapLogger,
	}, nil
}

// getLogLevel converts string level to zapcore.Level
func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// With adds structured context to the Logger
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
	}
}

// Named adds a sub-logger with the specified name
func (l *Logger) Named(name string) *Logger {
	return &Logger{
		Logger: l.Logger.Named(name),
	}
}

// Info logs a message at InfoLevel
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Debug logs a message at DebugLevel
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Warn logs a message at WarnLevel
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// Fatal logs a message at FatalLevel
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}
