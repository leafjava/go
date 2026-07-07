# 第23课：Docker + CI/CD

> 学习时间：3-4小时 | 难度：⭐⭐⭐

## 📋 本课目标

- 掌握 Go 项目的 Docker 化
- 学会编写多阶段构建 Dockerfile
- 使用 Docker Compose 编排服务
- 配置 GitHub Actions CI/CD
- 实现自动化测试和部署

## 1. Docker 基础

### Docker 核心概念

| 概念 | 说明 | 类比 |
|------|------|------|
| Image（镜像） | 只读模板，包含运行环境和代码 | 类（Class） |
| Container（容器） | 镜像的运行实例 | 对象（Instance） |
| Dockerfile | 构建镜像的指令文件 | 构建脚本 |
| Docker Compose | 多容器编排工具 | docker-compose.yml |
| Registry | 镜像仓库 | Docker Hub / GHCR |

## 2. Go 项目 Dockerfile

### 多阶段构建（推荐）

```dockerfile
# Dockerfile
# ============================================
# 阶段1: 构建阶段（builder）
# ============================================
FROM golang:1.22-alpine AS builder

# 设置工作目录
WORKDIR /build

# 安装构建依赖
RUN apk add --no-cache gcc musl-dev

# 设置 Go 代理（中国大陆镜像加速）
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# 先复制依赖文件，利用 Docker 缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制文件
# -ldflags="-s -w" 去除调试信息，减小体积
RUN go build -ldflags="-s -w" -o /build/server ./cmd/server

# ============================================
# 阶段2: 运行阶段（runner）
# ============================================
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN adduser -D -g '' appuser

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/server .
COPY --from=builder /build/config ./config

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动服务
ENTRYPOINT ["./server"]
```

### .dockerignore

```
# .dockerignore
# Git
.git
.gitignore

# 文档
*.md
docs/

# IDE
.vscode/
.idea/

# 临时文件
*.exe
*.log
logs/
tmp/

# 测试
*_test.go

# Docker
Dockerfile
.dockerignore
docker-compose.yml

# 依赖
vendor/
```

### 最小化 Dockerfile（scratch 基础镜像）

```dockerfile
# Dockerfile.scratch（适用于纯 Go 静态编译项目）
FROM golang:1.22-alpine AS builder

WORKDIR /build
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o /build/server ./cmd/server

# 使用 scratch 作为最终镜像（约 0MB）
FROM scratch

# 复制 CA 证书（HTTPS 请求需要）
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# 复制时区信息
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY --from=builder /build/server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
```

## 3. Docker Compose

### docker-compose.yml

```yaml
version: '3.8'

services:
  # 主应用服务
  app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: blockchain-service
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - BC_SERVER_PORT=8080
      - BC_SERVER_MODE=release
      - BC_DATABASE_DSN=postgres://postgres:password@db:5432/blockchain?sslmode=disable
      - BC_REDIS_ADDR=redis:6379
      - BC_ETHEREUM_RPC_URL=https://eth.llamarpc.com
      - BC_ETHEREUM_WS_URL=wss://eth.llamarpc.com
      - BC_JWT_SECRET=${JWT_SECRET:-change-me}
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./config:/app/config:ro
      - ./logs:/app/logs
    networks:
      - blockchain-net

  # PostgreSQL 数据库
  db:
    image: postgres:16-alpine
    container_name: blockchain-db
    restart: unless-stopped
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${DB_PASSWORD:-password}
      - POSTGRES_DB=blockchain
    ports:
      - "5432:5432"
    volumes:
      - db_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - blockchain-net

  # Redis
  redis:
    image: redis:7-alpine
    container_name: blockchain-redis
    restart: unless-stopped
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD:-}
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    networks:
      - blockchain-net

volumes:
  db_data:
  redis_data:

networks:
  blockchain-net:
    driver: bridge
```

### .env 文件

```bash
# .env（不提交到 Git）
DB_PASSWORD=your_secure_password
REDIS_PASSWORD=
JWT_SECRET=your_jwt_secret_key
```

### 常用 Docker Compose 命令

```bash
# 启动所有服务
docker compose up -d

# 查看日志
docker compose logs -f app

# 仅重建 app 服务
docker compose up -d --build app

# 停止所有服务
docker compose down

# 停止并删除卷数据
docker compose down -v

# 查看运行状态
docker compose ps

# 进入 app 容器
docker compose exec app sh
```

## 4. 构建和推送镜像

```bash
# 构建镜像
docker build -t blockchain-service:latest .

# 查看镜像大小
docker images blockchain-service

# 打标签（准备推送）
docker tag blockchain-service:latest your-registry.com/blockchain-service:v1.0.0
docker tag blockchain-service:latest your-registry.com/blockchain-service:latest

# 推送到镜像仓库
docker push your-registry.com/blockchain-service:v1.0.0
docker push your-registry.com/blockchain-service:latest

# 运行容器
docker run -d \
  --name blockchain-service \
  -p 8080:8080 \
  -e BC_SERVER_PORT=8080 \
  -e BC_ETHEREUM_RPC_URL=https://eth.llamarpc.com \
  blockchain-service:latest
```

## 5. GitHub Actions CI/CD

### .github/workflows/ci.yml

```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  # ============================================
  # 任务1: 代码检查和测试
  # ============================================
  lint-and-test:
    name: Lint & Test
    runs-on: ubuntu-latest

    services:
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379

    steps:
      - name: Checkout 代码
        uses: actions/checkout@v4

      - name: 设置 Go 环境
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: 代码格式检查
        run: |
          go install golang.org/x/tools/cmd/goimports@latest
          test -z "$(goimports -l .)"

      - name: 静态分析
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m

      - name: 运行测试
        run: go test -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: 上传覆盖率
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: unittests

      - name: 检查测试覆盖率
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "代码覆盖率: ${COVERAGE}%"
          if (( $(echo "$COVERAGE < 70" | bc -l) )); then
            echo "⚠️ 警告: 覆盖率低于 70%"
          fi

  # ============================================
  # 任务2: 构建 Docker 镜像
  # ============================================
  build:
    name: Build Docker Image
    needs: lint-and-test
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest

    steps:
      - name: Checkout 代码
        uses: actions/checkout@v4

      - name: 设置 Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: 登录 GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: 提取元数据
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ghcr.io/${{ github.repository }}
          tags: |
            type=sha,prefix=
            type=ref,event=branch
            type=semver,pattern={{version}}
            latest

      - name: 构建并推送
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

### .github/workflows/deploy.yml

```yaml
name: Deploy

on:
  workflow_dispatch:  # 手动触发
  push:
    tags:
      - 'v*'         # v1.0.0 等标签推送时自动部署

jobs:
  deploy:
    name: Deploy to Production
    runs-on: ubuntu-latest
    environment: production

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: 部署到服务器
        uses: appleboy/ssh-action@v1.0.0
        with:
          host: ${{ secrets.DEPLOY_HOST }}
          username: ${{ secrets.DEPLOY_USER }}
          key: ${{ secrets.DEPLOY_SSH_KEY }}
          script: |
            cd /opt/blockchain-service

            # 拉取最新代码
            git pull origin main

            # 拉取最新镜像
            docker compose pull

            # 重启服务
            docker compose up -d --remove-orphans

            # 清理旧镜像
            docker image prune -f

            echo "✅ 部署完成"
```

## 6. 生产环境部署清单

```yaml
# docker-compose.prod.yml（生产环境配置）
version: '3.8'

services:
  app:
    image: ghcr.io/your-org/blockchain-service:latest
    restart: always
    ports:
      - "8080:8080"
    environment:
      - BC_SERVER_MODE=release
      - BC_DATABASE_DRIVER=postgres
      - BC_DATABASE_DSN=postgres://${DB_USER}:${DB_PASS}@db:5432/blockchain?sslmode=disable
      - BC_REDIS_ADDR=redis:6379
      - BC_REDIS_PASSWORD=${REDIS_PASS}
      - BC_JWT_SECRET=${JWT_SECRET}
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
    networks:
      - blockchain-net

  db:
    image: postgres:16-alpine
    restart: always
    volumes:
      - db_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_USER=${DB_USER}
      - POSTGRES_PASSWORD=${DB_PASS}
      - POSTGRES_DB=blockchain
    networks:
      - blockchain-net

  redis:
    image: redis:7-alpine
    restart: always
    command: redis-server --appendonly yes --requirepass ${REDIS_PASS}
    volumes:
      - redis_data:/data
    networks:
      - blockchain-net

  # Nginx 反向代理（可选）
  nginx:
    image: nginx:alpine
    restart: always
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
    depends_on:
      - app
    networks:
      - blockchain-net

volumes:
  db_data:
  redis_data:

networks:
  blockchain-net:
    driver: bridge
```

## 7. CI/CD 完整流程

```
代码推送 (git push)
    │
    ▼
┌─────────────────────────────────────┐
│  CI Pipeline (.github/workflows/ci.yml) │
├─────────────────────────────────────┤
│  1. Checkout 代码                      │
│  2. 设置 Go 环境                       │
│  3. go vet 静态分析                    │
│  4. golangci-lint 代码检查             │
│  5. go test -race 运行测试             │
│  6. 上传覆盖率报告                     │
│  7. 构建 Docker 镜像                   │
│  8. 推送到镜像仓库                     │
└─────────────────────────────────────┘
    │
    ▼ (tag 推送时触发)
┌─────────────────────────────────────┐
│  CD Pipeline (deploy.yml)              │
├─────────────────────────────────────┤
│  1. SSH 连接到生产服务器               │
│  2. 拉取最新镜像                       │
│  3. docker compose up -d              │
│  4. 健康检查                           │
│  5. 清理旧镜像                         │
└─────────────────────────────────────┘
```

## 📝 作业

### 作业1：Docker 化你的区块链服务

```bash
# TODO: 完成以下任务
# 1. 为 day20-21 的区块链服务编写 Dockerfile
# 2. 实现多阶段构建
# 3. 镜像大小控制在 50MB 以内
# 4. 编写 .dockerignore 文件
# 5. 验证 docker build 成功
```

### 作业2：编写 Docker Compose

```yaml
# TODO: 编写 docker-compose.yml
# 1. 包含 app + redis + postgres 三个服务
# 2. 配置健康检查
# 3. 配置环境变量
# 4. 配置网络和数据卷
# 5. 验证 docker compose up 成功
```

### 作业3：配置 GitHub Actions

```yaml
# TODO: 编写 .github/workflows/ci.yml
# 1. 包含 lint 检查
# 2. 包含单元测试
# 3. 包含 Docker 镜像构建
# 4. 测试覆盖率门槛 70%
# 5. 在 PR 时自动运行
```

## 🎯 检查点

- ✅ 掌握 Go 项目 Dockerfile 编写
- ✅ 理解多阶段构建
- ✅ 能够使用 Docker Compose 编排服务
- ✅ 掌握 GitHub Actions CI/CD
- ✅ 了解生产环境部署最佳实践

## ⏭️ 下一课

[第24课：Go 面试高频题](./day24-interview.md)
