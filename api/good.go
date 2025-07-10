package api

// 提供商品服务的api端口文件
import (
	"FlashSale/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 获取商品信息的API
func GetGoodInfo(context *gin.Context) {
	gid, _ := strconv.Atoi(context.Query("gid"))
	result := service.GetGoodInfo(gid)
	// 发送回前端
	context.JSON(result.Status, result)
}
