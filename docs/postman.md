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

第五阶段验证 Mock 支付成功和电子票生成：

```text
发起支付
UNPAID → PAYING
↓
模拟支付成功
Payment: PAYING → SUCCESS
TicketOrder: WAITING_PAYMENT → TICKETED
Seat: LOCKED → SOLD
↓
查询支付单和订单详情
↓
查询电子票
```

对应接口：

```text
POST /api/v1/payments/{payment_id}/pay
POST /api/v1/payments/{payment_id}/mock/success
GET  /api/v1/payments/{payment_id}
GET  /api/v1/orders/{order_id}
GET  /api/v1/tickets/{order_id}
```

第五阶段依赖第四阶段刚刚创建的 `payment_id` 和 `order_id`，并且必须在 10 分钟支付截止时间内开始执行。Mock 成功接口只用于本地开发联调，不代表真实支付渠道。

第六阶段验证已出票订单退票和电子票失效：

```text
确认订单为 TICKETED
↓
用户发起退票
TicketOrder: TICKETED → RETURNED
Seat: SOLD → AVAILABLE
↓
查询订单和支付单
↓
查询历史电子票，确认 valid = false
↓
查询车次，确认对应席别余票恢复
```

对应接口：

```text
GET  /api/v1/orders/{order_id}
POST /api/v1/orders/{order_id}/return
GET  /api/v1/payments/{payment_id}
GET  /api/v1/tickets/{order_id}
GET  /api/v1/trains/{train_id}
```

第六阶段直接使用第五阶段成功出票的 `order_id`、`payment_id` 和原车次变量。退票必须发生在发车前 10 分钟以前。

当前后端的 Mock 退款实现比 `api.md` 中的目标响应简单：退票接口实际返回 `data.status = RETURNED`；Payment 暂时保持 `SUCCESS`，并将票价累计到 `refunded_amount`。当前代码尚未实现 `REFUNDING → REFUNDED` 的异步资金退款状态流转，Postman 按当前真实行为验证。

第七阶段验证用户主动取消待支付订单：

```text
记录当前余票
↓
创建一张新的待支付订单
TicketOrder = WAITING_PAYMENT
Seat = LOCKED
Payment = UNPAID
↓
用户主动取消订单
TicketOrder: WAITING_PAYMENT → CANCELLED
Seat: LOCKED → AVAILABLE
Payment: payable_amount → 0.00
↓
查询订单、支付单和车次余票
```

对应接口：

```text
GET  /api/v1/trains/{train_id}
POST /api/v1/orders
POST /api/v1/orders/{order_id}/cancel
GET  /api/v1/orders/{order_id}
GET  /api/v1/payments/{payment_id}
```

第七阶段会创建新订单，并覆盖 Environment 中用于后续请求的 `order_id` 和 `payment_id`；这只是更新 Postman 当前指向的资源，不会删除之前已经退票的数据库记录。乘车人信息仍由用户直接在创建订单 Body 中填写。

当前后端只支持一位乘车人，因此取消这张订单后 Payment 的 `payable_amount` 应为 `0.00`，状态仍为 `UNPAID`。取消接口的实际响应仅包含 `order_id` 和 `status`，没有 `api.md` 示例中的嵌套 `payment` 字段，Postman 按当前真实响应验证。

第八阶段验证支付失败后重新支付：

```text
记录当前余票并创建新订单
↓
第一次发起支付
Payment: UNPAID → PAYING
↓
模拟支付失败
Payment: PAYING → FAILED
TicketOrder 仍为 WAITING_PAYMENT
Seat 仍为 LOCKED
↓
再次发起支付
Payment: FAILED → PAYING
↓
模拟支付成功
Payment: PAYING → SUCCESS
TicketOrder: WAITING_PAYMENT → TICKETED
Seat: LOCKED → SOLD
↓
查询支付单和电子票
```

对应接口：

```text
GET  /api/v1/trains/{train_id}
POST /api/v1/orders
POST /api/v1/payments/{payment_id}/pay
POST /api/v1/payments/{payment_id}/mock/fail
GET  /api/v1/payments/{payment_id}
POST /api/v1/payments/{payment_id}/mock/success
GET  /api/v1/tickets/{order_id}
```

第八阶段会再创建一张新订单，并覆盖 Environment 中当前的 `order_id` 和 `payment_id`。支付失败不会延长或重置订单原有的 10 分钟支付窗口，因此本阶段请求应连续执行。

第九阶段对应文件夹 `10 管理员用户、订单与审计查询`，包含 6 个只读请求：

```text
读取当前用户 ID（用户 Token）
→ 管理员查询用户列表
→ 管理员查询该用户详情及历史订单
→ 按手机号、车次号查询订单
→ 管理员查看当前 order_id 的订单详情
→ 查询管理员审计日志
```

执行前分别运行 `01 / 03 用户登录` 和 `03 / 01 管理员登录`，使用创建订单的用户账号。保留前面保存的 `order_id`、`train_id`、`train_no`。本模块不要求订单仍为 TICKETED，取消、退票和行程完成后的历史订单也可以查询。

第一个请求从 `GET /api/v1/users/me` 自动保存 `admin_query_user_id` 和 `user_phone`。后续管理员请求统一携带 `admin_token`。用户详情按 ID 查询，不依赖用户恰好出现在列表第一页。

在请求的 Params 中修改 `page` 和 `page_size` 可翻页。订单列表默认用 `phone={{user_phone}}`、`train_no={{train_no}}` 筛选，另提供默认不勾选的 `status=TICKETED` 参数。勾选后按该状态筛选；空列表可能只是没有匹配记录，测试检查返回记录符合筛选条件，不要求指定订单必须在当前页。

审计接口 `GET /api/v1/admin/audit-logs` 当前只支持分页，返回操作人、动作、资源类型、资源 ID 和时间。空数组是合法结果；只读查询本身不生成操作日志，本阶段不为了制造日志而修改业务数据。非空时检查每条记录的字段，不假定某条历史操作一定出现在第一页。

实际管理员订单详情包含完整身份证号，联调请继续使用虚构乘车人信息。

第十阶段对应 `11 订单超时自动取消`，分成两个子文件夹，共 9 个请求：

1. `A 创建并检查`：记录车次余票、创建未支付订单、确认锁座、确认支付单 UNPAID。
2. 人工等待创建响应中的 `payment_deadline` 到达，再等待约 30 秒，让后台任务执行。
3. `B 到期后验证`：确认订单 CANCELLED、支付金额归零、余票恢复、发起支付被拒绝、没有电子票。

执行前刷新创建订单用户的登录，确保 `train_id` 对应车次 ON_SALE、有所选席别余票，且测试结束前不会停售，可以使用第 04 组创建的明日车次。本组仅测试从未发起支付的 UNPAID 订单，不调用主动取消接口；PAYING 超时留待后续场景。

先只执行 A，再等待，最后执行 B。不要一次运行整个文件夹或整个 Collection：脚本不自动等待 10 分钟。Go 服务和 MySQL 在等待期间保持运行。后端当前默认每 30 秒检查超时订单；若已过截止时间但仍 WAITING_PAYMENT，稍等后重发 B 的第一个请求，持续不变化时检查 Linux 后端日志中的 `cancel expired orders failed`。

本组自动保存 `timeout_order_id`、`timeout_payment_id`、`timeout_train_id`、`timeout_payment_deadline`、`timeout_ordered_seat_class` 和两种席别的 `timeout_before_*_available`，不覆盖原来的订单变量。截止时间以创建响应为准，不修改服务器时间或数据库。等待期间不要切换用户账号，也不要对同车次做其他购票、取消或退票，否则余票对比会受到影响。

预期结果：订单 CANCELLED 且有 cancelled_at；支付单仍 UNPAID，payable_amount 为 0；余票恢复到 A 下单前；发起支付返回 HTTP 409 / code 44002；查询电子票返回 HTTP 409 / code 45001。这两个预期失败请求的 Tests 应显示绿色。

---

第十一阶段对应 `12 管理员修改与停用站点`，共 7 个请求：

```text
创建独立测试站点（ACTIVE）
→ 公开列表确认可见
→ 修改站点名称
→ 公开列表确认新名称和原 ID
→ 停用站点（DISABLED）
→ 公开列表确认不再可见
→ 管理员列表确认仍保留停用记录
```

执行前运行 `03 管理员与站点 / 01 管理员登录`。本阶段无需用户登录、车次或订单。创建和修改请求的 Body 分别填写测试名称，默认是 `联调维护站01` 和 `联调维护站01改名`。后端会去掉名称两端空格。

创建成功自动保存 `maintenance_station_id`、`maintenance_station_name`；修改成功更新该名称变量。只使用本阶段独立 ID，不覆盖原来的出发站和到达站变量。请勿将测试站点用于创建车次。遇到失败先停下处理，再继续后续请求。

重复整组测试时，创建和修改 Body 都换一组未使用的名称（例如 02）。停用并非删除，停用站点的名称仍受唯一约束；重名会返回 HTTP 409 / code 42004。不要重复执行停用请求，已停用站点再次停用会返回 HTTP 409 / code 42007。

对应接口：`POST /api/v1/admin/stations`、`PATCH /api/v1/admin/stations/{station_id}`、`POST /api/v1/admin/stations/{station_id}/disable`、`GET /api/v1/stations`、`GET /api/v1/admin/stations`。创建、修改、停用操作会分别产生 CREATE_STATION、UPDATE_STATION、DISABLE_STATION 审计日志，可用第 10 组查看。

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

`POST /api/v1/payments/{payment_id}/mock/fail` 只用于本地支付失败与重试联调，不代表真实支付渠道回调。

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
admin_query_user_id
user_nickname
user_token
admin_token
maintenance_station_id
maintenance_station_name
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
timeout_order_id
timeout_payment_id
timeout_train_id
timeout_payment_deadline
timeout_ordered_seat_class
timeout_before_first_class_available
timeout_before_second_class_available
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
管理员草稿车次修改与归档
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
