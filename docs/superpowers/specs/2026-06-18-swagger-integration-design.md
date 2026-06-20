# Swagger 集成设计

## 概述

为 dd-frame 项目集成 swaggo/swag + gin-swagger，为所有现有 HTTP handler 添加 Swagger 注解，提供在线 API 文档 UI。

## 依赖

| 包 | 用途 |
|----|------|
| `github.com/swaggo/swag/cmd/swag` | CLI 工具，解析注解生成 swagger 文档 |
| `github.com/swaggo/gin-swagger` | Gin 中间件，serve Swagger UI |
| `github.com/swaggo/files` | Swagger UI 静态资源 |

## 全局注解（main.go）

```go
// @title dd-frame API
// @version 1.0
// @description DDD 模块化单体项目 API 文档
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer JWT Token
```

## Handler 注解模式

每个 handler 添加标准 swag 注解：

```go
// LoginHandler 登录
// @Summary 用户登录
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body service.LoginInput true "登录参数"
// @Success 200 {object} response.Response{data=service.LoginOutput}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/login [post]
```

## DTO 注解

Input/Output 结构体添加 `@Description` 和字段 `example` tag。

## 路由

- `/swagger/*any` — 公开路由，注册在 `public` 路由组，无需认证
- 仅在 `server.mode != "release"` 时启用（生产环境可关闭）

## 文件变更

| 文件 | 变更 |
|------|------|
| `main.go` | 全局 swagger 注解 + import docs |
| `internal/auth/api/http_handler.go` | 所有 handler 添加注解 |
| `internal/auth/service/app_service.go` | DTO 结构体添加 example tag |
| `example/order/api/http_handler.go` | 所有 handler 添加注解 |
| `app/wire.go` | 注册 swagger UI 路由 |
| `Makefile` | 添加 `swagger` + `swagger-deps` 命令 |
| `docs/swagger.json` | 生成产物（提交 git） |
| `docs/swagger.yaml` | 生成产物（提交 git） |

## Makefile 命令

```makefile
swagger-deps:  ## 安装 swag CLI
	go install github.com/swaggo/swag/cmd/swag@latest

swagger:       ## 生成 Swagger 文档
	swag init -g main.go -o docs --parseDependency
```
