# 个人生产增强版开发计划

> 目标：一个人把 `go-xxl-admin` 从当前原型推进到个人生产增强版。这里不写泛泛而谈的内容，只保留自己真正要做的模块、优先级和工期。

## 总目标

把项目做成：

```text
Go 版轻量 XXL-JOB Admin
支持任务管理、分组管理、执行器治理、日志追踪、鉴权、监控告警和分布式调度增强
```

整体预估：

```text
30 ~ 50 人日
```

如果业余时间开发，按每天 2~3 小时算，大约：

```text
2 ~ 3 个月+
```

## 总排期

| 阶段 | 目标 | 工期 |
|------|------|------|
| 第 1 阶段 | 补齐控制台真实可用能力 | 7 ~ 10 人日 |
| 第 2 阶段 | 做到个人内部可用 | 8 ~ 12 人日 |
| 第 3 阶段 | 做生产稳定性增强 | 8 ~ 15 人日 |
| 第 4 阶段 | 做监控、告警、审计和多数据库 | 10 ~ 18 人日 |
| 第 5 阶段 | 测试、文档、验收 | 5 ~ 8 人日 |

总计：

```text
38 ~ 63 人日
```

实际执行时先按 **30 ~ 50 人日压缩版** 做，后面再补细节。

## 第 1 阶段：补齐控制台真实可用能力

目标：前端不再只能展示“暂无接口”，核心资源都能真实管理。

工期：

```text
7 ~ 10 人日
```

要做：

| 模块 | 内容 | 工期 |
|------|------|------|
| job_group 管理 | 增删改查、搜索、删除前检查关联任务、前端分组页面 | 2 ~ 3 人日 |
| job_log 查询 | 日志分页、按任务查、按状态筛选、日志详情、前端日志中心 | 2.5 ~ 4 人日 |
| executor 查询 | 展示 appName、节点地址、心跳时间、在线/离线 | 2 ~ 3 人日 |
| 当前 bug 修复 | Redis 空节点 panic、registryGroup 写入、CreateJob 停用状态、UpdateJob 零值更新 | 1 ~ 2 人日 |

接口计划：

```text
POST   /api/job-group
GET    /api/job-group
GET    /api/job-group/:id
PUT    /api/job-group/:id
DELETE /api/job-group/:id

GET    /api/job-log
GET    /api/job-log/:id
GET    /api/job/:id/logs

GET    /api/executor
GET    /api/executor/:appName
```

## 第 2 阶段：个人内部可用版

目标：自己日常可以真正使用，而不是只能演示。

工期：

```text
8 ~ 12 人日
```

要做：

| 模块 | 内容 | 工期 |
|------|------|------|
| 登录鉴权 | 登录、退出、当前用户、API 鉴权中间件、前端登录页 | 3 ~ 4 人日 |
| 手动触发 | `POST /api/job/:id/trigger`，立即触发一次任务 | 1.5 ~ 2 人日 |
| 执行日志拉取 | 暴露已有 `FetchLog`，从 Executor `/log` 拉真实日志 | 2 ~ 3 人日 |
| 失败重试 | 基于历史日志重新下发，生成新日志，不覆盖旧日志 | 2 ~ 3 人日 |

接口计划：

```text
POST /api/login
POST /api/logout
GET  /api/me

POST /api/job/:id/trigger
GET  /api/job-log/:id/content
POST /api/job-log/:id/retry
```

## 第 3 阶段：生产稳定性增强

目标：系统能长期运行，避免重复调度、异常崩溃和节点状态混乱。

工期：

```text
8 ~ 15 人日
```

要做：

| 模块 | 内容 | 工期 |
|------|------|------|
| 分布式调度锁 | 多 Admin 实例下同一任务不重复触发，优先 Redis 锁 | 3 ~ 5 人日 |
| 调度器健壮性 | cron 错误处理、next_time 初始化、misfire、panic recover、批量扫描 | 3 ~ 5 人日 |
| 注册中心增强 | 离线节点、内存/Redis 抽象统一、路由策略扩展 | 2 ~ 5 人日 |

路由策略后续支持：

```text
FIRST
ROUND
RANDOM
CONSISTENT_HASH
LEAST_RECENTLY_USED
```

## 第 4 阶段：生产能力增强

目标：具备基础运维能力。

工期：

```text
10 ~ 18 人日
```

要做：

| 模块 | 内容 | 工期 |
|------|------|------|
| Prometheus 指标 | `/metrics`，任务数、执行器数、调度次数、失败次数、耗时 | 2 ~ 3 人日 |
| 告警 | Webhook 告警，任务失败、连续失败、执行器离线、Redis/MQ 异常 | 3 ~ 5 人日 |
| 操作审计 | 记录创建、修改、删除、触发、停用、注册等操作 | 2 ~ 3 人日 |
| 多数据库 | SQLite / MySQL / PostgreSQL driver 切换 | 3 ~ 5 人日 |

指标草案：

```text
go_xxl_admin_jobs_total
go_xxl_admin_jobs_running
go_xxl_admin_executors_online
go_xxl_admin_trigger_total
go_xxl_admin_trigger_success_total
go_xxl_admin_trigger_failed_total
go_xxl_admin_callback_total
go_xxl_admin_callback_failed_total
go_xxl_admin_scheduler_scan_duration_seconds
go_xxl_admin_executor_request_duration_seconds
```

## 第 5 阶段：测试、文档、验收

目标：项目能长期维护，也能作为完整个人项目展示。

工期：

```text
5 ~ 8 人日
```

要做：

| 模块 | 内容 | 工期 |
|------|------|------|
| 测试 | registry、scheduler、executor client、handler、config、job CRUD、log query | 3 ~ 5 人日 |
| 文档 | 部署、配置、API、Executor 接入、Redis/MQ、生产建议、故障排查 | 1.5 ~ 2.5 人日 |
| 验收 | 单机、Redis、MQ、前端、回调、手动触发、鉴权、指标、告警 | 0.5 ~ 1 人日 |

## 优先级

### P0：先做

```text
job_group 管理
job_log 查询
executor 查询
现有 bug 修复
统一错误响应
```

### P1：个人可用

```text
登录鉴权
手动触发
失败重试
日志拉取
```

### P2：稳定运行

```text
分布式调度锁
调度器健壮性
注册中心增强
监控指标
告警
```

### P3：完善交付

```text
审计日志
多数据库
测试体系
文档
```

## 一个人执行排期

```text
Day 1-2:   job_group API
Day 3-5:   job_log API
Day 6-7:   executor 查询 API
Day 8:     bugfix
Day 9-10:  前端接入和联调

Day 11-13: 登录鉴权
Day 14-15: 手动触发
Day 16-18: 日志拉取
Day 19-20: 失败重试

Day 21-25: 分布式调度锁
Day 26-30: 调度器健壮性
Day 31-35: 注册中心增强

Day 36-38: Prometheus 指标
Day 39-43: 告警
Day 44-46: 审计日志
Day 47-50: 多数据库支持

Day 51-55: 测试
Day 56-58: 文档和验收
```

## 当前最先开始

先做 P0，顺序如下：

```text
1. 修当前明显 bug
2. job_group API + 前端
3. job_log API + 前端
4. executor 查询 API + 前端
```

这四个做完，项目就从“原型”进入“个人可用控制台”的阶段。
