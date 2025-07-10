package service

// 处理与商品属性有关的服务
import (
	"FlashSale/model"
	"FlashSale/pkg/e"
	"FlashSale/serializer"
)

// 获取商品信息的函数
func GetGoodInfo(gid int) serializer.Response {
	var code int
	good, err := model.FindGoodsById(gid)
	if err != nil {
		code = e.ERROR
		return serializer.Response{
			Status: code,
			Data:   nil,
			Msg:    e.GetMsg(code),
			Error:  err.Error(),
		}
	}
	code = e.SUCCESS
	return serializer.Response{
		Status: code,
		Data:   good,
		Msg:    e.GetMsg(code),
		Error:  "",
	}
}
