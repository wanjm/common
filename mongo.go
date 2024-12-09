package common

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfig struct {
	Debug            bool
	Uri              string
	Database         string
	MaxPoolSize      uint64
	MaxConnecting    uint64
	MaxConnIdleTime  int
	CollectionPrefix string
}

func ConnectMongo(cfg *MongoConfig) *mongo.Client {
	// fmt.Printf("ConnectMongo '%s'\n", cfg.Uri)
	optionsClient := options.Client().ApplyURI(cfg.Uri).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMaxConnecting(cfg.MaxConnecting).
		SetMaxConnIdleTime(time.Duration(cfg.MaxConnIdleTime) * time.Second)

	// if cfg.Debug {
	// 	sink := zapr.NewLogger(cfg.Logger).GetSink()
	// 	//sink := &CustomLogger{}
	// 	// Create a client with our logger options.
	// 	loggerOptions := options.
	// 		Logger().
	// 		SetSink(sink).
	// 		SetMaxDocumentLength(25).
	// 		SetComponentLevel(options.LogComponentCommand, options.LogLevelInfo)
	// 	optionsClient.SetLoggerOptions(loggerOptions).
	// 		SetMonitor(withMonitorOption())
	// }
	ctx := context.Background()
	client, err := mongo.Connect(ctx, optionsClient)
	if err != nil {
		panic(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		panic(err)
	}
	return client

}

// func withMonitorOption() *event.CommandMonitor {
// 	// log monitor
// 	return &event.CommandMonitor{
// 		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
// 			Info(ctx, fmt.Sprintf("Mongo %s Started", startedEvent.CommandName),
// 				Int64("req_id", startedEvent.RequestID),
// 				String("sql", startedEvent.Command.String()))
// 		},
// 		Succeeded: func(ctx context.Context, succeededEvent *event.CommandSucceededEvent) {
// 			Info(ctx, fmt.Sprintf("[%dms] Mongo %s Succeeded", succeededEvent.Duration.Milliseconds(), succeededEvent.CommandName),
// 				Int64("req_id", succeededEvent.RequestID),
// 				String("duration", fmt.Sprintf("%d", succeededEvent.Duration.Milliseconds())))
// 		},
// 		Failed: func(ctx context.Context, failedEvent *event.CommandFailedEvent) {
// 			Info(ctx, fmt.Sprintf("[%dms] Mongo %s Failed", failedEvent.Duration.Milliseconds(), failedEvent.CommandName),
// 				Int64("req_id", failedEvent.RequestID),
// 				String("duration", fmt.Sprintf("%d", failedEvent.Duration.Milliseconds())))
// 		},
// 	}
// }
