.PHONY: all build run dev clean

all: build

# 构建前端
build-web:
	cd web && npm install && npm run build

# 构建后端
build-backend:
	go build -o bin/gost-panel ./cmd/panel
	go build -o bin/gost-agent ./cmd/agent

# 完整构建
build: build-web build-backend

# 开发模式运行后端
run:
	go run ./cmd/panel

# 开发模式 (前后端分离)
dev:
	@echo "Starting backend on :8080..."
	@go run ./cmd/panel &
	@echo "Starting frontend on :5173..."
	@cd web && npm run dev

# 清理
clean:
	rm -rf bin/
	rm -rf web/node_modules
	rm -rf internal/api/dist

# 安装依赖
deps:
	go mod tidy
	cd web && npm install

# Docker 构建
docker-build:
	docker build -t gost-panel .

# Docker 运行
docker-run:
	docker-compose up -d
