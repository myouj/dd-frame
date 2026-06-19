package app

import (
	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/example/order"
)

// Wire 全局 Composition Root
//
// 装配所有模块，注册路由，返回 *gin.Engine。
// 这是唯一知道所有模块的地方。新增模块在此注册。
func Wire(cfg *Config) *gin.Engine {
	r := gin.New()

	// API 版本组
	v1 := r.Group("/api/v1")

	// ---------- 注册业务模块 ----------

	// 订单模块
	orderAPI := order.Wire()
	orderAPI.RegisterRoutes(v1)

	// ---------- 新增模块在此追加 ----------
	// example:
	// productAPI := product.Wire()
	// productAPI.RegisterRoutes(v1)

	return r
}
