# dd-frame 框架功能差距分析

> 基于生产级 Go 通用框架标准，对当前项目已实现功能与缺失功能的全面梳理。

## 当前已实现功能

| 类别 | 功能 | 状态 |
|------|------|------|
| HTTP 框架 | Gin + 优雅关闭 | ✅ |
| 架构 | DDD 六边形分层 + 模块化单体 + `internal/` 编译器隔离 | ✅ |
| 配置 | Viper + YAML + 环境变量覆盖 | ✅ |
| 数据库 | GORM + MySQL + 条件初始化 | ✅ |
| 缓存 | Redis 连接 + 条件初始化 | ✅ |
| 日志 | Zap 结构化日志 + 上下文注入 + 敏感数据脱敏 | ✅ |
| 认证 | JWT + AuthUser 实体注入 | ✅ |
| 权限 | RBAC（用户→角色→权限）+ RequirePermission 中间件 | ✅ |
| 中间件 | Recovery / CORS / RequestID / Logger / Auth / RBAC / RateLimit | ✅ |
| API 文档 | Swagger（swaggo/swag + gin-swagger UI） | ✅ |
| 统一响应 | pkg/response + pkg/errors | ✅ |
| 分页 | pkg/pagination | ✅ |
| CLI | dbinit（迁移 + 种子数据） | ✅ |
| 构建 | Makefile（build/run/test/vet/proto/swagger/db/port-check） | ✅ |
| 模块模板 | internal/_template 完整骨架 | ✅ |

---

## 缺失功能清单

### P0 — 生产必需

#### 1. 健康检查端点
- **现状**：无任何 `/health` 或 `/ready` 端点
- **需要**：
  - `GET /health` — 存活检查（liveness）
  - `GET /ready` — 就绪检查（readiness），检测 DB/Redis 连通性
- **用途**：K8s livenessProbe / readinessProbe，负载均衡器健康检测

#### 2. 分布式追踪（Tracing）
- **现状**：日志无 trace ID 关联，RequestID 中间件仅生成不传播
- **需要**：
  - 集成 OpenTelemetry SDK
  - 中间件注入 trace span
  - RequestID 与 trace ID 统一
  - 日志自动携带 traceID/spanID
- **用途**：微服务链路追踪，故障定位

#### 3. Prometheus 指标暴露
- **现状**：无可观测性指标
- **需要**：
  - `GET /metrics` 端点（Prometheus 格式）
  - HTTP 请求计数器/延迟直方图中间件
  - 业务指标注册辅助函数
- **用途**：Grafana 监控，告警

#### 4. 数据库迁移工具
- **现状**：仅 GORM AutoMigrate（不支持回滚，不支持 SQL 迁移）
- **需要**：
  - 集成 golang-migrate 或 goose
  - 版本化 SQL 迁移文件（`migrations/001_create_users.up.sql`）
  - Makefile 命令：`make migrate-up`, `make migrate-down`, `make migrate-status`
- **用途**：生产环境数据库变更管理，CI/CD 集成

#### 5. Docker 支持
- **现状**：无 Dockerfile，无 docker-compose
- **需要**：
  - 多阶段构建 Dockerfile（builder → scratch/alpine）
  - docker-compose.yaml（app + MySQL + Redis）
  - `.dockerignore`
- **用途**：容器化部署，本地开发环境一键启动

### P1 — 强烈推荐

#### 6. 请求限流（Rate Limiting） ✅
- **实现**：`middleware/ratelimit.go`，支持 memory（令牌桶）+ redis（滑动窗口）两种后端
- **需要**（已完成）：
  - ✅ 基于 IP 的全局限流中间件
  - 基于用户/角色的分级限流（待实现）
  - ✅ 支持内存 + Redis 两种后端
- **用途**：防止 API 滥用，保护服务稳定性

#### 7. 输入验证增强
- **现状**：仅依赖 Gin binding tag（`binding:"required"`）
- **需要**：
  - 自定义验证器注册（如手机号、邮箱格式）
  - 统一验证错误响应（字段级错误详情）
  - 请求参数自动 trim/sanitize
- **用途**：减少 boilerplate，统一错误格式

#### 8. 缓存工具层
- **现状**：Redis 仅初始化连接，无缓存操作封装
- **需要**：
  - 通用缓存 helper：`Get/Set/Delete` with TTL
  - Cache-aside 模式封装
  - 缓存 key 命名规范工具
  - 防缓存穿透/击穿（singleflight）
- **用途**：减少重复缓存代码，防止常见缓存问题

#### 9. 配置验证
- **现状**：配置加载后无校验，缺失字段可能导致运行时 panic
- **需要**：
  - 配置结构体添加 `validate` tag
  - `LoadConfig` 后自动校验必填字段
  - 启动时打印脱敏配置摘要
- **用途**：快速发现配置错误，避免运行时故障

#### 10. 统一日志增强 ✅
- **实现**：`pkg/log/context.go` + `pkg/log/sanitize.go` + `middleware/logger.go`
- **需要**（已完成）：
  - ✅ Logger 中间件注入 context（含 requestID、userID、traceID、clientIP）
  - ✅ 业务日志自动携带上下文字段（`log.FromContext`）
  - ✅ 敏感数据脱敏（`log.Sanitize`）
- **用途**：请求级日志关联，安全合规

#### 11. CI/CD 流水线
- **现状**：无 CI 配置
- **需要**：
  - GitHub Actions（lint → test → build → docker push）
  - 代码覆盖率报告
  - Release 自动打 tag
- **用途**：自动化质量保障，持续交付

### P2 — 推荐实现

#### 12. gRPC 集成
- **现状**：Connect RPC 仅预留目录和 buf 配置，无实际集成
- **需要**：
  - gRPC server 与 Gin HTTP server 双协议启动
  - 共享中间件（认证/日志/追踪）
  - Connect RPC handler 与 Gin handler 共存示例
- **用途**：高性能内部通信，对外 REST + 对内 gRPC

#### 13. 事件总线 / 异步任务
- **现状**：无异步处理机制
- **需要**：
  - 进程内事件总线（发布/订阅模式）
  - 可选集成消息队列（Redis Stream / RabbitMQ）
  - 后台任务队列 + Worker
- **用途**：解耦模块间通信，异步处理耗时操作

#### 14. 测试基础设施
- **现状**：无测试文件，无测试辅助工具
- **需要**：
  - 测试辅助包：HTTP 请求构造器、断言增强
  - 数据库测试：test fixtures、事务回滚
  - Mock 生成工具（mockery 或 moq）
  - 集成测试框架
  - `make test` 区分单元测试与集成测试
- **用途**：保障代码质量，支持 TDD

#### 15. 请求超时控制
- **现状**：无超时中间件
- **需要**：
  - 全局请求超时中间件（可配置）
  - 按路由组自定义超时
  - Context 级超时传播
- **用途**：防止慢请求拖垮服务

#### 16. 优雅重启（Hot Restart）
- **现状**：仅支持优雅关闭，不支持零停机重启
- **需要**：
  - 集成 tableflip 或 overseer
  - `SIGUSR2` 信号触发平滑重启
- **用途**：生产环境零停机部署

### P3 — 锦上添花

#### 17. 文件上传/存储 ✅
- **实现**：`pkg/storage/` + `middleware/upload.go`，支持本地磁盘 + 阿里云 OSS
- **需要**（已完成）：
  - ✅ 文件上传中间件（大小限制、类型校验）
  - ✅ 存储抽象层（本地 / OSS）
- **用途**：通用文件处理需求

#### 18. 国际化（i18n）
- **需要**：
  - 多语言错误消息
  - Accept-Language 头解析
  - 响应消息模板
- **用途**：面向国际用户的 API

#### 19. API 版本管理
- **现状**：`/api/v1` 硬编码
- **需要**：
  - 版本路由策略（URL path / Header）
  - 版本弃用通知机制
- **用途**：API 演进不破坏兼容

#### 20. 代码生成器
- **需要**：
  - `make gen-module name=xxx` — 基于 _template 生成新模块
  - `make gen-crud entity=xxx` — 生成 CRUD 骨架代码
- **用途**：加速新模块开发，减少复制粘贴

#### 21. 通知系统
- **需要**：
  - 邮件发送封装
  - 短信发送封装
  - 模板化通知
- **用途**：用户注册验证、密码重置等

#### 22. 定时任务（Cron）
- **需要**：
  - 集成 robfig/cron
  - 任务注册 + 分布式锁防重复执行
- **用途**：定期数据清理、报表生成

---

## 优先级总结

| 优先级 | 功能 | 核心价值 | 状态 |
|--------|------|----------|------|
| **P0** | 健康检查 | K8s 部署基础 | ✅ |
| **P0** | 分布式追踪 | 可观测性基础 | ✅ |
| **P0** | Prometheus 指标 | 监控告警基础 | ✅ |
| **P0** | 数据库迁移工具 | 生产变更管理 | ❌ |
| **P0** | Docker 支持 | 容器化部署 | ✅ |
| **P1** | 请求限流 | 服务保护 | ✅ |
| **P1** | 输入验证增强 | 安全 + 统一格式 | ❌ |
| **P1** | 缓存工具层 | 减少重复代码 | ❌ |
| **P1** | 配置验证 | 启动时快速发现错误 | ❌ |
| **P1** | 统一日志增强 | 请求级关联 + 脱敏 | ✅ |
| **P1** | CI/CD 流水线 | 自动化质量保障 | ❌ |
| **P2** | gRPC 集成 | 高性能内部通信 | ❌ |
| **P2** | 事件总线 | 模块解耦 | ❌ |
| **P2** | 测试基础设施 | 质量保障 | ❌ |
| **P2** | 请求超时 | 服务稳定性 | ❌ |
| **P2** | 优雅重启 | 零停机部署 | ❌ |
| **P3** | 文件上传 | 通用需求 | ✅ |
| **P3** | 国际化 | 多语言支持 | ❌ |
| **P3** | API 版本管理 | 向后兼容 | ❌ |
| **P3** | 代码生成器 | 开发效率 | ❌ |
| **P3** | 通知系统 | 业务通知 | ❌ |
| **P3** | 定时任务 | 周期性操作 | ❌ |
