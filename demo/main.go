package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wanjm/common"
)

func main() {
	// 演示了从studykey向amq.direct发送消息,然后转发到mqstudy队列的过程
	var p = flag.Bool("p", false, "producer")
	var s = flag.Bool("s", false, "simple mode")
	flag.Parse()
	common.InitLogger()
	if *s {
		if *p {
			fmt.Printf("producer\n")
			producer()
		} else {
			fmt.Printf("custom 2\n")
			go custom("commissionQueue")
			// custom("mqstudy2")
		}
	} else {
		var a = common.RabbitMqConfig{
			ConsumerTag:     "tag",
			Addr:            "XXX:5672",
			Exchange:        "amq.direct",
			ExchangeType:    "direct",
			ResourceOwnerId: "XXX",
			AccessKeyId:     "XXX",
			AccessKeySecret: "XXX",
			Vhost:           "dev",
			QueueName:       "commissionQueue",
			RoutingKey:      "commissionRouteKey",
		}
		if *p {
			queue := common.ConnectProducer(&a)
			fmt.Printf("producer\n")
			produce(queue)
		} else {
			queue := common.ConnectConsumer(&a)
			fmt.Printf("custom use our library\n")
			consume(queue)
		}

	}
}
func consume(queue *common.RabbitMqClient) {
	queue.Consume(context.Background(), "test",
		func(ctx context.Context, a amqp.Delivery) {

		},
	)
}

func produce(queue *common.RabbitMqClient) {
	message := []byte(`{"type":1,"info":{"orderSn":"b98cbacdb04219637100e7cb36e10d44"}}`)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
	defer cancel()
loop:
	for {
		select {
		// Attempt to push a message every 2 seconds
		case <-time.After(time.Second * 2):
			if err := queue.Push(message); err != nil {
				log.Printf("Push failed: %s\n", err)
			} else {
				log.Println("Push succeeded!")
			}
		case <-ctx.Done():
			if err := queue.Close(); err != nil {
				log.Printf("Close failed: %s\n", err)
			}
			break loop
		}
	}
}
