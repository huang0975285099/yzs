# 货柜结算值守 - 异常订单处理模块产品文档

## 产品概述

### 产品定位
货柜结算值守是 AI云值守智能运营平台的核心业务模块之一，主要服务于智能货柜（无人零售）场景。系统通过 AI 识别算法自动检测异常交易订单，由云端坐席人员进行人工审核与处理，实现高效的异常订单闭环管理。

### 核心价值
- **效率提升**: 一名坐席可同时处理多个货柜的异常订单，大幅降低人工成本
- **AI 辅助**: 自动识别异常类型，智能推送待处理订单
- **质量保障**: 双模式质检机制，确保处理结果准确可靠
- **数据追溯**: 全流程数据留存，支持审计与复盘

---

## 业务流程

### 整体流程图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           异常订单处理流程                                    │
└─────────────────────────────────────────────────────────────────────────────┘

     ┌──────────────┐
     │ 外部系统 API │
     │ (uboxol.com) │
     └──────┬───────┘
            │ 每30分钟同步
            ▼
     ┌──────────────┐
     │  数据同步    │ ──────► 异常订单入库 (trade_abnormals)
     │  Scheduler   │
     └──────┴───────┘
            │
            ▼
┌───────────────────────────────────────────────────────────────┐
│                        操作员处理                              │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    │
│  │ 开始处理 │───►│ 查看视频 │───►│ 选择商品 │───►│ 提交/挂起│    │
│  └─────────┘    └─────────┘    └─────────┘    └─────────┘    │
│       │              │              │              │         │
│       ▼              ▼              ▼              ▼         │
│  记录点击次数    确认消费情况   选择商品明细    生成处理结果    │
└───────────────────────────┬───────────────────────────────────┘
                            │
              ┌─────────────┴─────────────┐
              │                           │
              ▼                           ▼
    ┌─────────────────┐         ┌─────────────────┐
    │   审核模式      │         │   直通模式      │
    │ (ReviewEnabled) │         │ (!ReviewEnabled)│
    └─────────┬───────┘         └─────────┬───────┘
              │                           │
              ▼                           ▼
    ┌─────────────────┐         ┌─────────────────┐
    │  存入质检队列   │         │  直接调外部API  │
    │ (trade_reviews) │         │ (submit/pend)   │
    └─────────┬───────┘         └─────────┬───────┘
              │                           │
              ▼                           ▼
    ┌─────────────────┐         ┌─────────────────┐
    │  质检员审核     │         │  质检员复查     │
    │  (通过/驳回)    │         │  (正常/异常)    │
    └─────────┬───────┘         └─────────┬───────┘
              │                           │
              ▼                           ▼
    ┌─────────────────┐         ┌─────────────────┐
    │  调外部API提交  │         │  记录复查结果   │
    └─────────┬───────┘         └─────────┬───────┘
              │                           │
              └─────────────┬─────────────┘
                            ▼
                   ┌─────────────────┐
                   │  处理完成       │
                   │  统计数据更新   │
                   └─────────────────┘
```

---

## 功能模块详解

### 1. 订单状态检查 (CheckTradeStatus)

#### 功能说明
在操作员打开处理表单前，系统会先向外部系统确认订单是否仍处于"未处理"状态，避免重复处理。

#### 业务逻辑

```go
// 状态检查流程
1. 查询本地数据库，检查订单是否已标记为已处理
   - 如果已处理 → 返回 "该订单已处理"
   
2. 调用外部 API 查询订单是否仍在未处理列表中
   - 构造查询参数：
     - operatingModeList: [21] (货柜模式)
     - handleStatus: "NOT_HANDLED"
     - pendStatus: "NO_PENDING"
     - outOrderNo: 订单商户单号
     - 时间范围: 订单创建时间 ~ 当天
   
3. 判断返回结果
   - total > 0 → 订单仍存在，可处理
   - total = 0 → 订单已被外部处理，更新本地状态
   
4. 更新本地状态（如果外部已处理）
   - is_handled: true
   - handled_by_name: "外部系统"
   - handle_status_desc: "客服已处理"
```

#### API 接口

| 接口 | 方法 | 路径 |
|------|------|------|
| 订单状态检查 | GET | `/api/trades/:id/check` |

#### 返回数据

```json
{
  "code": 200,
  "data": {
    "alreadyHandled": false,  // 是否已被处理
    "message": ""             // 提示信息（如果已处理）
  }
}
```

---

### 2. 订单提交处理 (SubmitTrade)

#### 功能说明
操作员确认订单中的商品消费明细后，提交处理结果。根据系统配置，可选择"审核模式"或"直通模式"。

#### 请求数据结构

```typescript
interface SubmitRequest {
  orderGoodsDetailList: GoodsItem[];  // 商品明细列表
  duration: number;                   // 作业耗时（秒）
  remark: string;                     // 处理备注
}

interface GoodsItem {
  goodsId: number;      // 商品ID
  goodsName: string;    // 商品名称
  goodsPrice: number;   // 商品单价
  goodsImage: string;   // 商品图片URL
  type: number;         // 商品类型（默认1）
  goodsCount: number;   // 商品数量
}
```

#### 业务逻辑

##### 审核模式 (ReviewEnabled = true)

```
┌─────────────────────────────────────────────────────────────┐
│                     审核模式流程                             │
└─────────────────────────────────────────────────────────────┘

1. 参数验证
   - 检查订单是否存在
   - 检查订单是否已处理
   - 检查订单是否已在质检队列
   
2. 创建质检记录 (trade_reviews)
   - trade_id: 订单ID
   - action_type: "submit"
   - goods_json: 商品明细JSON
   - operator_remark: 操作员备注
   - duration: 作业耗时
   - submitted_by_id: 操作员ID
   - submitted_by_name: 操作员姓名
   - submitted_at: 提交时间
   - review_status: "pending"
   
3. 更新订单状态 (trade_abnormals)
   - review_status: "pending"
   - handled_by_id: 操作员ID
   - handled_by_name: 操作员姓名
   - handled_at: 当前时间
   - handle_duration: 作业耗时
   - handle_goods: 商品明细JSON
   - handle_remark: 处理备注
   - 解锁订单（清除 locked_by_id, locked_at）
   
4. 记录统计数据
   - incrementDailyStats(userID, "submit")
   
5. 返回结果
   - "已提交质检审核"
```

##### 直通模式 (ReviewEnabled = false)

```
┌─────────────────────────────────────────────────────────────┐
│                     直通模式流程                             │
└─────────────────────────────────────────────────────────────┘

1. 参数验证（同审核模式）

2. 直接调用外部 API
   - POST https://api.uboxol.com/lotus/trade/abnormal/handle
   - 请求参数：
     {
       "orderGoodsDetailList": [...],
       "outOrderNo": "商户单号",
       "handleUsername": "prisonProject"
     }
   
3. 根据外部返回结果更新本地状态
   - success = true:
     - is_handled: true
     - handled_by_name: 操作员姓名
     - 记录统计数据
     - 返回 "本订单处理成功"
     
   - success = false:
     - is_handled: true
     - handled_by_name: "外部系统"
     - 返回外部系统的错误信息
     
4. 解锁订单
```

#### API 接口

| 接口 | 方法 | 路径 |
|------|------|------|
| 提交处理 | POST | `/api/trades/:id/submit` |

---

### 3. 订单挂起 (PendTrade)

#### 功能说明
当操作员无法确认订单的消费情况时，可将订单挂起，等待后续处理。

#### 请求数据结构

```typescript
interface PendRequest {
  duration: number;   // 作业耗时（秒）
  remark: string;     // 挂起原因/备注
}
```

#### 业务逻辑

##### 审核模式

```
1. 创建质检记录 (trade_reviews)
   - action_type: "pend"
   - operator_remark: 挂起原因
   - review_status: "pending"
   
2. 更新订单状态
   - review_status: "pending"
   - pend_status: "PENDING"
   - pend_status_desc: "已挂起"
   
3. 记录统计数据
   - incrementDailyStats(userID, "pend")
   
4. 返回 "已提交质检审核"
```

##### 直通模式

```
1. 调用外部 API
   - POST https://api.uboxol.com/lotus/trade/abnormal/pend
   - 请求参数：
     {
       "id": tradeId,
       "pendStatus": "PENDING"
     }
   
2. 更新本地状态
   - is_handled: true
   - pend_status: "PENDING"
   - pend_status_desc: "已挂起"
   
3. 记录统计数据
   
4. 返回结果
```

#### API 接口

| 接口 | 方法 | 路径 |
|------|------|------|
| 挂起订单 | POST | `/api/trades/:id/pend` |

---

### 4. 内部标记处理 (HandleTrade)

#### 功能说明
简单的内部标记处理，不调用外部 API。主要用于特殊场景下的快速标记。

#### 业务逻辑

```
1. 查询订单是否存在
2. 检查订单是否已处理
3. 直接更新本地状态
   - is_handled: true
   - handled_by_id: 操作员ID
   - handled_by_name: 操作员姓名
   - handled_at: 当前时间
   - handle_remark: 备注信息
4. 保存并返回
```

#### API 接口

| 接口 | 方法 | 路径 |
|------|------|------|
| 内部标记 | POST | `/api/trades/:id/handle` |

---

### 5. 我的处理记录查询 (ListMyHandled)

#### 功能说明
操作员查看自己已处理的订单列表，包括待审核和已完成的订单。

#### 查询参数

| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码（默认1） |
| size | int | 每页数量（默认20） |
| date | string | 日期过滤（YYYY-MM-DD） |

#### 返回数据结构

```typescript
interface MyHandledResponse {
  code: 200;
  data: {
    records: MyHandledRecord[];
    total: number;
    page: number;
    size: number;
    totalAmount: number;        // 当日处理金额
    cumulativeSubmit: number;   // 累计提交数
    cumulativePend: number;     // 累计挂起数
    cumulativeAmount: number;   // 累计处理金额
  };
}

interface MyHandledRecord {
  id: number;
  tradeId: number;
  outOrderNo: string;
  nodeName: string;
  createTime: string;
  handledAt: string;
  handledByName: string;
  handleRemark: string;
  handleGoods: string;          // JSON 商品明细
  pendStatusDesc: string;
  reviewStatus: string;         // '' | 'pending'
  actionType: string;           // 'submit' | 'pend'
  isHandled: boolean;
}
```

#### 统计计算逻辑

系统使用 MySQL JSON_TABLE 函数在 SQL 层直接计算商品金额总和，避免将 JSON 数据传回 Go 层解析：

```sql
SELECT COALESCE(SUM(jt.price * jt.cnt), 0) AS total_amount
FROM trade_abnormals t
CROSS JOIN JSON_TABLE(
  t.handle_goods,
  '$[*]' COLUMNS (
    price DOUBLE PATH '$.goodsPrice',
    cnt INT PATH '$.goodsCount'
  )
) jt
WHERE t.handle_goods != '' 
  AND t.handle_goods IS NOT NULL
  AND t.handled_by_id = ?
  AND t.handled_by_name != '外部系统'
```

#### API 接口

| 接口 | 方法 | 路径 |
|------|------|------|
| 我的处理记录 | GET | `/api/trades/my-handled` |

---

### 6. 数据看板统计 (GetStats)

#### 功能说明
为管理后台提供数据统计，用于数据看板展示。

#### 返回数据结构

```typescript
interface StatsResponse {
  code: 200;
  data: {
    total: number;              // 异常订单总数
    todayCount: number;         // 今日新增
    handledCount: number;       // 已处理数量
    unhandledCount: number;     // 待处理数量
    pendingReviewCount: number; // 待质检数量
    dailyCounts: DayCount[];    // 近7天趋势
    typeCounts: TypeCount[];    // 异常类型分布
    opStats: OpStat[];          // 操作员处理排行（管理员/统计员可见）
  };
}

interface DayCount {
  day: string;    // 日期 YYYY-MM-DD
  count: number;  // 数量
}

interface TypeCount {
  name: string;   // 异常类型描述
  value: number;  // 数量
}

interface OpStat {
  name: string;   // 操作员姓名
  value: number;  // 处理数量
}
```

#### API 接口

| 接口 | 方法 | 路径 |
|------|------|------|
| 数据统计 | GET | `/api/stats` |

---

## 两种处理模式对比

### 模式配置

系统通过环境变量 `REVIEW_ENABLED` 控制处理模式：

```
REVIEW_ENABLED=true   → 审核模式
REVIEW_ENABLED=false  → 直通模式
```

### 流程对比表

| 对比项 | 审核模式 | 直通模式 |
|--------|----------|----------|
| **提交后状态** | 存入质检队列，等待审核 | 直接调用外部 API |
| **质检时机** | 处理前审核 | 处理后复查 |
| **质检员角色** | 审核员（需批准后才生效） | 复查员（事后标记正常/异常） |
| **统计数据** | 提交时即计入（不等审核） | 提交成功后计入 |
| **外部 API 调用** | 质检通过后调用 | 提交时立即调用 |
| **适用场景** | 高风险订单、新手操作员 | 低风险订单、熟练操作员 |

### 审核模式详细流程

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   操作员     │────►│   质检队列   │────►│   质检员     │
│  提交处理    │     │ (pending)    │     │   审核      │
└──────────────┘     └──────────────┘     └──────┬───────┘
                                                  │
                            ┌─────────────────────┴─────┐
                            │                           │
                            ▼                           ▼
                   ┌──────────────┐           ┌──────────────┐
                   │   审核通过   │           │   外部已处理  │
                   │  调外部API  │           │   未外发     │
                   └──────┬───────┘           └──────┬───────┘
                          │                          │
                          ▼                          ▼
                   ┌──────────────┐           ┌──────────────┐
                   │  订单完成    │           │  自动标记    │
                   │  状态更新    │           │  已处理      │
                   └──────────────┘           └──────────────┘
```

### 直通模式详细流程

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   操作员     │────►│  外部 API    │────►│  订单完成    │
│  提交处理    │     │  直接调用    │     │              │
└──────────────┘     └──────────────┘     └──────┬───────┘
                                                  │
                                                  ▼
                                         ┌──────────────┐
                                         │   质检员     │
                                         │   事后复查   │
                                         └──────┬───────┘
                                                │
                          ┌─────────────────────┴─────┐
                          │                           │
                          ▼                           ▼
                 ┌──────────────┐           ┌──────────────┐
                 │   复查正常   │           │   复查异常   │
                 │  inspect_   │           │  inspect_   │
                 │  status=    │           │  status=    │
                 │  normal     │           │  abnormal   │
                 └──────────────┘           └──────────────┘
```

---

## 外部系统集成

### API 接口列表

| 功能 | 外部 API 地址 | 调用时机 |
|------|---------------|----------|
| 订单列表查询 | `/lotus/trade/abnormal/page` | 状态检查、数据同步 |
| 提交处理 | `/lotus/trade/abnormal/handle` | 直通模式提交、审核通过后 |
| 挂起订单 | `/lotus/trade/abnormal/pend` | 直通模式挂起、审核通过后 |

### 请求参数示例

#### 查询未处理订单

```json
{
  "operatingModeList": [21],
  "handleStatus": "NOT_HANDLED",
  "pendStatus": "NO_PENDING",
  "outOrderNo": "2024010112345678",
  "current": 1,
  "size": 5,
  "startCreateTime": "2024-01-01 00:00:00",
  "endCreateTime": "2024-01-01 23:59:59"
}
```

#### 提交处理

```json
{
  "orderGoodsDetailList": [
    {
      "goodsId": 10001,
      "goodsName": "可乐",
      "goodsPrice": 3.5,
      "goodsImage": "https://...",
      "type": 1,
      "goodsCount": 2
    }
  ],
  "outOrderNo": "2024010112345678",
  "handleUsername": "prisonProject"
}
```

#### 挂起订单

```json
{
  "id": 12345678,
  "pendStatus": "PENDING"
}
```

### 返回结果结构

```json
{
  "code": 200,
  "success": true,
  "message": "处理成功",
  "data": {
    "total": 10,
    "records": [...]
  }
}
```

---

## 数据模型

### trade_abnormals（异常订单表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| trade_id | int64 | 外部订单ID（唯一索引） |
| inner_code | string | 机器编号 |
| node_name | string | 货柜名称 |
| out_order_no | string | 商户单号 |
| transaction_id | string | 交易流水号 |
| abnormal_type_desc | string | 异常类型描述 |
| abnormal_desc | string | 异常详情描述 |
| create_time | string | 订单创建时间 |
| door_open_time | string | 开门时间 |
| door_close_time | string | 关门时间 |
| is_handled | bool | 是否已处理 |
| handled_by_id | uint | 处理人ID |
| handled_by_name | string | 处理人姓名 |
| handled_at | datetime | 处理时间 |
| handle_duration | int | 处理耗时（秒） |
| handle_goods | text | 商品明细JSON |
| handle_remark | string | 处理备注 |
| pend_status | string | 挂起状态 |
| pend_status_desc | string | 挂起状态描述 |
| review_status | string | 质检状态（'' 或 'pending'） |
| inspect_status | string | 复查状态（'' 或 'normal' 或 'abnormal'） |
| inspect_remark | string | 复查备注 |
| inspected_by_id | uint | 复查人ID |
| inspected_by_name | string | 复查人姓名 |
| inspected_at | datetime | 复查时间 |
| locked_by_id | uint | 锁定人ID |
| locked_at | datetime | 锁定时间 |
| synced_at | datetime | 同步时间 |

### trade_reviews（质检记录表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| trade_id | uint | 订单ID |
| action_type | string | 操作类型（submit/pend） |
| goods_json | text | 商品明细JSON |
| operator_remark | string | 操作员备注 |
| duration | int | 作业耗时 |
| submitted_by_id | uint | 提交人ID |
| submitted_by_name | string | 提交人姓名 |
| submitted_at | datetime | 提交时间 |
| review_status | string | 审核状态（pending/approved） |
| reviewed_by_id | uint | 审核人ID |
| reviewed_by_name | string | 审核人姓名 |
| reviewed_at | datetime | 审核时间 |
| review_remark | string | 审核备注 |

---

## 统计数据

### 统计维度

| 维度 | 说明 | 数据来源 |
|------|------|----------|
| 点击开始 | 点击"开始处理"按钮次数 | daily_stats.start_count |
| 点击跳过 | 点击"跳过"按钮次数 | daily_stats.skip_count |
| 提交处理 | 确认提交的订单数 | trade_reviews + trade_abnormals |
| 挂起订单 | 挂起的订单数 | trade_reviews + trade_abnormals |
| 处理金额 | 提交处理的商品总金额 | JSON_TABLE 计算 |

### 统计更新时机

```
操作员提交 → incrementDailyStats(userID, "submit" | "pend")
                ↓
           daily_stats 表更新
                ↓
           对应字段 +1
```

### 统计查询接口

| 接口 | 说明 |
|------|------|
| `/api/me/daily-stats` | 当前用户每日统计 |
| `/api/stats/operator-records` | 操作员处理记录 |
| `/api/stats/operator-range` | 操作员时间段统计 |
| `/api/stats/daily` | 每日统计汇总 |

---

## 错误处理

### 错误码定义

| 错误码 | 说明 | 处理建议 |
|--------|------|----------|
| 400 | 参数错误 | 检查请求参数格式 |
| 401 | 未授权 | 重新登录 |
| 403 | 无权限 | 检查用户角色 |
| 404 | 记录不存在 | 检查订单ID |
| 500 | 服务器内部错误 | 联系技术人员 |
| 502 | 外部接口网络异常 | 等待后重试 |
| 503 | 无法确认订单状态 | 联系管理人员 |

### 外部系统异常处理

当外部 API 返回 `success=false` 时：

1. **订单已被外部处理**: 自动标记为已处理，处理人记为"外部系统"
2. **网络超时**: 返回 502 错误，提示用户稍后重试
3. **其他错误**: 显示外部系统返回的错误信息

---

## 安全机制

### 订单锁定

防止多操作员同时处理同一订单：

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  打开订单    │────►│  锁定订单    │────►│   处理中     │
│              │     │ locked_by_id │     │              │
└──────────────┘     └──────────────┘     └──────┬───────┘
                                                  │
                           ┌──────────────────────┴─────┐
                           │                            │
                           ▼                            ▼
                  ┌──────────────┐            ┌──────────────┐
                  │  提交/挂起   │            │   跳过/退出  │
                  │  自动解锁    │            │   手动解锁   │
                  └──────────────┘            └──────────────┘
```

### 权限控制

| 角色 | 权限 |
|------|------|
| admin | 全部权限 |
| operator | 处理订单、查看自己的记录 |
| inspector | 质检审核/复查 |
| statistician | 查看统计数据 |

---

## 性能优化

### JSON_TABLE 金额计算

使用 MySQL JSON_TABLE 函数在数据库层直接计算金额总和，避免将大量 JSON 数据传回应用层解析：

```sql
-- 传统方式：Go 层解析 JSON 循环累加（性能差）
-- 优化方式：SQL 层直接计算

SELECT SUM(jt.price * jt.cnt) AS total_amount
FROM trade_abnormals t
CROSS JOIN JSON_TABLE(
  t.handle_goods,
  '$[*]' COLUMNS (
    price DOUBLE PATH '$.goodsPrice',
    cnt INT PATH '$.goodsCount'
  )
) jt
```

### 索引设计

| 表 | 索引字段 | 说明 |
|------|----------|------|
| trade_abnormals | trade_id | 外部订单唯一索引 |
| trade_abnormals | handled_by_id | 处理人查询 |
| trade_abnormals | handled_at | 时间范围查询 |
| trade_abnormals | review_status | 质检状态查询 |
| trade_reviews | trade_id | 订单关联 |
| trade_reviews | submitted_by_id | 提交人查询 |
| trade_reviews | submitted_at | 时间范围查询 |

---

## 配置项

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| REVIEW_ENABLED | true | 是否启用审核模式 |
| DB_HOST | 127.0.0.1 | 数据库地址 |
| DB_PORT | 3306 | 数据库端口 |
| DB_USER | root | 数据库用户 |
| DB_PASSWORD | 123@123qwe | 数据库密码 |
| DB_NAME | go_yzs | 数据库名 |
| SERVER_PORT | 18881 | 服务端口 |
| JWT_SECRET | go-yzs-secret-key-2026 | JWT 密钥 |
| REDIS_HOST | 127.0.0.1 | Redis 地址 |
| REDIS_PORT | 6379 | Redis 端口 |

---

## 附录

### API 接口汇总

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 订单状态检查 | GET | `/api/trades/:id/check` | 检查订单是否可处理 |
| 提交处理 | POST | `/api/trades/:id/submit` | 提交商品明细处理 |
| 挂起订单 | POST | `/api/trades/:id/pend` | 挂起订单等待后续处理 |
| 内部标记 | POST | `/api/trades/:id/handle` | 内部快速标记 |
| 我的处理记录 | GET | `/api/trades/my-handled` | 查询自己的处理记录 |
| 数据统计 | GET | `/api/stats` | 看板统计数据 |

### 操作快捷键

| 快捷键 | 功能 |
|--------|------|
| Esc | 跳过当前订单 |
| G | 挂起当前订单 |
| Enter | 提交处理 |

---

## 版本信息

- **文档版本**: v1.0
- **更新日期**: 2026-04-16
- **适用系统**: AI云值守 v1.0