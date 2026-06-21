# 云值守系统 - 部署指南

## 概述

本系统采用 Docker 容器化部署，包含以下服务：

| 服务 | 镜像 | 说明 |
|------|------|------|
| `yzs-backend` | 自构建 | Go 后端 API，端口 18881 |
| `yzs-frontend` | 自构建 | Vue3 前端，端口 18880 |
| `yzs-nginx` | nginx:alpine | 反向代理，对外暴露 80/443 |
| `yzs-mysql` | mysql:8.1.0 | 数据库，自动建库 |
| `yzs-redis` | redis:7.2 | Session 缓存，限制 256mb |

---

## 一、本地构建（开发环境 WSL）

### 前置条件

- WSL2（Ubuntu）已安装并配置 Docker
- 项目根目录存在 `.env` 文件（见下方环境变量说明）

### 构建命令

```bash
cd /path/to/go-yzs
./build.sh
```

构建脚本会自动完成：

1. 构建后端镜像 `yzs-backend:latest` → 导出 `dist/yzs-backend.tar.gz`
2. 构建前端镜像 `yzs-frontend:latest` → 导出 `dist/yzs-frontend.tar.gz`
3. 复制 `nginx.conf`、`.env` 到 `dist/`
4. 生成服务器端加载脚本 `dist/load.sh`
5. 打包为 `yzs-deploy-<时间戳>.tar.gz`

构建完成后输出类似：

```
dist/
├── yzs-backend.tar.gz
├── yzs-frontend.tar.gz
├── nginx.conf
├── docker-compose.yml
├── .env
└── load.sh
yzs-deploy-20260412120000.tar.gz
```

---

## 二、服务器部署

### 前置要求

- Ubuntu 服务器，已安装 Docker 和 Docker Compose
- 域名 SSL 证书已准备（`.pem` + `.key`）

### 步骤

#### 1. 上传构建包

```bash
scp yzs-deploy-<版本号>.tar.gz root@<服务器IP>:/opt/yzs/
```

#### 2. 解压

```bash
mkdir -p /opt/yzs
cd /opt/yzs
tar -xzf yzs-deploy-<版本号>.tar.gz
```

#### 3. 上传 SSL 证书

```bash
# 本地执行
scp www.yzs88.com.pem root@<服务器IP>:/etc/nginx/ssl/
scp www.yzs88.com.key root@<服务器IP>:/etc/nginx/ssl/
```

#### 4. 加载镜像并启动

```bash
cd /opt/yzs
chmod +x load.sh
./load.sh   # 需要 root 权限
```

`load.sh` 会自动完成：加载镜像 → 停止旧容器 → 启动所有服务。

#### 5. 验证

```bash
docker compose ps
# 期望所有服务状态均为 Up
```

---

## 三、环境变量（.env）

在项目根目录创建 `.env` 文件，内容如下：

```env
# 数据库
DB_HOST=yzs-mysql
DB_PORT=3306
DB_USER=root
DB_PASSWORD=你的数据库密码
DB_NAME=go_yzs

# Redis（Docker 内通过容器名访问，无需修改）
REDIS_HOST=yzs-redis
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT 密钥（生产环境请修改为随机强密钥）
JWT_SECRET=your-secret-key

# 后端端口
SERVER_PORT=18881

# 审核模式：true=开启质检审核，false=直通模式
REVIEW_ENABLED=true
```

> **注意**：`.env` 文件包含敏感信息，不要提交到 Git。

---

## 四、端口说明

| 端口 | 服务 | 说明 |
|------|------|------|
| 80 | Nginx | HTTP，自动重定向到 HTTPS |
| 443 | Nginx | HTTPS 入口 |
| 18880 | Frontend | 前端内部端口（容器间通信，不对外暴露） |
| 18881 | Backend | 后端 API 端口（容器间通信，不对外暴露） |

---

## 五、Nginx 域名路由

| 域名 | 说明 |
|------|------|
| `www.yzs88.com` | 云值守主系统 |
| `yzs88.com` | 重定向到 `www.yzs88.com` |

SSL 证书路径（服务器上）：

```
/etc/nginx/ssl/
├── www.yzs88.com.pem
├── www.yzs88.com.key
```

---

## 六、常用运维命令

```bash
# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f backend
docker compose logs -f frontend
docker compose logs -f nginx

# 重启单个服务
docker compose restart backend

# 停止所有服务
docker compose down

# 强制重建（不删数据卷）
docker compose up -d --force-recreate
```

---

## 七、数据持久化

MySQL 和 Redis 数据通过 Docker Volume 持久化，不会因容器重启丢失：

```bash
# 查看数据卷
docker volume ls | grep yzs

# 备份 MySQL
docker exec yzs-mysql mysqldump -uroot -p<密码> go_yzs > backup.sql

# 恢复 MySQL
docker exec -i yzs-mysql mysql -uroot -p<密码> go_yzs < backup.sql
```
