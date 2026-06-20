package app

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/example/dd-frame/example/order"
	authmodule "github.com/example/dd-frame/internal/auth"
	"github.com/example/dd-frame/middleware"
	"github.com/example/dd-frame/pkg/auth"
)

// Wire 全局 Composition Root
//
// 装配所有模块，注册路由，返回 *gin.Engine。
// 这是唯一知道所有模块的地方。新增模块在此注册。
func Wire(cfg *Config) *gin.Engine {
	r := gin.New()

	// 全局中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())

	jwtMgr := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpiresIn)

	// 公开路由组（无需认证）
	public := r.Group("/api/v1")

	// 认证路由组（需要 JWT）
	v1 := r.Group("/api/v1")
	v1.Use(middleware.RequireAuth(jwtMgr))

	// ---------- 注册业务模块 ----------

	// auth 模块
	if GlobalDB != nil {
		authAPI, permChecker := authmodule.Wire(GlobalDB, jwtMgr, cfg.RBAC.SeedEnabled)
		authAPI.RegisterPublicRoutes(public)
		authAPI.RegisterRoutes(v1)

		// permChecker 可供其他模块路由使用，例如：
		// orderGroup.Use(middleware.RequirePermission(permChecker, "order:read"))
		_ = permChecker
	}

	// 订单模块
	orderAPI := order.Wire()
	orderAPI.RegisterRoutes(v1)

	// ---------- 新增模块在此追加 ----------
	// example:
	// productAPI := product.Wire()
	// productAPI.RegisterRoutes(v1)

	// Swagger UI（非 release 模式可用）
	if cfg.Server.Mode != "release" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return r
}
