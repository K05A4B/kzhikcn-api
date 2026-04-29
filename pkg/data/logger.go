package data

import (
	"context"
	"fmt"
	"kzhikcn/pkg/log"
	"time"

	glogger "gorm.io/gorm/logger"
)

type gormLogger struct {
	logger log.Logger
	level  glogger.LogLevel
}

func (g *gormLogger) LogMode(l glogger.LogLevel) glogger.Interface {
	g.level = l
	return g
}

func (g *gormLogger) Info(ctx context.Context, format string, v ...any) {
	if g.level < glogger.Info {
		return
	}

	g.logger.With("type", "database_log").Infof(format, v...)
}

func (g *gormLogger) Warn(ctx context.Context, format string, v ...any) {
	if g.level < glogger.Warn {
		return
	}

	g.logger.With("type", "database_log").Warnf(format, v...)
}

func (g *gormLogger) Error(ctx context.Context, format string, v ...any) {
	if g.level < glogger.Info {
		return
	}

	g.logger.With("type", "database_log").Errorf(format, v...)
}

func (g *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sqlText, row := fc()
	l := g.logger.With("type", "database_log_trace").
		With("row", row)

	if err != nil {
		l.With("sql", sqlText).Error(err)
	} else {
		l.Debug("execute sql: ", fmt.Sprintf("\033[4m%s\033[0m", sqlText))
	}
}
