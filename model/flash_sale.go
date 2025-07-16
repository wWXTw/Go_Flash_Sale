package model

// 对有关秒杀的数据库进行操作(DAO)

import (
	"time"

	"gorm.io/gorm"
)

// GoodCounts表结构
type GoodCounts struct {
	GoodsId        int64 `gorm:"primaryKey"`
	Counts         int64
	LastUpdateTime time.Time `gorm:"autoUpdateTime"`
	Version        int64
}

// GoodOrder表结构
type GoodOrders struct {
	GoodsId  int64
	UserId   int64
	SoldTime time.Time `gorm:"autoCreateTime;autoUpdateTime"`
}

// 查询商品数量
func GetCountByGoodsId(gid int) (int, error) {
	var gc int
	err := DB.Model(&GoodCounts{}).Select("counts").Where("goods_id=?", gid).Scan(&gc).Error
	return gc, err
}

// 查询商品
func GetGoodByGoodId(gid int) (GoodCounts, error) {
	var good GoodCounts
	err := DB.Model(&GoodCounts{}).Where("goods_id=?", gid).First(&good).Error
	return good, err
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

// 悲观锁读锁查询商品数量
func PccReadGetCountByGoodId(tx *gorm.DB, gid int) (int, error) {
	var gc int
	// 设置悲观读锁
	sql := `SELECT counts FROM good_counts WHERE goods_id = ? FOR UPDATE`
	err := tx.Raw(sql, gid).Scan(&gc).Error
	return gc, err
}

// 悲观锁写锁更新商品数量
func PccWriteReduceOneByGoodsId(tx *gorm.DB, gid int) (int, error) {
	counts := 0
	// SQL原子化操作(本质上用到了行级锁,悲观锁的核心)
	sqlStr := `UPDATE good_counts SET counts = counts - 1 where counts > 0 AND goods_id = ?`
	res := tx.Exec(sqlStr, gid)
	if err := res.Error; err != nil {
		return counts, err
	}
	counts = int(res.RowsAffected)
	return counts, nil
}

// 乐观锁情况下更新商品数量
func OccReduceOneByGoodsID(tx *gorm.DB, gid int, need int, version int) (int, error) {
	counts := 0
	sqlStr := `UPDATE good_counts SET counts = counts - ?, version = version + 1 WHERE version = ? AND goods_id = ?`
	res := tx.Exec(sqlStr, need, version, gid)
	if err := res.Error; err != nil {
		return counts, err
	}
	counts = int(res.RowsAffected)
	return counts, nil
}
