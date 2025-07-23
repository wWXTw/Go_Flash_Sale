package service

import (
	"FlashSale/cache"
	"FlashSale/config"
	"FlashSale/model"
	"FlashSale/mq"
	"FlashSale/pkg/e"
	"FlashSale/serializer"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// 重置Redis数据库的函数
func ResetRedis(gid int) {
	// 设置测试key对应的值为40,过期时间为永不
	cache.RedisClusterClient.Set(cache.RedisCtx, strconv.Itoa(gid), 40, 0)
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
			val, err := cache.RedisClusterClient.Get(cache.RedisCtx, key).Result()
			if err == nil && val == uuid {
				cache.RedisClusterClient.Expire(cache.RedisCtx, key, ttl)
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
		hasLock, err = cache.RedisClusterClient.SetNX(cache.RedisCtx, gidStr, Uuid, 3*time.Second).Result()
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
	_, err = cache.RedisClusterClient.Eval(cache.RedisCtx, luaScript, []string{gidStr}, Uuid).Result()
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
	counts, err := cache.RedisClusterClient.Decr(cache.RedisCtx, gidStr).Result()
	if err != nil {
		return err
	}
	// 剩余量如果小于0则进行回滚
	if counts < 0 {
		cache.RedisClusterClient.Incr(cache.RedisCtx, gidStr)
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
		cache.RedisClusterClient.Incr(cache.RedisCtx, gidStr)
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
	// Redis数据库的初始化
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

// 定义ETCD锁结构
type ETCDMutex struct {
	Ttl     int64              // 锁有效时间
	Conf    clientv3.Config    // 客户端连接配置
	Key     string             // Key值
	Cancel  context.CancelFunc // 取消续约的上下文函数
	lease   clientv3.Lease     // ETCD的租借接口
	leaseID clientv3.LeaseID   // 租约ID
	txn     clientv3.Txn       // ETCD事务
}

// ETCD锁的构造函数
func (em *ETCDMutex) Init() error {
	// 创建一个ETCD客户端
	client, err := clientv3.New(em.Conf)
	if err != nil {
		fmt.Println("初始化ETCD客户端失败")
		return err
	}
	// 创建KV接口(key-value)的原子事务对象
	em.txn = clientv3.NewKV(client).Txn(context.Background())
	// 创建租约接口
	em.lease = clientv3.NewLease(client)
	// 申请一个固定ttl的租约
	var leaseResp *clientv3.LeaseGrantResponse
	leaseResp, err = em.lease.Grant(context.Background(), em.Ttl) // context.Background()类似于空,适合补足不重要context
	if err != nil {
		fmt.Println("申请租约失败")
		return err
	}
	// 保存租约ID
	em.leaseID = leaseResp.ID
	// 创建取消函数
	ctx, cancel := context.WithCancel(context.Background())
	em.Cancel = cancel
	// 开启自动续约 KeepAlive一直监听ctx的Done()信号,收到时则停止续约
	_, err = em.lease.KeepAlive(ctx, em.leaseID)
	if err != nil {
		fmt.Println("ETCD开启自动续约失败")
		return err
	}
	return nil
}

// ETCD锁的上锁函数
func (em *ETCDMutex) Lock() (bool, error) {
	// 初始化
	err := em.Init()
	if err != nil {
		fmt.Println("ETCD锁初始化失败")
		return false, err
	}
	// 设置原子加锁事务
	// 如果在ETCD数据库中没找到以em.Key为键值的行则进行创建并绑定租约
	em.txn.If(clientv3.Compare(clientv3.CreateRevision(em.Key), "=", 0)).
		Then(clientv3.OpPut(em.Key, "", clientv3.WithLease(em.leaseID))).
		Else()
	// 执行事务
	var txnResp *clientv3.TxnResponse
	txnResp, err = em.txn.Commit()
	if err != nil {
		fmt.Println("ETCD事务执行失败")
		return false, err
	}
	// 如果锁已被别人抢占
	if !txnResp.Succeeded {
		return false, nil
	}
	// 成功得到锁
	return true, nil
}

// ETCD锁的解锁函数
func (em *ETCDMutex) UnLock() error {
	// 调用取消函数 发送ctx.Done()信号
	em.Cancel()
	// 释放租约 自动删除对应Key行
	_, err := em.lease.Revoke(context.Background(), em.leaseID)
	if err != nil {
		return err
	}
	return nil
}

// 利用ETCD分布式锁购买商品
func WithETCDBuyGoodById(gid int, userid int) error {
	var conf = clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}
	// 设置ETCD锁的配置
	em := &ETCDMutex{
		Key:  strconv.Itoa(gid),
		Conf: conf,
		Ttl:  10,
	}
	var locked bool
	var err error
	// 尝试上锁
	for i := 0; i < config.MaxRetry; i++ {
		locked, err = em.Lock()
		if err != nil {
			return err
		}
		// 抢到锁退出重试
		if locked {
			break
		}
		// 延迟一段时间后进行重试
		time.Sleep(100 * time.Millisecond)
	}
	if !locked {
		return errors.New("ETCD锁无法抢到")
	}
	// 购买商品函数
	err = BuyGoodById(gid, userid)
	if err != nil {
		return err
	}
	// 解锁
	err = em.UnLock()
	if err != nil {
		return err
	}
	return nil
}

// 利用ETCD缓存分布式锁的服务
func WithETCDService(gid int) serializer.Response {
	// 初始化
	var code int
	ResetDataBase(gid)
	ProposedNum := 50
	wg.Add(ProposedNum)
	// 开启购买线程
	for i := 0; i < ProposedNum; i++ {
		go func(gid int) {
			userid := i
			err := WithETCDBuyGoodById(gid, userid)
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
