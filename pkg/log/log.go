package log

import (
	"fmt"
	"io"
	"kzhikcn/pkg/config"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	zeroLogger     zerolog.Logger
	zeroLoggerOnce sync.Once
)

func GetLogger() *Logger {
	zeroLoggerOnce.Do(func() {
		conf := config.Conf()
		logConf := conf.Log

		var writer io.Writer

		consoleWriter := zerolog.ConsoleWriter{
			Out:          os.Stdout,
			TimeLocation: time.Local,
		}

		if logConf.Enable {
			lumberjack := logConf.Lumberjack
			writer = zerolog.MultiLevelWriter(consoleWriter, lumberjack)
		} else {
			writer = consoleWriter
		}

		logLevel, err := zerolog.ParseLevel(logConf.LogLevel)
		if err != nil {
			panic(err)
		}

		zeroLogger = zerolog.New(writer).With().Timestamp().Logger().Level(logLevel)
	})

	return NewLogger(&zeroLogger)
}

func GetZeroLogger() *zerolog.Logger {
	return &zeroLogger
}

func WithTraceID(traceID string) *Logger {
	return GetLogger().WithTraceID(traceID)
}

func With(key string, value any) *Logger {
	return GetLogger().With(key, value)
}

func Debug(v ...any) {
	GetLogger().Debug(v...)
}

func Debugf(format string, v ...any) {
	GetLogger().Debugf(format, v...)
}

func Info(v ...any) {
	GetLogger().Info(v...)
}

func Infof(format string, v ...any) {
	GetLogger().Infof(format, v...)
}

func Warn(v ...any) {
	GetLogger().Warn(v...)
}

func Warnf(format string, v ...any) {
	GetLogger().Warnf(format, v...)
}

func Error(v ...any) {
	_, file, line, _ := runtime.Caller(1)
	_, file = path.Split(file)

	GetLogger().
		With("caller", fmt.Sprintf("%s:%d", file, line)).
		Error(v...)
}

func Errorf(format string, v ...any) {
	_, file, line, _ := runtime.Caller(1)
	_, file = path.Split(file)

	GetLogger().
		With("caller", fmt.Sprintf("%s:%d", file, line)).
		Errorf(format, v...)
}

func Fatal(v ...any) {
	GetLogger().Fatal(v...)
}

func Fatalf(format string, v ...any) {
	GetLogger().Fatalf(format, v...)
}
