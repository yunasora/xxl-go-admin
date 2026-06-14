# go-xxl-admin

> Go 版轻量 XXL-JOB Admin 调度中心原型。目标是保留 XXL-JOB Admin 的核心调度链路，用更轻、更容易二次开发的 Go 代码实现一个可继续生产增强的调度平台基础。

## 项目定位

这个项目不是为了 1:1 复刻官方 XXL-JOB Admin，而是做一个 **Go 技术栈下的轻量调度中心内核**。

当前已经打通核心链路：

```text
Executor 注册/心跳
  -> Admin 维护可用节点
  -> 创建任务
  -> Scheduler 扫描到期任务
  -> 选择 Executor
  -> HTTP 或 MQ 下发 /run
  -> Executor 回调 /api/callback
  -> Admin 更新 job_log
```

适合：

- 学习 XXL-JOB Admin 和 Executor 的协议交互
- 作为 Go 技术栈下的调度中心二次开发基础
- 做个人项目、课程项目、PoC 或内部轻量调度平台雏形

不建议当前版本直接用于生产。

## 相比传统 XXL-JOB 的优势

| 优势 | 说明 |
|------|------|
| Go 技术栈 | 对 Go 项目和 Go 团队更友好，不依赖 Java/Spring 体系 |
| 更轻量 | Gin + GORM + SQLite 起步，部署和理解成本低 |
| 主链路清晰 | 注册、调度、下发、回调、日志代码比较集中，适合学习和改造 |
| 二次开发空间大 | 不被官方完整控制台和复杂架构绑定，方便改成自己的内部平台 |
| 前后端更自由 | 前端可完全自定义，目前已使用 React + Vite 单独实现控制台 |
| 可选 MQ | 已预留 RabbitMQ 异步投递方向，便于把调度决策和任务下发解耦 |
| 可选 Redis | 已预留 Redis 注册中心和轮询计数，方便后续做多实例增强 |
| 本地原型友好 | SQLite 模式下可以快速启动和验证核心逻辑 |

一句话：官方 XXL-JOB 更完整；本项目更轻、更容易读、更适合 Go 方向自研扩展。

## 当前不足

| 不足 | 说明 |
|------|------|
| 功能不完整 | `job_group`、真实 `job_log` 查询、执行器列表等接口还没补齐 |
| 鉴权不足 | 有 `access_token` 配置，但 Admin API 还没有统一鉴权中间件 |
| 生产稳定性不足 | 还缺分布式调度锁、misfire 策略、panic recover、监控告警等能力 |
| 数据库较轻 | 当前以 SQLite 为主，更适合原型，生产建议支持 MySQL/PostgreSQL |
| 测试不足 | 单测、集成测试、API 契约测试还需要补 |
| 兼容性待完善 | 目前只覆盖 XXL-JOB 核心协议链路，不是官方完整替代品 |

## 已实现能力

| 功能 | 状态 | 说明 |
|------|------|------|
| 执行器注册 | ✅ | `POST /api/registry` |
| 心跳续约 | ✅ | 刷新执行器 `update_time` |
| 节点剔除 | ✅ | 后台循环剔除超时节点 |
| 节点选举 | ✅ | 基于轮询选择 Executor |
| 任务 CRUD | ✅ | 创建、查询、更新、删除任务 |
| 启停任务 | ✅ | 启用/停用调度任务 |
| 调度扫描 | ✅ | 定时扫描到期任务 |
| HTTP 下发 | ✅ | 调 Executor `/run` |
| 执行回调 | ✅ | `POST /api/callback` 更新日志 |
| 强杀任务 | ✅ | 调 Executor `/kill`，当前主要是测试入口 |
| 日志拉取客户端 | ✅ | 已有调用 Executor `/log` 的客户端能力 |
| RabbitMQ | ✅ | 可选异步投递任务 |
| Redis 注册中心 | ✅ | 可选 Redis 维护节点和轮询计数 |
| React 前端 | ✅ | 已有管理台基础页面 |
| 登录鉴权 | ❌ | 待做 |
| 监控告警 | ❌ | 待做 |
| 分布式调度锁 | ❌ | 待做 |

## 技术栈

- Go 1.26.2
- Gin
- GORM
- SQLite
- Redis（可选）
- RabbitMQ（可选）
- robfig/cron
- React + Vite

## 目录结构

```text
go-xxl-admin/
├── main.go              # 应用入口、初始化、路由注册
├── config.json          # 运行配置
├── xxl_job.db           # SQLite 数据库
├── config/              # 配置加载
├── global/              # DB 初始化
├── core/                # 调度器、注册中心、Executor HTTP 客户端
├── handlers/            # Gin API handler
├── models/              # GORM 模型和 XXL 协议 DTO
├── mq/                  # RabbitMQ 连接、发布、消费
├── redis/               # Redis 连接
├── frontend/            # React + Vite 前端源码
├── web/                 # 前端构建产物，由 Gin 托管
└── DEVREADME.md         # 个人生产增强版开发排期
```

## 快速启动

建议本地先关闭 Redis 和 RabbitMQ：

```json
{
  "mq_enabled": false,
  "redis_enabled": false
}
```

启动 Admin：

```bash
go run .
```

默认访问：

```text
http://127.0.0.1:8081/
```

## 前端开发

安装依赖：

```bash
npm install
```

开发模式：

```bash
npm run dev
```

构建前端到 `web/`：

```bash
npm run build
```

Go 服务会通过 `/` 托管 `web/index.html`，通过 `/web/*` 托管静态资源。

## 主要 API

### Executor 协议相关

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/registry` | Executor 注册/心跳 |
| `POST` | `/api/callback` | Executor 执行结果回调 |

### 任务管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/job` | 创建任务 |
| `GET` | `/api/job` | 查询任务列表 |
| `GET` | `/api/job/:id` | 查询单个任务 |
| `PUT` | `/api/job/:id` | 更新任务 |
| `DELETE` | `/api/job/:id` | 删除任务 |
| `PUT` | `/api/job/:id/start` | 启用任务 |
| `PUT` | `/api/job/:id/stop` | 停用任务 |

### UI 配置

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/ui-config` | 前端读取部分运行配置 |

## 创建任务前的注意事项

当前还没有 `job_group` 管理接口，所以创建任务前需要先准备一条 `job_group` 数据，并保证：

```text
job_group.app_name = Executor 注册时的 registryKey
job_info.job_group = job_group.id
```

这是后续生产增强版里优先要补的能力。

## 后续计划

详细个人开发排期见 [DEVREADME.md](DEVREADME.md)。

优先级概括：

```text
P0: job_group / job_log / executor 查询 / bugfix
P1: 登录鉴权 / 手动触发 / 失败重试 / 日志拉取
P2: 分布式调度锁 / 调度器健壮性 / 监控告警
P3: 审计 / 多数据库 / 测试 / 文档
```
