package api

// 单机环境下的API

import (
	"FlashSale/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 不加锁的情况
func WithoutLock(ctx *gin.Context) {
	var gid int
	gid, _ = strconv.Atoi(ctx.Query("gid"))
	resp := service.WithoutLockService(gid)
	ctx.JSON(resp.Status, resp)
}
