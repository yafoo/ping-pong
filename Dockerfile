# 第一阶段：构建Go应用
FROM golang:1.21-alpine AS builder

# 安装UPX压缩工具
RUN apk add --no-cache upx

# 设置工作目录
WORKDIR /app

# 先复制go.mod（利用Docker缓存层）
COPY go.mod ./

# 下载依赖（如果有外部依赖会在这里缓存，纯标准库项目会跳过）
RUN go mod download 2>/dev/null || true

# 复制源代码
COPY . .

# 编译Go应用（静态链接，优化大小）
# TARGETPLATFORM 和 TARGETARCH 由 Docker Buildx 自动设置
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o ping-pong main.go

# 使用UPX压缩二进制文件
RUN upx --best --lzma ping-pong

# 第二阶段：最小运行时镜像
FROM scratch

# 设置工作目录
WORKDIR /app

# 从构建阶段复制压缩后的二进制文件
COPY --from=builder /app/ping-pong .

# 暴露端口
EXPOSE 10101

# 设置环境变量
ENV WEBHOOK=""

# 启动应用
ENTRYPOINT ["/app/ping-pong"]
