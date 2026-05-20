package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// Interface -.
type Interface interface {
	Debug(ctx context.Context, message string, args ...interface{})
	Info(ctx context.Context, message string, args ...interface{})
	Warn(ctx context.Context, message string, args ...interface{})
	Error(ctx context.Context, message string, args ...interface{})
	Fatal(ctx context.Context, message string, args ...interface{})
}

// Logger -.
type Logger struct {
	logger *slog.Logger
}

// New -.
func NewLogger(sl *slog.Logger) Interface {
	return &Logger{
		logger: sl,
	}
}

// Debug -.
func (l *Logger) Debug(ctx context.Context, message string, args ...interface{}) {
	l.logger.Log(ctx, slog.LevelDebug, fmt.Sprintf(message, args...))
}

// Info -.
func (l *Logger) Info(ctx context.Context, message string, args ...interface{}) {
	l.logger.Log(ctx, slog.LevelInfo, fmt.Sprintf(message, args...))
}

// Warn -.
func (l *Logger) Warn(ctx context.Context, message string, args ...interface{}) {
	l.logger.Log(ctx, slog.LevelWarn, fmt.Sprintf(message, args...))
}

// Error -.
func (l *Logger) Error(ctx context.Context, message string, args ...interface{}) {
	l.logger.Log(ctx, slog.LevelError, fmt.Sprintf(message, args...))
}

// Fatal -.
func (l *Logger) Fatal(ctx context.Context, message string, args ...interface{}) {
	l.logger.Log(ctx, slog.LevelError, fmt.Sprintf(message, args...))

	os.Exit(1)
}
