package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"
)

func logWithTime(message string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s\n", now, message)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func main() {
	// 检查WEBHOOK环境变量
	webhook := os.Getenv("WEBHOOK")
	if webhook == "" {
		logWithTime("警告: WEBHOOK 环境变量未设置，将跳过URL访问步骤")
	} else {
		logWithTime("正在访问指定的WEBHOOK")

		// 创建自定义 HTTP 客户端，跳过证书验证
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{
			Timeout:   10 * time.Second,
			Transport: tr,
		}

		resp, err := client.Get(webhook)
		if err != nil {
			logWithTime(fmt.Sprintf("访问WEBHOOK失败: %v（继续启动服务）", err))
		} else {
			resp.Body.Close()
			logWithTime("WEBHOOK访问完成")
		}
	}

	// 启动HTTP服务
	port := "10101"
	logWithTime(fmt.Sprintf("启动ping-pong HTTP服务（端口%s）...", port))

	http.HandleFunc("/", pingHandler)

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	logWithTime(fmt.Sprintf("服务已启动，监听端口 %s", port))
	if err := server.ListenAndServe(); err != nil {
		logWithTime(fmt.Sprintf("服务错误: %v", err))
		os.Exit(1)
	}
}