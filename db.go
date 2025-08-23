package common

import (
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
			LogLevel:      level,
			SlowThreshold: time.Duration(cfg.SlowThreshold) * time.Millisecond,
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
