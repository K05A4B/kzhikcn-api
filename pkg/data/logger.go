package data

import (
	"context"
	"fmt"
	"kzhikcn/pkg/log"
	"time"

	glogger "gorm.io/gorm/logger"
)

// gormLogger 将 GORM 的日志桥接到 pkg/log。
// 不缓存 logger 实例，始终通过 log.GetLogger() 获取当前 logger，以便级别热更新生效。
type gormLogger struct {
	level glogger.LogLevel
}

func (g *gormLogger) LogMode(l glogger.LogLevel) glogger.Interface {
	g.level = l
	return g
}

func (g *gormLogger) Info(ctx context.Context, format string, v ...any) {
	if g.level < glogger.Info {
		return
	}

	log.GetLogger().With("type", "database_log").Infof(format, v...)
}

func (g *gormLogger) Warn(ctx context.Context, format string, v ...any) {
	if g.level < glogger.Warn {
		return
	}

	log.GetLogger().With("type", "database_log").Warnf(format, v...)
}

func (g *gormLogger) Error(ctx context.Context, format string, v ...any) {
	if g.level < glogger.Error {
		return
	}

	log.GetLogger().With("type", "database_log").Errorf(format, v...)
}

func (g *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if g.level <= glogger.Silent {
		return
	}

	sqlText, row := fc()
	l := log.GetLogger().
		With("type", "database_log_trace").
		With("row", row)

	if err != nil {
		if g.level >= glogger.Error {
			l.With("sql", sqlText).Error(err)
		}
		return
	}

	if g.level >= glogger.Info {
		l.Debug("execute sql: ", fmt.Sprintf("\033[4m%s\033[0m", sqlText))
	}
}
