package common

import (
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

type MySqlConfig struct {
	Debug         bool
	DSN           string   //主库
	Replicas      []string //从库
	MaxLifetime   int
	MaxIdleTime   int
	MaxOpenConns  int
	MaxIdleConns  int
	TablePrefix   string
	SingularTable bool
	SlowThreshold int // milliseconds

	// Resolver     []ResolverConfig
}

func ConnectGorm(cfg *MySqlConfig) (gormdb *gorm.DB) {
	var level logger.LogLevel
	if cfg.Debug {
		level = logger.Info
	} else {
		level = logger.Warn
	}
	gormdb, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: NewGormLogger(logger.Config{
			LogLevel:                  level,
			SlowThreshold:             time.Duration(cfg.SlowThreshold) * time.Millisecond,
			IgnoreRecordNotFoundError: true,
		}),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   cfg.TablePrefix,
			SingularTable: cfg.SingularTable,
		},
	})
	if err != nil {
		panic(err)
	}
	var replicas = make([]gorm.Dialector, len(cfg.Replicas))
	for i, replica := range cfg.Replicas {
		replicas[i] = mysql.Open(replica)
	}
	gormdb.Use(dbresolver.Register(dbresolver.Config{
		Replicas: replicas,
	}).
		SetConnMaxIdleTime(time.Duration(cfg.MaxIdleTime) * time.Second).
		SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second).
		SetMaxIdleConns(cfg.MaxIdleConns).
		SetMaxOpenConns(cfg.MaxOpenConns),
	)
	return gormdb
}

// 这两个函数会被编译器内连，如果传入是const，会直接替换为字符串；
// 不影响性能；
func C2(a, b string) string {
	return a + "." + b
}

// 这三个函数会被编译器内连，如果传入是const，会直接替换为字符串；
// 不影响性能；
func C3(a, b, c string) string {
	return a + "." + b + "." + c
}
func C4(a, b, c, d string) string {
	return a + "." + b + "." + c + "." + d
}

// 这个函数会被编译器内连，如果传入是const, 也不会优化，所以尽量不要使用；
func Cn(a ...string) string {
	return strings.Join(a, ".")
}

// return max(xx)
func CMax(a string) string {
	return "max(" + a + ")"
}

// return min(xx)
func CMin(a string) string {
	return "min(" + a + ")"
}

// return sum(xx)
func CSum(a string) string {
	return "sum(" + a + ")"
}

// return avg(xx)
func CAvg(a string) string {
	return "avg(" + a + ")"
}
