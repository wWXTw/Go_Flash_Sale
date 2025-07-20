package routers

// 接收前端请求并分发给后端API
import (
	"FlashSale/api"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	var r = gin.Default()
	r.StaticFile("/favicon.ico", "./static/favicon")

	r.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "pong"})
	})

	// 商品API入口
	r.GET("/good", api.GetGoodInfo)

	// 单机API入口
	// 单机分组
	LocalGroup := r.Group("/api/local")
	{
		LocalGroup.GET("/without-lock", api.WithoutLock)
		LocalGroup.GET("/with-lock", api.WithLock)
		LocalGroup.GET("/pcc-read-lock", api.PccReadLock)
		LocalGroup.GET("/pcc-write-lock", api.PccWriteLock)
		LocalGroup.GET("/occ-lock", api.OccLock)
		LocalGroup.GET("/channel", api.Channel)
	}

	// 分布式API入口
	// 分布式分组
	DistributedGroup := r.Group("/api/distributed")
	{
		DistributedGroup.GET("/rush", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{"msg": "success"})
		})
		DistributedGroup.GET("/with-redis-lock", api.WithRedisLock)
		DistributedGroup.GET("/with-mq", api.WithMQ)
		DistributedGroup.GET("/with-etcd", api.WithETCD)
	}

	return r
}
