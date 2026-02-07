# 构建前端
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci --prefer-offline --no-audit
COPY web/ ./
RUN npm run build

# 构建后端
FROM golang:1.23-alpine AS backend
WORKDIR /app

# 安装编译依赖 (SQLite 需要 CGO)
RUN apk add --no-cache gcc musl-dev sqlite-dev

# 复制 Go 模块文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 复制前端构建产物 (vite outDir: ../internal/api/dist)
COPY --from=frontend /app/internal/api/dist /app/internal/api/dist

# 构建后端 (带版本信息)
ARG VERSION=dev
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -X 'github.com/AliceNetworks/gost-panel/internal/api.CurrentAgentVersion=${VERSION}' -X 'github.com/AliceNetworks/gost-panel/internal/api.AgentBuildTime=${BUILD_TIME}'" \
    -o gost-panel cmd/panel/main.go

# 运行镜像
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata sqlite

WORKDIR /app

# 复制二进制文件和脚本
COPY --from=backend /app/gost-panel .
COPY --from=backend /app/scripts ./scripts/

# 创建数据目录
RUN mkdir -p /app/data

EXPOSE 8080

VOLUME ["/app/data"]

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/api/health || exit 1

ENTRYPOINT ["/app/gost-panel"]
CMD ["-db", "/app/data/panel.db"]
