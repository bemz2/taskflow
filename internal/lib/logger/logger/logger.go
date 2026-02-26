package logger

import (
	"context"
	"log/slog"
)

type Logger interface {
	Info(msg string, keysAndValues ...any)
	InfoContext(ctx context.Context, msg string, keysAndValues ...any)

	Debug(msg string, keysAndValues ...any)
	DebugContext(ctx context.Context, msg string, keysAndValues ...any)

	Warn(msg string, keysAndValues ...any)
	WarnContext(ctx context.Context, msg string, keysAndValues ...any)

	Error(msg string, keysAndValues ...any)
	ErrorContext(ctx context.Context, msg string, keysAndValues ...any)
}

type SlogLogger struct {
	l *slog.Logger
}

func NewSlogLogger() Logger {
	base := slog.Default()
	handler := ContextHandler{Handler: base.Handler()}

	return &SlogLogger{
		l: slog.New(handler),
	}
}

func (s *SlogLogger) Info(msg string, keysAndValues ...any) {
	s.l.Info(msg, keysAndValues...)
}

func (s *SlogLogger) InfoContext(ctx context.Context, msg string, keysAndValues ...any) {
	s.l.InfoContext(ctx, msg, keysAndValues...)
}

func (s *SlogLogger) Debug(msg string, keysAndValues ...any) {
	s.l.Debug(msg, keysAndValues...)
}

func (s *SlogLogger) DebugContext(ctx context.Context, msg string, keysAndValues ...any) {
	s.l.DebugContext(ctx, msg, keysAndValues...)
}

func (s *SlogLogger) Warn(msg string, keysAndValues ...any) {
	s.l.Warn(msg, keysAndValues...)
}

func (s *SlogLogger) WarnContext(ctx context.Context, msg string, keysAndValues ...any) {
	s.l.WarnContext(ctx, msg, keysAndValues...)
}

func (s *SlogLogger) Error(msg string, keysAndValues ...any) {
	s.l.Error(msg, keysAndValues...)
}

func (s *SlogLogger) ErrorContext(ctx context.Context, msg string, keysAndValues ...any) {
	s.l.ErrorContext(ctx, msg, keysAndValues...)
}
