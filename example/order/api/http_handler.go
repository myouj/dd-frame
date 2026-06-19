package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/example/order/service"
	"github.com/example/dd-frame/pkg/response"
)

// OrderAPI 订单 HTTP Handler
type OrderAPI struct {
	svc *service.OrderAppService
}

// NewOrderAPI 创建订单 API handler
func NewOrderAPI(svc *service.OrderAppService) *OrderAPI {
	return &OrderAPI{svc: svc}
}

// RegisterRoutes 注册订单路由
func (a *OrderAPI) RegisterRoutes(rg *gin.RouterGroup) {
	orderGroup := rg.Group("/order")
	{
		orderGroup.POST("", a.CreateOrderHandler)
		orderGroup.POST("/:orderNo/submit", a.SubmitOrderHandler)
		orderGroup.POST("/:orderNo/cancel", a.CancelOrderHandler)
	}
}

// CreateOrderHandler 创建订单 HTTP handler
func (a *OrderAPI) CreateOrderHandler(c *gin.Context) {
	var input service.CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40000, "invalid request body")
		return
	}

	output, err := a.svc.CreateOrder(c.Request.Context(), &input)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, output)
}

// SubmitOrderHandler 提交订单 HTTP handler
func (a *OrderAPI) SubmitOrderHandler(c *gin.Context) {
	orderNo := c.Param("orderNo")
	if err := a.svc.SubmitOrder(c.Request.Context(), orderNo); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}

// CancelOrderHandler 取消订单 HTTP handler
func (a *OrderAPI) CancelOrderHandler(c *gin.Context) {
	orderNo := c.Param("orderNo")
	reason := c.PostForm("reason")
	if err := a.svc.CancelOrder(c.Request.Context(), orderNo, reason); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}
