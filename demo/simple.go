package main

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var url = "amqp://MDphbXFwLWNuLTA5azF3NnZjaDAwOTpMVEFJNEc0RFFMeGVZanBOMkZQbVlGd2M=:RUY4OTNERTc0NDYwNzkxN0EyNjYwOTVFQjU1Q0E5QUY1OUM1RURDMToxNzMzNzQyNjA2MTQ3@amqp-cn-09k1w6vch009.mq-amqp.cn-hangzhou-249959-a-internal.aliyuncs.com:5672/dev?heartbeat=0"

func custom(key string) {
	fmt.Printf("connect to %s\n", url)
	conn, err := amqp.Dial(url)
	if err != nil {
		fmt.Printf("connect to %s failed: %s\n", url, err)
	}
	fmt.Printf("create channel\n")
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("create channel failed: %s\n", err)
	}
	delivery, err := ch.Consume(key, "wanjm", false, false, false, false, nil)
	if err != nil {
		fmt.Printf("consume failed: %s\n", err)
	}
	fmt.Printf("wait for consume\n")
	for {
		d, ok := <-delivery
		if ok {
			log.Printf("Received a message from %s : %s", key, d.Body)
			d.Ack(true)
		} else {
			log.Printf("channel closed,go reto reconnect\n")

		}
	}
}

func producer() {
	fmt.Printf("connect to %s\n", url)
	conn, err := amqp.Dial(url)
	if err != nil {
		fmt.Printf("connect to %s failed: %s\n", url, err)
	}
	fmt.Printf("create channel\n")
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("create channel failed: %s\n", err)
	}
	ch.Confirm(false)
	var confirm chan amqp.Confirmation = make(chan amqp.Confirmation, 1)
	ch.NotifyPublish(confirm)
	var i = 0
	for {
		message := fmt.Sprintf("hello world  %d", i)
		i++
		fmt.Printf("publish %s\n", message)
		err := ch.Publish("amq.direct", "studykey", false, false, amqp.Publishing{
			ContentType: "text/plain",
			MessageId:   fmt.Sprintf("%d", i),
			Body:        []byte(message),
		})
		if err != nil {
			fmt.Printf("publish failed: %s\n", err)
		}
		fmt.Printf("wait for confirm\n")
		info := <-confirm
		if info.Ack {
			fmt.Printf("publish %s success\n", message)
		} else {
			fmt.Printf("publish %s failed\n", message)
		}

		time.Sleep(time.Second)
	}
}
