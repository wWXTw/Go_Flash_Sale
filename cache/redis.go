package cache

import (
	"context"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
	"gopkg.in/ini.v1"
)

var (
	// redis集群单例
	RedisClusterClient *redis.ClusterClient
	RedisAddrs         []string
	RedisName          string
	RedisPassword      string
	RedisCtx           = context.Background()
)

// 初始化Redis数据库
func InitRedis() {
	file, err := ini.Load("config/config.ini")
	if err != nil {
		fmt.Println("配置文件读取错误，请检查文件路径:", err)
		panic(err)
	}
	LoadRedis(file)
	StartRedis()
}

// 从config文件中加载Redis参数
func LoadRedis(file *ini.File) {
	AddrStr := file.Section("redis").Key("RedisAddrs").String()
	// 拆分Redis集群地址
	RedisAddrs = strings.Split(AddrStr, ",")
	RedisName = file.Section("redis").Key("RedisName").String()
	RedisPassword = file.Section("redis").Key("RedisPassword").String()
}

// 启动Redis数据库
func StartRedis() {
	RedisClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    RedisAddrs,
		Password: RedisPassword,
	})
	pong, err := RedisClusterClient.Ping(RedisCtx).Result()
	if err != nil {
		fmt.Println("Redis数据库连接失败!")
		panic(err)
	}
	fmt.Println("Redis数据库连接成功", pong)
}
