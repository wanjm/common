package common

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMqClient is the base struct for handling connection recovery, consumption and
// publishing. Note that this struct has an internal mutex to safeguard against
// data races. As you develop and iterate over this example, you may need to add
// further locks, or safeguards, to keep your application safe from data races
type RabbitMqClient struct {
	m               *sync.Mutex
	queueName       string
	logger          *log.Logger
	connection      *amqp.Connection
	channel         *amqp.Channel
	done            chan bool
	notifyConnClose chan *amqp.Error
	notifyChanClose chan *amqp.Error
	notifyConfirm   chan amqp.Confirmation
	isReady         bool
}

const (
	reconnectDelay = 5 * time.Second

	reInitDelay = 2 * time.Second

	resendDelay = 5 * time.Second
)

var (
	errNotConnected  = errors.New("not connected to a server")
	errAlreadyClosed = errors.New("already closed: not connected to the server")
	errShutdown      = errors.New("client is shutting down")
)

// 以下代码添加了https://pkg.go.dev/github.com/rabbitmq/amqp091-go的example；
// 对Consume做了改造，自动检测网络联通；对于push的自动检测能力保持不变，看原来的代码挺好的；
// 后续添加ctx的配置，增加优雅退出，优雅退出时，首先停止接受消息，然后等待消息处理完毕，最后关闭连接，退出程序；

// New creates a new consumer state instance, and automatically
// attempts to connect to the server.
func New(queueName, addr string) *RabbitMqClient {
	client := RabbitMqClient{
		m:         &sync.Mutex{},
		logger:    log.New(os.Stdout, "", log.LstdFlags),
		queueName: queueName,
		done:      make(chan bool),
	}
	go client.handleReconnect(addr)
	return &client
}

func (client *RabbitMqClient) IsReady() bool {
	client.m.Lock()
	defer client.m.Unlock()
	return client.isReady
}

// handleReconnect will wait for a connection error on
// notifyConnClose, and then continuously attempt to reconnect.
func (client *RabbitMqClient) handleReconnect(addr string) {
	for {
		client.m.Lock()
		client.isReady = false
		client.m.Unlock()

		client.logger.Println("Attempting to connect " + addr)

		conn, err := client.connect(addr)

		if err != nil {
			client.logger.Println("Failed to connect. Retrying...")

			select {
			case <-client.done:
				return
			case <-time.After(reconnectDelay):
			}
			continue
		}

		if done := client.handleReInit(conn); done {
			break
		}
	}
}

// connect will create a new AMQP connection
func (client *RabbitMqClient) connect(addr string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(addr)

	if err != nil {
		return nil, err
	}

	client.changeConnection(conn)
	client.logger.Println("Connected!")
	return conn, nil
}

// handleReInit will wait for a channel error
// and then continuously attempt to re-initialize both channels
func (client *RabbitMqClient) handleReInit(conn *amqp.Connection) bool {
	for {
		client.m.Lock()
		client.isReady = false
		client.m.Unlock()

		err := client.init(conn)

		if err != nil {
			client.logger.Printf("Failed to initialize channel. Retrying... %v\n", err)

			select {
			case <-client.done:
				return true
			case <-client.notifyConnClose:
				client.logger.Println("Connection closed. Reconnecting...")
				return false
			case <-time.After(reInitDelay):
			}
			continue
		}

		select {
		case <-client.done:
			return true
		case <-client.notifyConnClose:
			client.logger.Println("Connection closed. Reconnecting...")
			return false
		case <-client.notifyChanClose:
			client.logger.Println("Channel closed. Re-running init...")
		}
	}
}

// init will initialize channel & declare queue
func (client *RabbitMqClient) init(conn *amqp.Connection) error {
	ch, err := conn.Channel()

	if err != nil {
		fmt.Printf("Failed to open a channel: %v\n", err)
		return err
	}

	err = ch.Confirm(false)

	if err != nil {
		fmt.Printf("Failed to confirm: %v\n", err)
		return err
	}

	_, err = ch.QueueDeclarePassive(
		client.queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		fmt.Printf("Failed to queueDeclare %s: %v\n", client.queueName, err)
		return err
	}

	client.changeChannel(ch)
	client.m.Lock()
	client.isReady = true
	client.m.Unlock()
	client.logger.Println("Setup!")

	return nil
}

// changeConnection takes a new connection to the queue,
// and updates the close listener to reflect this.
func (client *RabbitMqClient) changeConnection(connection *amqp.Connection) {
	client.connection = connection
	client.notifyConnClose = make(chan *amqp.Error, 1)
	client.connection.NotifyClose(client.notifyConnClose)
}

// changeChannel takes a new channel to the queue,
// and updates the channel listeners to reflect this.
func (client *RabbitMqClient) changeChannel(channel *amqp.Channel) {
	client.channel = channel
	client.notifyChanClose = make(chan *amqp.Error, 1)
	client.notifyConfirm = make(chan amqp.Confirmation, 1)
	client.channel.NotifyClose(client.notifyChanClose)
	client.channel.NotifyPublish(client.notifyConfirm)
}

// Push will push data onto the queue, and wait for a confirmation.
// This will block until the server sends a confirmation. Errors are
// only returned if the push action itself fails, see UnsafePush.
func (client *RabbitMqClient) Push(data []byte) error {
	client.m.Lock()
	if !client.isReady {
		client.m.Unlock()
		return errors.New("failed to push: not connected")
	}
	client.m.Unlock()
	for {
		err := client.UnsafePush(data)
		if err != nil {
			client.logger.Println("Push failed. Retrying...")
			select {
			case <-client.done:
				return errShutdown
			case <-time.After(resendDelay):
			}
			continue
		}
		confirm := <-client.notifyConfirm
		if confirm.Ack {
			client.logger.Printf("Push confirmed [%d]!", confirm.DeliveryTag)
			return nil
		}
	}
}

// UnsafePush will push to the queue without checking for
// confirmation. It returns an error if it fails to connect.
// No guarantees are provided for whether the server will
// receive the message.
func (client *RabbitMqClient) UnsafePush(data []byte) error {
	client.m.Lock()
	if !client.isReady {
		client.m.Unlock()
		return errNotConnected
	}
	client.m.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return client.channel.PublishWithContext(
		ctx,
		"",
		client.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		},
	)
}
func (queue *RabbitMqClient) Consume(customTag string, deal func(a amqp.Delivery)) {
reconnect:
	for {
		deliveries, err := queue.consume(customTag)
		if err != nil {
			log.Printf("Could not start consuming: %s\n", err)
			<-time.After(time.Second * 2)
			continue
		}
		for {
			select {
			// case <-ctx.Done():
			// 	err := queue.Close()
			// 	if err != nil {
			// 		log.Printf("Close failed: %s\n", err)
			// 	}
			// 	return
			// 当connection关闭时，deliveries也会关闭；
			case delivery, ok := <-deliveries:
				if ok {
					deal(delivery)
					if err := delivery.Ack(false); err != nil {
						log.Printf("Error acknowledging message: %s\n", err)
					}
				} else {
					log.Printf("channel closed,go reto reconnect\n")
					continue reconnect
				}
			}
		}
	}

}

// Consume will continuously put queue items on the channel.
// It is required to call delivery.Ack when it has been
// successfully processed, or delivery.Nack when it fails.
// Ignoring this will cause data to build up on the server.
func (client *RabbitMqClient) consume(customTag string) (<-chan amqp.Delivery, error) {
	client.m.Lock()
	if !client.isReady {
		client.m.Unlock()
		return nil, errNotConnected
	}
	client.m.Unlock()

	if err := client.channel.Qos(
		1,
		0,
		false,
	); err != nil {
		return nil, err
	}

	return client.channel.Consume(
		client.queueName,
		customTag,
		false,
		false,
		false,
		false,
		nil,
	)
}

// Close will cleanly shut down the channel and connection.
func (client *RabbitMqClient) Close() error {
	client.m.Lock()

	defer client.m.Unlock()

	if !client.isReady {
		return errAlreadyClosed
	}
	close(client.done)
	err := client.channel.Close()
	if err != nil {
		return err
	}
	err = client.connection.Close()
	if err != nil {
		return err
	}

	client.isReady = false
	return nil
}
