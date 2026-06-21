package app

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	applog "github.com/example/dd-frame/pkg/log"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

// HealthHandler 存活检查（liveness）
//
// 始终返回 200，表示进程存活。
// K8s livenessProbe 使用。
//
//	@Summary	存活检查
//	@Tags		health
//	@Produce	json
//	@Success	200	{object}	HealthResponse
//	@Router		/health [get]
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}

// ReadyHandler 就绪检查（readiness）
//
// 检测 DB / Redis 连通性，任一不通返回 503。
// K8s readinessProbe 使用。
//
//	@Summary	就绪检查
//	@Tags		health
//	@Produce	json
//	@Success	200	{object}	HealthResponse
//	@Failure	503	{object}	HealthResponse
//	@Router		/ready [get]
func ReadyHandler(c *gin.Context) {
	checks := make(map[string]string)
	degraded := false

	// 检查数据库
	if GlobalDB != nil {
		sqlDB, err := GlobalDB.DB()
		if err != nil {
			checks["database"] = "fail: " + err.Error()
			degraded = true
		} else if err := sqlDB.PingContext(context.Background()); err != nil {
			checks["database"] = "fail: " + err.Error()
			degraded = true
		} else {
			checks["database"] = "ok"
		}
	} else {
		checks["database"] = "skipped"
	}

	// 检查 Redis
	if GlobalRedis != nil {
		if err := GlobalRedis.Ping(context.Background()).Err(); err != nil {
			checks["redis"] = "fail: " + err.Error()
			degraded = true
		} else {
			checks["redis"] = "ok"
		}
	} else {
		checks["redis"] = "skipped"
	}

	status := "ok"
	httpCode := http.StatusOK
	if degraded {
		status = "degraded"
		httpCode = http.StatusServiceUnavailable
		applog.Warn("readiness check degraded", "checks", checks)
	}

	c.JSON(httpCode, HealthResponse{
		Status: status,
		Checks: checks,
	})
}

// RegisterHealthRoutes 注册健康检查路由
func RegisterHealthRoutes(r *gin.Engine) {
	r.GET("/health", HealthHandler)
	r.GET("/ready", ReadyHandler)
}
