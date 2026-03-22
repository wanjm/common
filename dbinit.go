package common

import (
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type DbConfig struct {
	MySqlConfig MySqlConfig
	MongoConfig MongoConfig
}

var Cfg *DbConfig

// @gos type="initiator"
func ConnectMysqlDB() (mysqlDB *gorm.DB) {
	return ConnectGorm(&Cfg.MySqlConfig)
}

// @gos type="initiator"
func GetMongoClient() (mongoDB *mongo.Client) {
	return ConnectMongo(&Cfg.MongoConfig)
}

// @gos type="initiator"
func ConnectMongoDB(mongo *mongo.Client) (mongoDb *mongo.Database) {
	db := mongo.Database(Cfg.MongoConfig.Database)
	return db
}
