# Ping Pong Docker Service

一个最小的Docker镜像，能接受WEBHOOK环境变量，在容器启动时访问该URL，并提供ping-pong HTTP服务。

## 功能特性

- ✅ 基于Alpine Linux，镜像体积最小
- ✅ 支持WEBHOOK环境变量，启动时自动访问指定URL
- ✅ 提供简单的HTTP ping-pong服务（监听8080端口）
- ✅ 支持多平台构建（amd64, arm64, arm/v7）
- ✅ GitHub Actions自动构建和推送到DockerHub

## 本地构建和运行

### 构建镜像

```bash
docker build -t ping-pong .
```

### 运行容器

不带WEBHOOK：
```bash
docker run -d -p 10101:10101 --name ping-pong ping-pong
```

带WEBHOOK环境变量：
```bash
docker run -d -p 10101:10101 -e WEBHOOK=https://www.example.com --name ping-pong ping-pong
```

### 测试服务

```bash
curl http://localhost:10101
# 返回: pong
```

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| WEBHOOK | 容器启动时要访问的URL | 空（跳过） |

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
├── entrypoint.sh           # 容器启动脚本
└── .github/
    └── workflows/
        └── docker-build.yml # GitHub Actions配置
```

## 许可证

暂无
