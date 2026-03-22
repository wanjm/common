package common

import (
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type DbConfig struct {
	MySqlConfig    MySqlConfig
	MongoConfig    MongoConfig
	RedisConfig    RedisConfig
	RabbitMqConfig RabbitMqConfig
}

var Cfg *DbConfig

// @gos type="initiator"
func ConnectMysqlDB() (mysqlDB *gorm.DB) {
	return ConnectGorm(&Cfg.MySqlConfig)
}

// @gos type="initiator"
func GetMongoClient() (mongo *mongo.Client) {
	return ConnectMongo(&Cfg.MongoConfig)
}

// @gos type="initiator"
func ConnectMongoDB(mongo *mongo.Client) (mongoDB *mongo.Database) {
	db := mongo.Database(Cfg.MongoConfig.Database)
	return db
}

// @gos type="initiator"
func ConnectRedisClient() (redis *redis.Client) {
	return ConnectRedis(&Cfg.RedisConfig)
}

// @gos type="initiator"
func ConnectRabbitConsumer() (normalConsumer *RabbitMqClient) {
	return ConnectConsumer(&Cfg.RabbitMqConfig)
}

// @gos type="initiator"
func ConnectRabbitFanoutConsumer() (fanoutConsumer *RabbitMqClient) {
	return ConnectFanoutConsumer(&Cfg.RabbitMqConfig)
}
