# 类 12306 高并发铁路票务系统｜API 设计 v1

# 1. 文档目的

本文件定义第一版 HTTP API，包括：

* 接口路径
* HTTP Method
* 请求参数
* 返回数据
* 身份要求
* 核心业务行为
* 基础错误码

本文件不定义：

* Go 具体实现
* MySQL SQL
* Redis
* 分布式锁
* MQ
* Kubernetes

API 必须遵守：

```text
PRD.md
business-rules.md
database.md
```

中的业务规则。

任何 Agent 不得为了方便实现而擅自修改业务状态机。

---

# 2. API 基础约定

统一前缀：

```text
/api/v1
```

数据格式：

```text
Content-Type: application/json
```

用户认证：

```text
Authorization: Bearer <token>
```

管理员接口：

```text
/api/v1/admin/*
```

普通用户不能访问管理员 API。

---

# 3. 统一响应格式

成功：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

失败：

```json
{
  "code": 40001,
  "message": "invalid request",
  "data": null
}
```

原则：

```text
HTTP 状态码
表达 HTTP 层结果

业务 code
表达具体业务错误
```

例如：

```text
HTTP 400
code = 42001
message = "ticket order expired"
```

---

# 4. 基础 HTTP 状态码

使用：

```text
200 OK
请求成功

201 Created
资源创建成功

400 Bad Request
参数错误 / 业务请求非法

401 Unauthorized
未登录 / Token 无效

403 Forbidden
没有权限

404 Not Found
资源不存在

409 Conflict
资源状态冲突

500 Internal Server Error
服务器内部错误
```

---

# 5. 注册

```http
POST /api/v1/auth/register
```

无需登录。

---

## 5.1 Request

```json
{
  "phone": "13800138000",
  "password": "123456",
  "nickname": "zhangsan"
}
```

规则：

```text
phone
必须唯一

password
至少 6 位

nickname
不能为空
```

---

## 5.2 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 10001,
    "phone": "13800138000",
    "nickname": "zhangsan"
  }
}
```

---

# 6. 登录

```http
POST /api/v1/auth/login
```

---

## 6.1 Request

```json
{
  "phone": "13800138000",
  "password": "123456"
}
```

---

## 6.2 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "xxx",
    "expires_in": 604800,
    "user": {
      "id": 10001,
      "phone": "13800138000",
      "nickname": "zhangsan"
    }
  }
}
```

登录有效期：

```text
7 天
```

---

# 7. 退出登录

```http
POST /api/v1/auth/logout
```

需要登录。

Response：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

# 8. 获取当前用户

```http
GET /api/v1/users/me
```

需要登录。

Response：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 10001,
    "phone": "13800138000",
    "nickname": "zhangsan",
    "created_at": "2026-09-12T10:00:00+08:00"
  }
}
```

---

# 9. 修改个人信息

```http
PATCH /api/v1/users/me
```

第一版只允许修改：

```text
nickname
```

Request：

```json
{
  "nickname": "new-name"
}
```

---

# 10. 修改密码

```http
PUT /api/v1/users/me/password
```

Request：

```json
{
  "old_password": "123456",
  "new_password": "654321"
}
```

---

# 11. 查询站点

```http
GET /api/v1/stations
```

无需登录。

只返回：

```text
ACTIVE
```

状态的站点。

Response：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "南京南"
    },
    {
      "id": 2,
      "name": "上海虹桥"
    }
  ]
}
```

---

# 12. 查询车次

```http
GET /api/v1/trains
```

---

## 12.1 按路线查询

示例：

```http
GET /api/v1/trains?departure_station_id=1&arrival_station_id=2&date=2026-09-20&page=1&page_size=10&sort=time_asc
```

参数：

```text
departure_station_id
出发站

arrival_station_id
目的站

date
出发日期

page
页码

page_size
每页条数，默认 10

sort
排序方式
```

---

## 12.2 排序

支持：

```text
time_asc
time_desc
price_asc
price_desc
```

其中价格排序第一版以：

```text
当前车次最低可售座位类型价格
```

作为排序依据。

---

## 12.3 按车次号查询

```http
GET /api/v1/trains?train_no=G101&date=2026-09-20
```

---

# 13. 查询车次返回

开售前：

```json
{
  "id": 101,
  "train_no": "G101",
  "departure_date": "2026-09-20",
  "departure_station": {
    "id": 1,
    "name": "南京南"
  },
  "arrival_station": {
    "id": 2,
    "name": "上海虹桥"
  },
  "departure_time": "2026-09-20T08:30:00+08:00",
  "arrival_time": "2026-09-20T10:10:00+08:00",
  "sale_start_time": "2026-09-12T10:00:00+08:00",
  "status": "WAITING_SALE",
  "first_class": {
    "price": "199.00",
    "available_count": null
  },
  "second_class": {
    "price": "99.00",
    "available_count": null
  }
}
```

开售后：

```json
{
  "id": 101,
  "train_no": "G101",
  "status": "ON_SALE",
  "first_class": {
    "price": "199.00",
    "available_count": 18
  },
  "second_class": {
    "price": "99.00",
    "available_count": 82
  }
}
```

---

# 14. 车次详情

```http
GET /api/v1/trains/{train_id}
```

用于：

```text
购票页
订单确认页
车次详情
```

返回完整车次信息。

---

# 15. 创建订单

```http
POST /api/v1/orders
```

需要登录。

这是核心交易接口。

---

## 15.1 Request

购买一张：

```json
{
  "train_id": 101,
  "seat_class": "SECOND_CLASS",
  "passengers": [
    {
      "name": "张三",
      "id_card": "320101200001011234"
    }
  ]
}
```

购买两张：

```json
{
  "train_id": 101,
  "seat_class": "SECOND_CLASS",
  "passengers": [
    {
      "name": "张三",
      "id_card": "320101200001011234"
    },
    {
      "name": "李四",
      "id_card": "320101200002022345"
    }
  ]
}
```

---

# 16. 创建订单业务行为

接口成功意味着：

```text
车次符合购票条件
乘车人符合购票条件
系统已经分配座位
座位已经 LOCKED
TicketOrder 已创建
Payment 已创建
10 分钟支付窗口已经开始
```

注意：

> 创建订单成功不代表出票成功。

此时：

```text
TicketOrder = WAITING_PAYMENT
Seat = LOCKED
Payment = UNPAID
```

---

# 17. 创建订单 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "payment_id": "a1111111-b222-c333-d444-e55555555555",
    "payment_deadline": "2026-09-12T10:10:00+08:00",
    "payable_amount": "198.00",
    "orders": [
      {
        "order_id": "11111111-2222-3333-4444-555555555555",
        "passenger": {
          "name": "张三",
          "id_card_masked": "3201********1234"
        },
        "seat_class": "SECOND_CLASS",
        "seat_no": 21,
        "ticket_price": "99.00",
        "status": "WAITING_PAYMENT"
      },
      {
        "order_id": "66666666-7777-8888-9999-000000000000",
        "passenger": {
          "name": "李四",
          "id_card_masked": "3201********2345"
        },
        "seat_class": "SECOND_CLASS",
        "seat_no": 22,
        "ticket_price": "99.00",
        "status": "WAITING_PAYMENT"
      }
    ]
  }
}
```

---

# 18. 获取用户订单列表

```http
GET /api/v1/orders
```

需要登录。

第一版：

```text
不筛选
不分类
按创建时间倒序
```

Response：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "order_id": "uuid",
      "train_no": "G101",
      "departure_station": "南京南",
      "arrival_station": "上海虹桥",
      "departure_time": "2026-09-20T08:30:00+08:00",
      "passenger_name": "张三",
      "seat_class": "SECOND_CLASS",
      "seat_no": 21,
      "ticket_price": "99.00",
      "status": "TICKETED",
      "created_at": "2026-09-12T10:00:00+08:00"
    }
  ]
}
```

---

# 19. 获取订单详情

```http
GET /api/v1/orders/{order_id}
```

只能查看：

```text
属于当前登录用户
```

的订单。

返回：

```text
订单状态
乘车人
车次
座位
价格
支付截止时间
出票时间
取消时间
退票时间
完成时间
```

---

# 20. 取消待支付车票

```http
POST /api/v1/orders/{order_id}/cancel
```

需要登录。

只允许：

```text
WAITING_PAYMENT
```

状态调用。

成功后：

```text
TicketOrder
WAITING_PAYMENT → CANCELLED

Seat
LOCKED → AVAILABLE
```

如果这个 Payment 原本包含两张票：

```text
重新计算 payable_amount
```

---

# 21. Cancel Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_id": "uuid",
    "status": "CANCELLED",
    "payment": {
      "payment_id": "uuid",
      "payable_amount": "99.00"
    }
  }
}
```

---

# 22. 用户退票

```http
POST /api/v1/orders/{order_id}/return
```

需要登录。

只允许：

```text
TICKETED
```

且必须满足：

```text
当前时间
<
发车时间 - 10分钟
```

---

# 23. 退票业务行为

退票确认后：

```text
TicketOrder
TICKETED → RETURNED

Seat
SOLD → AVAILABLE

电子票
立即失效
```

然后：

```text
发起资金退款
```

注意：

> API 成功表示退票业务已经完成。

并不代表：

```text
资金已经真正退款到账
```

---

# 24. Return Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_id": "uuid",
    "order_status": "RETURNED",
    "refund": {
      "payment_status": "REFUNDING",
      "refund_amount": "99.00"
    }
  }
}
```

---

# 25. 查询电子票

```http
GET /api/v1/tickets/{order_id}
```

需要登录。

可以查询：

```text
TICKETED
RETURNED
COMPLETED
```

状态的历史票据。

---

# 26. 电子票 Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_id": "uuid",
    "status": "TICKETED",
    "train_no": "G101",
    "departure_station": "南京南",
    "arrival_station": "上海虹桥",
    "departure_time": "2026-09-20T08:30:00+08:00",
    "arrival_time": "2026-09-20T10:10:00+08:00",
    "passenger": {
      "name": "张三",
      "id_card_masked": "3201********1234"
    },
    "seat": {
      "class": "SECOND_CLASS",
      "number": 21
    },
    "ticket_price": "99.00",
    "ticketed_at": "2026-09-12T10:02:15+08:00",
    "valid": true
  }
}
```

退票后：

```text
valid = false
```

---

# 27. 查询支付单

```http
GET /api/v1/payments/{payment_id}
```

需要登录。

只能查看当前用户自己的 Payment。

Response：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "payment_id": "uuid",
    "status": "UNPAID",
    "original_amount": "198.00",
    "payable_amount": "198.00",
    "refunded_amount": "0.00",
    "payment_deadline": "2026-09-12T10:10:00+08:00",
    "provider": "MOCK"
  }
}
```

---

# 28. 发起支付

```http
POST /api/v1/payments/{payment_id}/pay
```

需要登录。

作用：

```text
UNPAID / FAILED
→ PAYING
```

Mock 阶段返回模拟支付信息。

支付宝 Sandbox 阶段：

```text
返回支付宝沙箱支付参数 / 支付地址
```

---

# 29. Mock 支付成功

仅开发环境启用：

```http
POST /api/v1/payments/{payment_id}/mock/success
```

作用：

```text
Payment
PAYING → SUCCESS
```

同时：

```text
所有参与支付的有效 TicketOrder
WAITING_PAYMENT → TICKETED

对应 Seat
LOCKED → SOLD
```

并生成电子票。

---

# 30. Mock 支付失败

仅开发环境启用：

```http
POST /api/v1/payments/{payment_id}/mock/fail
```

作用：

```text
Payment
PAYING → FAILED
```

TicketOrder：

```text
仍然 WAITING_PAYMENT
```

Seat：

```text
仍然 LOCKED
```

如果还没超时，可以重新支付。

---

# 31. 支付宝 Sandbox 回调

第二阶段增加：

```http
POST /api/v1/payments/alipay/notify
```

这是：

```text
支付宝服务器
→ 本系统后端
```

的异步通知接口。

它不是普通用户接口。

核心作用：

```text
验签
确认支付结果
更新 Payment
更新 TicketOrder
更新 Seat
```

具体签名和支付宝字段在支付接入阶段另写。

---

# 32. 支付宝退款

第二阶段由：

```text
POST /orders/{order_id}/return
```

内部触发。

用户端不需要直接调用：

```text
/payment/refund
```

这种接口。

因为：

> 用户操作的是“退票”，不是直接操作支付系统。

---

# 33. 管理员登录

```http
POST /api/v1/admin/auth/login
```

Request：

```json
{
  "username": "admin",
  "password": "xxxxxx"
}
```

返回管理员 Token。

---

# 34. 管理员查询站点

```http
GET /api/v1/admin/stations
```

管理员可查看：

```text
ACTIVE
DISABLED
```

全部站点。

---

# 35. 管理员创建站点

```http
POST /api/v1/admin/stations
```

Request：

```json
{
  "name": "南京南"
}
```

自动记录：

```text
CREATE_STATION
```

审计日志。

---

# 36. 管理员修改站点

```http
PATCH /api/v1/admin/stations/{station_id}
```

第一版主要用于修改：

```text
站点名称
```

已产生历史业务数据后应谨慎使用。

---

# 37. 停用站点

```http
POST /api/v1/admin/stations/{station_id}/disable
```

作用：

```text
ACTIVE → DISABLED
```

停用后不能用于创建新车次。

历史车次不受影响。

---

# 38. 管理员查询车次

```http
GET /api/v1/admin/trains
```

支持基本分页。

可查看所有状态：

```text
DRAFT
WAITING_SALE
ON_SALE
STOPPED
DEPARTED
ARCHIVED
```

---

# 39. 管理员创建车次

```http
POST /api/v1/admin/trains
```

Request：

```json
{
  "train_no": "G101",
  "departure_date": "2026-09-20",
  "departure_station_id": 1,
  "arrival_station_id": 2,
  "departure_time": "2026-09-20T08:30:00+08:00",
  "arrival_time": "2026-09-20T10:10:00+08:00",
  "sale_start_time": "2026-09-12T10:00:00+08:00",
  "first_class_price": "199.00",
  "second_class_price": "99.00",
  "first_class_seat_count": 200,
  "second_class_seat_count": 1000
}
```

创建后：

```text
Train = DRAFT
```

---

# 40. 管理员修改车次

```http
PATCH /api/v1/admin/trains/{train_id}
```

仅允许：

```text
DRAFT
```

状态修改核心字段。

进入：

```text
WAITING_SALE
ON_SALE
STOPPED
DEPARTED
ARCHIVED
```

后禁止修改核心业务字段。

---

# 41. 发布车次

```http
POST /api/v1/admin/trains/{train_id}/publish
```

作用：

```text
DRAFT
→ WAITING_SALE
```

发布前必须完成：

```text
路线
日期
时间
票价
座位数量
开售时间
```

等必要配置。

发布时根据座位数量创建对应 Seat 数据。

例如：

```text
一等座 200
→ FIRST_CLASS 1～200

二等座 1000
→ SECOND_CLASS 1～1000
```

---

# 42. 归档车次

```http
POST /api/v1/admin/trains/{train_id}/archive
```

作用：

```text
DEPARTED
→ ARCHIVED
```

不能物理删除。

---

# 43. 管理员查询用户

```http
GET /api/v1/admin/users
```

第一版支持简单分页。

返回：

```text
用户 ID
手机号
昵称
注册时间
```

---

# 44. 管理员查看用户详情

```http
GET /api/v1/admin/users/{user_id}
```

可查看：

```text
用户信息
用户历史订单
```

---

# 45. 管理员重置密码

```http
POST /api/v1/admin/users/{user_id}/reset-password
```

第一版管理员直接重置用户密码。

该操作必须写入：

```text
admin_audit_logs
```

---

# 46. 管理员查询订单

```http
GET /api/v1/admin/orders
```

支持：

```text
status
phone
train_no
page
page_size
```

示例：

```http
GET /api/v1/admin/orders?status=TICKETED&train_no=G101&page=1&page_size=20
```

---

# 47. 管理员订单详情

```http
GET /api/v1/admin/orders/{order_id}
```

管理员可以看到：

```text
完整身份证号
用户
乘车人
车次
座位
订单状态
支付信息
时间信息
```

---

# 48. 强制取消异常订单

```http
POST /api/v1/admin/orders/{order_id}/force-cancel
```

这个接口只作为异常人工处理入口。

不得被正常业务流程调用。

执行后必须写入审计日志。

具体允许操作哪些异常状态，由后续异常处理设计定义。

---

# 49. 退款重试

```http
POST /api/v1/admin/orders/{order_id}/retry-refund
```

用于：

```text
车票已经 RETURNED
但资金退款仍存在异常
```

的情况。

它只能重新尝试资金退款。

绝对不能：

```text
重新激活车票
重新锁回座位
RETURNED → TICKETED
```

---

# 50. 管理员审计日志

```http
GET /api/v1/admin/audit-logs
```

支持分页。

Response：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 10001,
      "admin_id": 1,
      "action": "PUBLISH_TRAIN",
      "resource_type": "TRAIN",
      "resource_id": "101",
      "created_at": "2026-09-12T09:00:00+08:00"
    }
  ]
}
```

---

# 51. 第一版业务错误码

用户相关：

```text
41001
PHONE_ALREADY_REGISTERED

41002
INVALID_PHONE_OR_PASSWORD

41003
UNAUTHORIZED
```

车次相关：

```text
42001
TRAIN_NOT_FOUND

42002
TRAIN_NOT_ON_SALE

42003
TRAIN_SALE_STOPPED

42004
INVALID_STATION

42005
TRAIN_ALREADY_EXISTS
```

座位相关：

```text
43001
NO_AVAILABLE_SEAT

43002
INSUFFICIENT_SEATS

43003
SEAT_LOCK_FAILED
```

订单相关：

```text
44001
ORDER_NOT_FOUND

44002
ORDER_STATUS_INVALID

44003
ORDER_EXPIRED

44004
USER_HAS_PENDING_ORDER

44005
PASSENGER_ALREADY_HAS_TICKET
```

支付相关：

```text
45001
PAYMENT_NOT_FOUND

45002
PAYMENT_STATUS_INVALID

45003
PAYMENT_EXPIRED

45004
PAYMENT_FAILED
```

退票相关：

```text
46001
RETURN_NOT_ALLOWED

46002
RETURN_DEADLINE_PASSED
```

管理员相关：

```text
47001
ADMIN_UNAUTHORIZED

47002
ADMIN_OPERATION_NOT_ALLOWED
```

---

# 52. 前端最核心调用链

用户查票：

```text
GET /stations
↓
GET /trains
↓
GET /trains/{id}
```

用户购票：

```text
POST /orders
↓
GET /payments/{payment_id}
↓
POST /payments/{payment_id}/pay
↓
Mock / Sandbox 支付
↓
GET /orders/{order_id}
↓
GET /tickets/{order_id}
```

用户取消：

```text
POST /orders/{order_id}/cancel
```

用户退票：

```text
POST /orders/{order_id}/return
```

---

# 53. 管理员最核心调用链

```text
登录
↓
创建站点
↓
创建车次
↓
修改草稿
↓
发布车次
↓
等待定时开售
↓
查看订单
↓
必要时异常人工处理
```

---

# 54. API 设计核心原则

## 原则一

API 不能绕开业务状态机。

例如不能存在：

```text
POST /orders/{id}/set-status
```

让前端随便指定：

```text
TICKETED
RETURNED
COMPLETED
```

---

## 原则二

前端提交：

```text
动作
```

后端决定：

```text
状态怎么变化
```

例如前端只能请求：

```text
支付
取消
退票
```

不能请求：

```text
把订单改成 TICKETED
```

---

## 原则三

支付接口只负责：

```text
钱
```

订单接口负责：

```text
票
```

退票 API：

```text
/orders/{id}/return
```

由订单业务内部触发退款。

用户不直接操作底层退款接口。

---

## 原则四

管理员接口也不能破坏核心状态机。

管理员只是：

```text
异常兜底
```

不是：

```text
数据库状态编辑器
```

---

# 55. v1 核心 API 总览

用户：

```text
POST   /auth/register
POST   /auth/login
POST   /auth/logout

GET    /users/me
PATCH  /users/me
PUT    /users/me/password

GET    /stations

GET    /trains
GET    /trains/{id}

POST   /orders
GET    /orders
GET    /orders/{id}
POST   /orders/{id}/cancel
POST   /orders/{id}/return

GET    /tickets/{order_id}

GET    /payments/{id}
POST   /payments/{id}/pay
```

开发环境：

```text
POST /payments/{id}/mock/success
POST /payments/{id}/mock/fail
```

管理员：

```text
POST   /admin/auth/login

GET    /admin/stations
POST   /admin/stations
PATCH  /admin/stations/{id}
POST   /admin/stations/{id}/disable

GET    /admin/trains
POST   /admin/trains
PATCH  /admin/trains/{id}
POST   /admin/trains/{id}/publish
POST   /admin/trains/{id}/archive

GET    /admin/users
GET    /admin/users/{id}
POST   /admin/users/{id}/reset-password

GET    /admin/orders
GET    /admin/orders/{id}
POST   /admin/orders/{id}/force-cancel
POST   /admin/orders/{id}/retry-refund

GET    /admin/audit-logs
```

第二阶段：

```text
POST /payments/alipay/notify
```

这些接口组成第一版完整业务闭环。

