# Go Order Lite (V2.0)

> 一个基于 Go (Gin) + MySQL + Redis + RocketMQ 构建的轻量级、高可用电商订单系统后端。

本项目在常规 Web 后端的基础上，融入了**消息队列（事件驱动）**架构。不仅实现了基础的 JWT 鉴权与缓存策略，更针对高并发场景下的“重复下单”、“订单超时取消的性能瓶颈”以及“支付下游业务耦合”等经典痛点，提供了完整的工业级解决方案。

## 核心技术亮点

- **微服务异步解耦 (Pub/Sub)**：重构支付成功链路，将原有的同步 HTTP 逻辑拆分为事件发布。支付网关仅负责更新 DB 并向 RocketMQ 投递 `ORDER_PAID` 事件（耗时 < 50ms），下游的积分服务、短信服务通过不同的 Consumer Group 并行订阅消费，彻底消除跨服务调用的单点故障与延迟瓶颈。
- **零内存压力的延迟队列**：彻底废弃基于 Redis ZSet 的轮询方案，全面接入 **RocketMQ 原生延迟消息**实现订单超时自动取消 (`ORDER_DELAY_CANCEL`)。将千万级未支付订单的倒计时任务下沉至 MQ 磁盘，彻底释放宝贵的 Redis 内存，并消除应用层定时轮询的 CPU 损耗。
- **全链路严密防重与幂等性保证**：
  - **API 接口层**：通过客户端传递 `X-Request-Id` 并在 HTTP 头校验，结合 Redis `SetNX` 护航，防止前端手抖造成的重复下单。
  - **MQ 消费层**：针对网络抖动导致的 MQ 重复投递问题，在下游消费者（如积分发放）中注入基于订单 ID 的 Redis 分布式锁 (`consume:point:order:{id}`)。配合 `continue` 机制与 DB 状态机兜底，实现严格的消费端幂等，确保用户的积分“一分不多，一分不少”。
- **缓存穿透与雪崩防御**：在用户信息查询场景中引入 Redis 缓存，针对空值缓存 `null` 设置较短过期时间，有效防止缓存穿透攻击。
- **优雅启停与资源回收**：通过 `os/signal` 监听 `SIGINT`/`SIGTERM` 信号，结合 `context.WithTimeout`，确保 HTTP Server 以及底层的 **RocketMQ 生产者/消费者** 在容器销毁前安全排空并退出，拒绝丢失任何处理中的消息。

## 技术栈

- **Web 框架**: [Gin](https://github.com/gin-gonic/gin) (v1.11.0)
- **数据库 ORM**: [GORM](https://gorm.io/) (v1.31.1) + MySQL 8.0
- **缓存组件**: [go-redis](https://github.com/redis/go-redis) (v9) + Redis 7.0
- **消息队列**: **[RocketMQ](https://rocketmq.apache.org/) (4.x) + apache/rocketmq-client-go/v2**
- **身份认证**: Golang-JWT
- **日志组件**: Uber Zap
- **配置管理**: Viper (YAML)

## 核心目录结构

```text
.
├── cmd/server/          # 程序入口 (注入组件，优雅启停)
├── config/              # 配置文件 (config.yaml)
├── docs/                # Swagger API 自动生成文档
├── internal/
│   ├── dao/             # 数据访问层 (MySQL DB 操作)
│   ├── handler/         # HTTP 路由处理与入参校验
│   ├── middleware/      # Gin 全局中间件 (JWT, RequestID, ErrorHandler)
│   ├── model/           # 数据库映射实体 (GORM Model)
│   ├── mq/              # RocketMQ 消费者群组 (订单取消、短信、积分)
│   ├── server/          # 路由注册与 HTTP 服务组装
│   └── service/         # 核心业务逻辑层 (状态机、缓存、发送 MQ 消息)
└── pkg/                 # 公共基础设施 (Logger, Redis, MySQL, JWT, Config, MQ 客户端)
```

## 快速部署

### 1. 环境依赖

- Go 1.24+
- MySQL 5.7+
- Redis 6.0+
- RocketMQ 4.x (NameServer & Broker)

### 2. 克隆与配置

```bash
git clone https://github.com/Phantomor/go-order-lite.git
cd go-order-lite

# 安装依赖
go mod tidy
```

修改配置文件 `config/config.yaml`，填入你的中间件连接信息：

```yaml
server:
  port: 8080
log:
  level: info
mysql:
  dsn: "root:123456@tcp(127.0.0.1:3306)/go_order_lite?charset=utf8mb4&parseTime=True&loc=Local"
rocketmq:
  name_servers:
    - "127.0.0.1:9876"
  retry: 2
```

### 3. 启动服务

数据库表结构会在启动时通过 GORM `AutoMigrate` 自动创建。

```bash
go run cmd/server/main.go
```

看到 `[INFO] http server start` 以及 `RocketMQ Producer started successfully` 即代表启动成功。

## 接口文档

项目集成了 Swagger UI。服务启动后，在浏览器访问以下地址即可在线调试 API： 👉 **http://localhost:8080/swagger/index.html**

主要接口：

- `POST /register`: 用户注册
- `POST /login`: 获取 JWT Token
- `GET /user/info`: 获取用户信息
- `POST /api/order`: 创建订单 (需 Header: `Authorization: Bearer <Token>` & `X-Request-Id`)
- `GET /api/order`: 查询订单列表 (支持分页)(需 Query:page和page_size)
- `GET /api/order/:id/pay`: 模拟支付订单 (触发异步积分与短信下发)
- `GET /api/order/:id/cancel`: 模拟主动取消订单