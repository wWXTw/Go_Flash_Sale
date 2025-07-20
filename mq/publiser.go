package mq

// MQ的生产者--向队列发送订单消息

import (
	"FlashSale/model"
	"encoding/json"

	"github.com/streadway/amqp"
)

// 发送消息到通道的函数
func PulishOrderMsg(msg model.OrderMsg) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	err = MQChan.Publish(
		"",
		"OrderQueue", // 选择设置好的订单队列
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: 2, // 持久化
			Body:         body,
		},
	)
	if err != nil {
		return err
	}
	return nil
}
