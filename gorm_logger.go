package common

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm/logger"
)

type GormLogger struct {
	logger.Config
}

func NewGormLogger(config logger.Config) *GormLogger {
	return &GormLogger{
		Config: config,
	}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l *GormLogger) Info(ctx context.Context, s string, i ...any) {
	message := getMessage(s, i)
	Info(ctx, message, String("gen", "gorm"))
}

func (l *GormLogger) Warn(ctx context.Context, s string, i ...any) {
	message := getMessage(s, i)
	Warn(ctx, message, String("gen", "gorm"))
}

func (l *GormLogger) Error(ctx context.Context, s string, i ...any) {
	message := getMessage(s, i)
	Error(ctx, message, String("gen", "gorm"))
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	var fields = InitFields(
		// String("lineNo", utils.FileWithLineNum()),
		Float64("elapsed", float64(elapsed.Nanoseconds())/1e6),
		String("gen", "gorm"),
		String("sql", sql),
		Int("rows", int(rows)),
	)
	switch {

	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		Error(ctx, err.Error(), fields...)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		Warn(ctx, "SLOW SQL", fields...)
	case l.LogLevel == logger.Info:
		Info(ctx, "debug sql", fields...)
	}
}

// getMessage format with Sprint, Sprintf, or neither. copy from sugar.go of zap;
func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}
