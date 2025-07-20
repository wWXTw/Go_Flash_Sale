package mq

// 消息队列消费者,用于消费生产者发送的订单消息,进行异步入库

import (
	"FlashSale/model"
	"encoding/json"
	"fmt"
)

// 消费订单消息的函数
func ConsumeOrderMsg() {
	msgs, err := MQChan.Consume(
		"OrderQueue",
		"",    // 消费者标签
		true,  // 自动回复成功ack -- 之后可改为手动发送->高可靠
		false, // 不为专用
		false,
		false,
		nil, // 暂不设置高级参数
	)
	if err != nil {
		fmt.Println("MQ消费者初始化失败")
		panic(err)
	}
	// 启动一个专门的消费者线程进行服务
	go func() {
		// 堵塞式的从队列中取消息/类似channel
		for msgJson := range msgs {
			var msg model.OrderMsg
			err = json.Unmarshal(msgJson.Body, &msg)
			if err != nil {
				fmt.Println("消息解析失败!")
				continue
			}
			// 将订单落到MySQL数据库上
			gid := msg.Gid
			userid := msg.Uid
			tx := model.DB.Begin()
			// 获取商品数目
			counts, err := model.GetCountByGoodsId(gid)
			if err != nil {
				fmt.Println("消费过程中获取MySQL库存失败")
				continue
			}
			if counts > 0 {
				// 对商品数目减一
				// 优化: MQ消费者多实例的情况?
				err = model.ReduceCountByGoodsId(tx, gid, int64(counts-1))
				if err != nil {
					tx.Rollback()
					fmt.Println("消费过程中MySQL库存减一失败")
					continue
				}
				// 生成订单元组
				var Order = model.GoodOrders{
					GoodsId: int64(gid),
					UserId:  int64(userid),
				}
				err = model.AddOrder(tx, Order)
				if err != nil {
					tx.Rollback()
					fmt.Println("消费过程中MySQL订单添加失败", err.Error())
					continue
				}
			}
			tx.Commit()
			fmt.Println("一份消息被消费者处理完毕")
		}
	}()
}
