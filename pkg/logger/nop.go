package logger

import "context"

type nopLogger struct{}

func (n nopLogger) Debug(ctx context.Context, message string, args ...interface{}) {
}

func (n nopLogger) Info(ctx context.Context, message string, args ...interface{}) {
}

func (n nopLogger) Warn(ctx context.Context, message string, args ...interface{}) {
}

func (n nopLogger) Error(ctx context.Context, message string, args ...interface{}) {
}

func (n nopLogger) Fatal(ctx context.Context, message string, args ...interface{}) {
}

func NewNopLogger() Interface {
	return &nopLogger{}
}
