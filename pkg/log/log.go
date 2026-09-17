package log

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

var (
	current atomic.Pointer[zerolog.Logger]

	consoleWriter = zerolog.ConsoleWriter{
		Out:          os.Stdout,
		TimeLocation: time.Local,
	}
)

// init 安装安全默认 logger（标准输出 + info 级别）。
// 保证在配置加载前发生的任何日志调用都不会 panic。
func init() {
	l := zerolog.New(consoleWriter).With().Timestamp().Logger().Level(zerolog.InfoLevel)
	current.Store(&l)
}

// Configure 依据 opts 重建全局 logger，并以原子方式替换。
// 可重复调用（例如配置热重载）。
func Configure(opts Options) error {
	level, err := parseLevel(opts.Level)
	if err != nil {
		return err
	}

	writers := make([]io.Writer, 0, len(opts.Writers)+1)
	if opts.Console || len(opts.Writers) == 0 {
		writers = append(writers, consoleWriter)
	}
	for _, w := range opts.Writers {
		if w != nil {
			writers = append(writers, w)
		}
	}

	l := zerolog.New(zerolog.MultiLevelWriter(writers...)).
		With().Timestamp().Logger().Level(level)
	current.Store(&l)

	return nil
}

// SetLevel 仅更新当前 logger 的级别，保留已配置的输出目标。
func SetLevel(level string) error {
	lvl, err := parseLevel(level)
	if err != nil {
		return err
	}

	updated := current.Load().Level(lvl)
	current.Store(&updated)

	return nil
}

func parseLevel(level string) (zerolog.Level, error) {
	if level == "" {
		return zerolog.InfoLevel, nil
	}

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.InfoLevel, fmt.Errorf("非法的日志级别: %s", level)
	}

	return lvl, nil
}

// GetLogger 返回当前全局 logger 的包装。
func GetLogger() *Logger {
	return NewLogger(current.Load())
}

// GetZeroLogger 返回当前全局 zerolog logger。
func GetZeroLogger() *zerolog.Logger {
	return current.Load()
}

func caller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	return path.Base(file) + ":" + strconv.Itoa(line)
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
	GetLogger().log(zerolog.ErrorLevel, 3, fmt.Sprint(v...))
}

func Errorf(format string, v ...any) {
	GetLogger().logf(zerolog.ErrorLevel, 3, format, v...)
}

func Fatal(v ...any) {
	GetLogger().log(zerolog.FatalLevel, 3, fmt.Sprint(v...))
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	GetLogger().logf(zerolog.FatalLevel, 3, format, v...)
	os.Exit(1)
}
