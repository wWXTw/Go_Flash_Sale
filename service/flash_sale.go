package service

import (
	"FlashSale/model"
	"FlashSale/pkg/e"
	"FlashSale/serializer"
	"errors"
	"fmt"
	"math/rand"
	"sync"
)

var wg sync.WaitGroup
var lock sync.Mutex

// 重置数据库环境的函数
func ResetDataBase(gid int) error {
	// 开启一个事务
	tx := model.DB.Begin()
	// 清空订单表
	err := model.ClearOrderByGoodsId(tx, gid)
	// 发生错误即回滚
	if err != nil {
		tx.Rollback()
		return err
	}
	// 重置商品数量与版本
	err = model.ResetCountByGoodsId(tx, gid)
	if err != nil {
		tx.Rollback()
		return err
	}
	// 事务提交
	tx.Commit()
	return nil
}

// 购买商品并创建订单的函数
func BuyGoodById(gid int, userid int) error {
	tx := model.DB.Begin()
	// 捕获panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 获取商品数目
	counts, err := model.GetCountByGoodsId(gid)
	if err != nil {
		return err
	}
	if counts > 0 {
		// 对商品数目减一
		err = model.ResetCountByGoodsId(tx, int(counts-1))
		if err != nil {
			tx.Rollback()
			return err
		}
		// 生成订单元组
		var Order = model.GoodOrders{
			GoodId: int64(gid),
			UserId: int64(userid),
		}
		err = model.AddOrder(tx, Order)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

// 不带锁购买商品的服务函数
func WithoutLockService(gid int) serializer.Response {
	var code int
	// 设置为购买50个(超过预设的40个)
	ProposedNum := 50
	// 重置预设环境
	ResetDataBase(gid)
	// 设置等待组
	wg.Add(ProposedNum)
	for i := 0; i < ProposedNum; i++ {
		// 开启线程进行购买
		go func(gid int) {
			// 随机生成用户ID
			userid := rand.Intn(100)
			err := BuyGoodById(gid, userid)
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

// 带锁购买商品的服务函数
func WithLockService(gid int) serializer.Response {
	// 初始化
	var code int
	ProposedNum := 50
	ResetDataBase(gid)
	wg.Add(ProposedNum)
	// 开启购买线程
	for i := 0; i < ProposedNum; i++ {
		go func(gid int) {
			userid := rand.Intn(100)
			// 购买商品时进行加锁
			lock.Lock()
			err := BuyGoodById(gid, userid)
			lock.Unlock()
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

// 悲观锁加读锁购买商品并创建订单
func PccReadBuyGoodById(gid int, userid int) error {
	tx := model.DB.Begin()
	// 捕获panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 利用悲观锁查询
	counts, err := model.PccReadGetCountByGoodId(tx, gid)
	if err != nil {
		tx.Rollback()
		return err
	}
	if counts > 0 {
		// 对商品数目减一
		err = model.ResetCountByGoodsId(tx, int(counts-1))
		if err != nil {
			tx.Rollback()
			return err
		}
		// 生成订单元组
		var Order = model.GoodOrders{
			GoodId: int64(gid),
			UserId: int64(userid),
		}
		err = model.AddOrder(tx, Order)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

// 悲观锁加读锁购买商品的服务
func PccReadService(gid int) serializer.Response {
	// 初始化
	var code int
	ProposedNum := 50
	ResetDataBase(gid)
	wg.Add(ProposedNum)
	// 开启购买线程
	for i := 0; i < ProposedNum; i++ {
		go func(gid int) {
			userid := rand.Intn(100)
			err := PccReadBuyGoodById(gid, userid)
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

// 悲观锁加写锁购买商品并创建订单
func PccWriteBuyGoodById(gid int, userid int) error {
	tx := model.DB.Begin()
	counts, err := model.PccWriteReduceOneByGoodsId(tx, gid)
	if err != nil {
		tx.Rollback()
		return err
	}
	// 如果商品已被购买(counts > 0 -> rowAffected > 0)
	if counts > 0 {
		var Order = model.GoodOrders{
			GoodId: int64(gid),
			UserId: int64(userid),
		}
		err = model.AddOrder(tx, Order)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

// 悲观锁加写锁购买商品的服务
func PccWriteService(gid int) serializer.Response {
	// 初始化
	var code int
	ResetDataBase(gid)
	ProposedNum := 50
	wg.Add(ProposedNum)
	// 开启购买线程
	for i := 0; i < ProposedNum; i++ {
		go func(gid int) {
			userid := rand.Intn(100)
			err := PccWriteBuyGoodById(gid, userid)
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

// 乐观锁购买商品并创建订单
func OccBuyGoodById(gid int, userid int, need int) error {
	tx := model.DB.Begin()
	good, err := model.GetGoodByGoodId(gid)
	if err != nil {
		return err
	}
	// 不加这个if判断语句乐观锁会一直重试下去
	if good.Counts >= int64(need) {
		// 版本控制,传入获取到商品信息的版本号
		counts, err := model.OccReduceOneByGoodsID(tx, gid, need, int(good.Version))
		if err != nil {
			// 版本老旧或者数据库问题,进行重试
			// 重试体现在何处呢?
			tx.Rollback()
			return err
		}
		if counts > 0 {
			// 购买成功添加订单
			var Order = model.GoodOrders{
				GoodId: int64(gid),
				UserId: int64(userid),
			}
			err = model.AddOrder(tx, Order)
			if err != nil {
				tx.Rollback()
				return err
			}
		} else {
			tx.Rollback()
		}
	} else {
		tx.Rollback()
		return errors.New("内存已经不足")
	}
	tx.Commit()
	return nil
}

// 乐观锁购买商品的服务
func OccService(gid int) serializer.Response {
	// 初始化
	var code int
	ResetDataBase(gid)
	ProposedNum := 50
	wg.Add(ProposedNum)
	// 开启购买线程
	for i := 0; i < ProposedNum; i++ {
		go func(gid int) {
			userid := rand.Intn(100)
			err := OccBuyGoodById(gid, userid, 1)
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

// 传递通道购买请求的函数
func ChannelSend(gid int, userid int) error {
	// 创建通道中的请求([2]int)结构
	var req = [2]int{gid, userid}
	// 利用通道传递
	channel := GetInstance()
	*channel <- req
	return nil
}

// 处理通道购买请求的函数
func ChannelRecv() {
	for {
		// 从通道中获取请求
		req, ok := <-(*GetInstance())
		if !ok {
			continue
		}
		err := BuyGoodById(req[0], req[1])
		if err != nil {
			fmt.Println("Error!", err.Error())
		} else {
			fmt.Printf("处理完用户%d对商品%d的购买请求\n", req[0], req[1])
		}
		wg.Done()
	}
}

// 通道购买商品的服务
func ChannelService(gid int) serializer.Response {
	// 初始化
	var code int
	ResetDataBase(gid)
	ProposedNum := 50
	wg.Add(ProposedNum)
	// 开启通道接收线程
	go ChannelRecv()
	// 开启购买线程
	for i := 0; i < ProposedNum; i++ {
		go func(gid int) {
			userid := rand.Intn(100)
			err := ChannelSend(gid, userid)
			if err != nil {
				fmt.Println("Error!", err.Error())
			} else {
				fmt.Printf("发送完用户%d对商品%d的购买请求\n", userid, gid)
			}
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
