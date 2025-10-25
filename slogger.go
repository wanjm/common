//go:build slogger
// +build slogger

package common

import (
	"context"
	"log/slog"
	"os"
)

var selfLogger *slog.Logger

type LogField = any

func InitLogger() *slog.Logger {
	if selfLogger == nil {
		selfLogger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}
	return selfLogger
}

func addCommonFields(context context.Context, fields []any) []any {
	var trace_id = context.Value(TraceIdNameInContext)
	if trace_id != nil {
		fields = append(fields, String(TraceId, trace_id.(string)))
	}

	// , String(HTTPURL, context.Value(HttpUrl{}).(string)))
	var url = context.Value(HttpUrl{})
	if url != nil {
		fields = append(fields, String(HTTPURL, url.(string)))
	}
	return fields
}
func Info(context context.Context, msg string, fields ...any) {
	// 程序启动过程中，如gorm链接数据库失败，此时没有trace_id；
	// 其他情况下，如gorm可能忘记了withContext，也会出现没有trace_id的情况，但这是不可以接受的，应该修复，此处这么写，只是为了不影响程序运行
	fields = addCommonFields(context, fields)
	selfLogger.Info(msg, fields...)
}

func Warn(context context.Context, msg string, fields ...any) {
	// 程序启动过程中，如gorm链接数据库失败，此时没有trace_id；
	// 其他情况下，如gorm可能忘记了withContext，也会出现没有trace_id的情况，但这是不可以接受的，应该修复，此处这么写，只是为了不影响程序运行
	fields = addCommonFields(context, fields)
	selfLogger.Warn(msg, fields...)
}

func Error(context context.Context, msg string, fields ...any) {
	// 程序启动过程中，如gorm链接数据库失败，此时没有trace_id；
	// 其他情况下，如gorm可能忘记了withContext，也会出现没有trace_id的情况，但这是不可以接受的，应该修复，此处这么写，只是为了不影响程序运行
	fields = addCommonFields(context, fields)
	selfLogger.Error(msg, fields...)
}

func Debug(context context.Context, msg string, fields ...any) {
	// 程序启动过程中，如gorm链接数据库失败，此时没有trace_id；
	// 其他情况下，如gorm可能忘记了withContext，也会出现没有trace_id的情况，但这是不可以接受的，应该修复，此处这么写，只是为了不影响程序运行
	fields = addCommonFields(context, fields)
	selfLogger.Debug(msg, fields...)
}

// warpper for slog.String，主要是屏蔽zap包
// 而且这样写，代码编译时会inline，并不影响性能
func String(key, val string) LogField {
	return slog.String(key, val)
}

func Int64(key string, val int64) LogField {
	return slog.Int64(key, val)
}

func Int(key string, val int) LogField {
	return slog.Int(key, val)
}
func Err(err error) LogField {
	return slog.String("error", err.Error())
}

func Float64(key string, val float64) LogField {
	return slog.Float64(key, val)
}

// 本函数存在的目的是让外界的模块不需要看见LogField的定义,如代码可以写为：
// var fiedls = InitFields( String("key", "value"), Int("key2",10));
func InitFields(fileds ...any) []any {
	return fileds
}
