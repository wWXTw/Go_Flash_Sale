package mq

import (
	"fmt"

	"github.com/streadway/amqp"
	"gopkg.in/ini.v1"
)

var (
	MQConn     *amqp.Connection
	MQChan     *amqp.Channel
	MQName     string
	MQPassword string
	MQAddr     string
)

// 初始化RabbitMQ的函数
func InitRabbitMQ() {
	file, err := ini.Load("config/config.ini")
	if err != nil {
		fmt.Println("配置文件读取错误，请检查文件路径:", err)
		panic(err)
	}
	LoadRabbitMQ(file)
	StartRabbitMQ()
}

// 读取MQ参数的函数
func LoadRabbitMQ(file *ini.File) {
	MQName = file.Section("rabbitmq").Key("MQName").String()
	MQPassword = file.Section("rabbitmq").Key("MQPassword").String()
	MQAddr = file.Section("rabbitmq").Key("MQAddr").String()
}

// 启动MQ的函数
func StartRabbitMQ() {
	var err error
	mq_link := fmt.Sprintf("amqp://%s:%s@%s",
		MQName,
		MQPassword,
		MQAddr)
	MQConn, err = amqp.Dial(mq_link)
	if err != nil {
		fmt.Println("MQ连接失败")
		panic(err)
	}
	MQChan, err = MQConn.Channel()
	if err != nil {
		fmt.Println("MQ通道打开失败")
		panic(err)
	}
	// 配置消息队列
	_, err = MQChan.QueueDeclare(
		"OrderQueue",
		true,  // 持久化队列/重启后仍存在原数据
		false, // 自动删除设置为否/无消费者时也不删除队列
		false, // 不是排他队列/多个连接可共享
		false, // 等待Rabbit服务器返回ok
		nil,   // 暂不设计高级参数
	)
	if err != nil {
		fmt.Println("MQ队列初始化失败")
		panic(err)
	}
	fmt.Println("MQ连接成功,队列设置成功")
}
