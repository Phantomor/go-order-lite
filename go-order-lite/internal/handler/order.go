package handler

import (
	"go-order-lite/internal/handler/dto"
	"go-order-lite/internal/service"
	"go-order-lite/pkg/errno"
	"go-order-lite/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateOrder godoc
// @Summary 创建订单
// @Description 用户创建订单（JWT 鉴权 + RequestID 幂等）
// @Tags order
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param X-Request-Id header string true "请求唯一ID（用于防重复下单）"
// @Param data body dto.CreateOrderRequest true "创建订单参数"
// @Success 200 {object} response.Response{data=dto.OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /order [post]
func CreateOrder(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetUint("user_id")
	var req dto.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errno.InvalidParam)
		return
	}
	requestID := c.GetHeader("X-Request-Id")
	if requestID == "" {
		c.Error(errno.MissingRequestId)
		return
	}
	order, err := service.CreateOrder(ctx, userID, req.Amount, requestID)
	if err != nil {
		c.Error(err)
		return
	}

	orderResp := dto.OrderResponse{
		ID:        order.ID,
		UserID:    order.UserID,
		Amount:    order.Amount,
		Status:    dto.OrderStatus(order.Status),
		CreatedAt: order.CreatedAt,
	}

	c.JSON(http.StatusOK, response.Success(orderResp))
}

// ListMyOrders godoc
// @Summary 查询我的订单列表
// @Description 查询当前登录用户的订单（分页）
// @Tags order
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=[]dto.OrderResponse}
// @Failure 401 {object} response.Response
// @Router /order [get]
func ListMyOrders(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.Error(errno.InvalidUser)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	orders, err := service.ListMyOrders(userID, page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(orders))
}

// PayOrder godoc
// @Summary 支付订单
// @Description 模拟用户支付订单
// @Tags order
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "订单ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /order/{id}/pay [get]
func PayOrder(c *gin.Context) {
	userID := c.GetUint("user_id")
	orderID, _ := strconv.Atoi(c.Param("id"))

	if err := service.PayOrder(userID, uint(orderID)); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// CancelOrder godoc
// @Summary 取消订单
// @Description 用户主动取消订单
// @Tags order
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "订单ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /order/{id}/cancel [get]
func CancelOrder(c *gin.Context) {
	userID := c.GetUint("user_id")
	orderID, _ := strconv.Atoi(c.Param("id"))

	if err := service.CancelOrder(userID, uint(orderID)); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}
