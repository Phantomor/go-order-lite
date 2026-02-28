package handler

import (
	"go-order-lite/internal/service"
	"go-order-lite/pkg/errno"
	"go-order-lite/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
}

// Register godoc
// @Summary 用户注册
// @Description 注册新用户，需提供用户名和密码
// @Tags user
// @Accept json
// @Produce json
// @Param data body RegisterRequest true "注册参数"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /../register [post]
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errno.InternalError)
		return
	}

	err := service.Register(req.Username, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success("register ok"))
}

// UserInfo godoc
// @Summary 获取用户信息
// @Description 通过 JWT Token 获取当前登录用户的详细信息
// @Tags user
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /user/info [get]
func UserInfo(c *gin.Context) {
	userID := c.GetUint("user_id") // 从 JWT 中间件来的
	if userID == 0 {
		c.Error(errno.Unauthorized)
		return
	}

	user, err := service.GetUserInfo(userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, response.Success(user))
}
