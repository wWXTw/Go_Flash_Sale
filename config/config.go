package config

import (
	"FlashSale/model"
	"fmt"

	"gopkg.in/ini.v1"
)

// 配置服务器与MySql数据库与其他配置
var (
	AppMode  string
	HttpPort string

	Db         string
	DbHost     string
	DbPort     string
	DbUser     string
	DbPassWord string
	DbName     string

	MaxRetry int
)

// 初始化函数
func Init() {
	file, err := ini.Load("./config/config.ini")
	if err != nil {
		fmt.Println("配置文件读取错误，请检查文件路径:", err)
		panic(err)
	}
	LoadServer(file)
	LoadMysql(file)
	// 拼接DSN
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		DbUser,
		DbPassWord,
		DbHost,
		DbPort,
		DbName,
	)
	// 开启数据库
	model.DataBase(dsn)
	// 配置其他参数
	MaxRetry = 30
}

// 读取服务器参数的函数
func LoadServer(file *ini.File) {
	AppMode = file.Section("service").Key("AppMode").String()
	HttpPort = file.Section("service").Key("HttpPort").String()
}

// 读取MySql数据库参数的函数
func LoadMysql(file *ini.File) {
	Db = file.Section("mysql").Key("Db").String()
	DbHost = file.Section("mysql").Key("DbHost").String()
	DbPort = file.Section("mysql").Key("DbPort").String()
	DbUser = file.Section("mysql").Key("DbUser").String()
	DbPassWord = file.Section("mysql").Key("DbPassWord").String()
	DbName = file.Section("mysql").Key("DbName").String()
}
