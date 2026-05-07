# Ping Pong Docker Service

一个最小的Docker镜像，在启动时访问指定WEBHOOK，并提供ping-pong HTTP服务，同时提供多个URL监控通知功能。

**GitHub仓库**: [https://github.com/yafoo/ping-pong](https://github.com/yafoo/ping-pong)

## 功能特性

- ✅ 基于scratch镜像，极致最小化（使用Go + UPX压缩）
- ✅ 支持WEBHOOK配置（环境变量或命令行参数）
- ✅ 支持自定义端口（环境变量或命令行参数）
- ✅ **支持URL监控轮询**（定时检查目标URL健康状态，失败时发送webhook通知）
- ✅ 提供简单的HTTP ping-pong服务
- ✅ 支持多平台构建（amd64, arm64, arm/v7）
- ✅ GitHub Actions自动构建和推送到DockerHub
- ✅ 日志带时间戳，方便调试

## 本地构建和运行

### 构建镜像

```bash
docker build -t yafoo/ping-pong .
```

### 运行容器

#### 方式1：使用默认配置

```bash
docker run -d -p 10101:10101 --name ping-pong yafoo/ping-pong
```

#### 方式2：使用环境变量

```bash
# 仅设置WEBHOOK
docker run -d -p 10101:10101 \
  -e WEBHOOK=https://www.example.com \
  --name ping-pong yafoo/ping-pong

# 同时设置WEBHOOK和端口
docker run -d -p 8080:8080 \
  -e WEBHOOK=https://www.example.com \
  -e PORT=8080 \
  --name ping-pong yafoo/ping-pong

# 启用URL监控功能
docker run -d -p 10101:10101 \
  -e PING_URL="https://api1.com,https://api2.com" \
  -e PING_INTERVAL="5,15" \
  -e WEBHOOK_PARAMS="service=api1,service=api2" \
  -e WEBHOOK="https://notify.example.com/alert" \
  --name ping-pong yafoo/ping-pong
```

#### 方式3：使用命令行参数

```bash
# 仅设置端口
docker run -d -p 8080:8080 \
  --name ping-pong yafoo/ping-pong \
  --port 8080

# 同时设置WEBHOOK和端口
docker run -d -p 8080:8080 \
  --name ping-pong yafoo/ping-pong \
  --webhook https://www.example.com \
  --port 8080

# 使用短参数格式
docker run -d -p 8080:8080 \
  --name ping-pong yafoo/ping-pong \
  -w https://www.example.com \
  -p 8080

# 启用URL监控功能（长参数）
docker run -d -p 10101:10101 \
  --name ping-pong yafoo/ping-pong \
  --ping-url "https://api1.com,https://api2.com" \
  --ping-interval "5,15" \
  --webhook-params "service=api1,service=api2" \
  --webhook "https://notify.example.com/alert"

# 启用URL监控功能（短参数）
docker run -d -p 10101:10101 \
  --name ping-pong yafoo/ping-pong \
  -u "https://api1.com;https://api2.com" \
  -i "5;15" \
  -wp "service=api1;service=api2" \
  -w "https://notify.example.com/alert"
```

### 直接运行（非Docker）

```bash
# 编译
go build -o ping main.go

# 使用默认配置
./ping

# 使用命令行参数
./ping --port 8080 --webhook https://www.example.com
./ping -p 8080 -w https://www.example.com

# 使用环境变量
export PORT=8080
export WEBHOOK=https://www.example.com
./ping

# 启用URL监控功能
./ping --ping-url "https://api1.com,https://api2.com" \
       --ping-interval "5,15" \
       --webhook-params "service=api1,service=api2" \
       --webhook "https://notify.example.com/alert"

# 使用短参数
./ping -u "https://api1.com;https://api2.com" \
       -i "5;15" \
       -wp "service=api1;service=api2" \
       -w "https://notify.example.com/alert"
```

### 最佳实践（以PushMe为例）
```bash
# 实现功能：
# 1. 系统启动或异常重启时，会给你PushMe客户端发送系统启动的通知
# 2. 在默认端口10101上启动ping-pong HTTP服务，以被其他工具监控
# 3. 启动URL监控服务，每隔5分钟和15分钟分别检查https://api1.com和https://api2.com接口状态
# 4. 当接口访问失败时，会给你PushMe客户端发送接口访问失败+错误内容的通知
# 以命令行短参数为例
./ping -w 'https://push.i-i.me/?push_key=YourPushKey&title=ping-pong&content=系统已启动：SystemName' \
       -u 'https://api1.com,https://api2.com' \
       -i '5,15' \
       -wp 'content=api1接口异常%0A{$err},content=api2接口异常%0A{$err}'
```


### 测试服务

```bash
curl http://localhost:10101
# 返回: pong
```

### 查看日志

```
docker logs ping-pong
# 输出示例（基础模式）:
# [2026-04-24 17:30:15] 正在访问指定的WEBHOOK
# [2026-04-24 17:30:16] WEBHOOK访问完成
# [2026-04-24 17:30:16] 启动ping-pong HTTP服务（端口10101）...
# [2026-04-24 17:30:16] 服务已启动，监听端口 10101

# 输出示例（启用URL监控）:
# [2026-04-24 17:30:15] 正在访问指定的WEBHOOK
# [2026-04-24 17:30:16] WEBHOOK访问完成
# [2026-04-24 17:30:16] 启动URL监控服务,共 2 个监控目标
# [2026-04-24 17:30:16] 开始监控URL: https://api1.com (间隔: 5分钟)
# [2026-04-24 17:30:16] 开始监控URL: https://api2.com (间隔: 15分钟)
# [2026-04-24 17:30:16] 启动ping-pong HTTP服务（端口10101）...
# [2026-04-24 17:30:16] 服务已启动，监听端口 10101
# [2026-04-24 17:35:16] 正在检查URL: https://api1.com
# [2026-04-24 17:35:17] URL访问成功: https://api1.com - 状态码: 200
```

## 配置说明

### 配置优先级

配置加载遵循以下优先级顺序（从高到低）：

1. **命令行短参数**（`-p`, `-w`, `-u`, `-i`, `-wp`）
2. **命令行长参数**（`--port`, `--webhook`, `--ping-url`, `--ping-interval`, `--webhook-params`）
3. **环境变量**（`PORT`, `WEBHOOK`, `PING_URL`, `PING_INTERVAL`, `WEBHOOK_PARAMS`）
4. **默认值**（端口：10101，其他：空或默认值）

**示例**：如果同时设置了 `-p 8080` 和 `--port 9090`，实际使用的端口是 `8080`（短参数优先）。

### 可用参数

| 参数类型 | 短参数 | 长参数 | 环境变量 | 说明 | 默认值 |
|---------|--------|--------|----------|------|--------|
| 端口 | `-p` | `--port` | `PORT` | HTTP服务监听端口 | `10101` |
| Webhook | `-w` | `--webhook` | `WEBHOOK` | 启动时访问的URL，失败通知的目标URL | 空（跳过） |
| 监控URL | `-u` | `--ping-url` | `PING_URL` | 要监控的URL（支持多个，用逗号/分号/竖线分隔） | 空（不启用监控） |
| 监控间隔 | `-i` | `--ping-interval` | `PING_INTERVAL` | 监控间隔时间（分钟，支持多个，与URL一一对应） | `10` |
| Webhook参数 | `-wp` | `--webhook-params` | `WEBHOOK_PARAMS` | 失败通知时追加的参数（支持多个，与URL一一对应） | 空 |

### URL监控功能详解

**功能说明**：
- 系统启动后，如果设置了 `--ping-url`，会启动后台服务定时访问这些URL
- 每个URL可以配置独立的轮询间隔时间
- 如果访问失败（网络错误或HTTP状态码非2xx），则调用webhook URL发送通知
- webhook通知时可以追加自定义参数（用于标识哪个URL失败、失败原因等）

**多值分隔符**：
支持三种分隔符来设置多个值：
- 逗号 `,`
- 分号 `;`
- 竖线 `|`

**使用示例**：

```
# 单个URL监控（每5分钟检查一次）
./ping -u "https://api.example.com" -i "5" -wp "service=api"

# 多个URL监控（不同间隔和参数）
./ping -u "https://api1.com,https://api2.com,https://api3.com" \
       -i "5,10,15" \
       -wp "url=api1&status=down,url=api2&status=down,url=api3&status=down"

# 使用不同分隔符
./ping -u "https://api1.com;https://api2.com" \
       -i "5|10" \
       -wp "service=api1|service=api2"

# 使用变量占位符（推荐用单引号包裹参数，避免shell转义问题）
./ping -u "https://api.example.com" \
       -wp 'msg=服务异常&url={$url}&error={$err}&time={$time}'

# 包含换行的消息（使用真正的换行符或\n的URL编码%0A）
./ping -u "https://api.example.com" \
       -wp $'msg=访问失败\\nURL: {$url}\\n错误: {$err}'

# 使用完整URL作为通知地址（webhookParam以http://或https://开头时）
./ping -u "https://api.example.com" \
       -wp 'https://emergency-notify.example.com/critical?service=api&error={$err}'

# Docker环境
docker run -d -p 10101:10101 \
  -e PING_URL="https://api1.com,https://api2.com" \
  -e PING_INTERVAL="5,15" \
  -e WEBHOOK_PARAMS="service=api1&error={\$err},service=api2&error={\$err}" \
  -e WEBHOOK="https://notify.example.com/alert?token=xxx" \
  yafoo/ping-pong
```

**支持的变量占位符**：

| 占位符 | 说明 | 示例值 |
|--------|------|--------|
| `{$err}` | 错误信息 | `dial tcp: connection refused` |
| `{$url}` | 监控的URL | `https://api.example.com/health` |
| `{$time}` | 当前时间 | `2026-05-06 17:30:15` |

**工作流程**：
1. 解析配置的监控URL列表
2. 为每个URL启动独立的goroutine进行监控
3. 按照配置的间隔时间定时访问URL
4. 检测访问结果（网络错误或非2xx状态码视为失败）
5. 失败时解析webhook参数中的变量占位符（`{$err}`, `{$url}`, `{$time}`）
6. **判断webhook参数格式**：
   - 如果以 `http://` 或 `https://` 开头 → 直接作为通知URL使用
   - 否则 → 与基础webhook URL合并参数
7. 调用webhook URL发送通知
8. 持续循环监控

**高级用法：动态通知URL**

当webhookParam以 `http://` 或 `https://` 开头时，系统会将其视为完整的通知URL，直接使用该URL发送通知，不再与基础webhook URL合并。这允许你根据监控目标动态选择不同的通知端点：

```bash
# 示例1：不同服务使用不同的通知URL
./ping -u "https://api1.com,https://api2.com" \
       -wp 'https://notify1.example.com/alert?error={$err},https://notify2.example.com/critical?error={$err}'

# 示例2：根据错误类型路由到不同的通知渠道
./ping -u "https://payment-api.com" \
       -wp 'https://slack-webhook.example.com/payment-alert?msg=支付接口异常: {$err}'

# 示例3：结合变量占位符构建动态URL
./ping -u "https://api.example.com" \
       -wp 'https://monitor.example.com/incident/create?source={$url}&error={$err}&time={$time}'

# 示例4：不配置基础webhook，仅使用webhookParam提供完整URL
./ping -u "https://api.example.com" \
       -wp 'https://emergency-notify.example.com/critical?error={$err}'
       # 注意：这里不需要设置 --webhook 参数
```

**重要说明**：

1. **完整URL模式优先级最高**：
   - 如果webhookParam是完整URL格式，即使没有配置 `--webhook` 参数，也能正常发送通知
   - 此时 `--webhook` 参数会被完全忽略

2. **参数合并模式需要基础webhook**：
   - 如果webhookParam不是完整URL格式，则必须配置 `--webhook` 参数
   - 否则系统会跳过通知并输出警告日志

3. **灵活组合**：
   - 可以为不同的监控URL配置不同的通知策略
   - 有的使用参数合并，有的使用完整URL，互不影响

**注意事项**：

1. **Shell转义问题**：
   - Linux/Mac：建议使用单引号 `'...'` 包裹包含占位符的参数，防止shell提前展开
   - Windows PowerShell：使用双引号 `"..."`，但需要对 `$` 进行转义 `` `$ ``
   - 或者使用 `$'...'` 语法（bash/zsh）来支持 `\n` 等特殊字符

2. **换行符处理**：
   - 直接在参数中使用 `\n` 不会被识别为换行
   - 使用 `$'...\n...'` 语法（bash/zsh）插入真实换行
   - 或使用URL编码 `%0A` 表示换行：`msg=line1%0Aline2`

3. **特殊字符**：
   - 如果错误信息包含特殊字符（如 `&`, `=`, `?`），会被自动URL编码
   - 例如：`connection refused & timeout` → `connection%20refused%20%26%20timeout`

### 参数格式示例

```
# 长参数格式
./ping --port 8080 --webhook https://example.com/hook

# 短参数格式
./ping -p 8080 -w https://example.com/hook

# 等号格式
./ping --port=8080 --webhook=https://example.com/hook
./ping -p=8080 -w=https://example.com/hook

# 混合格式
./ping --port 8080 -w https://example.com/hook

# URL监控完整示例
./ping -u "https://api1.com,https://api2.com" \
       -i "5,15" \
       -wp "service=api1,service=api2" \
       -w "https://notify.example.com/alert"

# 查看帮助
./ping -h
```

## GitHub Actions 配置

### 设置步骤

1. 在GitHub仓库中创建Secrets：
   - `DOCKERHUB_USERNAME`: 你的DockerHub用户名
   - `DOCKERHUB_TOKEN`: 你的DockerHub Access Token

2. 获取DockerHub Access Token：
   - 登录DockerHub
   - 进入 Account Settings → Security
   - 点击 "New Access Token"
   - 赋予读写权限

3. 推送代码到GitHub：
   ```bash
   git init
   git add .
   git commit -m "Initial commit"
   git branch -M main
   git remote add origin https://github.com/YOUR_USERNAME/your-repo.git
   git push -u origin main
   ```

4. 自动触发构建：
   - 推送到main分支会自动触发构建
   - 创建tag（如v1.0.0）会发布版本镜像

### 构建的平台

- linux/amd64 (x86_64)
- linux/arm64 (ARM 64-bit)
- linux/arm/v7 (ARM 32-bit, Raspberry Pi等)

### 镜像标签策略

- `latest`: 最新构建（仅main分支）
- `v1.0.0`: 语义化版本标签
- `1.0`: 次版本号标签
- `sha-{commit}`: Git commit SHA

## 项目结构

```
.
├── Dockerfile              # Docker镜像定义
├── main.go                 # Go应用程序主文件
└── .github/
    └── workflows/
        └── docker-build.yml # GitHub Actions配置
```

## 许可证

暂无
