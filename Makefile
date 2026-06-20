# =============================================================================
# dd-frame Makefile
# =============================================================================

.PHONY: help build run test vet clean proto proto-gen proto-deps lint port-check db-init db-seed swagger swagger-deps

# 默认目标
.DEFAULT_GOAL := help

# ---------------------------------------------------------------------------
# 变量
# ---------------------------------------------------------------------------
APP_NAME     := dd-frame
BUILD_DIR    := ./bin
PROTO_DIR    := ./proto
PROTO_OUT    := ./proto/gen
CONFIG_FILE  := config/config.yaml

# 从配置文件读取端口，默认 8080
HTTP_PORT    := $(shell grep 'http_port' $(CONFIG_FILE) 2>/dev/null | head -1 | awk '{print $$2}' | tr -d '[:space:]')
HTTP_PORT    := $(or $(HTTP_PORT),8080)

# Go 相关
GOCMD        := go
GOBUILD      := $(GOCMD) build
GOTEST       := $(GOCMD) test
GOVET        := $(GOCMD) vet
GOMOD        := $(GOCMD) mod
GOTIDY       := $(GOMOD) tidy

# Proto 相关
PROTOC       := protoc
PROTOC_FLAGS := --proto_path=$(PROTO_DIR)
GO_OUT       := --go_out=$(PROTO_OUT) --go_opt=paths=source_relative
CONNECT_OUT  := --connect-go_out=$(PROTO_OUT) --connect-go_opt=paths=source_relative

# ---------------------------------------------------------------------------
# 开发命令
# ---------------------------------------------------------------------------

help: ## 显示帮助信息
	@echo "dd-frame — DDD 模块化单体项目框架"
	@echo ""
	@echo "用法: make <target>"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

build: ## 编译项目
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) ./main.go

run: port-check ## 运行项目
	$(GOCMD) run main.go

port-check: ## 检查端口是否被占用
	@PID=$$(lsof -ti:$(HTTP_PORT) 2>/dev/null); \
	if [ -n "$$PID" ]; then \
		echo "⚠  端口 $(HTTP_PORT) 被 PID=$$PID 占用"; \
		echo "   执行 'make kill-port' 终止，或修改 $(CONFIG_FILE) 中的 http_port"; \
		exit 1; \
	fi

kill-port: ## 终止占用端口的进程
	@PID=$$(lsof -ti:$(HTTP_PORT) 2>/dev/null); \
	if [ -n "$$PID" ]; then \
		kill -9 $$PID; \
		echo "已终止 PID=$$PID（端口 $(HTTP_PORT)）"; \
	else \
		echo "端口 $(HTTP_PORT) 未被占用"; \
	fi

test: ## 运行测试
	$(GOTEST) -race -cover ./...

vet: ## 静态分析
	$(GOVET) ./...

lint: ## 代码检查（需要 golangci-lint）
	golangci-lint run ./...

tidy: ## 整理依赖
	$(GOTIDY)

clean: ## 清理构建产物
	rm -rf $(BUILD_DIR)
	rm -rf $(PROTO_OUT)

# ---------------------------------------------------------------------------
# Proto 命令
# ---------------------------------------------------------------------------

proto-deps: ## 安装 proto 代码生成工具
	@echo "安装 protoc-gen-go..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@echo "安装 protoc-gen-connect-go..."
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	@echo "安装 buf..."
	go install github.com/bufbuild/buf/cmd/buf@latest
	@echo "Proto 工具安装完成"

proto-gen: ## 使用 protoc 生成代码（不依赖 buf）
	@echo "生成 proto 代码..."
	@mkdir -p $(PROTO_OUT)
	@find $(PROTO_DIR) -name "*.proto" | while read proto; do \
		echo "  编译: $${proto}"; \
		$(PROTOC) $(PROTOC_FLAGS) \
			$(GO_OUT) \
			$(CONNECT_OUT) \
			$${proto}; \
	done
	@echo "代码生成完成 → $(PROTO_OUT)"

proto-buf: ## 使用 buf 生成代码（需先安装 buf）
	@echo "使用 buf 生成代码..."
	cd $(PROTO_DIR) && buf generate
	@echo "代码生成完成"

proto-lint: ## 使用 buf 检查 proto 规范
	cd $(PROTO_DIR) && buf lint

proto-breaking: ## 使用 buf 检查 proto 兼容性变更
	cd $(PROTO_DIR) && buf breaking --against .git

proto: proto-gen ## proto-gen 的别名

# ---------------------------------------------------------------------------
# Swagger 命令
# ---------------------------------------------------------------------------

swagger-deps: ## 安装 Swagger CLI 工具
	go install github.com/swaggo/swag/cmd/swag@latest

swagger: ## 生成 Swagger API 文档
	swag init -g main.go -o docs --parseDependency --exclude internal/_template

# ---------------------------------------------------------------------------
# 数据库命令
# ---------------------------------------------------------------------------

db-init: ## 数据库迁移（创建/更新表结构）
	$(GOCMD) run cmd/dbinit/main.go --config $(CONFIG_FILE)

db-seed: ## 数据库初始化种子数据（admin 角色/用户/权限）
	$(GOCMD) run cmd/dbinit/main.go --seed --config $(CONFIG_FILE)

# ---------------------------------------------------------------------------
# 组合命令
# ---------------------------------------------------------------------------

check: vet test ## 编译检查 + 测试

all: tidy proto-gen build test ## 完整构建流程（整理依赖 + proto生成 + 编译 + 测试）
