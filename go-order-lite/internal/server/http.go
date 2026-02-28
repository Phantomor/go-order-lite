package server

import (
	"fmt"
	_ "go-order-lite/docs" // 必须匿名导入
	"go-order-lite/internal/handler"
	"go-order-lite/internal/middleware"
	"net/http"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

func NewHTTPServer(port int) *http.Server {
	r := SetupRouter()

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}
}
func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.RequestContext())
	r.POST("/register", handler.Register)
	r.POST("/login", handler.Login)

	auth := r.Group("/api")
	auth.Use(middleware.JWTAuth())
	{
		auth.GET("/user/info", handler.UserInfo)
		auth.POST("/order", handler.CreateOrder)
		auth.GET("/order", handler.ListMyOrders)
		auth.GET("/order/:id/pay", handler.PayOrder)
		auth.GET("/order/:id/cancel", handler.CancelOrder)
	}

	return r
}
