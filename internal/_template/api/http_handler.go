package api

import (
	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/internal/_template/service"
	"github.com/example/dd-frame/pkg/response"
)

// EntityAPI HTTP Handler（替换为实际名称）
type EntityAPI struct {
	svc *service.EntityAppService
}

// NewEntityAPI 创建 API handler
func NewEntityAPI(svc *service.EntityAppService) *EntityAPI {
	return &EntityAPI{svc: svc}
}

// RegisterRoutes 注册模块路由
func (a *EntityAPI) RegisterRoutes(rg *gin.RouterGroup) {
	group := rg.Group("/entity")
	{
		group.POST("", a.CreateHandler)
		// group.GET("/:id", a.GetHandler)
		// group.POST("/:id/submit", a.SubmitHandler)
	}
}

// CreateHandler 创建实体 handler
func (a *EntityAPI) CreateHandler(c *gin.Context) {
	var input service.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, 400, 40000, "invalid request body")
		return
	}

	output, err := a.svc.CreateEntity(c.Request.Context(), &input)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, output)
}
