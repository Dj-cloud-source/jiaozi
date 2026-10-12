# Postman 联调说明 v1

# 1. 文档目的

本文档定义 Postman 联调文件的范围、变量和执行顺序。

在创建或修改 Postman Collection、Environment 之前，必须先更新本文档。

Postman 只用于调用和验证 HTTP API，不定义业务规则。发生冲突时，依次以以下内容为准：

```text
business-rules.md
PRD.md
database.md
api.md
当前后端代码
```

如果文档和代码不一致，应先指出差异，不在 Postman 脚本中隐藏或绕过问题。

---

# 2. 运行环境

当前开发和联调环境：

```text
Mac
负责代码修改、Git commit 和 push

Windows VM 中的 Linux
负责 git pull、Go 编译、启动后端和连接 MySQL

Windows
运行 Postman 或 Apifox，调用 Linux 中的 Go 后端
```

Postman 通过 `base_url` 访问后端，不在集合中写死主机地址和端口。

---

# 3. 当前进度

第一阶段已经验证最小 HTTP 链路：

```text
健康检查
↓
用户注册
↓
用户登录
↓
获取当前用户
```

对应当前已经实现的接口：

```text
GET  /health
POST /api/v1/auth/register
POST /api/v1/auth/login
GET  /api/v1/users/me
```

第二阶段已经验证管理员登录和创建车次前需要的站点数据：

```text
管理员登录
↓
创建出发站
↓
创建到达站
↓
查询公开站点列表
↓
查询管理员站点列表
```

对应接口：

```text
POST /api/v1/admin/auth/login
POST /api/v1/admin/stations
GET  /api/v1/stations
GET  /api/v1/admin/stations
```

站点修改和停用暂不加入本阶段，避免影响下一阶段创建车次。

第三阶段验证车次从草稿到自动开售的状态流转：

```text
按名称加载出发站和到达站 ID
↓
自动生成明天的车次时间
↓
管理员创建车次
↓
管理员确认草稿
↓
发布车次
↓
等待最多 30 秒自动开售
↓
管理员确认 ON_SALE
↓
普通用户查询车次和详情
```

对应接口：

```text
POST /api/v1/admin/trains
GET  /api/v1/admin/trains
POST /api/v1/admin/trains/{train_id}/publish
GET  /api/v1/stations
GET  /api/v1/trains
GET  /api/v1/trains/{train_id}
```

创建车次请求的 Pre-request Script 自动生成：

```text
train_no
train_date
departure_time
arrival_time
sale_start_time
```

车次安排在明天，开售时间设置为当前时间前 1 分钟。发布后，后端每 30 秒运行一次的车次状态任务会将其从 `WAITING_SALE` 更新为 `ON_SALE`。

重新导入 Environment 后，站点 ID 可能为空。第三阶段第一个请求会根据 `departure_station_name` 和 `arrival_station_name` 查询公开站点列表，并重新保存两个站点 ID，不需要手工复制。

第四阶段验证用户创建订单后的锁座结果：

```text
确认目标车次仍为 ON_SALE
↓
用户填写一位乘车人并创建订单
↓
自动保存订单 ID 和支付单 ID
↓
查询订单列表和订单详情
↓
查询尚未支付的支付单
↓
再次查询车次，确认对应席别余票减少 1
```

对应接口：

```text
GET  /api/v1/trains/{train_id}
POST /api/v1/orders
GET  /api/v1/orders
GET  /api/v1/orders/{order_id}
GET  /api/v1/payments/{payment_id}
```

当前后端创建订单只支持一位乘车人，因此 Postman 的请求 Body 只放一位乘车人。创建成功后应看到：

```text
TicketOrder = WAITING_PAYMENT
Seat = LOCKED
Payment = UNPAID
```

执行第四阶段前，需要先运行“01 基础链路 / 03 用户登录”刷新 `user_token`，并确保 `train_id` 对应的车次仍然是 `ON_SALE`。如果车次已经发车、停售或变量不存在，重新完整执行第三阶段，创建一趟明天开行的新车次。

订单的支付窗口为 10 分钟。第四阶段的 6 个请求应连续执行；超过支付截止时间后，定时任务会关闭订单并释放座位。重复练习同一趟车时，如果上一张订单仍然有效，需要在创建订单 Body 中换一个身份证号，否则后端会按“一位乘客不能重复购买同一车次”返回冲突。

---

# 4. 第二阶段前置条件

管理员不能通过 HTTP 接口注册，需要先在 VMware Linux 中初始化一次：

```bash
go run ./cmd/admin-init -username admin -password 123456
```

该命令必须使用与后端服务相同的 MySQL 环境变量。示例账号只用于本地联调；如果管理员已经存在，不要重复初始化。

---

# 5. 后续完善接口

以下接口当前不加入 Postman 主流程：

```text
POST  /api/v1/auth/logout
PATCH /api/v1/users/me
PUT   /api/v1/users/me/password
```

原因：当前后端尚未实现，相关功能已经在产品和 API 文档中标记为后续完善。

以下接口也不加入当前阶段：

```text
POST /api/v1/admin/orders/{order_id}/force-cancel
POST /api/v1/admin/orders/{order_id}/retry-refund
POST /api/v1/payments/alipay/notify
```

前两个接口的异常处理规则尚未定义完整；支付宝回调属于第二阶段。

---

# 6. 输入与环境变量

为了让操作更直观，业务输入直接在请求的 `Body` 中修改：

```text
手机号、密码、昵称
管理员用户名、密码
站点名称
车次票价、座位数量
后续的乘车人和座位类型
```

Environment 模板只预置：

```text
base_url
```

示例：

```text
base_url = http://127.0.0.1:8080
```

实际地址由 Windows 到 Linux 后端的网络连接方式决定。

以下运行状态由 Postman 脚本自动写入 Environment，不要求用户手工填写：

```text
user_phone
user_nickname
user_token
admin_token
departure_station_name
arrival_station_name
departure_station_id
arrival_station_id
train_id
train_no
train_date
departure_time
arrival_time
sale_start_time
first_class_seat_count
second_class_seat_count
before_first_class_available
before_second_class_available
ordered_seat_class
order_id
payment_id
```

已经导入过 Environment 的用户只替换 Collection，保留现有 Environment，不要再次导入模板。这样 `base_url`、Token 和已保存的 ID 不会丢失。

密码、Token 和其他敏感值不得提交到 Git。仓库中的 Environment 模板不得保存真实凭据。

---

# 7. 自动保存变量

登录成功后，Postman 测试脚本可以从响应中读取：

```text
data.access_token
```

并保存到：

```text
user_token
```

管理员登录成功后保存：

```text
data.access_token
→ admin_token
```

创建两个站点后分别保存：

```text
data.id
→ departure_station_id

data.id
→ arrival_station_id
```

创建车次成功后保存：

```text
data.id
→ train_id
```

创建订单成功后保存：

```text
data.orders[0].order_id
→ order_id

data.payment_id
→ payment_id
```

后续创建资源时，可以按业务链路保存站点 ID、车次 ID、订单 ID 和支付单 ID。

自动保存变量只是减少重复复制，不得代替用户理解请求和响应。

---

# 8. 请求约定

JSON 请求统一使用：

```http
Content-Type: application/json
```

需要用户登录的接口统一使用：

```http
Authorization: Bearer {{user_token}}
```

管理员接口使用：

```http
Authorization: Bearer {{admin_token}}
```

接口地址统一写成：

```text
{{base_url}}/api/v1/...
```

---

# 9. 基础验证

每个 Postman 请求至少检查：

```text
HTTP 状态码符合预期
响应是 JSON
业务 code 等于 0，或等于本次异常场景的预期错误码
关键 data 字段存在
```

第一阶段不仅测试成功请求，还应保留少量容易理解的异常场景：

```text
重复手机号注册
错误密码登录
未携带 Token 获取当前用户
```

---

# 10. 后续联调顺序

完整业务链按照以下顺序逐步加入：

```text
发起 Mock 支付
↓
模拟支付成功
↓
查询订单和电子票
↓
取消待支付订单 / 退票
```

每次只扩展一个小阶段，并在 Windows VM 中实际验证后再继续。

---

# 11. 配置文件规则

后续计划生成：

```text
postman/jiaozi.postman_collection.json
postman/jiaozi.postman_environment.json
```

规则：

```text
先更新本文档
再修改 Postman 配置文件
不写入真实密码、Token 或服务器敏感信息
不添加当前后端不存在的接口
不使用测试脚本绕过后端业务状态机
```
