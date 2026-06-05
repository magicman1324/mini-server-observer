.PHONY: help agent test-agent test clean dev-up dev-down proto

GO_FLAGS   := -ldflags="-s -w"
CGO_ENABLED := 0

help: ## 显示帮助
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
	awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ============ 构建 ============

agent: ## 编译 Agent
	cd agent && CGO_ENABLED=$(CGO_ENABLED) go build $(GO_FLAGS) -o bin/hydra-agent ./cmd/agent/

regional: ## 编译 Regional Collector
	cd backend && CGO_ENABLED=$(CGO_ENABLED) go build $(GO_FLAGS) -o bin/hydra-regional ./cmd/regional/

central: ## 编译 Central Aggregator
	cd backend && CGO_ENABLED=$(CGO_ENABLED) go build $(GO_FLAGS) -o bin/hydra-central ./cmd/central/

proto: ## 生成 protobuf
	cd backend && protoc --go_out=. --go-grpc_out=. internal/proto/hydra.proto

frontend: ## 构建前端
	cd frontend && npm ci && npm run build

# ============ 测试 ============

test-agent: ## 测试 Agent
	cd agent && go test -v -race -count=1 ./...

test-backend: ## 测试 Backend
	cd backend && go test -v -race -count=1 ./...

test-frontend: ## 测试前端
	cd frontend && npm run test -- --run

test: test-agent test-backend ## 运行全部测试

# ============ Docker ============

docker-agent: ## 构建 Agent 镜像
	docker build -t hydra-agent:latest -f agent/Dockerfile agent/

docker-central: ## 构建 Central 镜像
	docker build -t hydra-central:latest -f backend/Dockerfile backend/

docker-frontend: ## 构建前端镜像
	docker build -t hydra-dashboard:latest frontend/

# ============ 开发环境 ============

dev-up: ## 启动开发环境
	docker compose -f deploy/docker-compose.yml up -d

dev-down: ## 停止开发环境
	docker compose -f deploy/docker-compose.yml down

dev-logs: ## 查看日志
	docker compose -f deploy/docker-compose.yml logs -f

# ============ 清理 ============

clean: ## 清理构建产物
	rm -rf agent/bin backend/bin frontend/dist frontend/node_modules
