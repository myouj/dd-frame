package app

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/example/dd-frame/example/order"
	authmodule "github.com/example/dd-frame/internal/auth"
	"github.com/example/dd-frame/middleware"
	"github.com/example/dd-frame/pkg/auth"
	"github.com/example/dd-frame/pkg/storage"

	applog "github.com/example/dd-frame/pkg/log"
)

// GlobalStore 全局存储实例（启动时初始化）
var GlobalStore storage.Store

// Wire 全局 Composition Root
//
// 装配所有模块，注册路由，返回 *gin.Engine。
// 这是唯一知道所有模块的地方。新增模块在此注册。
func Wire(cfg *Config) *gin.Engine {
	r := gin.New()

	// 全局中间件（按顺序注册）
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 分布式追踪中间件（在 RequestID 前，traceID 优先）
	if cfg.Tracing.Enabled {
		r.Use(otelgin.Middleware(cfg.Tracing.ServiceName))
	}

	r.Use(middleware.RequestID())

	// 请求限流中间件
	r.Use(middleware.RateLimiter(
		cfg.RateLimit.Enabled,
		cfg.RateLimit.Backend,
		cfg.RateLimit.Rate,
		cfg.RateLimit.Burst,
		GlobalRedis,
		cfg.RateLimit.KeyPrefix,
	))

	// Prometheus 指标中间件
	if cfg.Metrics.Enabled {
		r.Use(middleware.Metrics())
	}

	r.Use(middleware.Logger())

	jwtMgr := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpiresIn)

	// 基础设施路由（无需认证）
	RegisterHealthRoutes(r)
	RegisterMetricsRoute(r, &cfg.Metrics)

	// 初始化存储
	initStore(&cfg.Storage)

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

	// 文件上传路由示例（需要认证）
	if GlobalStore != nil {
		v1.POST("/upload", middleware.FileUpload(GlobalStore, 10<<20, nil), func(c *gin.Context) {
			result := middleware.GetUploadResult(c)
			c.JSON(200, gin.H{
				"code": 0,
				"data": result,
			})
		})
	}

	// Swagger UI（非 release 模式可用）
	if cfg.Server.Mode != "release" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return r
}

// initStore 根据配置初始化存储后端
func initStore(cfg *StorageConfig) {
	if cfg.Driver == "" {
		return
	}

	switch cfg.Driver {
	case "local":
		store, err := storage.NewLocalStore(cfg.Local.BaseDir, cfg.Local.BaseURL)
		if err != nil {
			applog.Error("init local storage failed", "err", err)
			return
		}
		GlobalStore = store
		applog.Info("storage: local", "base_dir", cfg.Local.BaseDir)

	case "oss":
		store, err := storage.NewOSSStore(storage.OSSConfig{
			Endpoint:        cfg.OSS.Endpoint,
			AccessKeyID:     cfg.OSS.AccessKeyID,
			AccessKeySecret: cfg.OSS.AccessKeySecret,
			Bucket:          cfg.OSS.Bucket,
			Prefix:          cfg.OSS.Prefix,
			CDNURL:          cfg.OSS.CDNURL,
		})
		if err != nil {
			applog.Error("init oss storage failed", "err", err)
			return
		}
		GlobalStore = store
		applog.Info("storage: oss", "bucket", cfg.OSS.Bucket)

	default:
		applog.Warn("unknown storage driver", "driver", cfg.Driver)
	}
}
