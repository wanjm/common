package common

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	ACCESS_FROM_USER = 0
	COLON            = ":"
)

type RabbitMqConfig struct {
	AccessKeySecret string
	ResourceOwnerId string
	AccessKeyId     string
	ConsumerTag     string
	Vhost           string
	Addr            string
	Exchange        string
	ExchangeType    string
	QueueName       string
	RoutingKey      string
}

func (cfg *RabbitMqConfig) GetAddr() string {
	username := getUserName(cfg.AccessKeyId, cfg.ResourceOwnerId)
	password := getPassword(cfg.AccessKeySecret)
	return fmt.Sprintf("amqp://%s:%s@%s/%s?heartbeat=0", username, password, cfg.Addr, cfg.Vhost)
}

func getUserName(ak string, instanceId string) string {
	var buffer bytes.Buffer
	buffer.WriteString(strconv.Itoa(ACCESS_FROM_USER))
	buffer.WriteString(COLON)
	buffer.WriteString(instanceId)
	buffer.WriteString(COLON)
	buffer.WriteString(ak)
	return base64.StdEncoding.EncodeToString(buffer.Bytes())
}
func getPassword(sk string) string {
	now := time.Now()
	currentMillis := strconv.FormatInt(now.UnixNano()/1000000, 10)
	var buffer bytes.Buffer
	buffer.WriteString(strings.ToUpper(HmacSha1(sk, currentMillis)))
	buffer.WriteString(COLON)
	buffer.WriteString(currentMillis)
	return base64.StdEncoding.EncodeToString(buffer.Bytes())
}

func ConnectProducer(cfg *RabbitMqConfig) *RabbitMqClient {
	queue := New(cfg.Exchange, cfg.RoutingKey, cfg.GetAddr())
	return queue
}

func ConnectConsumer(cfg *RabbitMqConfig) *RabbitMqClient {
	queue := New("", cfg.QueueName, cfg.GetAddr())
	return queue
}
