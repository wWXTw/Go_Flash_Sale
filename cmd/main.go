package main

// 入口/启动文件
import (
	"FlashSale/config"
	"FlashSale/routers"
)

func main() {
	config.Init()
	r := routers.NewRouter()
	_ = r.Run(config.HttpPort)
}
