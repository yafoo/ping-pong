package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
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

// encodeWebhookURL 智能处理 webhook URL 编码
func encodeWebhookURL(rawURL string) (string, error) {
	// 解析 URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("URL解析失败: %v", err)
	}

	// 检查查询参数是否已经编码过
	query := parsedURL.RawQuery
	if query == "" {
		return rawURL, nil
	}

	// 尝试解码后再重新编码，确保正确编码
	decodedQuery, err := url.QueryUnescape(query)
	if err != nil {
		// 如果解码失败，说明可能已经正确编码或部分编码
		// 直接对每个参数值进行编码
		values, _ := url.ParseQuery(query)
		encodedValues := make(url.Values)
		for k, v := range values {
			encodedValues[k] = v // 保持原样
		}
		parsedURL.RawQuery = encodedValues.Encode()
		return parsedURL.String(), nil
	}

	// 成功解码后，重新编码以确保一致性
	values, err := url.ParseQuery(decodedQuery)
	if err != nil {
		return rawURL, nil
	}

	parsedURL.RawQuery = values.Encode()
	return parsedURL.String(), nil
}

func main() {
	// 检查WEBHOOK环境变量
	webhook := os.Getenv("WEBHOOK")
	if webhook == "" {
		logWithTime("警告: WEBHOOK 环境变量未设置，将跳过URL访问步骤")
	} else {
		logWithTime("正在访问指定的WEBHOOK")

		// 智能处理 URL 编码
		encodedWebhook, err := encodeWebhookURL(webhook)
		if err != nil {
			logWithTime(fmt.Sprintf("URL编码处理警告: %v（使用原始URL）", err))
			encodedWebhook = webhook
		} else if encodedWebhook != webhook {
			logWithTime("已对WEBHOOK URL进行编码处理")
		}

		// 创建自定义 HTTP 客户端，跳过证书验证
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{
			Timeout:   10 * time.Second,
			Transport: tr,
		}

		resp, err := client.Get(encodedWebhook)
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
