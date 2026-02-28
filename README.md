# Go Order Lite

> 一个基于 Go (Gin) + MySQL + Redis 构建的轻量级、高性能订单系统后端。

本项目旨在实现一个具备高可用特性和规范工程结构的现代 Web 后端系统，涵盖了 JWT 鉴权、并发安全控制、缓存策略以及优雅启停等核心微服务架构实践。

## 核心技术亮点

- **防重复下单（幂等性）**：通过客户端传递 `X-Request-Id` 并在 HTTP 头校验，结合 MySQL 唯一索引与 Redis `SetNX` 分布式锁，彻底解决高并发下的重复下单问题。
- **高性能延迟队列**：基于 Redis `ZSet` (Sorted Set) 实现可靠的订单超时自动取消机制，替代传统的数据库轮询，大幅降低 DB 压力。
- **缓存穿透与雪崩防御**：在用户信息查询场景中引入 Redis 缓存，针对空值缓存 `null` 设置较短过期时间，有效防止缓存穿透攻击。
- **全局链路追踪与统一异常处理**：实现自定义中间件，自动向 Context 注入 `request_id` 并串联 Zap 日志；封装全局错误拦截器，标准化 API 响应结构 (`code`, `msg`, `data`)。
- **优雅启停：通过 `os/signal` 监听 `SIGINT`/`SIGTERM` 信号，结合 `context.WithTimeout`，确保 HTTP Server 和后台队列消费者在容器销毁前安全退出，不丢失正在处理的请求。

## 技术栈

- **Web 框架**: [Gin](https://github.com/gin-gonic/gin) (v1.11.0)
- **数据库 ORM**: [GORM](https://gorm.io/) (v1.31.1) + MySQL 8.0
- **缓存 & 队列**: [go-redis](https://github.com/redis/go-redis) (v9) + Redis 7.0
- **身份认证**: Golang-JWT
- **日志组件**: Uber Zap
- **配置管理**: Viper (YAML)
- **API 文档**: Swaggo (gin-swagger)

## 核心目录结构

```text
.
├── cmd/server/          # 程序入口 (main.go)
├── config/              # 配置文件 (config.yaml)
├── docs/                # Swagger API 自动生成文档
├── internal/
│   ├── dao/             # 数据访问层 (MySQL DB 操作)
│   ├── handler/         # HTTP 路由处理与入参校验
│   ├── middleware/      # Gin 全局中间件 (JWT, RequestID, ErrorHandler)
│   ├── model/           # 数据库映射实体 (GORM Model)
│   ├── server/          # 路由注册与 HTTP 服务组装
│   └── service/         # 核心业务逻辑层 (缓存、状态机、分布式锁)
└── pkg/                 # 公共基础设施 (Logger, Redis, MySQL, JWT, Config)
```

## 快速部署

### 1. 环境依赖

- Go 1.24+
- MySQL 5.7+
- Redis 6.0+

### 2. 克隆与配置

```bash
git clone [https://github.com/Phantomor/go-order-lite.git](https://github.com/Phantomor/go-order-lite.git)
cd go-order-lite

# 安装依赖
go mod tidy
```

修改配置文件 `config/config.yaml`，填入你的 MySQL 和 Redis 连接信息：

```YAML
server:
  port: 8080
log:
  level: info
mysql:
  dsn: "root:123456@tcp(127.0.0.1:3306)/go_order_lite?charset=utf8mb4&parseTime=True&loc=Local"
```

### 3. 启动服务

数据库表结构会在启动时通过 GORM `AutoMigrate` 自动创建。

```bash
go run cmd/server/main.go
```

看到 `[INFO] http server start {"port": 8080}` 即代表启动成功。

## 接口文档

项目集成了 Swagger UI。服务启动后，在浏览器访问以下地址即可在线调试 API： 👉 **http://localhost:8080/swagger/index.html**

主要接口：

- `POST /register`: 用户注册
- `POST /login`: 获取 JWT Token
- `GET /user/info`: 获取用户信息
- `POST /api/order`: 创建订单 (需 Header: `Authorization: Bearer <Token>` & `X-Request-Id`)
- `GET /api/order`: 查询订单列表 (支持分页)(需 Query:page和page_size)
- `GET /api/order/:id/pay`: 模拟支付订单
- `GET /api/order/:id/cancel`: 模拟取消订单
