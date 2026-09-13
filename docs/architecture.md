# 类 12306 高并发铁路票务系统｜架构设计 v1

# 1. 文档目的

本文件定义第一版系统架构，包括：

* 技术栈
* Go 项目结构
* 模块划分
* 分层规范
* MySQL / Redis / Nginx 的位置
* 模块之间的依赖关系
* 开发阶段演进路线
* Docker / Kubernetes 后续接入方式

本文件不重新定义业务规则。

所有实现必须遵守：

```text
PRD.md
business-rules.md
database.md
api.md
```

如果架构实现与业务规则冲突：

> 以业务规则为准。

---

# 2. 第一版总体原则

第一版采用：

```text
模块化单体
Modular Monolith
```

不一开始拆微服务。

原因：

```text
业务还没跑通
没有真实性能瓶颈
没有必要提前引入分布式复杂度
```

第一阶段目标：

```text
先把完整票务交易链跑通
```

然后再逐步加入：

```text
Redis
Docker
Kubernetes
高并发优化
监控
CI/CD
```

---

# 3. 第一版技术栈

后端：

```text
Go
Gin
```

数据库：

```text
MySQL 8.x
```

缓存：

```text
Redis
```

反向代理：

```text
Nginx
```

数据库访问：

```text
database/sql
或
sqlx
```

第一版建议：

```text
sqlx
```

本项目本身需要学习：

```text
SQL
索引
事务
锁
```

---

# 4. 第一阶段系统架构

```text
┌─────────────────────┐
│      Browser        │
│   Web Frontend      │
└──────────┬──────────┘
           │ HTTP
           ▼
┌─────────────────────┐
│        Nginx        │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│      Go Backend     │
│        Gin          │
└───────┬─────┬───────┘
        │     │
        │     │
        ▼     ▼
     MySQL   Redis
```

其中：

```text
Nginx
负责入口、反向代理

Go Backend
负责业务逻辑

MySQL
作为核心业务数据真值

Redis
作为缓存和临时状态辅助组件
```

---

# 5. 数据真值原则

第一版必须明确：

```text
MySQL
=
核心业务真值
```

Redis 不能在第一阶段成为唯一数据源。

例如：

```text
订单状态
支付状态
座位最终状态
车次配置
```

最终都必须能够从 MySQL 中恢复。

Redis 第一阶段主要承担：

```text
缓存
会话
热点数据
临时状态
```

后续做高并发优化时，再进一步讨论 Redis 是否参与抢票核心链路。

---

# 6. Go 项目目录

建议目录：

```text
ticket-system/
│
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   │
│   ├── auth/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── request.go
│   │   └── response.go
│   │
│   ├── user/
│   │
│   ├── station/
│   │
│   ├── train/
│   │
│   ├── seat/
│   │
│   ├── order/
│   │
│   ├── payment/
│   │
│   ├── ticket/
│   │
│   └── admin/
│
├── internal/
│   ├── model/
│   ├── middleware/
│   ├── platform/
│   ├── scheduler/
│   └── common/
│
├── configs/
│
├── migrations/
│
├── scripts/
│
├── deployments/
│   ├── docker/
│   └── kubernetes/
│
├── docs/
│
├── tests/
│
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

---

# 7. cmd/server

入口：

```text
cmd/server/main.go
```

负责：

```text
读取配置
初始化日志
连接 MySQL
连接 Redis
初始化 Repository
初始化 Service
初始化 Handler
注册路由
启动 HTTP Server
```

禁止把业务逻辑写进：

```text
main.go
```

main.go 只负责组装应用。

---

# 8. 模块划分

第一版核心模块：

```text
auth
user
station
train
seat
order
payment
ticket
admin
```

---

# 9. auth 模块

负责：

```text
注册
登录
退出
身份认证
Token
```

不负责：

```text
订单
车次
支付
座位
```

---

# 10. user 模块

负责：

```text
用户信息
昵称
密码修改
```

第一版不负责：

```text
常用乘车人
实名中心
头像
风控
```

---

# 11. station 模块

负责：

```text
站点查询
创建站点
修改站点
停用站点
```

---

# 12. train 模块

负责：

```text
创建车次
修改草稿车次
发布车次
开售状态
停售状态
发车状态
归档
车次查询
```

Train 模块不直接处理：

```text
支付
出票
退票
```

---

# 13. seat 模块

Seat 是核心模块之一。

负责：

```text
创建座位
查询座位
座位分配
座位锁定
座位出售
座位释放
余票统计
```

必须遵守：

```text
AVAILABLE
LOCKED
SOLD
```

状态机。

Seat 模块不能自己决定：

```text
订单是否支付成功
```

它只接受业务动作：

```text
lock
sell
release
```

---

# 14. order 模块

Order 是整个系统最核心模块。

负责：

```text
创建车票订单
取消待支付订单
订单查询
退票
订单状态流转
订单超时
订单完成
```

核心状态：

```text
WAITING_PAYMENT
TICKETED
CANCELLED
RETURNED
COMPLETED
```

OrderService 是主要业务编排层。

---

# 15. payment 模块

Payment 负责：

```text
创建支付单
发起支付
支付成功
支付失败
退款
支付状态查询
Mock Payment
支付宝 Sandbox
```

核心状态：

```text
UNPAID
PAYING
SUCCESS
FAILED
REFUNDING
REFUNDED
```

Payment 模块负责：

```text
钱
```

Order 模块负责：

```text
票
```

两者禁止混为同一个状态。

---

# 16. ticket 模块

Ticket 模块主要负责：

```text
电子票查询
电子票展示数据组装
身份证脱敏
票有效性判断
```

第一版电子票本质上来源于：

```text
TicketOrder
Train
Passenger
Seat
```

不一定需要额外创建一张独立 `tickets` 数据表。

---

# 17. admin 模块

管理员模块负责：

```text
管理员认证
管理员后台接口
异常操作入口
审计日志
```

管理员不能直接修改数据库状态。

禁止提供：

```text
setOrderStatus(...)
setSeatStatus(...)
```

这种万能接口。

所有状态变化必须走合法业务动作。

---

# 18. 分层设计

每个模块遵循：

```text
Handler
↓
Service
↓
Repository
↓
MySQL / Redis
```

---

# 19. Handler

负责：

```text
接收 HTTP Request
参数绑定
基础参数校验
调用 Service
返回 HTTP Response
```

Handler 不写：

```text
SQL
事务
复杂业务判断
座位分配算法
状态流转
```

错误示例：

```text
OrderHandler
直接 UPDATE seats
```

禁止。

---

# 20. Service

Service 是业务核心。

负责：

```text
业务规则
状态校验
事务边界
模块编排
```

例如：

```text
OrderService.CreateOrder()
```

应该负责：

```text
检查用户是否有待支付订单
检查车次状态
检查乘车人购票资格
分配座位
锁座
创建 TicketOrder
创建 Payment
```

---

# 21. Repository

负责：

```text
数据库读写
SQL
数据持久化
```

例如：

```text
SeatRepository
TrainRepository
OrderRepository
PaymentRepository
```

Repository 不负责：

```text
业务决策
```

例如 Repository 不应该决定：

```text
这个人能不能退票
```

它只负责：

```text
读取订单
更新订单
```

---

# 22. Model

`internal/model/`

负责数据库实体结构。

例如：

```go
type TicketOrder struct {
    ID              string
    UserID          uint64
    PassengerID     uint64
    TrainID         uint64
    SeatID          uint64
    TicketPrice     decimal.Decimal
    Status          string
    PaymentDeadline time.Time
}
```

Model 不负责业务逻辑。

---

# 23. Request / Response DTO

API 请求和数据库 Model 不直接共用。

例如：

```text
CreateOrderRequest
```

和：

```text
TicketOrder
```

是两个不同结构。

原因：

```text
HTTP 输入模型
≠
数据库模型
```

这样可以避免前端直接控制数据库字段。

---

# 24. 创建订单调用链

核心流程：

```text
POST /orders
      ↓
OrderHandler
      ↓
OrderService.CreateOrder()
      ↓
检查 User
      ↓
检查 Train
      ↓
检查 Passenger
      ↓
SeatService.Allocate()
      ↓
锁定 Seat
      ↓
创建 TicketOrder
      ↓
PaymentService.Create()
      ↓
Commit
```

最终得到：

```text
TicketOrder = WAITING_PAYMENT
Seat = LOCKED
Payment = UNPAID
```

---

# 25. 创建订单事务边界

第一版创建订单时：

```text
分配座位
锁座
创建 TicketOrder
创建 Payment
建立 payment_orders
```

这些核心动作应当作为一个完整业务事务考虑。

原则：

```text
要么全部成功
要么全部失败
```

不能出现：

```text
座位已经锁住
但订单没创建
```

或者：

```text
订单创建成功
但没有 Payment
```

具体事务实现放到开发阶段。

---

# 26. 支付成功调用链

Mock 阶段：

```text
POST /payments/{id}/mock/success
              ↓
PaymentHandler
              ↓
PaymentService.MarkSuccess()
              ↓
Payment = SUCCESS
              ↓
OrderService.MarkTicketed()
              ↓
TicketOrder = TICKETED
              ↓
SeatService.MarkSold()
              ↓
Seat = SOLD
```

这几个状态必须作为一个一致性操作处理。

---

# 27. 取消订单调用链

```text
POST /orders/{id}/cancel
          ↓
OrderHandler
          ↓
OrderService.Cancel()
          ↓
验证 WAITING_PAYMENT
          ↓
TicketOrder = CANCELLED
          ↓
SeatService.Release()
          ↓
Seat = AVAILABLE
          ↓
重新计算 Payment payable_amount
```

---

# 28. 退票调用链

```text
POST /orders/{id}/return
          ↓
OrderHandler
          ↓
OrderService.Return()
          ↓
验证 TICKETED
          ↓
验证退票时间
          ↓
TicketOrder = RETURNED
          ↓
Seat = AVAILABLE
          ↓
PaymentService.Refund()
```

注意：

```text
退票业务完成
和
资金退款完成
```

不能混成一个动作。

---

# 29. 定时任务

系统需要定时处理几类业务。

第一版：

```text
车次定时开售
车次自动停售
车次发车状态变化
订单自动完成
待支付订单超时
```

放在：

```text
internal/scheduler/
```

例如：

```text
sale_scheduler.go
order_timeout_scheduler.go
train_status_scheduler.go
```

---

# 30. 待支付订单超时任务

业务：

```text
WAITING_PAYMENT
+
超过 payment_deadline
```

则：

```text
TicketOrder → CANCELLED
Seat → AVAILABLE
```

第一版可以先通过：

```text
定时扫描 MySQL
```

实现。

暂时不需要一开始就上：

```text
延迟队列
MQ
Redis Keyspace Notification
```

先跑通。

---

# 31. Redis 第一阶段职责

人类开发者有话说：不一定一上来就上redis，先跑通才是关键，像lnmp一样，能跑是关键。你只负责开发代码

Redis 第一阶段可以用于：

```text
登录状态
热点车次缓存
余票查询缓存
```

暂时不要一开始就把：

```text
座位真实状态
订单真实状态
支付真实状态
```

只放 Redis。

---

# 32. Redis Key 示例

登录：

```text
session:{token}
```

热点车次：

```text
train:{train_id}
```

路线查询：

```text
trains:{date}:{departure_station_id}:{arrival_station_id}
```

余票：

```text
availability:{train_id}:{seat_class}
```

具体缓存结构后续再设计。

---

# 33. Redis 故障原则

第一阶段设计目标：

> Redis 挂掉，系统可以变慢，但核心业务数据不能丢。

因此：

```text
MySQL 是真值
Redis 是辅助
```

这是 v1 的基本可靠性边界。

---

# 34. Nginx 第一阶段职责

人类开发者有话说：Nginx、redis、mysql不是你要担心的，你的任务是写代码，我负责nginx这些，咱两配合

Nginx：

```text
监听 80 / 443
反向代理 Go Backend
静态资源
基础日志
```

例如：

```text
client
↓
Nginx
↓
Go :8080
```

第一阶段不要求：

```text
复杂限流
WAF
多级缓存
```

后续压测后再加。

---

# 35. 配置管理

程序配置不硬编码。

例如：

```text
configs/config.yaml
```

或环境变量。

配置项：

```text
server.port

mysql.host
mysql.port
mysql.database
mysql.username
mysql.password

redis.host
redis.port
redis.password

auth.token_expire

payment.provider
```

生产环境密码：

```text
禁止提交 Git
```

---

# 36. 日志

后端至少记录：

```text
request_id
HTTP method
path
status
latency
user_id
error
```

核心业务操作额外记录：

```text
order_id
payment_id
train_id
seat_id
```

方便以后排障。

---

# 37. Request ID

每一个请求生成：

```text
request_id
```

例如：

```text
req-fab17c...
```

日志链：

```text
Nginx
↓
Go Handler
↓
Service
↓
Repository
```

尽量带同一个 request_id。

这对后面的：

```text
日志排障
链路追踪
```

很重要。

---

# 38. 错误处理

Repository 层：

```text
返回数据库错误
```

Service 层：

```text
转成业务错误
```

Handler 层：

```text
转成 HTTP Response
```

例如：

```text
数据库查不到 Train
↓
Service
TRAIN_NOT_FOUND
↓
Handler
HTTP 404
```

---

# 39. 前端

前端不是本项目重点。

第一版目标：

```text
能正常完成业务流程
```

页面至少包括：

```text
登录
注册
查票
车次列表
购票确认
待支付
订单列表
订单详情
电子票
个人信息
管理员后台
```

前端可以主要交给 AI Agent 实现。

---

# 40. 前端与后端边界

前端负责：

```text
展示
输入
交互
```

后端负责：

```text
业务规则
状态校验
权限
价格
座位
订单
支付
```

前端不能决定：

```text
订单状态
座位号
票价
支付金额
是否允许退票
```

所有这些必须以后端为准。

---

# 41. 第一阶段开发顺序

建议：

```text
Phase 1
Go 项目骨架
MySQL 连接（人类：只写接口？我负责数据库运维）
配置系统
日志

↓

Phase 2
User / Auth

↓

Phase 3
Station / Train

↓

Phase 4
Seat

↓

Phase 5
TicketOrder

↓

Phase 6
Mock Payment

↓

Phase 7
Cancel / Timeout

↓

Phase 8
Return / Refund

↓

Phase 9
电子票

↓

Phase 10
管理员后台
```

---

# 42. AI Agent 分工建议

机械 CRUD：

```text
Auth
User
Station
Train Admin
Admin List
前端页面
```

可以主要由 Agent 生成。

人需要重点参与：

```text
Seat
Order
Payment
事务
状态机
超时
退票
退款
```

因为这些是项目核心。

---

# 43. 人必须理解的代码

即使 Agent 写，也必须真正看懂：

```text
OrderService.CreateOrder()
SeatService.Allocate()
SeatService.Lock()
PaymentService.Pay()
PaymentService.Refund()
OrderService.Cancel()
OrderService.Return()
订单超时任务
```

这些代码以后很可能是面试重点。

---

# 44. Docker 阶段（先别理这个，先把go写好，我自会打包docker）

单机业务跑通后：

```text
Go Backend
MySQL
Redis
Nginx
```

分别容器化。

结构：

```text
docker compose
│
├── nginx
├── backend
├── mysql
└── redis
```

目标：

```text
docker compose up -d
```

可以完整启动系统。

---

# 45. Go Docker 编译

使用多阶段构建：

```text
golang builder
↓
go build
↓
得到二进制
↓
复制进入运行镜像
```

最终运行：

```text
./ticket-server
```

这样保留：

```text
编译
打包
部署
```

完整工程过程。

---

# 46. Kubernetes 阶段（下面同样，我自己会弄，你记住你是开发，和我运维配合）

Docker Compose 稳定后再迁移 Kubernetes。

第一阶段目标架构：

```text
Ingress
   ↓
Service
   ↓
Go Backend Deployment
      ├── Pod
      ├── Pod
      └── Pod
```

后端需要设计成：

```text
无状态
```

Pod 本地不能保存：

```text
订单状态
登录状态
座位状态
```

这些应该进入：

```text
MySQL
Redis
```

---

# 47. Kubernetes 第一批资源

后续需要：

```text
Deployment
Service
Ingress
ConfigMap
Secret
HPA
```

MySQL / Redis 第一阶段可以根据实验环境决定：

```text
集群内
或
外部服务
```

不在业务开发第一阶段决定。

---

# 48. 多副本问题

当 Go Backend 从：

```text
1 个实例
```

变成：

```text
3 个 Pod
```

后，本项目才真正进入高并发 / 分布式阶段。

这时候必须重新审视：

```text
座位锁
订单幂等
支付幂等
Redis
数据库锁
```

因为：

```text
Go 进程内 mutex
```

只能控制：

```text
一个进程
```

不能控制多个 Pod。

---

# 49. 高并发优化阶段

业务正确后再进行：

```text
压测
↓
找到瓶颈
↓
优化
```

可能涉及：

```text
Redis 缓存
数据库事务优化
悲观锁
乐观锁
Redis 原子操作
Lua
分布式锁
限流
连接池
MQ
```

但不能为了“技术栈多”提前全部塞进去。

---

# 50. 压测目标

后续至少压测：

```text
车次查询
余票查询
创建订单
抢票
支付
```

观察：

```text
QPS
平均响应时间
P95
P99
错误率
CPU
内存
数据库连接数
Redis
```

---

# 51. 可观测性阶段

后续加入：

```text
Prometheus
Grafana
Alertmanager
```

重点监控：

```text
HTTP QPS
HTTP latency
HTTP 5xx
Go runtime
CPU
Memory
Pod
MySQL
Redis
```

监控不属于产品管理员后台。

---

# 52. CI/CD 阶段

后续目标：

```text
git push
↓
CI
↓
go test
↓
go build
↓
docker build
↓
push image
↓
deploy Kubernetes
↓
RollingUpdate
```

新版本异常：

```text
kubectl rollout undo
```

---

# 53. 故障演练阶段

后期需要主动制造故障：

```text
杀 Pod
Redis 故障
MySQL 故障
Nginx upstream 错误
高 CPU
高并发
错误版本发布
```

验证：

```text
发现
告警
定位
处理
恢复
```

---

# 54. 第一版明确禁止的架构膨胀

在业务没有跑通之前，禁止为了炫技术主动加入：

```text
微服务
Kafka
RabbitMQ
Service Mesh
分库分表
Elasticsearch
复杂网关
分布式事务框架
多级缓存
事件总线
CQRS
Event Sourcing
```

后续只有真实问题出现时才引入。

---

# 55. 架构演进路线

完整路线：

```text
业务设计
↓
Go 模块化单体
↓
MySQL
↓
Mock Payment
↓
完整交易链
↓
Redis
↓
Docker
↓
Docker Compose
↓
压力测试
↓
问题暴露
↓
高并发优化
↓
Kubernetes
↓
多副本
↓
Prometheus / Grafana
↓
Alertmanager
↓
CI/CD
↓
故障演练
↓
支付宝 Sandbox
```

---

# 56. 第一版架构最终形态

业务开发阶段：

```text
Browser
   ↓
Nginx
   ↓
Go Backend
   ├── MySQL
   └── Redis
```

部署演进后：

```text
                    Internet
                        │
                        ▼
                 Nginx / Ingress
                        │
                        ▼
                 Kubernetes Service
                        │
              ┌─────────┼─────────┐
              ▼         ▼         ▼
           Go Pod     Go Pod     Go Pod
              │         │         │
              └─────────┼─────────┘
                        │
              ┌─────────┴─────────┐
              ▼                   ▼
            MySQL               Redis

                        │
                        ▼
                Prometheus
                        │
                        ▼
                  Grafana
                        │
                        ▼
                 Alertmanager
```

---

# 57. 架构核心原则

整个项目坚持：

```text
先业务正确
再高并发

先单体
再分布式

先出现问题
再引入技术

MySQL 保证核心数据
Redis 优化性能

订单管票
支付管钱
座位管库存
```

这就是 v1 的总体架构基线。

