# go-xxl-admin

> Go 语言实现的 XXL-JOB 管理控制台（Admin 端），负责执行器注册发现、任务下发调度、执行结果回调收集、以及强杀/日志拉取等运维操作。

---

## 1. 项目概览

本项目是 [XXL-JOB](https://www.xuxueli.com/xxl-job/) 分布式任务调度平台中 **Admin（调度中心）** 角色的 Go 语言移植版。它与标准的 Java XXL-JOB Executor（执行器）通过 HTTP + JSON 协议互通，遵循 XXL-JOB 的注册/回调/下发/强杀/日志拉取协议规范。

### 已实现功能

| 功能 | 状态 | 说明 |
|------|------|------|
| 执行器注册 | ✅ | Executor 通过 `/api/registry` 注册到 Admin，持久化到 SQLite + 内存 |
| 心跳保洁 & 剔除 | ✅ | 后台协程每 10s 扫描，剔除 90s 无心跳的执行器 |
| 轮询负载均衡 | ✅ | 下发任务时从存活节点中轮询选择 |
| 任务下发（触发） | ✅ | Admin 向 Executor 的 `/run` 端点 POST 任务参数 |
| 执行结果回调 | ✅ | Executor 完成后 POST `/api/callback`，Admin 记录结果 |
| 强杀任务 | ✅ | Admin 向 Executor 的 `/kill` 端点发送终止指令 |
| 日志拉取 | ✅ | Admin 从 Executor 的 `/log` 端点拉取执行日志 |
| Cron 调度器 | ❌ | 尚未实现定时触发，目前仅支持手动/API 触发 |
| Web 管理 UI | ❌ | 尚未实现前端界面 |

---

## 2. 目录结构

```
go-xxl-admin/
├── main.go                  # 应用入口：初始化 DB、启动清理协程、启动 Gin HTTP 服务
├── go.mod / go.sum          # Go 模块依赖管理
├── xxl_job.db               # SQLite 数据库文件（运行时生成）
│
├── global/                  # 全局共享资源
│   └── db.go                #   全局 DB 变量 + GORM/SQLite 初始化
│
├── models/                  # 数据模型层
│   ├── protocol.go          #   通信协议 DTO：XxlResponse、RegistryParam、CallbackRequest
│   ├── client_req.go        #   执行器请求/响应 DTO：KillRequest、LogRequest、LogResultContent
│   ├── job_info.go          #   job_info 表 GORM 模型（任务定义）
│   ├── job_log.go           #   job_log 表 GORM 模型（执行日志）
│   └── job_registry.go      #   job_registry 表 GORM 模型（执行器注册信息）
│
├── core/                    # 核心业务逻辑层
│   ├── registry.go          #   内存注册中心：节点存储、心跳剔除、轮询选举
│   └── executor.go          #   执行器 HTTP 客户端：下发触发、强杀、日志拉取
│
└── handlers/                # HTTP 处理层（Gin 路由控制器）
    ├── registry.go          #   POST /api/registry — 执行器注册处理
    ├── callback.go          #   POST /api/callback  — 执行结果回调处理
    └── KillJob.go           #   POST /test/kill      — 测试强杀手动端点
```

---

## 3. 整体架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Go XXL-Admin (调度中心)                          │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                   HTTP Layer (Gin :8081)                          │   │
│  │                                                                    │   │
│  │   POST /api/registry ──→ handlers.HandlerRegistry                 │   │
│  │   POST /api/callback  ──→ handlers.HandleCallBack                 │   │
│  │   POST /test/kill     ──→ handlers.HandlerKillJob                 │   │
│  └──────────┬─────────────────────────┬──────────────────────────────┘   │
│             │                         │                                   │
│  ┌──────────▼─────────────────────────▼──────────────────────────────┐   │
│  │                      Core Layer (业务核心)                         │   │
│  │                                                                    │   │
│  │  ┌─────────────────────┐  ┌──────────────────────────────────┐   │   │
│  │  │  RegisterCenter     │  │  Executor HTTP Client             │   │   │
│  │  │  (registry.go)      │  │  (executor.go)                    │   │   │
│  │  │                     │  │                                   │   │   │
│  │  │  • StoreNode()      │  │  • SendTrigger(appName)           │   │   │
│  │  │  • ElectNode()      │  │    → POST {addr}/run              │   │   │
│  │  │  • StartClearloop() │  │  • KillJob(addr, jobId)           │   │   │
│  │  │                     │  │    → POST {addr}/kill             │   │   │
│  │  │  内存存储:          │  │  • FetchLog(addr, ...)            │   │   │
│  │  │  sync.Map{          │  │    → POST {addr}/log              │   │   │
│  │  │    appName → {      │  │                                   │   │   │
│  │  │      addr → time    │  │  Header: XXL-JOB-ACCESS-TOKEN     │   │   │
│  │  │    }                │  │  Timeout: 5s                      │   │   │
│  │  │  }                  │  └──────────────────────────────────┘   │   │
│  │  └──────────┬──────────┘                                         │   │
│  │             │                                                     │   │
│  └─────────────┼─────────────────────────────────────────────────────┘   │
│                │                                                          │
│  ┌─────────────▼──────────────────────────────────────────────────────┐   │
│  │                    Persistence Layer                                │   │
│  │                                                                     │   │
│  │  global.DB (GORM + SQLite)                                         │   │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐           │   │
│  │  │ job_registry │ │  job_info    │ │    job_log       │           │   │
│  │  │ (执行器注册) │ │  (任务定义)  │ │   (执行日志)     │           │   │
│  │  └──────────────┘ └──────────────┘ └──────────────────┘           │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘

                                  │
                                  │  HTTP (JSON)
                                  │  Header: XXL-JOB-ACCESS-TOKEN: default_token
                                  │
           ┌──────────────────────┼──────────────────────┐
           │                      │                      │
           ▼                      ▼                      ▼
   ┌──────────────┐      ┌──────────────┐      ┌──────────────┐
   │  Executor A  │      │  Executor B  │      │  Executor C  │
   │  (Java/Go)   │      │  (Java/Go)   │      │  (Java/Go)   │
   │              │      │              │      │              │
   │  /run        │      │  /run        │      │  /run        │
   │  /kill       │      │  /kill       │      │  /kill       │
   │  /log        │      │  /log        │      │  /log        │
   └──────────────┘      └──────────────┘      └──────────────┘
```

---

## 4. 核心数据流

### 4.1 执行器注册流程

```
 Executor                    Go Admin                      SQLite
    │                           │                             │
    │  POST /api/registry       │                             │
    │  {                        │                             │
    │    registryGroup: "app1"  │                             │
    │    registryKey:   "app1"  │                             │
    │    registryValue: "http://10.0.0.1:9999/"              │
    │  }                        │                             │
    │ ─────────────────────────►│                             │
    │                           │                             │
    │                           │  1. 解析 JSON 参数          │
    │                           │  2. 规范化地址（补尾部 /）   │
    │                           │  3. Upsert 到 job_registry   │
    │                           │ ────────────────────────────►│
    │                           │      (ON CONFLICT 更新       │
    │                           │       update_time)          │
    │                           │                             │
    │                           │  4. StoreNode() 写入内存      │
    │                           │     RegsC.store[app1][addr]  │
    │                           │     = time.Now()             │
    │                           │                             │
    │  {"code": 200}            │                             │
    │ ◄─────────────────────────│                             │
```

### 4.2 任务下发流程

```
  (API/调度触发)          Go Admin Core             Executor 节点
       │                       │                         │
       │  SendTrigger("app1")  │                         │
       │ ─────────────────────►│                         │
       │                       │                         │
       │                       │  1. ElectNode("app1")    │
       │                       │     扫描存活节点          │
       │                       │     (心跳<90s)           │
       │                       │     轮询选择一个 addr    │
       │                       │                         │
       │                       │  2. POST {addr}/run      │
       │                       │  ┌─────────────────────► │
       │                       │  │ {                     │
       │                       │  │  jobId: 1,            │
       │                       │  │  executorHandler:     │
       │                       │  │    "demoJobHandler",  │
       │                       │  │  executorParams:      │
       │                       │  │    "Go callback test",│
       │                       │  │  logId: 20240420,     │
       │                       │  │  logDateTime: ...,    │
       │                       │  │  glueType: "BEAN"     │
       │                       │  │ }                     │
       │                       │  └─────────────────────► │
       │                       │                         │
       │                       │  Executor 执行任务...     │
       │                       │                         │
       │  [下发成功]            │                         │
       │ ◄─────────────────────│                         │
```

### 4.3 执行结果回调流程

```
 Executor                         Go Admin
    │                                 │
    │  POST /api/callback             │
    │  [{                             │
    │    logId: 20240420,             │
    │    logDateTime: 1690000000000,  │
    │    handleCode: 200,             │
    │    handleMsg: "执行成功"         │
    │  }]                             │
    │ ───────────────────────────────►│
    │                                 │
    │                                 │  1. 解析 []CallbackRequest
    │                                 │  2. 遍历打印每条的
    │                                 │     logId + status + msg
    │                                 │
    │  {"code": 200}                  │
    │ ◄───────────────────────────────│
```

### 4.4 强杀 & 日志拉取流程

```
  Go Admin Core                     Executor 节点
       │                                 │
       │  KillJob(addr, jobId)           │
       │  POST {addr}/kill               │
       │  {"jobId": 1}                   │
       │ ───────────────────────────────►│
       │                                 │
       │  {"code":200, "msg":"ok"}       │
       │ ◄───────────────────────────────│
       │                                 │
       │  FetchLog(addr, logDataTim,     │
       │           logId, fromLineNum)   │
       │  POST {addr}/log                │
       │  {"logDataTim":...,             │
       │   "logId":...,                  │
       │   "fromLineNum": 1}             │
       │ ───────────────────────────────►│
       │                                 │
       │  {code:200, content: {          │
       │    fromLineNum: 1,              │
       │    toLineNum: 50,               │
       │    isEnd: false,                │
       │    logContent: "..."            │
       │  }}                             │
       │ ◄───────────────────────────────│
```

---

## 5. 数据库设计

### 5.1 `job_registry` — 执行器注册表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增主键 |
| registry_group | VARCHAR(50) | 注册分组（复合索引） |
| registry_key | VARCHAR(255) | 应用名称 / AppName（复合索引） |
| registry_value | VARCHAR(255) | 执行器地址（唯一索引，如 `http://10.0.0.1:9999/`） |
| update_time | DATETIME | 最后心跳时间（自动更新） |

> Upsert 策略：`registry_value` 冲突时只更新 `update_time`。

### 5.2 `job_info` — 任务定义表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 任务 ID |
| job_group | INTEGER | 任务分组 |
| job_desc | VARCHAR(255) | 任务描述 |
| executor_handler | VARCHAR(255) | 执行器 Handler 名称 |
| job_cron | VARCHAR(128) | Cron 表达式 |
| executor_routing_strategy | VARCHAR(50) | 路由策略（默认 FIRST） |
| executor_param | TEXT | 执行参数 |
| executor_timeout | INTEGER | 超时时间（秒） |
| executor_fail_retry_count | INTEGER | 失败重试次数 |
| trigger_status | INTEGER | 触发状态 |
| trigger_last_time | BIGINT | 上次触发时间戳 |
| trigger_next_time | BIGINT | 下次触发时间戳 |
| create_time / update_time | DATETIME | GORM 自动管理 |

### 5.3 `job_log` — 执行日志表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 日志 ID |
| job_id | INTEGER | 关联任务 ID |
| executor_address | VARCHAR(255) | 执行节点地址 |
| executor_handler | VARCHAR(255) | Handler 名称 |
| executor_param | TEXT | 执行参数 |
| trigger_time | DATETIME | 触发时间 |
| trigger_code / trigger_msg | VARCHAR / TEXT | 触发结果 |
| handler_time | DATETIME | 执行完成时间 |
| handler_code / handler_msg | VARCHAR / TEXT | 执行结果 |

---

## 6. 技术栈

| 组件 | 技术选型 | 用途 |
|------|----------|------|
| Web 框架 | `gin-gonic/gin v1.12` | HTTP 路由、请求绑定、JSON 响应 |
| ORM | `gorm.io/gorm v1.31` | 数据库 ORM、自动迁移 |
| 数据库 | SQLite（`gorm.io/driver/sqlite`） | 嵌入式持久化存储 |
| JSON | Go 标准库 `encoding/json` | 序列化/反序列化 |
| 并发 | `sync.Map` + goroutine | 线程安全的内存注册中心 + 后台清理 |

---

## 7. 启动方式

```bash
# 确保 Go 1.26+ 已安装
cd go-xxl-admin

# 下载依赖
go mod tidy

# 运行（SQLite 文件 xxl_job.db 会自动创建）
go run main.go
```

启动后：
- Admin 监听 **`:8081`** 端口
- SQLite 数据库 **`xxl_job.db`** 在项目根目录自动生成
- 三张表（`job_registry`、`job_info`、`job_log`）自动建表

---

## 8. API 接口

### `POST /api/registry` — 执行器注册

Executor 启动后向 Admin 注册自己的地址。

```json
// Request
{
  "registryGroup": "xxl-job-executor-sample",
  "registryKey": "xxl-job-executor-sample",
  "registryValue": "http://192.168.1.100:9999"
}

// Response
{ "code": 200 }
```

### `POST /api/callback` — 执行结果回调

Executor 完成任务后回报结果（支持批量）。

```json
// Request (数组)
[{
  "logId": 20240420,
  "logDateTime": 1690000000000,
  "handleCode": 200,
  "handleMsg": "执行成功"
}]

// Response
{ "code": 200 }
```

### `POST /test/kill` — 测试强杀（手动调试用）

对硬编码的 `xxl-job-executor-sample` 应用：选一个节点 → 下发任务 → 等 3s → 强杀。

```json
// Response
{
  "code": 200,
  "message": "success",
  "data": { "targetAppName": "xxl-job-executor-sample" }
}
```

---

## 9. 关键设计决策与状态

1. **双写存储（内存 + SQLite）**：注册信息同时写入内存 `sync.Map` 和 SQLite。路由选举查内存（快），数据库作持久备份。注意：清理协程只清内存不清数据库，两者可能不一致。

2. **硬编码配置**：当前所有配置（端口、超时时间、Token、目标应用名）均硬编码在各 `.go` 文件中，无外部配置文件。

3. **XXL-JOB 协议兼容**：Admin 与 Executor 之间使用标准 XXL-JOB HTTP 协议（`/run`、`/kill`、`/log` 端点 + `XXL-JOB-ACCESS-TOKEN` 认证头），可与 Java 版 Executor 互通。

4. **调度器缺失**：`job_info` 表已设计好 cron 字段，但尚未实现 cron 解析和定时触发逻辑。当前任务下发仅支持程序内调用 `SendTrigger()` 或通过 `/test/kill` 间接触发。

5. **轮询负载均衡**：`ElectNode()` 使用全局递增计数器对存活节点数取模的方式实现轮询，简单有效。
