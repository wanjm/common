package common

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
}

// type options struct {
// 	Delimiter string
// }
// type Option func(*options)

func ConnectRedis(cfg *RedisConfig) *redis.Client {
	cli := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx := context.Background()
	if err := cli.Ping(ctx).Err(); err != nil {
		panic("redis ping error " + err.Error())
	}
	return cli
}
