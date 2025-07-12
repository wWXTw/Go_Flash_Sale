package service

// 单例模式 返回统一的通道instance
import "sync"

// 定义通道长度为2,传递gid与userid
type SingleChan chan [2]int

var instance *SingleChan
var once sync.Once

func GetInstance() *SingleChan {
	// once确保只初始化一次
	once.Do(func() {
		channel := make(SingleChan, 100)
		instance = &channel
	})
	return instance
}
