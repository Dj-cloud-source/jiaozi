# 类 12306 高并发铁路票务系统｜数据库设计 v1

# 1. 文档目的

本文件定义第一版 MySQL 数据模型，包括：

* 核心数据表
* 字段含义
* 主键
* 外键关系
* 唯一约束
* 必要索引
* 状态字段
* 表之间的关系

本文件暂不定义：

* Redis 数据结构
* 分布式锁
* Lua
* MQ
* 缓存一致性
* 分库分表
* 高并发优化

这些内容后续单独设计。

人类开发者有话说：这个表结构是给AI看的，具体创建和维护是人来

---

# 2. 核心实体

第一版核心数据表：

```text
users
admins
stations
trains
seats
passengers
ticket_orders
payments
payment_orders
admin_audit_logs
```

其中：

```text
users
普通用户

admins
管理员

stations
站点

trains
具体某一天的一趟车

seats
某趟车上的具体座位

passengers
乘车人信息

ticket_orders
一张具体车票

payments
一次支付

payment_orders
支付单与车票订单之间的关联

admin_audit_logs
管理员操作审计
```

---

# 3. 核心关系

整体关系：

```text
User
 │
 ├── Passenger
 │
 └── TicketOrder
         │
         ├── Train
         ├── Seat
         └── Payment
```

更准确地展开：

```text
users
  │
  ├── passengers
  │
  └── ticket_orders
          │
          ├── trains
          │     ├── stations
          │     └── seats
          │
          └── payment_orders
                  │
                  └── payments
```

因为：

> 一次 Payment 可以支付 1～2 张 TicketOrder。

所以 Payment 和 TicketOrder 是：

```text
多对多关系
```

但第一版业务限制下：

```text
一个 TicketOrder
最多属于一个有效 Payment

一个 Payment
最多关联两个 TicketOrder
```

---

# 4. users

用户账号表。

```sql
CREATE TABLE users (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    phone           VARCHAR(20) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    nickname        VARCHAR(50) NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                    ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_users_phone (phone)
);
```

---

## 4.1 字段说明

```text
id
用户内部主键

phone
登录手机号

password_hash
密码哈希值

nickname
用户昵称

created_at
注册时间

updated_at
最后更新时间
```

手机号必须唯一：

```text
UNIQUE(phone)
```

---

# 5. admins

管理员账号表。

第一版管理员账号由系统初始化。

```sql
CREATE TABLE admins (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    username        VARCHAR(50) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                    ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_admins_username (username)
);
```

第一版：

```text
不做多角色
不做 RBAC
不做管理员注册
```

---

# 6. stations

站点表。

```sql
CREATE TABLE stations (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name            VARCHAR(100) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                    ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_stations_name (name),
    KEY idx_stations_status (status)
);
```

---

## 6.1 status

允许：

```text
ACTIVE
DISABLED
```

规则：

```text
ACTIVE
可以用于创建新车次

DISABLED
不能用于创建新车次
历史数据继续保留
```

站点不物理删除。

---

# 7. trains

`trains` 表表示：

> 某一天的一趟具体车次。

例如：

```text
G101 + 2026-09-20
```

和：

```text
G101 + 2026-09-21
```

属于两个不同的 `train`。

---

## 7.1 表结构

```sql
CREATE TABLE trains (
    id                      BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,

    train_no                VARCHAR(10) NOT NULL,
    departure_date          DATE NOT NULL,

    departure_station_id    BIGINT UNSIGNED NOT NULL,
    arrival_station_id      BIGINT UNSIGNED NOT NULL,

    departure_time          DATETIME NOT NULL,
    arrival_time            DATETIME NOT NULL,
    sale_start_time         DATETIME NOT NULL,

    first_class_price       DECIMAL(10,2) NULL,
    second_class_price      DECIMAL(10,2) NULL,

    first_class_seat_count  INT UNSIGNED NOT NULL DEFAULT 0,
    second_class_seat_count INT UNSIGNED NOT NULL DEFAULT 0,

    status                  VARCHAR(20) NOT NULL DEFAULT 'DRAFT',

    created_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                            ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_train_date (
        train_no,
        departure_date
    ),

    KEY idx_train_route_date (
        departure_station_id,
        arrival_station_id,
        departure_date
    ),

    KEY idx_train_sale_start_time (sale_start_time),
    KEY idx_train_departure_time (departure_time),
    KEY idx_train_status (status),

    CONSTRAINT fk_train_departure_station
        FOREIGN KEY (departure_station_id)
        REFERENCES stations(id),

    CONSTRAINT fk_train_arrival_station
        FOREIGN KEY (arrival_station_id)
        REFERENCES stations(id)
);
```

---

# 8. Train 状态

允许：

```text
DRAFT
WAITING_SALE
ON_SALE
STOPPED
DEPARTED
ARCHIVED
```

含义：

```text
DRAFT
草稿，可编辑

WAITING_SALE
已发布，等待定时开售

ON_SALE
正在售票

STOPPED
已停止售票

DEPARTED
已发车

ARCHIVED
历史归档
```

---

# 9. 车次唯一约束

核心唯一约束：

```sql
UNIQUE(train_no, departure_date)
```

表示：

> 同一天，同一个车次号只能存在一趟车。

---

# 10. seats

Seat 是本项目最核心的数据之一。

原则：

> 座位本身就是库存。

不再额外维护：

```text
stock = 100
```

这种独立库存数字作为最终库存真值。

---

## 10.1 表结构

```sql
CREATE TABLE seats (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,

    train_id        BIGINT UNSIGNED NOT NULL,

    seat_class      VARCHAR(20) NOT NULL,
    seat_no         INT UNSIGNED NOT NULL,

    status          VARCHAR(20) NOT NULL DEFAULT 'AVAILABLE',

    locked_order_id CHAR(36) NULL,
    locked_at       DATETIME NULL,

    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                    ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_train_class_seat (
        train_id,
        seat_class,
        seat_no
    ),

    KEY idx_seat_available (
        train_id,
        seat_class,
        status,
        seat_no
    ),

    CONSTRAINT fk_seat_train
        FOREIGN KEY (train_id)
        REFERENCES trains(id)
);
```

---

# 11. seat_class

允许：

```text
FIRST_CLASS
SECOND_CLASS
```

例如：

```text
G101

FIRST_CLASS
1～200

SECOND_CLASS
1～1000
```

这两个：

```text
FIRST_CLASS 1号
SECOND_CLASS 1号
```

是两个完全不同的座位。

---

# 12. Seat 状态

允许：

```text
AVAILABLE
LOCKED
SOLD
```

---

## 12.1 AVAILABLE

```text
可以出售
```

---

## 12.2 LOCKED

```text
已经分配给待支付订单
暂时不能被其他人购买
```

同时记录：

```text
locked_order_id
locked_at
```

---

## 12.3 SOLD

```text
已经支付成功并出票
```

---

# 13. 为什么 seats 不存 user_id

不建议直接在 Seat 中存：

```text
user_id
passenger_id
```

因为 Seat 只负责表达：

```text
这个座位目前能不能卖
```

具体：

```text
谁买的
谁乘车
对应哪个订单
```

应该由 `ticket_orders` 表表达。

---

# 14. passengers

Passenger 表示购票时使用的乘车人。

第一版不做复杂常用乘车人系统。

---

## 14.1 表结构

```sql
CREATE TABLE passengers (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,

    user_id         BIGINT UNSIGNED NOT NULL,

    name            VARCHAR(100) NOT NULL,
    id_card         VARCHAR(30) NOT NULL,

    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    KEY idx_passenger_user (user_id),
    KEY idx_passenger_id_card (id_card),

    CONSTRAINT fk_passenger_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
);
```

---

# 15. Passenger 的定位

第一版 Passenger 不一定代表：

> 永久保存的常用联系人。

更准确是：

> 用户在购票过程中使用过的乘车人信息。

后续如果要做：

```text
常用乘车人
```

再扩展即可。

---

# 16. ticket_orders

这是整个业务最重要的表之一。

原则：

```text
一张票
=
一个 TicketOrder
=
一个乘车人
=
一个座位
```

---

## 16.1 表结构

```sql
CREATE TABLE ticket_orders (
    id                  CHAR(36) PRIMARY KEY,

    user_id             BIGINT UNSIGNED NOT NULL,
    passenger_id        BIGINT UNSIGNED NOT NULL,
    train_id            BIGINT UNSIGNED NOT NULL,
    seat_id             BIGINT UNSIGNED NOT NULL,

    ticket_price        DECIMAL(10,2) NOT NULL,

    status              VARCHAR(30) NOT NULL,

    payment_deadline    DATETIME NOT NULL,

    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ticketed_at         DATETIME NULL,
    cancelled_at        DATETIME NULL,
    returned_at         DATETIME NULL,
    completed_at        DATETIME NULL,

    KEY idx_order_user_created (
        user_id,
        created_at
    ),

    KEY idx_order_train (
        train_id
    ),

    KEY idx_order_status (
        status
    ),

    KEY idx_order_payment_deadline (
        status,
        payment_deadline
    ),

    KEY idx_order_passenger_train (
        passenger_id,
        train_id
    ),

    CONSTRAINT fk_order_user
        FOREIGN KEY (user_id)
        REFERENCES users(id),

    CONSTRAINT fk_order_passenger
        FOREIGN KEY (passenger_id)
        REFERENCES passengers(id),

    CONSTRAINT fk_order_train
        FOREIGN KEY (train_id)
        REFERENCES trains(id),

    CONSTRAINT fk_order_seat
        FOREIGN KEY (seat_id)
        REFERENCES seats(id)
);
```

---

# 17. TicketOrder 状态

允许：

```text
WAITING_PAYMENT
TICKETED
CANCELLED
RETURNED
COMPLETED
```

---

# 18. 为什么 ticket_price 必须存在

不能每次打开历史订单时重新读取：

```text
trains.second_class_price
```

因为 TicketOrder 必须保存：

> 用户买票时的实际成交价格。

所以每张票自己保存：

```text
ticket_price
```

---

# 19. 为什么订单使用 UUID

第一版用户已经决定：

```text
订单号使用 UUID
```

因此：

```sql
id CHAR(36)
```

例如：

```text
550e8400-e29b-41d4-a716-446655440000
```

后续如果需要优化 UUID 存储空间，可以再改：

```text
BINARY(16)
```

第一版先保持可读性。

---

# 20. 同一身份证同车次唯一票问题

业务规则：

> 同一身份证，在同一趟车上只能存在一张有效票。

但这里不能简单写：

```sql
UNIQUE(id_card, train_id)
```

因为用户退票之后允许重新购买。

例如：

```text
张三
G101
订单A = RETURNED
```

之后应该允许生成：

```text
订单B
```

所以这个规则最终需要由：

```text
业务层校验
+
数据库事务
```

共同保证。

不能直接使用普通永久 UNIQUE 约束解决。

---

# 21. 一名用户一个待支付组

业务规则：

> 一个用户同一时间只能存在一组待支付车票。

这个规则也不适合简单写：

```sql
UNIQUE(user_id, status)
```

因为数据库里会存在大量：

```text
CANCELLED
TICKETED
COMPLETED
```

历史订单。

第一版在 Service 层进行业务校验。

后续并发优化时再设计更严格的数据库/Redis约束。

---

# 22. payments

Payment 表示：

> 一次资金支付。

一笔 Payment 可以支付：

```text
1 张票
或
2 张票
```

---

## 22.1 表结构

```sql
CREATE TABLE payments (
    id                  CHAR(36) PRIMARY KEY,

    user_id             BIGINT UNSIGNED NOT NULL,

    original_amount     DECIMAL(10,2) NOT NULL,
    payable_amount      DECIMAL(10,2) NOT NULL,
    refunded_amount     DECIMAL(10,2) NOT NULL DEFAULT 0,

    status              VARCHAR(30) NOT NULL,

    payment_deadline    DATETIME NOT NULL,

    provider            VARCHAR(30) NOT NULL DEFAULT 'MOCK',
    provider_trade_no   VARCHAR(100) NULL,

    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    paid_at             DATETIME NULL,
    updated_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
                        ON UPDATE CURRENT_TIMESTAMP,

    KEY idx_payment_user (
        user_id
    ),

    KEY idx_payment_status (
        status
    ),

    KEY idx_payment_deadline (
        status,
        payment_deadline
    ),

    KEY idx_provider_trade_no (
        provider_trade_no
    ),

    CONSTRAINT fk_payment_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
);
```

---

# 23. Payment 状态

允许：

```text
UNPAID
PAYING
SUCCESS
FAILED
REFUNDING
REFUNDED
```

---

# 24. Payment 金额字段

三个金额：

```text
original_amount
payable_amount
refunded_amount
```

例如一次购买两张：

```text
张三 ¥99
李四 ¥99
```

创建支付单：

```text
original_amount = 198
payable_amount  = 198
refunded_amount = 0
```

如果支付前取消李四：

```text
original_amount = 198
payable_amount  = 99
refunded_amount = 0
```

支付成功后张三退票：

```text
original_amount = 198
payable_amount  = 99
refunded_amount = 99
```

---

# 25. provider

支付渠道。

第一版：

```text
MOCK
```

第二阶段：

```text
ALIPAY_SANDBOX
```

因此 Payment 模型不用因为以后接支付宝重新推倒。

---

# 26. provider_trade_no

用于保存第三方支付平台返回的交易号。

Mock 阶段可以为空。

支付宝 Sandbox 阶段保存类似：

```text
支付宝交易号
```

后续处理：

```text
支付查询
退款
回调
对账
```

都会用到。

---

# 27. 为什么不能在 ticket_orders 直接存 payment_id

表面上可以：

```text
ticket_orders.payment_id
```

但我们业务里：

```text
一个 Payment
可以关联多个 TicketOrder
```

并且未来支付重试、重新创建支付单后，关系可能变复杂。

所以第一版直接用：

```text
payment_orders
```

中间关联表。

结构更清楚。

---

# 28. payment_orders

关联：

```text
Payment
↔
TicketOrder
```

---

## 28.1 表结构

```sql
CREATE TABLE payment_orders (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,

    payment_id      CHAR(36) NOT NULL,
    order_id        CHAR(36) NOT NULL,

    amount          DECIMAL(10,2) NOT NULL,

    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uk_payment_order (
        payment_id,
        order_id
    ),

    KEY idx_payment_orders_order (
        order_id
    ),

    CONSTRAINT fk_payment_orders_payment
        FOREIGN KEY (payment_id)
        REFERENCES payments(id),

    CONSTRAINT fk_payment_orders_order
        FOREIGN KEY (order_id)
        REFERENCES ticket_orders(id)
);
```

---

# 29. payment_orders 示例

两张票一起支付：

```text
Payment P001
¥198
```

关联：

```text
P001 → Order A → ¥99
P001 → Order B → ¥99
```

数据库：

```text
payment_id | order_id | amount
-----------|----------|-------
P001       | A        | 99
P001       | B        | 99
```

---

# 30. 单独取消一张票

原来：

```text
Payment P001
A ¥99
B ¥99
```

取消 B：

```text
B → CANCELLED
```

同时：

```text
payments.payable_amount
198 → 99
```

Payment 继续只支付有效订单 A。

---

# 31. 单独退其中一张票

支付成功：

```text
Payment P001
SUCCESS
¥198
```

对应：

```text
A TICKETED
B TICKETED
```

A 退票：

```text
A → RETURNED
B → TICKETED
```

Payment：

```text
status = SUCCESS

original_amount = 198
refunded_amount = 99
```

第一版不需要增加：

```text
PARTIALLY_REFUNDED
```

---

# 32. 为什么 Payment 与 TicketOrder 分离

这是数据库模型中非常关键的设计。

TicketOrder 表达：

```text
票
```

Payment 表达：

```text
钱
```

所以可以存在：

```text
TicketOrder = RETURNED
Payment = REFUNDING
```

也可以存在：

```text
TicketOrder = WAITING_PAYMENT
Payment = FAILED
```

两者状态不应该绑成一个字段。

---

# 33. admin_audit_logs

管理员操作审计表。

---

## 33.1 表结构

```sql
CREATE TABLE admin_audit_logs (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,

    admin_id        BIGINT UNSIGNED NOT NULL,

    action          VARCHAR(100) NOT NULL,
    resource_type   VARCHAR(50) NOT NULL,
    resource_id     VARCHAR(100) NULL,

    before_data     JSON NULL,
    after_data      JSON NULL,

    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    KEY idx_audit_admin (
        admin_id,
        created_at
    ),

    KEY idx_audit_resource (
        resource_type,
        resource_id
    ),

    CONSTRAINT fk_audit_admin
        FOREIGN KEY (admin_id)
        REFERENCES admins(id)
);
```

---

# 34. 审计操作示例

例如：

```text
CREATE_STATION
DISABLE_STATION

CREATE_TRAIN
UPDATE_TRAIN
PUBLISH_TRAIN

RESET_USER_PASSWORD

FORCE_CANCEL_ORDER
RETRY_REFUND
```

记录：

```text
谁操作
操作什么
操作哪个资源
修改前是什么
修改后是什么
什么时候操作
```

---

# 35. 为什么不物理删除核心业务数据

第一版以下数据原则上不做物理删除：

```text
stations
trains
ticket_orders
payments
admin_audit_logs
```

原因：

```text
历史订单需要追溯
支付记录需要追溯
审计需要追溯
```

使用：

```text
状态字段
```

表达失效、停用、归档。

---

# 36. 基础 ER 图

```text
┌───────────┐
│   users   │
└─────┬─────┘
      │
      ├───────────────────┐
      │                   │
      ▼                   ▼
┌─────────────┐     ┌───────────────┐
│ passengers  │     │ ticket_orders │
└──────┬──────┘     └───────┬───────┘
       │                     │
       └────────────┐        │
                    ▼        ▼
                 ┌──────────────┐
                 │    trains    │
                 └──────┬───────┘
                        │
              ┌─────────┴─────────┐
              ▼                   ▼
        ┌──────────┐        ┌──────────┐
        │ stations │        │  seats   │
        └──────────┘        └──────────┘


ticket_orders
      │
      ▼
payment_orders
      │
      ▼
payments
```

---

# 37. 关键索引设计

第一版最重要的几个索引：

## 查询车次

```text
departure_station_id
arrival_station_id
departure_date
```

对应：

```sql
KEY idx_train_route_date (
    departure_station_id,
    arrival_station_id,
    departure_date
)
```

---

## 查询可售座位

```text
train_id
seat_class
status
seat_no
```

对应：

```sql
KEY idx_seat_available (
    train_id,
    seat_class,
    status,
    seat_no
)
```

这个索引后面对于：

```text
找最小座位号
找连续座位
抢票
```

都会非常重要。

---

## 用户订单列表

```text
user_id
created_at
```

---

## 超时订单

```text
status
payment_deadline
```

后面系统扫描：

```text
已经超过 10 分钟的 WAITING_PAYMENT
```

时会用到。

---

# 38. 金额字段原则

所有金额使用：

```sql
DECIMAL(10,2)
```

不使用：

```text
FLOAT
DOUBLE
```

因为金额不能接受浮点误差。

例如：

```text
99.00
199.00
```

---

# 39. 时间字段原则

第一版统一使用：

```sql
DATETIME
```

业务代码统一处理时区。

需要记录：

```text
created_at
updated_at
paid_at
ticketed_at
cancelled_at
returned_at
completed_at
payment_deadline
sale_start_time
departure_time
arrival_time
```

---

# 40. 身份证数据

`passengers.id_card` 第一版数据库保存完整身份证号。

用户端：

```text
脱敏展示
```

例如：

```text
3201********1234
```

数据库本身：

```text
保存完整值
```

后续如果项目继续强化安全，可以增加：

```text
字段加密
查询哈希
密钥管理
```

第一版暂不展开。

---

# 41. 第一版暂不做的数据库设计

暂时不做：

```text
分库
分表
读写分离
主从复制业务接入
历史表拆分
冷热数据分层
订单归档表
Redis 双写
分布式 ID
雪花算法
消息事件表
Outbox
CDC
```

先让 MySQL 单库模型正确运行。

---

# 42. 第一版核心表总览

```text
users
    用户账号

admins
    管理员账号

stations
    站点

trains
    某天具体车次

seats
    某趟车的具体座位 / 库存

passengers
    乘车人

ticket_orders
    一张车票

payments
    一次支付

payment_orders
    支付和车票订单关联

admin_audit_logs
    管理员审计日志
```

---

# 43. 最核心的数据关系

必须始终记住：

```text
一个 Train
→ 多个 Seat
```

```text
一个 User
→ 多个 TicketOrder
```

```text
一个 Passenger
→ 可以产生多个不同车次 TicketOrder
```

```text
一个 TicketOrder
→ 一个 Train
→ 一个 Seat
→ 一个 Passenger
```

```text
一个 Payment
→ 1～2 个 TicketOrder
```

以及：

```text
一张票
=
一个 TicketOrder
=
一个乘车人
=
一个具体座位
```

这是 v1 数据模型的核心。

