# 类 12306 高并发铁路票务系统｜业务规则 v1

# 1. 文档目的

本文件只定义业务规则，不定义：

* MySQL 表结构
* Redis 实现
* 分布式锁
* MQ
* API
* Go 代码
* Kubernetes

这些内容放到其他设计文档。

本文件负责回答：

```text
一张票现在是什么状态？
一个座位现在是什么状态？
一笔支付现在是什么状态？
什么动作会导致状态变化？
```

---

# 2. 核心业务对象

系统核心业务对象：

```text
Train        车次
Seat         座位
TicketOrder  车票订单
Payment      支付单
Passenger    乘车人
```

其中最重要的是三条状态线：

```text
TicketOrder 状态
Seat 状态
Payment 状态
```

三者必须分开管理。

不能使用一个字段同时表达：

```text
票有没有出票
座位有没有卖出
钱有没有支付成功
```

---

# 3. 车票订单状态

TicketOrder 表示：

> 一名乘车人在某一具体车次上的一张票。

一张票对应一个独立 TicketOrder。

状态：

```text
WAITING_PAYMENT
TICKETED
CANCELLED
RETURNED
COMPLETED
```

---

## 3.1 WAITING_PAYMENT

含义：

```text
订单已经创建
座位已经分配
座位已经锁定
用户还没有完成支付
```

用户此时可以：

```text
支付
取消订单
```

不能：

```text
修改乘车人
修改身份证号
修改车次
修改座位类型
修改座位号
```

---

## 3.2 TICKETED

含义：

```text
支付已经成功
车票已经正式出票
座位正式售出
```

此时：

```text
电子票有效
乘车人可以使用该票
```

在退票截止时间之前允许退票。

---

## 3.3 CANCELLED

含义：

```text
待支付车票已经取消
```

产生原因：

```text
用户主动取消
或
10 分钟未支付自动取消
```

进入 CANCELLED 后：

```text
电子票不存在
座位重新可售
订单不可恢复
```

如果用户还需要该车次，需要重新下单。

---

## 3.4 RETURNED

含义：

```text
已经出票的车票被用户确认退票
```

进入 RETURNED 后立即：

```text
电子票失效
座位重新可售
允许乘车人重新购买同一车次
```

注意：

```text
RETURNED
≠
资金已经退款到账
```

车票业务和资金退款必须分离。

---

## 3.5 COMPLETED

含义：

```text
该车票对应的行程已经结束
```

到达时间到达后：

```text
TICKETED
→ COMPLETED
```

进入 COMPLETED 后：

```text
不能退票
不能取消
只作为历史订单保留
```

---

# 4. TicketOrder 状态机

正常购票流程：

```text
WAITING_PAYMENT
      ↓
   支付成功
      ↓
   TICKETED
      ↓
  到达目的地
      ↓
   COMPLETED
```

取消流程：

```text
WAITING_PAYMENT
      ↓
用户取消 / 超时
      ↓
   CANCELLED
```

退票流程：

```text
TICKETED
   ↓
确认退票
   ↓
RETURNED
```

第一版不允许：

```text
CANCELLED → WAITING_PAYMENT
RETURNED  → TICKETED
COMPLETED → TICKETED
```

即：

> 已取消、已退票、已完成都是终态。

---

# 5. 座位状态

Seat 状态：

```text
AVAILABLE
LOCKED
SOLD
```

---

## 5.1 AVAILABLE

含义：

```text
该座位当前可以出售
```

用户查询余票时：

```text
AVAILABLE 数量
=
当前余票数量
```

---

## 5.2 LOCKED

含义：

```text
该座位已经被某个待支付订单占用
```

其他用户不能再次购买该座位。

座位进入 LOCKED 后：

```text
最长锁定 10 分钟
```

---

## 5.3 SOLD

含义：

```text
车票已经支付成功并出票
该座位已经正式售出
```

---

# 6. Seat 状态机

正常购票：

```text
AVAILABLE
   ↓
确认下单
   ↓
LOCKED
   ↓
支付成功
   ↓
SOLD
```

待支付取消：

```text
LOCKED
   ↓
取消 / 超时
   ↓
AVAILABLE
```

退票：

```text
SOLD
 ↓
确认退票
 ↓
AVAILABLE
```

第一版没有：

```text
REFUNDING
```

这种座位状态。

因为：

> 钱退没退到账，不影响座位业务状态。

---

# 7. 支付状态

Payment 表示：

> 一次资金支付行为。

一次 Payment 可以关联：

```text
1 张 TicketOrder
或
2 张 TicketOrder
```

支付状态：

```text
UNPAID
PAYING
SUCCESS
FAILED
REFUNDING
REFUNDED
```

---

# 8. UNPAID

含义：

```text
支付单已经创建
用户还没有正式发起支付
```

此时订单仍然：

```text
WAITING_PAYMENT
```

座位仍然：

```text
LOCKED
```

---

# 9. PAYING

含义：

```text
用户已经发起支付
当前正在等待支付结果
```

例如未来接支付宝 Sandbox：

```text
前端已经跳转支付
或
支付请求已经提交
正在等待最终支付结果
```

在 PAYING 期间：

```text
不能修改参与支付的订单
不能取消参与支付的订单
不能修改支付金额
```

---

# 10. SUCCESS

含义：

```text
该支付单已经支付成功
```

支付成功后：

```text
所有参与本次支付的 WAITING_PAYMENT TicketOrder
→ TICKETED
```

对应座位：

```text
LOCKED
→ SOLD
```

生成电子票。

---

# 11. FAILED

含义：

```text
本次支付尝试失败
```

支付失败不代表订单取消。

此时：

```text
TicketOrder
仍然 WAITING_PAYMENT

Seat
仍然 LOCKED
```

如果还没有到支付截止时间：

```text
用户可以再次发起支付
```

状态可以：

```text
FAILED
→ PAYING
```

---

# 12. REFUNDING

含义：

```text
系统已经向支付渠道发起资金退款
但资金退款尚未最终完成
```

此时可能出现：

```text
TicketOrder = RETURNED
Payment = REFUNDING
Seat = AVAILABLE
```

这是合法状态。

---

# 13. REFUNDED

含义：

```text
该支付单对应的退款流程已经完成
```

注意：

一笔支付可能对应两张票。

如果只退其中一张：

```text
支付单仍然可以保持 SUCCESS
```

同时记录：

```text
原支付金额
已退款金额
```

第一版不额外增加：

```text
PARTIALLY_REFUNDED
```

状态。

---

# 14. Payment 状态机

首次支付：

```text
UNPAID
   ↓
发起支付
   ↓
PAYING
  ├─ 成功 → SUCCESS
  └─ 失败 → FAILED
```

失败后重试：

```text
FAILED
  ↓
再次支付
  ↓
PAYING
```

退款：

```text
SUCCESS
   ↓
发起退款
   ↓
REFUNDING
   ↓
REFUNDED
```

---

# 15. 支付状态与订单状态联动

这是本项目最重要的规则之一。

---

## 15.1 下单完成

```text
TicketOrder = WAITING_PAYMENT
Seat        = LOCKED
Payment     = UNPAID
```

例如：

```text
张三
G101
二等座 21
```

此时：

```text
票：待支付
座位：被锁定
钱：未支付
```

---

## 15.2 发起支付

```text
TicketOrder = WAITING_PAYMENT
Seat        = LOCKED
Payment     = PAYING
```

订单此时不能修改。

---

## 15.3 支付失败

```text
TicketOrder = WAITING_PAYMENT
Seat        = LOCKED
Payment     = FAILED
```

只要没超时：

```text
允许重新支付
```

---

## 15.4 支付成功

```text
TicketOrder = TICKETED
Seat        = SOLD
Payment     = SUCCESS
```

同时生成：

```text
电子票
出票时间
```

---

# 16. 10 分钟支付规则

用户确认下单后开始计算：

```text
支付截止时间 = 下单时间 + 10 分钟
```

支付必须在截止时间之前确认成功。

例如：

```text
下单：10:00:00
截止：10:10:00
```

如果：

```text
10:09:00
支付成功
```

有效。

如果：

```text
10:10:01
才确认支付成功
```

则视为已经超过业务支付窗口。

第一版业务规则：

```text
超过截止时间的订单
不能正常出票
```

---

# 17. 待支付超时

如果到达支付截止时间仍未支付成功：

```text
TicketOrder
WAITING_PAYMENT
→ CANCELLED
```

座位：

```text
LOCKED
→ AVAILABLE
```

如果一次有两张待支付票：

```text
两张同时超时
→ 两张分别 CANCELLED
→ 两个座位分别释放
```

---

# 18. 用户主动取消待支付车票

用户在 WAITING_PAYMENT 状态下可以主动取消。

一张票取消：

```text
TicketOrder
WAITING_PAYMENT
→ CANCELLED
```

座位：

```text
LOCKED
→ AVAILABLE
```

---

# 19. 两张票的取消规则

假设：

```text
TicketOrder A
张三
21号
¥99

TicketOrder B
李四
22号
¥99
```

支付总额：

```text
¥198
```

在支付前允许单独取消 B。

取消后：

```text
A = WAITING_PAYMENT
B = CANCELLED

21号 = LOCKED
22号 = AVAILABLE
```

支付金额变为：

```text
¥99
```

剩余票的支付截止时间：

```text
保持原截止时间
```

不能重新延长 10 分钟。

---

# 20. 不支持部分支付

如果当前有两张 WAITING_PAYMENT 车票：

```text
A
B
```

用户不能：

```text
只支付 A
同时保留 B 待支付
```

如果只想购买 A：

```text
先取消 B
再支付 A
```

这样 Payment 的实际支付范围始终清楚。

---

# 21. 退票规则

只有：

```text
TICKETED
```

状态允许发起退票。

退票截止时间：

```text
发车前 10 分钟
```

进入停售/退票截止窗口后：

```text
不能再退票
```

---

# 22. 退票确认后的状态变化

用户确认退票后立即：

```text
TicketOrder
TICKETED
→ RETURNED
```

电子票：

```text
有效
→ 失效
```

座位：

```text
SOLD
→ AVAILABLE
```

支付：

```text
SUCCESS
→ 发起退款
```

然后：

```text
REFUNDING
→ REFUNDED
```

注意顺序：

```text
退票业务完成
≠
等待退款资金到账
```

---

# 23. 退款不影响票的有效性判断

如果：

```text
TicketOrder = RETURNED
Payment = REFUNDING
```

则：

```text
票已经无效
座位已经释放
用户可以重新购票
```

不能因为：

```text
退款还没到账
```

就重新把订单改成：

```text
TICKETED
```

---

# 24. 一次支付对应两张票

例如：

```text
Payment P001
金额 ¥198
```

对应：

```text
TicketOrder A
张三
¥99

TicketOrder B
李四
¥99
```

支付成功：

```text
Payment P001 = SUCCESS

A = TICKETED
B = TICKETED
```

之后只退 A：

```text
A = RETURNED
B = TICKETED
```

资金：

```text
原支付金额 = ¥198
已退款金额 = ¥99
```

Payment 第一版仍然可以保持：

```text
SUCCESS
```

不增加部分退款状态。

---

# 25. 订单与支付关系

重要原则：

> TicketOrder 管“票”。

> Payment 管“钱”。

不能混在一起。

例如：

```text
TicketOrder = RETURNED
Payment = REFUNDING
```

是正常状态。

又例如：

```text
TicketOrder = WAITING_PAYMENT
Payment = FAILED
```

也是正常状态。

---

# 26. 座位与支付关系

重要原则：

> Seat 不直接关心退款资金有没有到账。

座位只关心：

```text
当前这张票在业务上还有效吗？
```

因此：

```text
待支付
→ LOCKED

支付成功
→ SOLD

取消
→ AVAILABLE

退票确认
→ AVAILABLE
```

---

# 27. 乘车人有效票规则

同一身份证在同一具体车次上，只允许存在一张有效票。

以下状态视为有效占用：

```text
WAITING_PAYMENT
TICKETED
```

以下状态不再占用购票资格：

```text
CANCELLED
RETURNED
COMPLETED
```

因此退票后：

```text
乘车人可以重新购买同一车次
```

---

# 28. 一次下单最多 2 张

一次提交订单：

```text
1 张
或
2 张
```

如果购买 2 张：

```text
必须属于同一车次
必须属于同一座位类型
```

两张票：

```text
各自独立 TicketOrder
```

但：

```text
共享同一次 Payment
```

---

# 29. 座位分配规则（算法？）

购买 1 张：

```text
选择编号最小的 AVAILABLE 座位
```

购买 2 张：

第一优先：

```text
连续 AVAILABLE 座位
```

如果存在多组：

```text
优先编号较小的一组
```

例如：

```text
3、4
8、9
```

优先：

```text
3、4
```

如果不存在连续座位：

```text
选择编号最小的两个 AVAILABLE 座位
```

例如：

```text
3、8、15
```

购买 2 张：

```text
3、8
```

---

# 30. 一等座与二等座

两种座位类型完全独立。

例如：

```text
一等座：
1～20

二等座：
1～100
```

可以同时存在：

```text
一等座 1号
二等座 1号
```

它们不是同一个座位。

---

# 31. 售罄规则

某座位类型：

```text
AVAILABLE = 0
```

则：

```text
该座位类型显示售罄
```

例如：

```text
一等座：售罄
二等座：32 张
```

车次仍然可销售二等座。

只有：

```text
一等座 AVAILABLE = 0
且
二等座 AVAILABLE = 0
```

时，整个车次显示：

```text
已售罄
```

---

# 32. 退票后的重新销售

如果已售罄车次发生退票：

```text
Seat
SOLD
→ AVAILABLE
```

只要当前仍处于售票时间内：

```text
该座位立即重新可售
```

车次可以从：

```text
已售罄
```

恢复为：

```text
有票
```

---

# 33. 发车和完成

发车前 10 分钟：

```text
停止售票
停止退票
```

到达发车时间后：

```text
车次不再出现在正常查询结果
```

但已经出票订单仍然存在。

到达目的地时间后：

```text
TicketOrder
TICKETED
→ COMPLETED
```

---

# 34. 车次开售状态

车次主要状态：

```text
DRAFT
WAITING_SALE
ON_SALE
STOPPED
DEPARTED
ARCHIVED
```

业务含义：

```text
DRAFT
管理员仍可编辑

WAITING_SALE
已发布，等待开售时间

ON_SALE
正常售票

STOPPED
已进入发车前 10 分钟停售区间

DEPARTED
已发车

ARCHIVED
历史归档
```

---

# 35. 车次状态流转

```text
DRAFT
  ↓ 发布
WAITING_SALE
  ↓ 到达开售时间
ON_SALE
  ↓ 发车前10分钟
STOPPED
  ↓ 发车
DEPARTED
  ↓ 后续归档
ARCHIVED
```

开售后：

```text
车次核心业务数据不允许再修改
```

---

# 36. 第一版禁止的状态操作

第一版明确禁止：

```text
已出票直接改乘车人
已出票直接改座位
已出票直接改车次
已出票直接改座位类型
RETURNED 恢复为 TICKETED
CANCELLED 恢复为 WAITING_PAYMENT
支付成功后手动改成未支付
座位 SOLD 直接人工改 AVAILABLE
```

状态必须通过合法业务动作流转。

---

# 37. 核心原则

整个业务状态设计遵守四个原则。

## 原则一

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

## 原则二

```text
TicketOrder 管票
Payment 管钱
Seat 管座位占用
```

三者分开。

---

## 原则三

```text
下单成功
≠
买票成功
```

下单成功只代表：

```text
座位已经锁定
进入待支付
```

真正买票成功：

```text
支付成功
→ 出票
```

---

## 原则四

```text
退票成功
≠
退款资金已经到账
```

用户确认退票后：

```text
票立即失效
座位立即释放
```

资金退款异步继续处理。

---

# 38. v1 核心状态总览

## TicketOrder

```text
WAITING_PAYMENT
TICKETED
CANCELLED
RETURNED
COMPLETED
```

## Seat

```text
AVAILABLE
LOCKED
SOLD
```

## Payment

```text
UNPAID
PAYING
SUCCESS
FAILED
REFUNDING
REFUNDED
```

## Train

```text
DRAFT
WAITING_SALE
ON_SALE
STOPPED
DEPARTED
ARCHIVED
```

---

# 39. 最核心业务链

正常购票：

```text
Train = ON_SALE
Seat = AVAILABLE

↓ 用户确认下单

TicketOrder = WAITING_PAYMENT
Seat = LOCKED
Payment = UNPAID

↓ 发起支付

Payment = PAYING

↓ 支付成功

Payment = SUCCESS
TicketOrder = TICKETED
Seat = SOLD

↓ 到达目的地

TicketOrder = COMPLETED
```

取消：

```text
WAITING_PAYMENT
↓
CANCELLED

LOCKED
↓
AVAILABLE
```

退票：

```text
TICKETED
↓
RETURNED

SOLD
↓
AVAILABLE

SUCCESS
↓
REFUNDING
↓
REFUNDED
```

这三条链路就是第一版系统最核心的业务逻辑。

