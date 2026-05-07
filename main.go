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

func setupTimezone(timezoneOffset string) {
	if timezoneOffset == "" || timezoneOffset == "0" {
		return
	}

	// 将时区偏移量转换为整数
	timezoneOffsetInt, err := strconv.Atoi(timezoneOffset)
	if err != nil {
		logWithTime("[启动阶段] 警告: 时区偏移量转换失败，请检查输入: " + timezoneOffset)
		return
	}

	time.FixedZone("CST", timezoneOffsetInt)
	logWithTime("[启动阶段] 已设置时区偏移为: " + timezoneOffset)
}

func logWithTime(message string) {
	now := time.Now().Format("2006-01-02 15:04:05 MST")
	fmt.Printf("[%s] %s\n", now, message)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// encodeWebhookURL 智能处理 webhook URL 编码，确保参数正确编码且避免双重编码
func encodeWebhookURL(rawURL string) (string, error) {
	// 解析 URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("URL解析失败: %v", err)
	}

	// 如果没有查询参数，直接返回
	if parsedURL.RawQuery == "" {
		return rawURL, nil
	}

	// 利用 url.ParseQuery 自动解码的特性，统一处理编码
	values, err := url.ParseQuery(parsedURL.RawQuery)
	if err != nil {
		// 如果解析失败，保持原样返回
		return rawURL, nil
	}

	// 重新编码，确保所有参数都使用标准URL编码
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

	// 解析现有查询参数（自动处理已编码的参数）
	existingParams, _ := url.ParseQuery(parsedURL.RawQuery)

	// 解析新参数（支持未编码的中文等字符）
	newParamsParsed, _ := url.ParseQuery(newParams)

	// 合并参数：新参数会覆盖同名的旧参数
	for key, values := range newParamsParsed {
		existingParams[key] = values
	}

	// 重新编码查询字符串（Go标准库会自动对所有参数进行正确的URL编码）
	parsedURL.RawQuery = existingParams.Encode()
	return parsedURL.String()
}

// resolveWebhookParamVariables 解析webhook参数中的变量占位符
// 支持的占位符:
// - {$err}: 错误信息
// - {$url}: 监控的URL
// - {$time}: 当前时间
func resolveWebhookParamVariables(param string, targetURL string, checkErr error) string {
	if param == "" {
		return param
	}

	// 替换错误信息
	if checkErr != nil {
		param = strings.ReplaceAll(param, "{$err}", checkErr.Error())
	} else {
		param = strings.ReplaceAll(param, "{$err}", "unknown error")
	}

	// 替换URL
	param = strings.ReplaceAll(param, "{$url}", targetURL)

	// 替换时间
	param = strings.ReplaceAll(param, "{$time}", time.Now().Format("2006-01-02 15:04:05"))

	return param
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

			// 如果访问失败，尝试发送webhook通知
			if failed {
				var notificationURL string

				if webhookParam != "" {
					// 处理参数中的变量占位符
					resolvedParam := resolveWebhookParamVariables(webhookParam, targetURL, err)

					// 检查resolvedParam是否是完整的URL（以http://或https://开头）
					if strings.HasPrefix(resolvedParam, "http://") || strings.HasPrefix(resolvedParam, "https://") {
						// 如果是完整URL，直接作为通知URL使用（即使webhookURL为空也可以）
						notificationURL = resolvedParam
						logWithTime("检测到完整URL格式，直接使用作为通知地址")
					} else if webhookURL != "" {
						// 如果不是完整URL且有基础webhook URL，则合并参数
						notificationURL = mergeWebhookParams(webhookURL, resolvedParam)
					} else {
						// 既不是完整URL，也没有基础webhook URL，跳过通知
						logWithTime("警告: webhookParam不是完整URL且未配置基础webhook，跳过通知")
						continue
					}
				} else if webhookURL != "" {
					// 没有额外参数，但有基础webhook URL，直接使用
					notificationURL = webhookURL
				} else {
					// 既没有webhookParam也没有webhookURL，跳过通知
					logWithTime("警告: 未配置webhook通知地址，跳过通知")
					continue
				}

				// 统一对最终的通知URL进行编码处理
				encodedURL, encodeErr := encodeWebhookURL(notificationURL)
				if encodeErr == nil {
					notificationURL = encodedURL
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

	// 时区配置参数
	timezoneOffsetShort := flag.String("t", "", "Timezone offset in seconds from UTC (short)")
	timezoneOffsetLong := flag.String("tz-offset", "", "Timezone offset in seconds from UTC (long)")

	// 新增监控相关参数
	pingURLShort := flag.String("u", "", "URL to ping/monitor (comma/semicolon/pipe separated for multiple) (short)")
	pingURLLong := flag.String("ping-url", "", "URL to ping/monitor (comma/semicolon/pipe separated for multiple) (long)")
	pingIntervalShort := flag.String("i", "", "Ping interval in minutes (comma/semicolon/pipe separated, default: 10) (short)")
	pingIntervalLong := flag.String("ping-interval", "", "Ping interval in minutes (comma/semicolon/pipe separated, default: 10) (long)")
	webhookParamsShort := flag.String("wp", "", "Webhook parameters to append on failure (comma/semicolon/pipe separated) (short)")
	webhookParamsLong := flag.String("webhook-params", "", "Webhook parameters to append on failure (comma/semicolon/pipe separated) (long)")

	flag.Parse()

	// 设置时区（必须在其他操作之前）
	timezoneOffset := getConfigValue(timezoneOffsetShort, timezoneOffsetLong, "TZ_OFFSET", "")
	setupTimezone(timezoneOffset)

	// 使用通用函数获取配置
	webhook := getConfigValue(webhookShort, webhookLong, "WEBHOOK", "")
	port := getConfigValue(portShort, portLong, "PORT", "10101")
	pingURLValue := getConfigValue(pingURLShort, pingURLLong, "PING_URL", "")
	pingIntervalValue := getConfigValue(pingIntervalShort, pingIntervalLong, "PING_INTERVAL", "")
	webhookParamsValue := getConfigValue(webhookParamsShort, webhookParamsLong, "WEBHOOK_PARAMS", "")

	if webhook == "" {
		logWithTime("[启动阶段] 警告: WEBHOOK 未设置（既无命令行参数也无环境变量），将跳过URL访问步骤")
	} else {
		logWithTime("[启动阶段] 正在异步访问指定的WEBHOOK...")

		// 异步调用 webhook，不阻塞应用启动
		go func() {
			// 智能处理 URL 编码
			encodedWebhook, err := encodeWebhookURL(webhook)
			if err != nil {
				logWithTime(fmt.Sprintf("[启动阶段] URL编码处理警告: %v（使用原始URL）", err))
				encodedWebhook = webhook
			} else if encodedWebhook != webhook {
				logWithTime("[启动阶段] 已对WEBHOOK URL进行编码处理")
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
				logWithTime(fmt.Sprintf("[启动阶段] 访问WEBHOOK失败: %v", err))
				logWithTime("[启动阶段] 提示：请检查WEBHOOK URL是否可达，或考虑增加超时时间")
			} else {
				defer resp.Body.Close()
				logWithTime(fmt.Sprintf("[启动阶段] WEBHOOK访问完成 (状态码: %d)", resp.StatusCode))
			}
		}()

		logWithTime("[启动阶段] Webhook调用已在后台启动，继续初始化服务...")
	}

	logWithTime(fmt.Sprintf("[启动阶段] 启动ping-pong HTTP服务（端口%s）...", port))

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

	logWithTime(fmt.Sprintf("[启动阶段] 服务已启动，监听端口 %s", port))
	if err := server.ListenAndServe(); err != nil {
		logWithTime(fmt.Sprintf("[启动阶段] 服务错误: %v", err))
		os.Exit(1)
	}
}
