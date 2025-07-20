package main

// 入口/启动文件
import (
	"FlashSale/cache"
	"FlashSale/config"
	"FlashSale/mq"
	"FlashSale/routers"
)

func main() {
	// 初始化MySQL
	config.Init()
	// 初始化Redis
	cache.InitRedis()
	// 初始化RabbitMQ
	mq.InitRabbitMQ()
	// 初始化MQ消费者
	mq.ConsumeOrderMsg()
	// 设置router并运行
	r := routers.NewRouter()
	println("启动监听端口：", config.HttpPort)
	_ = r.Run(config.HttpPort)
}
