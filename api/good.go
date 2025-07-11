package api

// 提供商品服务的api端口文件
import (
	"FlashSale/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 获取商品信息的API
func GetGoodInfo(ctx *gin.Context) {
	gid, _ := strconv.Atoi(ctx.Query("gid"))
	result := service.GetGoodInfo(gid)
	// 发送回前端
	ctx.JSON(result.Status, result)
}
