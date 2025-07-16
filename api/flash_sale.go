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

// 加锁的情况
func WithLock(ctx *gin.Context) {
	var gid int
	gid, _ = strconv.Atoi(ctx.Query("gid"))
	resp := service.WithLockService(gid)
	ctx.JSON(resp.Status, resp)
}

// 悲观锁读锁的情况
func PccReadLock(ctx *gin.Context) {
	var gid int
	gid, _ = strconv.Atoi(ctx.Query("gid"))
	resp := service.PccReadService(gid)
	ctx.JSON(resp.Status, resp)
}

// 悲观锁写锁的情况
func PccWriteLock(ctx *gin.Context) {
	var gid int
	gid, _ = strconv.Atoi(ctx.Query("gid"))
	resp := service.PccWriteService(gid)
	ctx.JSON(resp.Status, resp)
}

// 乐观锁的情况
func OccLock(ctx *gin.Context) {
	var gid int
	gid, _ = strconv.Atoi(ctx.Query("gid"))
	resp := service.OccService(gid)
	ctx.JSON(resp.Status, resp)
}

// 通道的情况
func Channel(ctx *gin.Context) {
	var gid int
	gid, _ = strconv.Atoi(ctx.Query("gid"))
	resp := service.ChannelService(gid)
	ctx.JSON(resp.Status, resp)
}
