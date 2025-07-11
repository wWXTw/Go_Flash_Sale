package service

import (
	"FlashSale/model"
	"FlashSale/pkg/e"
	"FlashSale/serializer"
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
				fmt.Printf("用户%d购买了一件商品,ID为%d\n", userid, gid)
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
