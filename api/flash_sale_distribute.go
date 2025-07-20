package api

import (
	"FlashSale/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 使用Redis分布式锁的情况
func WithRedisLock(ctx *gin.Context) {
	gid, _ := strconv.Atoi(ctx.Query("gid"))
	resp := service.WithRedisLockService(gid)
	ctx.JSON(resp.Status, resp)
}

// 使用MQ+Redis原子化的情况
func WithMQ(ctx *gin.Context) {
	gid, _ := strconv.Atoi(ctx.Query("gid"))
	resp := service.WithMQService(gid)
	ctx.JSON(resp.Status, resp)
}
