package cache

import (
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/ini.v1"
)

// ETCD单例
var (
	ETCDClient *clientv3.Client
	ETCDAddr   string
)

// 初始化ETCD
func InitETCD() {
	file, err := ini.Load("config/config.ini")
	if err != nil {
		fmt.Println("配置文件读取错误，请检查文件路径:", err)
		panic(err)
	}
	LoadETCD(file)
	StartETCD()
}

// 读取ETCD参数
func LoadETCD(file *ini.File) {
	ETCDAddr = file.Section("etcd").Key("ETCDAddr").String()
}

// 启动ETCD
func StartETCD() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{ETCDAddr},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("连接ETCD客户端失败")
		panic(err)
	}
	ETCDClient = cli
}
