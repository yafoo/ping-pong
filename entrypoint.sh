#!/bin/sh

# 检查WEBHOOK环境变量是否设置
if [ -z "$WEBHOOK" ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 警告: WEBHOOK 环境变量未设置，将跳过URL访问步骤"
else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 正在访问指定的WEBHOOK"
    # 使用wget访问URL（Alpine自带）
    wget -q --spider "$WEBHOOK" || echo "[$(date '+%Y-%m-%d %H:%M:%S')] 访问WEBHOOK失败（继续启动服务）"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] WEBHOOK访问完成"
fi

# 创建HTTP响应处理脚本
cat > /tmp/http_handler.sh << 'EOF'
#!/bin/sh
# 读取并丢弃请求头
while read line; do
    if [ -z "$line" ] || [ "$line" = "" ] || [ "$line" = $'\r' ]; then
        break
    fi
done

# 发送响应
printf "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nConnection: close\r\nContent-Length: 4\r\n\r\npong"
EOF

chmod +x /tmp/http_handler.sh

# 启动ping-pong HTTP服务
echo "[$(date '+%Y-%m-%d %H:%M:%S')] 启动ping-pong HTTP服务（端口10101）..."

# 使用socat或改进的nc方式
if command -v socat >/dev/null 2>&1; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 使用socat启动服务"
    socat TCP-LISTEN:10101,reuseaddr,fork SYSTEM:"/tmp/http_handler.sh"
else
    # 使用nc的备用方案
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 使用netcat启动服务"
    while true; do
        printf "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nConnection: close\r\nContent-Length: 4\r\n\r\npong" | nc -l -p 10101 -w 1 -q 0 >/dev/null 2>&1
    done
fi