package model

// 定义连接的数据库模型
import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func DataBase(conn string) {
	// 连接数据库
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       conn,
		DefaultStringSize:         256,
		SkipInitializeWithVersion: false, // 老版本兼容设置
		DisableDatetimePrecision:  false,
		DontSupportRenameIndex:    false,
		DontSupportRenameColumn:   false,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 设置连接池
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetConnMaxLifetime(3 * time.Minute)

	DB = db

	// 初始化结构
	err = initDB()
	if err != nil {
		panic(fmt.Sprintf("数据库的结构初始化失败:%v", err))
	}
}

// 初始化结构(建表并插入初始数据)
func initDB() error {
	// 自动安全建表
	DB.AutoMigrate(&Good{}, &GoodCounts{}, &GoodOrders{})
	// 初始化数据
	var counts int64
	err := DB.Model(&Good{}).Where("goods_id=?", 520).Count(&counts).Error
	if err != nil {
		return err
	}
	if counts == 0 {
		var good = Good{
			GoodsId:       520,
			Title:         "雅马哈P48电子钢琴",
			SubTitle:      "88键重锤便携电子钢琴",
			OriginalPrice: 3294.0,
			CurrentPrice:  2964.0,
			CategoryId:    1,
		}
		err = DB.Model(&Good{}).Create(&good).Error
		if err != nil {
			return err
		}
	}
	err = DB.Model(&GoodCounts{}).Where("goods_id=?", 520).Count(&counts).Error
	if err != nil {
		return err
	}
	if counts == 0 {
		var goodCount = GoodCounts{
			GoodsId: 520,
			Counts:  40,
			Version: 0,
		}
		err = DB.Model(&GoodCounts{}).Create(&goodCount).Error
		if err != nil {
			return err
		}
	}
	return nil
}
