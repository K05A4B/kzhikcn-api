package log

import (
	"fmt"

	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog *zerolog.Logger
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
	l.zerolog.Error().Msg(fmt.Sprint(v...))
}

func (l *Logger) Errorf(format string, v ...any) {
	l.zerolog.Error().Msgf(format, v...)
}

func (l *Logger) Fatal(v ...any) {
	l.zerolog.Fatal().Msg(fmt.Sprint(v...))
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.zerolog.Fatal().Msgf(format, v...)
}

func NewLogger(logger *zerolog.Logger) *Logger {
	return &Logger{zerolog: logger}
}
