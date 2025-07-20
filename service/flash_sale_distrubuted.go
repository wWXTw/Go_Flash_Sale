package service

import (
	"FlashSale/cache"
	"FlashSale/config"
	"FlashSale/model"
	"FlashSale/mq"
	"FlashSale/pkg/e"
	"FlashSale/serializer"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// 重置Redis数据库的函数
func ResetRedis(gid int) {
	// 设置测试key对应的值为40,过期时间为永不
	cache.RedisClient.Set(strconv.Itoa(gid), 40, 0)
}

// 优化: 可考虑多实例
// 获取UUID的函数
func GetUuid(gidStr string) string {
	// 获取时间戳
	timeStamp := time.Now().Format("20060102-150405")
	// 随机生成8字节/16位十六进制
	randByte := make([]byte, 8)
	_, err := rand.Read(randByte)
	if err != nil {
		fmt.Println("生成uuid时发生错误!", err)
		panic(err)
	}
	randStr := hex.EncodeToString(randByte)
	Uuid := fmt.Sprintf("%s-%s-%s", timeStamp, randStr, gidStr)
	return Uuid
}

// 设置看门狗
// <-chan struct{} 只读通道,传输类型为空结构体
func SetWatchDog(key, uuid string, interval, ttl time.Duration, stopChan <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		// select多路复用,同时从ticker.C与stopChan两个通道监听信号
		select {
		// 如果是从ticker.C收到了信号
		case <-ticker.C:
			// 对锁进行续期
			val, err := cache.RedisClient.Get(key).Result()
			if err == nil && val == uuid {
				cache.RedisClient.Expire(key, ttl)
			}
		// 如果是从stopChan收到了信号 (var a = <-stopChan)
		case <-stopChan:
			return
		}
	}
}

// 使用Redis分布式锁购买商品
func RedisLockBuyGoodById(gid int, userid int) error {
	gidStr := strconv.Itoa(gid)
	Uuid := GetUuid(gidStr)
	var hasLock bool
	var err error
	for i := 0; i < config.MaxRetry; i++ {
		hasLock, err = cache.RedisClient.SetNX(gidStr, Uuid, 3*time.Second).Result()
		if err != nil {
			return errors.New("获取Redis锁出错")
		} else if hasLock {
			break
		} else {
			// 没抢到锁,等待一段时间后重试
			time.Sleep(25 * time.Millisecond)
		}
	}
	if !hasLock {
		return errors.New("无法抢到Redis锁")
	}
	// 启动看门狗
	stopChan := make(chan struct{})
	SetWatchDog(gidStr, Uuid, 1*time.Second, 3*time.Second, stopChan)
	// 普通购买商品
	err = BuyGoodById(gid, userid)
	if err != nil {
		return err
	}
	// 结束看门狗(向stopChan中发送了一个关闭信号)
	close(stopChan)
	// 利用Lua脚本进行原子解锁
	luaScript := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
    	return redis.call("del", KEYS[1])
	else
   	 return 0
	end
	`
	_, err = cache.RedisClient.Eval(luaScript, []string{gidStr}, Uuid).Result()
	if err != nil {
		fmt.Println("Redis解锁失败")
		return err
	}
	return nil
}

// 利用Redis分布式锁购买商品的服务
func WithRedisLockService(gid int) serializer.Response {
	// 初始化
	var code int
	ResetDataBase(gid)
	ProposedNum := 50
	wg.Add(ProposedNum)
	// 开启购买线程
	for i := 0; i < ProposedNum; i++ {
		go func(gid int) {
			userid := i
			err := RedisLockBuyGoodById(gid, userid)
			if err != nil {
				fmt.Println("Error!", err.Error())
			} else {
				fmt.Printf("处理完用户%d对商品%d的购买请求\n", userid, gid)
			}
			wg.Done()
		}(gid)
	}
	wg.Wait()
	// 获取成功订单数
	SuccessOrder, err := model.GetOrdersCountById(gid)
	if err != nil {
		code = e.ERROR
		return serializer.Response{
			Status: code,
			Data:   nil,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	fmt.Printf("一共完成了%d笔订单\n", SuccessOrder)
	code = e.SUCCESS
	return serializer.Response{
		Status: code,
		Data:   nil,
		Msg:    e.GetMsg(code),
		Error:  "",
	}
}

// MQ+Redis原子化+异步落库购买商品
func MQBuyGoodById(gid int, userid int) error {
	gidStr := strconv.Itoa(gid)
	// 利用Redis原子化操作减去库存
	counts, err := cache.RedisClient.Decr(gidStr).Result()
	if err != nil {
		return err
	}
	// 剩余量如果小于0则进行回滚
	if counts < 0 {
		cache.RedisClient.Incr(gidStr)
		return errors.New("库存已经不足")
	}
	// 库存正常则通过MQ生产者向消费者发送消息
	msg := model.OrderMsg{
		Gid: gid,
		Uid: userid,
	}
	err = mq.PulishOrderMsg(msg)
	if err != nil {
		// 回滚库存
		cache.RedisClient.Incr(gidStr)
		fmt.Println("消息发送失败")
		return err
	}
	return nil
}

// 利用MQ与Redis原子化操作+异步落库的服务
func WithMQService(gid int) serializer.Response {
	// 初始化
	var code int
	ResetDataBase(gid)
	ProposedNum := 50
	wg.Add(ProposedNum)
	// Redis数据库的初始化 ???
	ResetRedis(gid)
	// 开启购买线程
	for i := 0; i < ProposedNum; i++ {
		go func(gid int) {
			userid := i
			err := MQBuyGoodById(gid, userid)
			if err != nil {
				fmt.Println("Error!", err.Error())
			} else {
				fmt.Printf("处理完用户%d对商品%d的购买请求\n", userid, gid)
			}
			wg.Done()
		}(gid)
	}
	wg.Wait()
	// 获取成功订单数
	SuccessOrder, err := model.GetOrdersCountById(gid)
	if err != nil {
		code = e.ERROR
		return serializer.Response{
			Status: code,
			Data:   nil,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	fmt.Printf("一共完成了%d笔订单\n", SuccessOrder)
	code = e.SUCCESS
	return serializer.Response{
		Status: code,
		Data:   nil,
		Msg:    e.GetMsg(code),
		Error:  "",
	}
}
