#!/bin/sh

# 检查WEBHOOK环境变量是否设置
if [ -z "$WEBHOOK" ]; then
    echo "警告: WEBHOOK 环境变量未设置，将跳过URL访问步骤"
else
    echo "正在访问指定的WEBHOOK: $WEBHOOK"
    # 使用wget访问URL（Alpine自带）
    wget -q --spider "$WEBHOOK" || echo "访问WEBHOOK失败（继续启动服务）"
    echo "WEBHOOK访问完成"
fi

# 启动ping-pong HTTP服务
echo "启动ping-pong HTTP服务..."
while true; do
    echo -e "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nConnection: close\r\n\r\npong" | nc -l -p 10101 -w 1
done
