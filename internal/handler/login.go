package handler

import (
	"go-order-lite/internal/handler/dto"
	"go-order-lite/internal/service"
	"go-order-lite/pkg/errno"
	"go-order-lite/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Login godoc
// @Summary 用户登录
// @Description 用户名密码登录，返回 JWT Token
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body dto.LoginRequest true "登录参数"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Failure 400 {object} response.Response
// @Router /../login [post]
func Login(c *gin.Context) {
	var req dto.LoginRequest
	// 1. bind & validate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errno.LoginInvalid)
		return
	}
	// 2. call service
	token, err := service.Login(req.Username, req.Password)
	if err != nil {
		c.Error(err)
		return
	}
	// 3. return json
	// c.JSON(http.StatusOK, Success(req))
	c.JSON(http.StatusOK, response.Success(dto.LoginResponse{
		Token: token,
	}))
}
