package logger

import (
	"context"
	"fmt"
	"time"
)

// Level representa o nível de log
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

// Logger define a interface para logging
type Logger interface {
	Debug(ctx context.Context, msg string, args ...interface{})
	Info(ctx context.Context, msg string, args ...interface{})
	Warn(ctx context.Context, msg string, args ...interface{})
	Error(ctx context.Context, msg string, args ...interface{})
}

// DefaultLogger é a implementação padrão do Logger
type DefaultLogger struct {
	level Level
}

// NewDefaultLogger cria uma nova instância do DefaultLogger
func NewDefaultLogger(level Level) *DefaultLogger {
	return &DefaultLogger{level: level}
}

func (l *DefaultLogger) log(level Level, ctx context.Context, msg string, args ...interface{}) {
	if level < l.level {
		return
	}
	
	levelStr := "DEBUG"
	switch level {
	case INFO:
		levelStr = "INFO"
	case WARN:
		levelStr = "WARN"
	case ERROR:
		levelStr = "ERROR"
	}
	
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(msg, args...)
	
	fmt.Printf("[%s] %s: %s\n", timestamp, levelStr, message)
}

func (l *DefaultLogger) Debug(ctx context.Context, msg string, args ...interface{}) {
	l.log(DEBUG, ctx, msg, args...)
}

func (l *DefaultLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.log(INFO, ctx, msg, args...)
}

func (l *DefaultLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.log(WARN, ctx, msg, args...)
}

func (l *DefaultLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.log(ERROR, ctx, msg, args...)
} 