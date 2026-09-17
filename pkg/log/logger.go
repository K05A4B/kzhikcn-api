package log

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog *zerolog.Logger
}

func NewLogger(logger *zerolog.Logger) *Logger {
	return &Logger{zerolog: logger}
}

func (l *Logger) WithTraceID(traceID string) *Logger {
	log := l.zerolog.With().Str("traceID", traceID).Logger()
	return NewLogger(&log)
}

func (l *Logger) With(key string, val any) *Logger {
	log := l.zerolog.With().Any(key, val).Logger()
	return NewLogger(&log)
}

func (l *Logger) ZeroLogger() *zerolog.Logger {
	return l.zerolog
}

func (l *Logger) Debug(v ...any) {
	l.zerolog.Debug().Msg(fmt.Sprint(v...))
}

func (l *Logger) Debugf(format string, v ...any) {
	l.zerolog.Debug().Msgf(format, v...)
}

func (l *Logger) Info(v ...any) {
	l.zerolog.Info().Msg(fmt.Sprint(v...))
}

func (l *Logger) Infof(format string, v ...any) {
	l.zerolog.Info().Msgf(format, v...)
}

func (l *Logger) Warn(v ...any) {
	l.zerolog.Warn().Msg(fmt.Sprint(v...))
}

func (l *Logger) Warnf(format string, v ...any) {
	l.zerolog.Warn().Msgf(format, v...)
}

func (l *Logger) Error(v ...any) {
	l.log(zerolog.ErrorLevel, 3, fmt.Sprint(v...))
}

func (l *Logger) Errorf(format string, v ...any) {
	l.logf(zerolog.ErrorLevel, 3, format, v...)
}

func (l *Logger) Fatal(v ...any) {
	l.log(zerolog.FatalLevel, 3, fmt.Sprint(v...))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.logf(zerolog.FatalLevel, 3, format, v...)
	os.Exit(1)
}

// log 在错误及以上级别附加 caller 字段。
// skip 需与调用链匹配：caller -> log -> 调用方 -> 用户代码。
func (l *Logger) log(level zerolog.Level, skip int, msg string) {
	ev := l.zerolog.WithLevel(level)
	if ev != nil && level >= zerolog.ErrorLevel {
		ev = ev.Str("caller", caller(skip))
	}
	ev.Msg(msg)
}

func (l *Logger) logf(level zerolog.Level, skip int, format string, v ...any) {
	ev := l.zerolog.WithLevel(level)
	if ev != nil && level >= zerolog.ErrorLevel {
		ev = ev.Str("caller", caller(skip))
	}
	ev.Msgf(format, v...)
}
