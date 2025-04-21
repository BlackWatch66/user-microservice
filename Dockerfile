# 使用官方 Go 镜像作为构建环境
ARG GO_VERSION=1.23
FROM golang:${GO_VERSION}-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 Go 模块文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
# CGO_ENABLED=0 确保静态链接，以便在 Alpine 镜像中运行
# -ldflags="-s -w" 减小二进制文件大小
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o user-service ./cmd/server

# 使用轻量级的 Alpine 镜像作为最终镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/user-service /app/user-service

# 暴露 HTTP 和 gRPC 端口
EXPOSE 8080
EXPOSE 50051

# 运行应用程序
CMD ["/app/user-service"] 