package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
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

// getConfigValue 通用配置获取函数，按优先级获取：短参数 > 长参数 > 环境变量 > 默认值
func getConfigValue(shortParam *string, longParam *string, envKey string, defaultValue string) string {
	// 1. 短参数优先
	if shortParam != nil && *shortParam != "" {
		return *shortParam
	}

	// 2. 长参数次之
	if longParam != nil && *longParam != "" {
		return *longParam
	}

	// 3. 环境变量再次
	envValue := os.Getenv(envKey)
	if envValue != "" {
		return envValue
	}

	// 4. 默认值
	return defaultValue
}

// mergeWebhookParams 智能合并webhook参数，重名参数会被替换
func mergeWebhookParams(baseURL string, newParams string) string {
	if newParams == "" {
		return baseURL
	}

	// 解析基础URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		// 如果URL解析失败，直接追加
		separator := "?"
		if strings.Contains(baseURL, "?") {
			separator = "&"
		}
		return baseURL + separator + newParams
	}

	// 解析现有查询参数
	existingParams, _ := url.ParseQuery(parsedURL.RawQuery)

	// 解析新参数
	newParamsParsed, _ := url.ParseQuery(newParams)

	// 合并参数：新参数会覆盖同名的旧参数
	for key, values := range newParamsParsed {
		existingParams[key] = values
	}

	// 重新编码查询字符串
	parsedURL.RawQuery = existingParams.Encode()
	return parsedURL.String()
}

// parseMultiValue 解析支持分隔符的多值参数
func parseMultiValue(value string, defaultValue string) []string {
	if value == "" {
		if defaultValue != "" {
			return strings.Split(defaultValue, ",")
		}
		return []string{}
	}
	// 支持逗号、分号、竖线作为分隔符
	value = strings.ReplaceAll(value, ";", ",")
	value = strings.ReplaceAll(value, "|", ",")
	parts := strings.Split(value, ",")
	// 清理空白
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// monitorURL 监控单个URL
func monitorURL(targetURL string, intervalMinutes int, webhookURL string, webhookParam string) {
	interval := time.Duration(intervalMinutes) * time.Minute
	logWithTime(fmt.Sprintf("开始监控URL: %s (间隔: %d分钟)", targetURL, intervalMinutes))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 创建HTTP客户端
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	for {
		select {
		case <-ticker.C:
			logWithTime(fmt.Sprintf("正在检查URL: %s", targetURL))

			// 尝试访问目标URL
			resp, err := client.Get(targetURL)
			failed := false

			if err != nil {
				logWithTime(fmt.Sprintf("URL访问失败: %s - 错误: %v", targetURL, err))
				failed = true
			} else {
				resp.Body.Close()
				if resp.StatusCode < 200 || resp.StatusCode >= 300 {
					logWithTime(fmt.Sprintf("URL返回异常状态码: %s - 状态码: %d", targetURL, resp.StatusCode))
					failed = true
				} else {
					logWithTime(fmt.Sprintf("URL访问成功: %s - 状态码: %d", targetURL, resp.StatusCode))
				}
			}

			// 如果访问失败且有webhook配置,则调用webhook
			if failed && webhookURL != "" {
				notificationURL := webhookURL
				if webhookParam != "" {
					// 智能合并参数：解析现有参数，替换或新增
					notificationURL = mergeWebhookParams(webhookURL, webhookParam)
				}

				logWithTime(fmt.Sprintf("发送失败通知到: %s", notificationURL))
				webhookResp, webhookErr := client.Get(notificationURL)
				if webhookErr != nil {
					logWithTime(fmt.Sprintf("Webhook通知失败: %v", webhookErr))
				} else {
					webhookResp.Body.Close()
					logWithTime("Webhook通知发送完成")
				}
			}
		}
	}
}

// startMonitoring 启动所有URL的监控
func startMonitoring(monitorURLs []string, intervals []string, webhookURL string, webhookParams []string) {
	if len(monitorURLs) == 0 {
		return
	}

	logWithTime(fmt.Sprintf("启动URL监控服务,共 %d 个监控目标", len(monitorURLs)))

	for i, targetURL := range monitorURLs {
		// 获取对应的间隔时间
		intervalStr := "10" // 默认10分钟
		if i < len(intervals) && intervals[i] != "" {
			intervalStr = intervals[i]
		}

		intervalMinutes, err := strconv.Atoi(intervalStr)
		if err != nil || intervalMinutes <= 0 {
			logWithTime(fmt.Sprintf("警告: 无效的间隔值 '%s',使用默认值10分钟", intervalStr))
			intervalMinutes = 10
		}

		// 获取对应的webhook参数
		webhookParam := ""
		if i < len(webhookParams) {
			webhookParam = webhookParams[i]
		}

		// 为每个URL启动独立的goroutine进行监控
		go monitorURL(targetURL, intervalMinutes, webhookURL, webhookParam)
	}
}

func main() {
	// 定义命令行参数（同时支持长短格式）
	webhookShort := flag.String("w", "", "Webhook URL to call on startup (short)")
	webhookLong := flag.String("webhook", "", "Webhook URL to call on startup (long)")
	portShort := flag.String("p", "", "HTTP service port (default: 10101) (short)")
	portLong := flag.String("port", "", "HTTP service port (default: 10101) (long)")

	// 新增监控相关参数
	pingURLShort := flag.String("u", "", "URL to ping/monitor (comma/semicolon/pipe separated for multiple) (short)")
	pingURLLong := flag.String("ping-url", "", "URL to ping/monitor (comma/semicolon/pipe separated for multiple) (long)")
	pingIntervalShort := flag.String("i", "", "Ping interval in minutes (comma/semicolon/pipe separated, default: 10) (short)")
	pingIntervalLong := flag.String("ping-interval", "", "Ping interval in minutes (comma/semicolon/pipe separated, default: 10) (long)")
	webhookParamsShort := flag.String("wp", "", "Webhook parameters to append on failure (comma/semicolon/pipe separated) (short)")
	webhookParamsLong := flag.String("webhook-params", "", "Webhook parameters to append on failure (comma/semicolon/pipe separated) (long)")

	flag.Parse()

	// 使用通用函数获取配置
	webhook := getConfigValue(webhookShort, webhookLong, "WEBHOOK", "")
	port := getConfigValue(portShort, portLong, "PORT", "10101")
	pingURLValue := getConfigValue(pingURLShort, pingURLLong, "PING_URL", "")
	pingIntervalValue := getConfigValue(pingIntervalShort, pingIntervalLong, "PING_INTERVAL", "")
	webhookParamsValue := getConfigValue(webhookParamsShort, webhookParamsLong, "WEBHOOK_PARAMS", "")

	if webhook == "" {
		logWithTime("警告: WEBHOOK 未设置（既无命令行参数也无环境变量），将跳过URL访问步骤")
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

	logWithTime(fmt.Sprintf("启动ping-pong HTTP服务（端口%s）...", port))

	http.HandleFunc("/", pingHandler)

	// 启动监控服务
	if pingURLValue != "" {
		pingURLs := parseMultiValue(pingURLValue, "")
		intervals := parseMultiValue(pingIntervalValue, "10")
		webhookParams := parseMultiValue(webhookParamsValue, "")

		startMonitoring(pingURLs, intervals, webhook, webhookParams)
	}

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
