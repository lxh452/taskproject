package svc

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// MQClient RabbitMQ 客户端
type MQClient struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
}

// GetChannel 获取 channel（用于消费者绑定队列）
func (m *MQClient) GetChannel() *amqp.Channel {
	return m.channel
}

// GetExchange 获取 exchange 名称
func (m *MQClient) GetExchange() string {
	return m.exchange
}

// NewMQClient 创建 RabbitMQ 客户端
func NewMQClient(url, exchange string) (*MQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// 声明交换机（topic 类型，支持路由键匹配）
	err = channel.ExchangeDeclare(
		exchange, // 交换机名称
		"topic",  // 类型：topic 支持路由键匹配
		true,     // 持久化
		false,    // 自动删除
		false,    // 内部使用
		false,    // 不等待
		nil,      // 参数
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	return &MQClient{
		conn:     conn,
		channel:  channel,
		exchange: exchange,
	}, nil
}

// Publish 发布消息
func (m *MQClient) Publish(routingKey string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = m.channel.Publish(
		m.exchange, // 交换机
		routingKey, // 路由键
		false,      // 强制
		false,      // 立即
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	logx.Infof("Published message to exchange=%s, routingKey=%s", m.exchange, routingKey)
	return nil
}

// Close 关闭连接
func (m *MQClient) Close() error {
	if m.channel != nil {
		m.channel.Close()
	}
	if m.conn != nil {
		return m.conn.Close()
	}
	return nil
}
