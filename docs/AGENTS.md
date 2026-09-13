# AGENTS.md


人类：
“不要防御性编程，不要你写什么兜底、安全。我只要代码逻辑优秀、可读。一次只实现一个小功能，小模块，或者算法。”


# 1. 项目说明

本项目是一个：

```text
类 12306 高并发铁路票务系统
```

第一阶段目标不是完整复刻 12306，而是完成一套可以真实运行的核心票务交易系统。

当前开发重点：

```text
Go 后端
业务状态机
座位分配
订单
支付
退票
退款
```

项目 Owner 是人类开发者。

AI Agent 的角色是：

> 软件开发协作者。

AI Agent 不负责：

```text
服务器运维
Nginx 部署
MySQL 运维
Redis 运维
Docker 部署
Kubernetes
Prometheus
Grafana
CI/CD
服务器安全
生产基础设施
```

这些由人类负责。

---

# 2. 开始开发前必须阅读

任何 Agent 在修改代码前，必须先阅读：

```text
docs/PRD.md
docs/business-rules.md
docs/database.md
docs/api.md
docs/architecture.md
AGENTS.md
```

优先级：

```text
business-rules.md
    ↓
PRD.md
    ↓
database.md
    ↓
api.md
    ↓
architecture.md
    ↓
当前代码
```

如果文档之间存在冲突：

> 停止擅自推断。

应指出冲突，由人类决定。

不要私自“优化”业务规则。

---

# 3. AI 的主要职责

AI Agent 主要负责：

```text
Go 后端代码
HTTP API
业务 Service
Repository
数据 Model
DTO
基础测试
Mock Payment
前端基础页面
```

可以根据现有设计实现：

```text
Auth
User
Station
Train
Seat
Order
Payment
Ticket
Admin
```

---

# 4. 人类负责的内容

以下内容默认由人类处理。

## MySQL

人类负责：

```text
MySQL 安装
MySQL 配置
创建数据库
创建账号
权限
备份
恢复
数据库运维
```

AI 负责：

```text
Go 数据访问代码
Repository
SQL 查询
事务代码
```

`database.md` 中的 SQL 表结构主要是：

> 数据模型设计参考。

AI 不应该假设自己拥有生产数据库管理权限。

---

## Redis

人类负责：

```text
Redis 安装
Redis 配置
Redis 运维
```

AI 仅在明确要求后：

```text
编写 Redis 客户端代码
缓存逻辑
```

第一阶段业务没有跑通前：

> 不要主动引入 Redis 作为核心依赖。

---

## Nginx

人类负责：

```text
Nginx
反向代理
SSL
负载均衡
日志
限流配置
```

AI 不需要主动生成完整 Nginx 运维方案。

后端只需要：

```text
正确监听指定端口
正确处理 HTTP
正确输出日志
```

---

## Docker / Kubernetes

人类负责：

```text
Docker
Docker Compose
Kubernetes
Ingress
Deployment
Service
HPA
ConfigMap
Secret
滚动更新
```

除非明确要求，否则：

> AI 当前不要提前处理 Kubernetes 和容器部署。

第一目标是：

```text
go build
↓
启动 Go 服务
↓
业务跑通
```

---

# 5. 第一开发原则

第一原则：

> 先跑起来。

开发顺序：

```text
代码可以编译
↓
服务可以启动
↓
数据库可以连接
↓
接口可以调用
↓
业务主链可以跑
↓
再优化
```

不要为了所谓：

```text
企业级
最佳实践
高可用
高并发
优雅架构
```

在第一阶段引入不必要的复杂度。

---

# 6. 禁止架构膨胀

未经人类明确批准，不允许主动加入：

```text
微服务
Kafka
RabbitMQ
RocketMQ
Service Mesh
gRPC 微服务通信
Elasticsearch
分库分表
复杂分布式事务
CQRS
Event Sourcing
复杂 DDD
多级缓存
复杂网关
```

如果 Agent 认为某项技术有必要：

> 先说明当前遇到了什么具体问题。

不能因为：

```text
12306 是高并发项目
```

就直接加入整个分布式技术栈。

---

# 7. 当前技术栈

第一阶段：

```text
Go
Gin
MySQL
sqlx
```

Redis：

```text
后续按需要加入
```

前端技术可以由开发阶段决定，但：

```text
前端不是项目重点
```

不要让前端复杂度抢占核心业务开发时间。

---

# 8. 项目架构

第一版：

```text
模块化单体
Modular Monolith
```

禁止擅自拆微服务。

后端分层：

```text
Handler
↓
Service
↓
Repository
↓
Database
```

---

# 9. Handler 规则

Handler 只负责：

```text
HTTP 参数绑定
基础格式校验
身份获取
调用 Service
返回 Response
```

Handler 禁止：

```text
直接写 SQL
直接修改 Seat
直接修改 Order 状态
直接修改 Payment 状态
写复杂业务逻辑
```

错误：

```go
func CreateOrder(c *gin.Context) {
    db.Exec("UPDATE seats ...")
}
```

禁止这种实现。

---

# 10. Service 规则

Service 是主要业务逻辑层。

负责：

```text
业务规则
状态校验
模块编排
事务边界
```

例如：

```text
OrderService.CreateOrder()
OrderService.Cancel()
OrderService.Return()
PaymentService.Pay()
PaymentService.Refund()
SeatService.Allocate()
```

核心业务逻辑必须能够在 Service 层清晰读出来。

---

# 11. Repository 规则

Repository 负责：

```text
SQL
查询
插入
更新
持久化
```

Repository 不负责：

```text
业务决策
```

例如：

```text
Repository 可以问：

这张订单是什么状态？

Repository 不应该决定：

这张订单现在能不能退票？
```

后者属于 Service。

---

# 12. 不允许绕过状态机

禁止编写这种通用函数：

```go
SetOrderStatus(id, status)
```

供任何业务随意调用。

也禁止：

```go
SetSeatStatus(...)
SetPaymentStatus(...)
```

作为万能状态修改入口。

状态必须通过明确动作发生变化。

例如：

```text
CancelOrder()
ReturnTicket()
PaymentSuccess()
PaymentFailed()
ReleaseSeat()
MarkSeatSold()
```

代码必须能够表达：

> 为什么状态发生变化。

---

# 13. TicketOrder 状态

严格使用：

```text
WAITING_PAYMENT
TICKETED
CANCELLED
RETURNED
COMPLETED
```

合法主链：

```text
WAITING_PAYMENT
→ TICKETED
→ COMPLETED
```

合法取消：

```text
WAITING_PAYMENT
→ CANCELLED
```

合法退票：

```text
TICKETED
→ RETURNED
```

禁止：

```text
CANCELLED → WAITING_PAYMENT
RETURNED → TICKETED
COMPLETED → TICKETED
```

---

# 14. Seat 状态

严格使用：

```text
AVAILABLE
LOCKED
SOLD
```

合法变化：

```text
AVAILABLE
→ LOCKED
→ SOLD
```

以及：

```text
LOCKED
→ AVAILABLE
```

和：

```text
SOLD
→ AVAILABLE
```

只有明确业务动作才能改变座位状态。

---

# 15. Payment 状态

严格使用：

```text
UNPAID
PAYING
SUCCESS
FAILED
REFUNDING
REFUNDED
```

首次支付：

```text
UNPAID
→ PAYING
→ SUCCESS
```

失败：

```text
PAYING
→ FAILED
```

重新支付：

```text
FAILED
→ PAYING
```

退款：

```text
SUCCESS
→ REFUNDING
→ REFUNDED
```

---

# 16. 三套状态必须分开

这是强约束。

```text
TicketOrder
负责票

Seat
负责座位

Payment
负责钱
```

不能为了省代码把三套状态合并。

例如以下状态组合是允许存在的：

```text
TicketOrder = WAITING_PAYMENT
Seat = LOCKED
Payment = FAILED
```

以及：

```text
TicketOrder = RETURNED
Seat = AVAILABLE
Payment = REFUNDING
```

不要“修正”为一个统一状态。

---

# 17. 一张票规则

必须遵守：

```text
一个 TicketOrder
=
一个乘车人
=
一个具体座位
```

一次购买两张票：

```text
TicketOrder A
TicketOrder B
```

不是：

```text
一个 TicketOrder quantity=2
```

---

# 18. 支付关系

一次 Payment：

```text
可以关联 1～2 个 TicketOrder
```

通过：

```text
payment_orders
```

关联。

不要为了代码简单把两张票合成一个 TicketOrder。

---

# 19. 座位就是库存

当前模型：

```text
Seat
=
库存实体
```

余票：

```text
AVAILABLE Seat 数量
```

不要另外创建：

```text
stock
inventory
remaining_ticket
```

作为第二套库存真值。

如果后期为了性能增加缓存库存：

> 必须明确它只是缓存/优化层，不得悄悄改变当前业务模型。

---

# 20. 座位分配规则(算法？）

购买 1 张：

```text
分配编号最小的 AVAILABLE 座位
```

购买 2 张：

```text
优先寻找连续座位
```

如果存在多组连续座位：

```text
选择编号最小的一组
```

没有连续座位：

```text
选择编号最小的两个 AVAILABLE 座位
```

例如：

```text
可售：
3
8
15

购买 2 张：

3
8
```

不要随机分配。

---

# 21. 下单规则

CreateOrder 至少包含：

```text
检查用户
检查车次
检查车次是否可售
检查乘车人
检查重复购票
检查当前用户是否已有待支付订单
分配座位
锁座
创建 TicketOrder
创建 Payment
创建 payment_orders
设置支付截止时间
```

最终：

```text
TicketOrder = WAITING_PAYMENT
Seat = LOCKED
Payment = UNPAID
```

---

# 22. 下单必须考虑事务

创建订单涉及：

```text
Seat
TicketOrder
Payment
payment_orders
```

这些不能各自随意提交。

必须保证：

```text
要么全部成功
要么全部失败
```

禁止产生：

```text
座位锁了
订单不存在
```

或者：

```text
订单存在
支付单不存在
```

---

# 23. 支付截止时间

创建订单：

```text
payment_deadline
=
created_at + 10分钟
```

规则：

> 必须在截止时间前确认支付成功。

不能因为用户：

```text
已经点击支付
```

就无限延长锁座时间。

---

# 24. 支付成功

支付成功后：

```text
Payment → SUCCESS
TicketOrder → TICKETED
Seat → SOLD
```

这个动作必须保持一致性。

禁止出现长期状态：

```text
Payment = SUCCESS
TicketOrder = WAITING_PAYMENT
```

---

# 25. 支付失败

支付失败：

```text
Payment → FAILED
```

但是：

```text
TicketOrder
仍然 WAITING_PAYMENT

Seat
仍然 LOCKED
```

只要没有超时：

```text
允许重新支付
```

不要因为支付失败自动取消订单。

---

# 26. 待支付取消

用户取消：

```text
TicketOrder
WAITING_PAYMENT → CANCELLED
```

同时：

```text
Seat
LOCKED → AVAILABLE
```

如果一次有两张票：

```text
允许只取消其中一张
```

剩余票：

```text
支付截止时间不变
```

---

# 27. 不支持部分支付

如果一次有两张 WAITING_PAYMENT：

```text
A
B
```

用户不能只支付 A，同时保持 B WAITING_PAYMENT。

如果只要 A：

```text
先取消 B
再支付 A
```

---

# 28. 退票规则

只有：

```text
TICKETED
```

允许退票。

并且：

```text
当前时间
<
发车时间 - 10分钟
```

确认退票后立即：

```text
TicketOrder → RETURNED
Seat → AVAILABLE
电子票失效
```

然后再：

```text
发起资金退款
```

---

# 29. 退票与退款禁止混淆

最重要规则之一：

```text
退票成功
≠
钱已经退到账
```

退票后可能：

```text
TicketOrder = RETURNED
Seat = AVAILABLE
Payment = REFUNDING
```

这是正确状态。

退款失败不能让：

```text
TicketOrder
RETURNED → TICKETED
```

---

# 30. Train 规则

车次状态：

```text
DRAFT
WAITING_SALE
ON_SALE
STOPPED
DEPARTED
ARCHIVED
```

只有：

```text
DRAFT
```

允许修改核心配置。

发布：

```text
DRAFT
→ WAITING_SALE
```

定时开售：

```text
WAITING_SALE
→ ON_SALE
```

发车前 10 分钟：

```text
ON_SALE
→ STOPPED
```

发车：

```text
STOPPED
→ DEPARTED
```

---

# 31. 发布车次

发布车次时，根据配置生成 Seat。

例如：

```text
FIRST_CLASS = 200
SECOND_CLASS = 1000
```

生成：

```text
FIRST_CLASS 1～200
SECOND_CLASS 1～1000
```

Seat 创建完成后：

```text
status = AVAILABLE
```

---

# 32. Passenger

第一版 Passenger 不是完整联系人系统。

它主要用于：

```text
保存订单中的乘车人信息
```

字段：

```text
name
id_card
```

普通用户接口：

```text
身份证必须脱敏
```

管理员接口：

```text
允许查看完整值
```

---

# 33. UUID

TicketOrder：

```text
UUID
```

Payment：

```text
UUID
```

不要擅自改成：

```text
Snowflake
自增订单号
时间戳订单号
```

后期需要再改。

---

# 34. 金额

所有金额：

```text
DECIMAL
```

Go 代码不能使用：

```text
float32
float64
```

直接进行资金核心运算。

避免浮点精度问题。

---

# 35. API 契约

接口必须遵守：

```text
docs/api.md
```

不要因为实现方便擅自：

```text
改接口路径
改字段名称
删除接口
新增万能接口
```

如确实需要修改：

> 先说明原因。

---

# 36. 错误处理

禁止：

```go
panic()
```

处理普通业务错误。

业务错误例如：

```text
TRAIN_NOT_ON_SALE
NO_AVAILABLE_SEAT
ORDER_EXPIRED
PAYMENT_STATUS_INVALID
```

应该返回明确业务错误。

未知系统错误：

```text
记录日志
返回通用内部错误
```

不要向前端泄漏：

```text
SQL
数据库密码
堆栈
内部路径
```

---

# 37. 日志

核心交易日志必须包含必要上下文。

例如：

```text
request_id
user_id
train_id
order_id
payment_id
seat_id
```

不要只写：

```text
order failed
```

应该能够支持后期排障。

同时：

> 不要把完整身份证号、密码、Token 打进日志。

---

# 38. 安全

密码：

```text
必须哈希
禁止明文保存
```

Token / Secret：

```text
禁止写死进代码
禁止提交 Git
```

身份证：

```text
用户端脱敏
日志禁止完整输出
```

SQL：

```text
使用参数化查询
禁止字符串拼接用户输入
```

---

# 39. 测试

核心业务至少应该存在测试：

```text
座位分配
创建订单
取消订单
支付成功
支付失败
订单超时
退票
退款状态
```

Agent 在修改核心状态机时，应同步检查相关测试。

不要只测试：

```text
HTTP 200
```

要测试业务状态变化。

---

# 40. 开发顺序

当前推荐：

```text
Phase 1
项目骨架
配置
日志
MySQL 连接

Phase 2
Auth / User

Phase 3
Station / Train

Phase 4
Seat

Phase 5
Order

Phase 6
Mock Payment

Phase 7
Cancel / Timeout

Phase 8
Return / Refund

Phase 9
Ticket

Phase 10
Admin
```

没有明确要求时：

> 不要跨越多个 Phase 一次性生成整个系统。

---

# 41. 每次开发任务范围

一次任务尽量只处理：

```text
一个模块
或
一个完整小功能
```

例如：

```text
实现 Station CRUD
```

可以。

```text
实现 TicketOrder 创建流程
```

可以。

但不要：

```text
现在把整个 12306 系统全部完成
```

然后一次性修改几十个文件。

---

# 42. 修改前行为

开始任务前：

1. 阅读相关文档。
2. 阅读相关现有代码。
3. 明确本次要改哪些模块。
4. 不触碰无关模块。
5. 判断是否涉及业务状态机。

如果涉及：

```text
Seat
Order
Payment
```

修改必须更加谨慎。

---

# 43. 修改后行为

完成代码后至少检查：

```text
go fmt
go test
go build
```

如果当前仓库已有：

```text
golangci-lint
```

再运行 lint。

不要未经人类要求主动引入大型 lint / build 工具链。

---

# 44. 禁止大范围重构

如果任务只是：

```text
修复订单取消 Bug
```

不要顺手：

```text
重构整个项目结构
替换 Gin
更换 sqlx
更换日志库
改数据库模型
```

除非这是解决问题所必须的。

---

# 45. 不要擅自升级依赖

不要看到版本旧就自动：

```text
go get -u ./...
```

依赖升级必须：

```text
有明确原因
```

并说明可能影响。

---

# 46. 数据库 Schema

数据库结构以：

```text
docs/database.md
```

为设计基线。

注意：

> 实际数据库创建和运维由人类负责。

如果代码需要数据库新增字段：

1. 说明为什么。
2. 指出需要修改哪张表。
3. 提供 migration / SQL 建议。
4. 不假设已经执行。

---

# 47. Redis

当前阶段：

> Redis 不是必需前置条件。

不要因为 architecture.md 中出现 Redis 就要求系统必须依赖 Redis 才能启动。

第一版应尽量做到：

```text
Go + MySQL
```

就能跑通核心交易链。

后面由人类决定什么时候加入 Redis。

---

# 48. Nginx

不要把：

```text
Nginx
```

当成开发阻塞条件。

Go Backend 必须能够：

```text
直接监听端口
直接通过 HTTP 测试
```

Nginx 由人类后续接入。

---

# 49. Docker

当前阶段：

> Docker 不是代码开发阻塞条件。

AI 应首先保证：

```bash
go build
```

能够成功。

由人类负责后续：

```text
Dockerfile
镜像
Docker Compose
```

除非人类明确要求协助。

---

# 50. Kubernetes

当前开发阶段不要主动：

```text
写 Deployment.yaml
写 Service.yaml
写 Ingress.yaml
设计 HPA
```

Kubernetes 是后续 SRE 阶段。

人类会处理。

---

# 51. 高并发

不要在业务第一版还没跑通时提前使用：

```text
Redis 分布式锁
Lua
MQ
复杂协程池
异步订单
```

第一阶段目标：

> 先保证单实例下业务正确。

然后：

```text
压测
发现问题
优化
```

---

# 52. Go 并发

不要为了展示 Go 能力随意使用：

```go
go func()
channel
sync.Mutex
```

如果业务本身不需要并发，就不要硬加。

尤其：

```text
进程内 Mutex
```

不能被视为未来 Kubernetes 多 Pod 场景下的最终分布式锁方案。

---

# 53. Agent 不得伪造运行结果

禁止声明：

```text
测试已经通过
数据库连接正常
接口已经成功
```

除非真的运行并获得对应结果。

如果没有执行环境：

> 明确告诉人类“代码已生成，但尚未实际运行验证”。

---

# 54. Agent 不得擅自操作基础设施

除非人类明确要求，不主动：

```text
修改系统服务
修改防火墙
安装 MySQL
安装 Redis
修改 Nginx
操作 Kubernetes
删除数据库
清空表
重置服务器
```

AI 当前身份：

> 开发。

人类身份：

> DevOps / SRE / 基础设施负责人。

双方配合。

---

# 55. 人类优先

人类开发者的最新明确指令高于本文档。

例如本文档写：

```text
Redis 后面加入
```

但人类明确要求：

```text
现在接 Redis
```

则执行最新指令。

如果最新指令与核心业务状态机明显冲突：

> 指出冲突后再修改。

---

# 56. 最重要的开发目标

不要追求：

```text
代码最多
技术最多
架构最复杂
```

要追求：

```text
业务正确
代码能编译
服务能运行
核心流程能走通
问题能够定位
后续能够演进
```

---

# 57. 当前最核心链路

开发过程中始终优先保证：

```text
管理员创建车次
↓
发布
↓
定时开售
↓
用户查票
↓
提交订单
↓
分配座位
↓
锁座
↓
WAITING_PAYMENT
↓
支付
↓
TICKETED
↓
电子票
↓
取消 / 退票 / 完成
```

如果某个新功能会严重阻碍这条主链：

> 第一版可以不做。

---

# 58. 最终边界

AI Agent 当前负责：

```text
写好软件
```

人类负责：

```text
让软件真正跑在基础设施上
```

等业务代码稳定后，再一起进入：

```text
Redis
Docker
Nginx
压测
Kubernetes
Prometheus
CI/CD
故障演练
支付宝 Sandbox
```

不要提前跨阶段。

