FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git nodejs npm

# 复制 Go 模块文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 构建前端
WORKDIR /app/web
RUN npm install && npm run build

# 构建后端
WORKDIR /app
RUN CGO_ENABLED=1 go build -o gost-panel ./cmd/panel

# 运行镜像
FROM alpine:latest

RUN apk add --no-cache ca-certificates sqlite

WORKDIR /app

COPY --from=builder /app/gost-panel .

EXPOSE 8080

ENV DB_PATH=/app/data/panel.db
ENV LISTEN_ADDR=:8080

VOLUME ["/app/data"]

CMD ["./gost-panel"]
