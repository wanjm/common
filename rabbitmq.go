package common

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/xid"
)

// RabbitMqClient is the base struct for handling connection recovery, consumption and
// publishing. Note that this struct has an internal mutex to safeguard against
// data races. As you develop and iterate over this example, you may need to add
// further locks, or safeguards, to keep your application safe from data races
type RabbitMqClient struct {
	m            *sync.Mutex
	exchange     string
	queueName    string
	exchangeType string // 交换机类型："fanout", "direct", "topic" 等
	// logger          *log.Logger
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
// 发送消息时，queueName为routingKey，或者exchange为空，queueName为queueName；
// 接受消息时，queueName为queueName，exchange为空；
func New(exchange, queueName, addr string) *RabbitMqClient {
	return NewWithExchangeType(exchange, queueName, "", addr)
}

// NewWithExchangeType creates a new consumer state instance with exchange type.
// exchangeType: "fanout", "direct", "topic" 等，用于声明交换机和判断消费模式
func NewWithExchangeType(exchange, queueName, exchangeType, addr string) *RabbitMqClient {
	client := RabbitMqClient{
		m: &sync.Mutex{},
		// logger:    log.New(os.Stdout, "", log.LstdFlags),
		exchange:     exchange,
		queueName:    queueName,
		exchangeType: exchangeType,
		done:         make(chan bool),
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

		Info(context.Background(), "Attempting to connect ", String("addr", addr))

		conn, err := client.connect(addr)

		if err != nil {
			Info(context.Background(), "Failed to connect to rabbitmq. Retrying...")

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
	Info(context.Background(), "Rabbit Connected!")
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
			Info(context.Background(), "Failed to initialize rabbitmq channel. Retrying.", String("error", err.Error()))

			select {
			case <-client.done:
				return true
			case <-client.notifyConnClose:
				Info(context.Background(), "Connection closed. Reconnecting...")
				return false
			case <-time.After(reInitDelay):
			}
			continue
		}

		select {
		case <-client.done:
			return true
		case <-client.notifyConnClose:
			Info(context.Background(), "rabbitmq Connection closed. Reconnecting...")
			return false
		case <-client.notifyChanClose:
			Info(context.Background(), "rabbitmq Channel closed. Re-running init...")
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
	client.changeChannel(ch)
	client.m.Lock()
	client.isReady = true
	client.m.Unlock()
	Info(context.Background(), "rabbitmq Setup!")

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
			Info(context.Background(), "Push failed. Retrying...")
			select {
			case <-client.done:
				return errShutdown
			case <-time.After(resendDelay):
			}
			continue
		}
		confirm := <-client.notifyConfirm
		if confirm.Ack {
			Info(context.Background(), "Push confirmed", Int("id", int(confirm.DeliveryTag)))
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
		client.exchange,
		client.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		},
	)
}
func (queue *RabbitMqClient) Consume(ctx context.Context, customTag string, deal func(ctx context.Context, a amqp.Delivery)) {
reconnect:
	for {
		deliveries, err := queue.consume(customTag)
		if err != nil {
			Info(ctx, "Could not start consuming in rabbitmq", String("error", err.Error()))
			<-time.After(time.Second * 2)
			continue
		}
		// 获取实际使用的队列名（fanout 模式下已动态更新）
		queue.m.Lock()
		queueName := queue.queueName
		exchangeType := queue.exchangeType
		queue.m.Unlock()

		Info(ctx, "Start consuming",
			String("queue", queueName),
			String("exchangeType", exchangeType),
			String("exchange", queue.exchange))
		for {
			select {
			case <-ctx.Done():
				err := queue.Close()
				if err != nil {
					Info(ctx, "Close failed", String("error", err.Error()))
				}
				Info(ctx, "context done return")
				return
			// 当connection关闭时，deliveries也会关闭；
			case delivery, ok := <-deliveries:
				if ok {
					nc := context.WithValue(ctx, TraceIdNameInContext, xid.New().String())
					Info(nc, "Received mq message", String("id", delivery.MessageId), String("message", string(delivery.Body)))
					start := time.Now()
					deal(nc, delivery)
					if err := delivery.Ack(false); err != nil {
						Info(nc, "Error acknowledging message", String("error", err.Error()))
					}
					Info(nc, "deal mq message finished", Int("duraton", int(time.Since(start).Milliseconds())))
				} else {
					Info(ctx, "channel closed, go reto reconnect")
					continue reconnect
				}
			}
		}
	}
}

// isFanoutMode 判断是否为 fanout 模式（从 exchangeType 推导，避免使用魔法字符串）
func (client *RabbitMqClient) isFanoutMode() bool {
	client.m.Lock()
	defer client.m.Unlock()
	return client.exchangeType == "fanout"
}

// ensureExchange 确保交换机存在（在 consume 时调用）
func (client *RabbitMqClient) ensureExchange() error {
	if client.exchange == "" {
		return nil // 不需要交换机
	}

	client.m.Lock()
	ch := client.channel
	client.m.Unlock()

	if ch == nil {
		return errNotConnected
	}

	// 根据类型声明交换机
	return ch.ExchangeDeclare(
		client.exchange,
		client.exchangeType,
		true,  // durable - 持久化
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
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
	ch := client.channel
	client.m.Unlock()

	if ch == nil {
		return nil, errNotConnected
	}

	// 如果是 fanout 模式，需要特殊处理
	if client.isFanoutMode() {
		// 1. 确保交换机存在
		if err := client.ensureExchange(); err != nil {
			return nil, fmt.Errorf("failed to ensure fanout exchange: %w", err)
		}

		// 2. 声明临时队列并绑定
		queueName, err := client.fanoutQueueDeclare()
		if err != nil {
			return nil, fmt.Errorf("failed to declare fanout queue: %w", err)
		}

		// 3. 保存队列名（直接更新 queueName，复用字段避免重复）
		client.m.Lock()
		client.queueName = queueName
		client.m.Unlock()
	}

	if err := ch.Qos(
		1,
		0,
		false,
	); err != nil {
		return nil, err
	}

	client.m.Lock()
	queueName := client.queueName
	client.m.Unlock()

	return ch.Consume(
		queueName,
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

func (client *RabbitMqClient) fanoutQueueDeclare() (string, error) {
	client.m.Lock()
	ch := client.channel
	client.m.Unlock()

	if ch == nil {
		return "", errNotConnected
	}

	// 声明临时队列
	queue, err := ch.QueueDeclare(
		"",    // 自动生成队列名
		false, // 非持久化
		true,  // 自动删除
		true,  // 排他性
		false, // 非阻塞
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to declare queue: %w", err)
	}

	// 绑定队列到广播交换机
	if err := ch.QueueBind(
		queue.Name,      // 队列名
		"",              // fanout类型不需要路由键
		client.exchange, // 交换机名
		false,
		nil,
	); err != nil {
		return "", fmt.Errorf("failed to bind queue to exchange: %w", err)
	}

	return queue.Name, nil
}
