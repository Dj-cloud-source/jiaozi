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

第二阶段验证管理员登录和创建车次前需要的站点数据：

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

# 6. 环境变量

Postman Environment 计划包含：

```text
base_url
user_phone
user_password
user_nickname
user_token
admin_username
admin_password
admin_token
departure_station_name
arrival_station_name
station_id
departure_station_id
arrival_station_id
train_id
order_id
payment_id
```

初始阶段只需要：

```text
base_url
user_phone
user_password
user_nickname
user_token
admin_username
admin_password
admin_token
departure_station_name
arrival_station_name
departure_station_id
arrival_station_id
```

示例：

```text
base_url = http://127.0.0.1:8080
user_phone = 13800138000
user_password = 123456
user_nickname = zhangsan
admin_username = admin
admin_password = 123456
departure_station_name = 南京南
arrival_station_name = 上海虹桥
```

实际地址由 Windows 到 Linux 后端的网络连接方式决定。

上述账号和站点信息仅为本地联调示例。重复执行注册请求前，需要更换 `user_phone`；重复创建站点前，需要更换对应站点名称。

密码、Token 和其他敏感值不得提交到 Git。仓库中的 Environment 文件只能保存空值或示例值。

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
创建并发布车次
↓
等待车次开售
↓
用户查询车次
↓
创建订单并锁座
↓
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
