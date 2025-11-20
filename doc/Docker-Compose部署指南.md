# AI股票分析系统 - Docker Compose 部署指南

## 📋 目录

- [一、简介](#一简介)
- [二、前置要求](#二前置要求)
- [三、快速部署](#三快速部署)
- [四、文件说明](#四文件说明)
- [五、配置说明](#五配置说明)
- [六、启动与停止](#六启动与停止)
- [七、访问系统](#七访问系统)
- [八、常见问题](#八常见问题)

---

## 一、简介

本指南介绍如何使用 Docker Compose 部署 AI股票分析系统。系统使用远程 Docker 镜像（GitHub Container Registry），无需本地编译，只需下载必要的配置文件即可快速部署。

### 部署方式对比

| 方式 | 说明 | 优点 | 缺点 |
|------|------|------|------|
| **Git Clone** | 克隆整个项目仓库 | 包含完整源代码和文档 | 文件较多，下载时间长 |
| **Docker Compose** | 仅下载配置文件 | 快速部署，文件最少 | 不支持本地修改源代码 |

---

## 二、前置要求

### 必需软件

1. **Docker** (版本 20.10+)
   ```bash
   docker --version
   ```

2. **Docker Compose** (版本 2.0+ 或 1.29+)
   ```bash
   docker-compose --version
   # 或
   docker compose version
   ```

### 系统要求

- **操作系统**: Linux / macOS / Windows (WSL2)
- **内存**: 最低 512MB，推荐 1GB+
- **磁盘**: 最低 500MB 可用空间
- **网络**: 能够访问 GitHub Container Registry (ghcr.io)

---

## 三、快速部署

### 方式一：一键下载脚本（推荐）

#### Linux/macOS

```bash
# 下载并执行一键部署脚本
wget https://raw.githubusercontent.com/maxage/ai-ding-stock/main/download-deploy.sh -O download-deploy.sh
chmod +x download-deploy.sh
./download-deploy.sh
```

#### Windows (Git Bash / WSL)

```bash
# 下载脚本
curl -O https://raw.githubusercontent.com/maxage/ai-ding-stock/main/download-deploy.sh

# 执行脚本（需要 Git Bash 或 WSL）
bash download-deploy.sh
```

脚本会自动：
- ✅ 创建 `Ai-Ding-Stock` 目录
- ✅ 创建 `web` 子目录
- ✅ 下载所有必需文件
- ✅ 显示下一步操作说明

### 方式二：手动下载（如果脚本无法使用）

#### 1. 创建目录结构

```bash
# 创建部署目录
mkdir -p Ai-Ding-Stock/web
cd Ai-Ding-Stock
```

#### 2. 下载必需文件

```bash
# 仓库基础 URL
BASE_URL="https://raw.githubusercontent.com/maxage/ai-ding-stock/main"

# 下载 docker-compose.yml
wget ${BASE_URL}/docker-compose.yml -O docker-compose.yml

# 下载 nginx.conf
wget ${BASE_URL}/nginx.conf -O nginx.conf

# 下载配置文件（从示例文件复制）
wget ${BASE_URL}/config_stock.json.example -O config_stock.json

# 下载前端页面
wget ${BASE_URL}/web/config.html -O web/config.html
```

#### Windows PowerShell

```powershell
# 创建目录
New-Item -ItemType Directory -Force -Path "Ai-Ding-Stock\web"
cd Ai-Ding-Stock

# 仓库基础 URL
$BASE_URL = "https://raw.githubusercontent.com/maxage/ai-ding-stock/main"

# 下载文件
Invoke-WebRequest -Uri "$BASE_URL/docker-compose.yml" -OutFile "docker-compose.yml"
Invoke-WebRequest -Uri "$BASE_URL/nginx.conf" -OutFile "nginx.conf"
Invoke-WebRequest -Uri "$BASE_URL/config_stock.json.example" -OutFile "config_stock.json"
Invoke-WebRequest -Uri "$BASE_URL/web/config.html" -OutFile "web\config.html"
```

### 3. 验证文件结构

部署完成后，目录结构应该如下：

```
Ai-Ding-Stock/
├── docker-compose.yml      # Docker Compose 配置文件
├── config_stock.json       # 系统配置文件（需要编辑）
├── nginx.conf              # Nginx 配置文件
└── web/                    # Web 前端目录
    └── config.html         # 前端页面文件
```

---

## 四、文件说明

### 必需文件清单

| 文件名 | 说明 | 是否必须 |
|--------|------|----------|
| `docker-compose.yml` | Docker Compose 编排配置 | ✅ 必需 |
| `config_stock.json` | 系统配置文件 | ✅ 必需（需编辑） |
| `nginx.conf` | Nginx 反向代理配置 | ✅ 必需 |
| `web/config.html` | Web 前端界面 | ✅ 必需 |

### 自动生成的文件

| 目录/文件 | 说明 | 生成时机 |
|-----------|------|----------|
| `stock_analysis_logs/` | 日志目录 | 首次启动时自动创建 |

---

## 五、配置说明

### 1. 编辑配置文件

```bash
# Linux/macOS
vim config_stock.json
# 或
nano config_stock.json

# Windows
notepad config_stock.json
```

### 2. 必需配置项

#### 2.1 TDX API 地址

```json
{
  "tdx_api_url": "http://192.168.1.222:8181"
}
```

> **说明**: 替换为您的 TDX API 服务地址。如果 TDX API 在同一台服务器上，使用 `http://host.docker.internal:8181` 或 `http://172.17.0.1:8181`。

#### 2.2 AI 配置

**DeepSeek (推荐)**

```json
{
  "ai_config": {
    "provider": "deepseek",
    "deepseek_key": "sk-your-deepseek-api-key",
    "qwen_key": "",
    "custom_api_url": "",
    "custom_api_key": "",
    "custom_model_name": ""
  }
}
```

**通义千问 (Qwen)**

```json
{
  "ai_config": {
    "provider": "qwen",
    "deepseek_key": "",
    "qwen_key": "your-qwen-api-key",
    "custom_api_url": "",
    "custom_api_key": "",
    "custom_model_name": ""
  }
}
```

**自定义 API**

```json
{
  "ai_config": {
    "provider": "custom",
    "deepseek_key": "",
    "qwen_key": "",
    "custom_api_url": "https://your-api-endpoint.com/v1/chat/completions",
    "custom_api_key": "your-api-key",
    "custom_model_name": "your-model-name"
  }
}
```

#### 2.3 股票监控配置

```json
{
  "stocks": [
    {
      "code": "000001",
      "name": "平安银行",
      "enabled": true,
      "scan_interval_minutes": 5,
      "min_confidence": 70,
      "position_quantity": 0,
      "buy_price": 0,
      "buy_date": ""
    }
  ]
}
```

**字段说明**:
- `code`: 股票代码（6位数字，如 "000001"）
- `name`: 股票名称（仅用于显示）
- `enabled`: 是否启用监控（true/false）
- `scan_interval_minutes`: 扫描间隔（分钟）
- `min_confidence`: 最小信心度阈值（0-100，达到此值才发送通知）
- `position_quantity`: 持仓数量（可选，持仓模式）
- `buy_price`: 买入价格（可选，持仓模式）
- `buy_date`: 买入日期（可选，格式：YYYY-MM-DD）

#### 2.4 通知配置（可选）

**钉钉通知**

```json
{
  "notification": {
    "enabled": true,
    "dingtalk": {
      "enabled": true,
      "webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN",
      "secret": "YOUR_SECRET"
    },
    "feishu": {
      "enabled": false,
      "webhook_url": "",
      "secret": ""
    }
  }
}
```

**飞书通知**

```json
{
  "notification": {
    "enabled": true,
    "dingtalk": {
      "enabled": false,
      "webhook_url": "",
      "secret": ""
    },
    "feishu": {
      "enabled": true,
      "webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/YOUR_TOKEN",
      "secret": "YOUR_SECRET"
    }
  }
}
```

#### 2.5 其他配置

```json
{
  "api_server_port": 9090,
  "log_dir": "stock_analysis_logs",
  "api_token": "1122334455667788",
  "analysis_history_limit": 20,
  "trading_time": {
    "enable_check": true,
    "trading_hours": ["09:30-11:30", "13:00-15:00"],
    "timezone": "Asia/Shanghai"
  }
}
```

**字段说明**:
- `api_server_port`: 后端 API 端口（容器内端口，无需修改）
- `log_dir`: 日志目录（相对路径）
- `api_token`: API 认证 Token（用于前端重启后端等功能，建议修改为强密码）
- `analysis_history_limit`: 分析历史记录数量（3-100，默认 20）
- `trading_time`: 交易时间检查配置

### 3. 配置文件验证

系统启动时会自动验证配置文件。常见错误：

- ❌ `tdx_api_url不能为空`
- ❌ `至少需要配置一只股票`
- ❌ `ai_config.provider 必须是 deepseek、qwen 或 custom`

---

## 六、启动与停止

### 1. 拉取远程镜像

```bash
docker-compose pull
```

> **说明**: 首次部署或更新镜像时执行。系统会自动从 GitHub Container Registry 拉取最新镜像。

### 2. 启动服务

#### 后台启动（推荐）

```bash
docker-compose up -d
```

#### 前台启动（查看实时日志）

```bash
docker-compose up
```

按 `Ctrl+C` 停止服务。

### 3. 查看服务状态

```bash
docker-compose ps
```

**正常状态示例**:

```
NAME              IMAGE                               STATUS
stock-analyzer    ghcr.io/maxage/ai-ding-stock:latest   Up 2 minutes (healthy)
stock-web         nginx:1.25-alpine                    Up 2 minutes
```

### 4. 查看日志

#### 查看所有服务日志

```bash
docker-compose logs -f
```

#### 查看后端服务日志

```bash
docker-compose logs -f stock-analyzer
```

#### 查看 Nginx 服务日志

```bash
docker-compose logs -f stock-web
```

### 5. 停止服务

```bash
docker-compose down
```

**停止并删除数据卷**:

```bash
docker-compose down -v
```

> ⚠️ **警告**: 使用 `-v` 参数会删除日志目录，请谨慎使用。

### 6. 重启服务

```bash
docker-compose restart
```

或

```bash
docker-compose down
docker-compose up -d
```

### 7. 更新服务

```bash
# 1. 停止服务
docker-compose down

# 2. 拉取最新镜像
docker-compose pull

# 3. 启动服务
docker-compose up -d
```

---

## 七、访问系统

### 1. Web 界面

**访问地址**: http://localhost:53280

> **说明**: 通过 Nginx 反向代理访问，前端页面和 API 请求都会自动代理到后端服务。

### 2. API 接口

**直接访问后端**: http://localhost:53290

**通过前端访问**: http://localhost:53280/api/*

### 3. 健康检查

```bash
# 检查后端服务
curl http://localhost:53290/health

# 检查 Nginx 服务
curl http://localhost:53280/health
```

**正常响应**:

```json
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

---

## 八、常见问题

### 1. 端口冲突

**问题**: 端口 53280 或 53290 已被占用

**解决**: 修改 `docker-compose.yml` 中的端口映射

```yaml
ports:
  - "53281:9090"  # 修改主机端口
  - "53281:80"    # 修改主机端口
```

### 2. 无法连接 TDX API

**问题**: 后端日志显示 `连接 TDX API 失败`

**原因**: 
- TDX API 服务未启动
- 网络无法访问 TDX API 地址
- 防火墙阻止连接

**解决**:
1. 确认 TDX API 服务正在运行
2. 如果 TDX API 在同一台服务器上，使用以下地址之一：
   - `http://host.docker.internal:8181` (macOS/Windows)
   - `http://172.17.0.1:8181` (Linux，Docker 默认网关)
   - 使用宿主机的实际 IP 地址

### 3. 前端页面空白

**问题**: 访问 http://localhost:53280 显示空白页面

**原因**: 
- `web/config.html` 文件不存在或损坏
- Nginx 配置错误

**解决**:
1. 检查 `web/config.html` 文件是否存在
2. 查看 Nginx 容器日志: `docker-compose logs stock-web`
3. 重新下载 `web/config.html` 文件

### 4. 配置文件保存失败

**问题**: 在 Web 界面修改配置后保存失败

**原因**: 
- 文件权限问题
- 磁盘空间不足
- 配置文件格式错误

**解决**:
1. 检查 `config_stock.json` 文件权限
2. 检查磁盘空间: `df -h`
3. 验证 JSON 格式: `cat config_stock.json | jq .`

### 5. 镜像拉取失败

**问题**: `docker-compose pull` 失败，提示无法访问 ghcr.io

**原因**: 
- 网络无法访问 GitHub Container Registry
- 需要登录 GitHub

**解决**:
1. 检查网络连接
2. 配置 Docker 镜像代理（如需要）
3. 如果使用私有镜像，需要登录:
   ```bash
   echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
   ```

### 6. 日志文件过大

**问题**: 日志文件占用大量磁盘空间

**解决**:
1. 配置日志轮转（已在 `docker-compose.yml` 中配置）
2. 手动清理旧日志:
   ```bash
   docker-compose down
   rm -rf stock_analysis_logs/*.log
   docker-compose up -d
   ```

### 7. 服务启动失败

**问题**: `docker-compose up` 后服务立即退出

**解决**:
1. 查看详细日志: `docker-compose logs`
2. 检查配置文件格式: `cat config_stock.json | jq .`
3. 检查容器状态: `docker-compose ps -a`

---

## 九、高级配置

### 1. 环境变量配置

在 `docker-compose.yml` 中可以通过环境变量覆盖配置:

```yaml
environment:
  - API_TOKEN=your-custom-token
  - LOG_LEVEL=debug
  - TZ=Asia/Shanghai
```

### 2. 自定义日志目录

修改 `docker-compose.yml` 中的卷挂载:

```yaml
volumes:
  - /path/to/custom/logs:/app/stock_analysis_logs
```

### 3. 数据持久化

所有重要数据会自动持久化到宿主机:

- **配置文件**: `./config_stock.json` → `/app/config_stock.json`
- **日志文件**: `./stock_analysis_logs` → `/app/stock_analysis_logs`

> **提示**: 定期备份 `config_stock.json` 文件，防止配置丢失。

---

## 十、技术支持

### 问题反馈

如遇到问题，请检查：
1. ✅ Docker 和 Docker Compose 版本是否符合要求
2. ✅ 配置文件格式是否正确
3. ✅ 网络连接是否正常
4. ✅ 日志文件中的错误信息

### 相关文档

- [API 接口文档](./API_接口文档.md)
- [项目 README](../README.md)

---

## 📝 更新日志

- **2024-01-XX**: 首次发布 Docker Compose 部署指南

---

**祝您使用愉快！** 🎉

