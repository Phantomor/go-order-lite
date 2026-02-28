package main

// @title Go Order Lite API
// @version 1.0
// @description 一个用于学习的订单系统后端（Gin + JWT + Redis + MySQL）
// @termsOfService http://example.com

// @contact.name Pvr1sC
// @contact.email phantomor@163.com

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-order-lite/internal/server"
	"go-order-lite/internal/service"
	"go-order-lite/pkg/config"
	"go-order-lite/pkg/logger"
	"go-order-lite/pkg/mysql"
	"go-order-lite/pkg/redis"

	"go.uber.org/zap"
)

func main() {
	// 1. 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	// 2. 初始化日志
	if err := logger.Init(config.Cfg.Log.Level); err != nil {
		log.Fatalf("init logger failed: %v", err)
	}
	defer logger.Log.Sync()

	// 2.1 启动Mysql
	if err := mysql.Init(); err != nil {
		log.Fatalf("mysql init failed: %v", err)
	}
	// 2.2 启动Redis
	if err := redis.InitRedis(); err != nil {
		log.Fatalf("redis init failed: %v", err)
	}
	// 3 创建「根 context」
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4 启动后台任务（Redis 延迟队列消费者）
	go service.StartDelayQueueConsumer(rootCtx)

	// 5. 创建 HTTP Server
	srv := server.NewHTTPServer(config.Cfg.Server.Port)

	// 6 启动 HTTP 服务（goroutine）
	go func() {
		logger.Log.Info("http server start", zap.Int("port", config.Cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("listen failed", zap.Error(err))
		}
	}()

	// 7 监听退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Log.Info("shutdown signal received")

	// 8 优雅关闭（关键）
	ctx, timeoutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer timeoutCancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("server shutdown failed", zap.Error(err))
	}

	logger.Log.Info("server exited gracefully")
}
