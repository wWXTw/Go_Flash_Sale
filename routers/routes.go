package routers

import "github.com/gin-gonic/gin"

// ready to overhaul...
func NewRouter() *gin.Engine {
	var r = gin.Default()
	return r
}
