package model

// Good表结构
type Good struct {
	GoodsId       int64 `gorm:"primaryKey"`
	Title         string
	SubTitle      string
	OriginalPrice float64
	CurrentPrice  float64
	CategoryId    int64
}

func FindGoodsById(goodsid int) (Good, error) {
	var good Good
	// 根据id值查询商品数据
	err := DB.Where("goods_id=?", goodsid).First(&good).Error
	return good, err
}
