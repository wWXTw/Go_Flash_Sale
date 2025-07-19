package cache

import (
	"fmt"

	"github.com/go-redis/redis"
	"gopkg.in/ini.v1"
)

var (
	// redis单例
	RedisClient   *redis.Client
	RedisAddr     string
	RedisName     string
	RedisPassword string
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
	RedisAddr = file.Section("redis").Key("RedisAddr").String()
	RedisName = file.Section("redis").Key("RedisName").String()
	RedisPassword = file.Section("redis").Key("RedisPassword").String()
}

// 启动Redis数据库
func StartRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     RedisAddr,
		Password: RedisPassword,
		// 默认使用第0个数据库
		DB: 0,
	})
	// 测试redis连接效果
	pong, err := RedisClient.Ping().Result()
	if err != nil {
		fmt.Println("Redis数据库连接失败!")
		panic(err)
	}
	fmt.Println("Redis数据库连接成功", pong)
}
