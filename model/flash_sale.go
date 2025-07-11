package model

// 对有关秒杀的数据库进行操作(DAO)

import (
	"time"

	"gorm.io/gorm"
)

// GoodCounts表结构
type GoodCounts struct {
	GoodId         int64 `gorm:"primaryKey"`
	Counts         int64
	LastUpdateTime time.Time `gorm:"autoUpdateTime"`
	Version        int64
}

// GoodOrder表结构
type GoodOrders struct {
	GoodId   int64
	UserId   int64
	SoldTime time.Time `gorm:"autoCreateTime;autoUpdateTime"`
}

// 查询商品数量
func GetCountByGoodsId(gid int) (int64, error) {
	var gc int64
	err := DB.Model(&GoodCounts{}).Select("count").Where("goods_id=?", gid).Scan(&gc).Error
	return gc, err
}

// 重置商品数量与版本
func ResetCountByGoodsId(tx *gorm.DB, gid int) error {
	return tx.Model(&GoodCounts{}).Where("goods_id=?", gid).Updates(map[string]interface{}{
		"counts":  40,
		"version": 0,
	}).Error
}

// 减少商品为指定数量
func ReduceCountByGoodsId(tx *gorm.DB, gid int, newCount int64) error {
	return tx.Model(&GoodCounts{}).Where("goods_id=?", gid).Update("counts", newCount).Error
}

// 减少一个商品数量
func ReduceOneByGoodsId(tx *gorm.DB, gid int) error {
	var counts int64
	err := tx.Model(&GoodCounts{}).Select("counts").Where("goods_id=?", gid).Scan(&counts).Error
	if counts <= 0 || err != nil {
		return err
	}
	err = tx.Model(&GoodCounts{}).Where("goods_id=?", gid).Update("counts", counts-1).Error
	return err
}

// 添加订单
func AddOrder(tx *gorm.DB, order GoodOrders) error {
	return tx.Model(&GoodOrders{}).Create(&order).Error
}

// 清空商品订单
func ClearOrderByGoodsId(tx *gorm.DB, gid int) error {
	return tx.Where("goods_id", gid).Delete(&GoodOrders{}).Error
}

// 获取商品订单数
func GetOrdersCountById(gid int) (int64, error) {
	var counts int64
	err := DB.Model(&GoodOrders{}).Where("goods_id", gid).Count(&counts).Error
	return counts, err
}
