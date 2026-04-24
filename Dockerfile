# 使用Alpine作为基础镜像（最小化）
FROM alpine:latest

# 安装必要的工具：netcat-openbsd用于HTTP服务，wget用于访问URL，socat用于更好的TCP处理
RUN apk add --no-cache netcat-openbsd wget socat

# 设置工作目录
WORKDIR /app

# 复制启动脚本
COPY entrypoint.sh /app/entrypoint.sh

# 赋予执行权限
RUN chmod +x /app/entrypoint.sh

# 暴露端口
EXPOSE 10101

# 设置环境变量（可选，提供默认值）
ENV WEBHOOK=""

# 启动容器
ENTRYPOINT ["/app/entrypoint.sh"]
