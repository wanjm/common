//go:build zaplogger
// +build zaplogger

package common

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

var selfLogger *zap.Logger

func InitLogger() {
	if selfLogger != nil {
		return
	}
	var err error
	selfLogger, err = zap.NewProduction()
	if err != nil {
		fmt.Printf("init logger failed: %v\n", err)
	}
}

func Info(context context.Context, msg string, fields ...zap.Field) {
	// 程序启动过程中，如gorm链接数据库失败，此时没有trace_id；
	// 其他情况下，如gorm可能忘记了withContext，也会出现没有trace_id的情况，但这是不可以接受的，应该修复，此处这么写，只是为了不影响程序运行
	var trace_id = context.Value(TraceId{})
	if trace_id != nil {
		fields = append(fields, String(TRACEID, trace_id.(string)), String(HTTPURL, context.Value(HttpUrl{}).(string)))
	}
	selfLogger.Info(msg, fields...)
}

func Warn(context context.Context, msg string, fields ...zap.Field) {
	// 程序启动过程中，如gorm链接数据库失败，此时没有trace_id；
	// 其他情况下，如gorm可能忘记了withContext，也会出现没有trace_id的情况，但这是不可以接受的，应该修复，此处这么写，只是为了不影响程序运行
	var trace_id = context.Value(TraceId{})
	if trace_id != nil {
		fields = append(fields, String(TRACEID, trace_id.(string)), String(HTTPURL, context.Value(HttpUrl{}).(string)))
	}
	selfLogger.Warn(msg, fields...)
}

func Error(context context.Context, msg string, fields ...zap.Field) {
	// 程序启动过程中，如gorm链接数据库失败，此时没有trace_id；
	// 其他情况下，如gorm可能忘记了withContext，也会出现没有trace_id的情况，但这是不可以接受的，应该修复，此处这么写，只是为了不影响程序运行
	var trace_id = context.Value(TraceId{})
	if trace_id != nil {
		fields = append(fields, String(TRACEID, trace_id.(string)), String(HTTPURL, context.Value(HttpUrl{}).(string)))
	}
	selfLogger.Error(msg, fields...)
}

func Debug(context context.Context, msg string, fields ...zap.Field) {
	// 程序启动过程中，如gorm链接数据库失败，此时没有trace_id；
	// 其他情况下，如gorm可能忘记了withContext，也会出现没有trace_id的情况，但这是不可以接受的，应该修复，此处这么写，只是为了不影响程序运行
	var trace_id = context.Value(TraceId{})
	if trace_id != nil {
		fields = append(fields, String(TRACEID, trace_id.(string)), String(HTTPURL, context.Value(HttpUrl{}).(string)))
	}
	selfLogger.Debug(msg, fields...)
}

// warpper for zap.String，主要是屏蔽zap包
// 而且这样写，代码编译时会inline，并不影响性能
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

// 本函数存在的目的是让外界的模块不需要看见zap.Field的定义,如代码可以写为：
// var fiedls = InitFields( String("key", "value"), Int("key2",10));
func InitFields(fileds ...zap.Field) []zap.Field {
	return fileds
}
