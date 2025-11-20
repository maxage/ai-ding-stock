# 📈 AI股票分析系统

> 基于DeepSeek/Qwen大模型的智能股票分析系统，实时监控、AI分析、自动通知

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)

---

## 🌟 项目特点

- 🤖 **AI驱动分析** - 使用DeepSeek/Qwen大模型进行深度技术分析
- 📊 **实时监控** - 自动监控多只股票，定时分析
- 🎯 **智能信号** - 提供BUY/SELL/HOLD明确信号和目标价
- 📱 **即时通知** - 支持钉钉、飞书Webhook推送
- 🌐 **Web界面** - 实时查看分析结果和历史记录
- 🔌 **RESTful API** - 完整的API接口，易于集成
- 🐳 **容器化部署** - Docker一键部署，开箱即用
- 📈 **技术指标** - 支持MA、RSI、波动率等多种技术指标
- 💼 **持仓模式** - 支持监控模式和持仓模式（含止盈止损）

---

## 📋 目录

- [系统预览](#系统预览)
- [功能概述](#功能概述)
- [快速开始](#快速开始)
  - [模式一：开发者模式（本地运行）](#模式一开发者模式本地运行)
  - [模式二：Docker Compose部署（推荐）](#模式二docker-compose部署推荐)
- [配置说明](#配置说明)
- [使用指南](#使用指南)
- [API文档](#api文档)
- [通知配置](#通知配置)
- [常见问题](#常见问题)
- [项目结构](#项目结构)
- [文档索引](#文档索引)

---

## 📸 系统预览

### 系统首页
![系统首页](image/首页.png)

### 系统测试
![系统测试](image/系统测试.png)

---

## 🎯 功能概述

### 核心功能

#### 1. 实时股票监控
- ✅ 支持多只股票同时监控
- ✅ 可配置扫描间隔（1-60分钟）
- ✅ 自动获取实时行情和K线数据
- ✅ 24/7不间断运行

#### 2. AI深度分析
- ✅ 基于大语言模型的技术分析
- ✅ 综合考虑趋势、量价、盘口等多维度
- ✅ 给出BUY/SELL/HOLD明确信号
- ✅ 提供信心度评分（0-100）
- ✅ 给出目标价和止损价建议

#### 3. 技术指标计算
- ✅ **均线系统**: MA5、MA10、MA20、MA60
- ✅ **相对强弱**: RSI(14)
- ✅ **波动率**: 20日标准差
- ✅ **量价分析**: 成交量、成交额、内外盘比
- ✅ **盘口分析**: 买卖五档、委比

#### 4. 智能通知
- ✅ 钉钉机器人推送
- ✅ 飞书机器人推送
- ✅ 可配置信心度阈值
- ✅ 支持持仓模式（显示盈亏、止盈止损价）

#### 5. Web监控界面
- ✅ 实时显示分析结果
- ✅ 股票列表和状态
- ✅ AI分析详情
- ✅ 历史信号记录
- ✅ 响应式设计
- ✅ 系统配置管理
- ✅ 前端可重启后端（需Token认证）

#### 6. RESTful API
- ✅ 获取所有股票状态
- ✅ 查询单个股票分析
- ✅ 健康检查端点
- ✅ 标准JSON响应

---

## 🚀 快速开始

### 前置要求

#### 模式一：开发者模式（本地运行）
- **Go 1.24+** 环境
- **TDX股票数据API服务**（项目地址：https://github.com/oficcejo/tdx-api）
- **AI API密钥**（DeepSeek/Qwen/自定义）

#### 模式二：Docker Compose部署
- **Docker 20.10+** 和 **Docker Compose 2.0+**
- **TDX股票数据API服务**
- **AI API密钥**

---

## 模式一：开发者模式（本地运行）

适合开发者进行二次开发、调试和测试的场景。

### 1️⃣ 克隆或下载项目

```bash
# 如果你是从git仓库clone
git clone <your-repo-url>
cd ai-ding-stock

# 或者直接解压下载的zip包
unzip ai-ding-stock.zip
cd ai-ding-stock
```

### 2️⃣ 安装Go依赖

```bash
# 使用国内代理加速（可选）
go env -w GOPROXY=https://goproxy.cn,direct

# 下载依赖
go mod download
```

### 3️⃣ 配置系统

```bash
# 复制配置示例
cp config_stock.json.example config_stock.json

# 编辑配置文件（或使用Web界面配置）
vim config_stock.json
# Windows: notepad config_stock.json
```

**最小配置示例**：

```json
{
  "tdx_api_url": "http://your-tdx-api:8181",
  "ai_config": {
    "provider": "deepseek",
    "deepseek_key": "sk-your-deepseek-api-key"
  },
  "stocks": [
    {
      "code": "000001",
      "name": "平安银行",
      "enabled": true,
      "scan_interval_minutes": 5,
      "min_confidence": 70
    }
  ],
  "notification": {
    "enabled": false
  },
  "api_server_port": 9090,
  "log_dir": "stock_analysis_logs"
}
```

### 4️⃣ 编译项目

```bash
# 编译
go build -o stock_analyzer main_stock.go

# Windows
go build -o stock_analyzer.exe main_stock.go
```

### 5️⃣ 运行程序

#### Linux/macOS

```bash
# 直接运行
./stock_analyzer config_stock.json

# 或使用管理脚本（推荐）
chmod +x manage_backend.sh
./manage_backend.sh start
```

#### Windows

```bash
# 直接运行
stock_analyzer.exe config_stock.json

# 或使用批处理脚本
stock-manager.bat start
```

### 6️⃣ 访问系统

- **Web配置界面**: http://localhost:9090
- **API接口**: http://localhost:9090/api/stocks
- **健康检查**: http://localhost:9090/health

### 7️⃣ 管理服务

使用管理脚本：

```bash
# 启动
./manage_backend.sh start

# 停止
./manage_backend.sh stop

# 重启
./manage_backend.sh restart

# 查看状态
./manage_backend.sh status

# 查看日志（最近50行）
./manage_backend.sh logs

# 实时查看日志
./manage_backend.sh tail
```

**注意**：日志文件默认保存在 `stock_analysis_logs/stock_analyzer.log`（可在配置文件中修改 `log_dir`）。

---

## 模式二：Docker Compose部署（推荐）

适合生产环境部署，一键启动，易于管理。

### 1️⃣ 准备项目文件

```bash
# 克隆或解压项目
cd ai-ding-stock
```

### 2️⃣ 配置系统

#### 方式A：Git Clone（开发者模式）

```bash
# 复制配置示例
cp config_stock.json.example config_stock.json

# 编辑配置文件
vim config_stock.json
```

#### 方式B：仅下载部署文件（Docker Compose 模式）

**一键下载部署文件（推荐）**:

```bash
# Linux/macOS - 一键下载所有必需文件
bash <(curl -sL https://raw.githubusercontent.com/maxage/ai-ding-stock/main/download-deploy.sh)
```

**或者手动下载**:

```bash
# 创建目录
mkdir -p Ai-Ding-Stock/web && cd Ai-Ding-Stock

# 仓库基础 URL
BASE_URL="https://raw.githubusercontent.com/maxage/ai-ding-stock/main"

# 下载必需文件
wget ${BASE_URL}/docker-compose.yml -O docker-compose.yml
wget ${BASE_URL}/nginx.conf -O nginx.conf
wget ${BASE_URL}/config_stock.json.example -O config_stock.json
wget ${BASE_URL}/web/config.html -O web/config.html
```

然后编辑 `config_stock.json` 填写您的配置。

**配置说明**：
- 必需配置：`tdx_api_url`、`ai_config`（AI密钥）、`stocks`（股票代码列表）
- 可选配置：`notification`（通知配置）
- 也可以在部署完成后通过Web界面进行配置

### 3️⃣ 启动服务

#### 方式A：使用docker-compose（推荐）

```bash
# 拉取远程镜像（使用远程镜像时）
docker-compose pull

# 启动服务（后台运行）
docker-compose up -d

# 查看日志
docker-compose logs -f stock-analyzer

# 查看状态
docker-compose ps
```

#### 方式B：使用启动脚本

**Linux/macOS**：

```bash
chmod +x docker-start.sh
./docker-start.sh start
```

**Windows**：

```cmd
docker-start.bat start
```

### 4️⃣ 访问系统

- **Web配置界面**: http://localhost:53280（通过Nginx）
- **API接口**: http://localhost:53290/api/stocks（直接访问后端）
- **健康检查**: http://localhost:53290/health（直接访问后端）

### 5️⃣ 管理服务

```bash
# 查看状态
docker-compose ps

# 查看日志（实时）
docker-compose logs -f stock-analyzer

# 查看日志（最近100行）
docker-compose logs --tail=100 stock-analyzer

# 重启服务
docker-compose restart stock-analyzer

# 停止服务
docker-compose down

# 停止并删除数据卷
docker-compose down -v
```

**或使用启动脚本**：

```bash
# Linux/macOS
./docker-start.sh start    # 启动
./docker-start.sh stop     # 停止
./docker-start.sh restart  # 重启
./docker-start.sh status   # 状态
./docker-start.sh logs     # 日志

# Windows
docker-start.bat start
docker-start.bat stop
docker-start.bat restart
```

### 6️⃣ 持久化数据

Docker部署会自动挂载以下目录：

- **配置文件**: `./config_stock.json` → `/app/config_stock.json`
- **日志目录**: `./stock_analysis_logs` → `/app/stock_analysis_logs`
- **Web前端**: `./web` → `/app/web`

**重要**：修改配置文件后，需要重启容器才能生效：

```bash
docker-compose restart stock-analyzer
```

---

## ⚙️ 配置说明

### 配置文件结构

详细配置说明请参考：[doc/API_接口文档.md](doc/API_接口文档.md)

**完整配置示例**：

```json
{
  "tdx_api_url": "http://192.168.1.222:8181",
  "ai_config": {
    "provider": "deepseek",
    "deepseek_key": "sk-your-api-key",
    "qwen_key": "",
    "custom_api_url": "https://api.siliconflow.cn/v1",
    "custom_api_key": "sk-your-custom-key",
    "custom_model_name": "deepseek-ai/DeepSeek-V3"
  },
  "stocks": [
    {
      "code": "000001",
      "name": "平安银行",
      "enabled": true,
      "scan_interval_minutes": 5,
      "min_confidence": 70,
      "position_quantity": 1000,
      "buy_price": 12.50,
      "buy_date": "2025-01-20"
    }
  ],
  "notification": {
    "enabled": true,
    "dingtalk": {
      "enabled": true,
      "webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN",
      "secret": "股票通知"
    },
    "feishu": {
      "enabled": false,
      "webhook_url": "",
      "secret": ""
    }
  },
  "trading_time": {
    "enable_check": false,
    "trading_hours": ["09:30-11:30", "13:00-15:00"],
    "timezone": "Asia/Shanghai"
  },
  "api_server_port": 9090,
  "log_dir": "stock_analysis_logs",
  "api_token": "1122334455667788",
  "analysis_history_limit": 20
}
```

### 关键配置项

#### TDX API配置
- `tdx_api_url`: TDX股票数据API的基础URL（必需）

#### AI配置
- `provider`: AI提供商，支持 `deepseek` / `qwen` / `custom`
- `deepseek_key`: DeepSeek API密钥
- `qwen_key`: 通义千问API密钥
- `custom_api_url`: 自定义OpenAI兼容API地址
- `custom_api_key`: 自定义API密钥
- `custom_model_name`: 自定义模型名称

#### 股票配置
- `code`: 股票代码（如：000001）
- `name`: 股票名称（如：平安银行）
- `enabled`: 是否启用监控
- `scan_interval_minutes`: 扫描间隔（分钟），建议5-60
- `min_confidence`: 最小信心度阈值（0-100）
- `position_quantity`: 持仓数量（股），0或不填表示监控模式
- `buy_price`: 购买价格（元/股），与持仓数量配合使用
- `buy_date`: 购买日期（格式：YYYY-MM-DD），可选

#### 通知配置
- `enabled`: 是否启用通知
- `dingtalk.webhook_url`: 钉钉机器人Webhook地址
- `dingtalk.secret`: 钉钉机器人关键词（用于安全验证）
- `feishu.webhook_url`: 飞书机器人Webhook地址
- `feishu.secret`: 飞书签名密钥

#### 系统配置
- `api_server_port`: API服务器端口（默认9090）
- `log_dir`: 日志目录（默认：stock_analysis_logs）
- `api_token`: API认证Token（用于前端重启后端等功能，默认：1122334455667788，建议修改）
- `analysis_history_limit`: 分析历史记录数量（3-100，默认20）

---

## 📖 使用指南

### 两种监控模式

#### 模式A：监控模式（默认）
不填写 `position_quantity` 和 `buy_price`，系统仅进行技术分析，提供买卖信号。

```json
{
  "code": "000001",
  "name": "平安银行",
  "enabled": true,
  "scan_interval_minutes": 5,
  "min_confidence": 70
}
```

#### 模式B：持仓模式
填写 `position_quantity` 和 `buy_price`，系统会：
- 计算浮动盈亏
- 基于持仓成本和技术分析给出止盈止损价格建议
- 在通知中显示持仓信息

```json
{
  "code": "000001",
  "name": "平安银行",
  "enabled": true,
  "scan_interval_minutes": 5,
  "min_confidence": 70,
  "position_quantity": 1000,
  "buy_price": 12.50,
  "buy_date": "2025-01-20"
}
```

### 添加监控股票

1. 编辑 `config_stock.json` 或通过Web界面
2. 在 `stocks` 数组中添加新股票
3. 重启服务

### 调整分析频率

修改 `scan_interval_minutes`（单位：分钟）

### 调整通知阈值

修改 `min_confidence`（0-100，数值越高要求越严格）

### Web界面管理

访问 http://localhost:9090，可以：

- ✅ 查看系统状态
- ✅ 添加/删除/修改股票
- ✅ 配置通知设置
- ✅ 查看最近分析记录
- ✅ 手动触发分析测试
- ✅ 重启后端服务（需Token认证）

**访问地址**：
- 开发者模式: http://localhost:9090
- Docker模式: http://localhost:53280（Nginx）或 http://localhost:53290（后端直接）

---

## 🔌 API文档

详细API文档请参考：[doc/API_接口文档.md](doc/API_接口文档.md)

### 基础信息

- **Base URL**: 
  - 开发者模式: `http://localhost:9090`
  - Docker模式: `http://localhost:53290`（直接访问后端）或 `http://localhost:53280`（通过Nginx代理）
- **Content-Type**: `application/json`
- **响应格式**: JSON

### 主要端点

#### 1. 健康检查

```http
GET /health
```

#### 2. 获取所有股票状态

```http
GET /api/stocks
```

#### 3. 获取单个股票最新分析

```http
GET /api/stock/:code/latest
```

#### 4. 获取单个股票历史分析

```http
GET /api/stock/:code/history?limit=20
```

#### 5. 获取所有股票最近分析

```http
GET /api/analysis/recent?limit=10
```

#### 6. 手动触发分析

```http
POST /api/stock/:code/analyze
```

#### 7. 重启后端（需Token认证）

```http
POST /api/system/restart
Headers: X-API-Token: your-token
```

---

## 📱 通知配置

### 钉钉机器人

1. 打开钉钉群 → 群设置 → 智能群助手 → 添加机器人
2. 选择"自定义"机器人
3. 安全设置：选择"自定义关键词"，填写关键词（如：`股票通知`）
4. 复制Webhook地址
5. 在配置文件中填写：
   ```json
   {
     "notification": {
       "dingtalk": {
         "enabled": true,
         "webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN",
         "secret": "股票通知"
       }
     }
   }
   ```

### 飞书机器人

1. 打开飞书群 → 群设置 → 群机器人 → 添加机器人
2. 选择"自定义机器人"
3. 复制Webhook URL
4. 在配置文件中填写Webhook地址

---

## ❓ 常见问题

### 1. 端口被占用

```bash
# 查找占用端口的进程
lsof -i :9090

# 停止占用端口的程序，或修改配置文件中的端口
```

### 2. TDX API连接失败

- 确认TDX API服务已启动
- 检查 `tdx_api_url` 配置是否正确
- 测试TDX API可访问性：`curl http://your-tdx-api:8181/health`

### 3. AI API调用失败

- 检查API密钥是否正确
- 确认API服务是否正常
- 检查网络连接

### 4. 钉钉/飞书通知失败

- 确认Webhook URL正确
- 检查关键词配置（钉钉）
- 查看日志文件：`stock_analysis_logs/stock_analyzer.log`

### 5. Docker容器无法启动

```bash
# 查看容器日志
docker-compose logs stock-analyzer

# 检查配置文件格式
cat config_stock.json | jq .

# 确认端口未被占用
```

### 6. 配置文件修改后未生效

**开发者模式**：需要重启程序

```bash
./manage_backend.sh restart
```

**Docker模式**：需要重启容器

```bash
docker-compose restart stock-analyzer
```

---

## 📁 项目结构

```
ai-ding-stock/
├── main_stock.go              # 主程序入口
├── config_stock.json.example  # 配置示例
├── go.mod                     # Go模块定义
├── go.sum                     # 依赖校验
│
├── manage_backend.sh          # 后端管理脚本（开发者模式）
├── docker-compose.yml         # Docker编排配置
├── Dockerfile                 # Docker镜像构建
├── docker-entrypoint.sh       # Docker启动脚本
├── docker-start.sh            # Docker启动脚本（Linux）
├── docker-start.bat           # Docker启动脚本（Windows）
├── nginx.conf                 # Nginx配置（Docker模式）
│
├── api/                       # API服务层
│   └── stock_server.go        # HTTP API服务器
│
├── config/                    # 配置管理
│   └── stock_config.go        # 配置加载和验证
│
├── stock/                     # 股票分析核心
│   ├── tdx_client.go          # TDX API客户端
│   ├── analyzer.go            # 分析引擎
│   ├── ai_parser.go           # AI响应解析
│   ├── position.go            # 持仓信息计算
│   └── trading_time.go        # 交易时间检查
│
├── mcp/                       # AI通信层
│   └── client.go              # AI API客户端
│
├── notifier/                  # 通知系统
│   └── webhook.go             # Webhook通知器
│
├── web/                       # Web前端
│   └── config.html            # 配置管理界面
│
├── stock_analysis_logs/       # 日志目录（自动创建）
│   └── stock_analyzer.log     # 日志文件
│
└── doc/                       # 文档目录
    ├── API_接口文档.md
    ├── DOCKER_DEPLOY.md
    ├── TDX-API调用分析.md
    ├── 持仓模式功能实施计划.md
    └── ...
```

---

## 📚 文档索引

详细技术文档位于 `doc/` 目录：

- **[API_接口文档.md](doc/API_接口文档.md)** - 完整的API接口文档
- **[DOCKER_DEPLOY.md](doc/DOCKER_DEPLOY.md)** - Docker部署详细指南
- **[TDX-API调用分析.md](doc/TDX-API调用分析.md)** - TDX API集成分析
- **[持仓模式功能实施计划.md](doc/持仓模式功能实施计划.md)** - 持仓模式功能说明
- **[项目分析总结.md](doc/项目分析总结.md)** - 项目技术架构分析

---

## 🔄 更新日志

### v2.1.0 (2025-11-20)

#### ✨ 新功能
- ➕ 持仓模式支持（输入持仓数量和购买价格）
- ➕ AI分析结果包含持仓止盈止损价格
- ➕ 分析历史记录存储和查询
- ➕ Web界面支持前端重启后端（Token认证）
- ➕ 信号本地化（BUY/SELL/HOLD显示为中文）
- ➕ 可配置分析历史记录数量（3-100条）

#### 🔧 优化
- ⚡ 优化日志路径（使用配置文件中的log_dir）
- ⚡ 优化通知格式（移除表格，使用emoji列表）
- ⚡ 优化分析原因显示（多行格式）
- ⚡ 改进管理脚本（使用相对路径）
- ⚡ 完善README文档（两种部署模式说明）

#### 🐛 修复
- 🔨 修复配置保存时丢失其他配置项的问题
- 🔨 修复通知逻辑（支持HOLD信号通知）
- 🔨 修复重启功能（自动重启机制）

### v2.0.0 (2025-11-04)

#### 🎉 重大更新
- ✨ 完全重构为股票分析系统
- 🚀 移除所有加密货币相关代码
- 🐳 优化Docker部署方案

---

## 📄 许可证

本项目采用 MIT 许可证

---

## ⚠️ 免责声明

本系统提供的分析结果仅供参考，不构成投资建议。

- ❌ 不保证分析准确性
- ❌ 不承担投资损失责任
- ❌ AI分析存在局限性
- ✅ 请独立思考，谨慎决策
- ✅ 投资有风险，入市需谨慎

---

---

## 🐳 Docker 镜像构建

### GitHub Container Registry (GHCR)

项目提供了 GitHub Actions 工作流，可以手动触发构建 Docker 镜像并推送到 GitHub Container Registry (GHCR)。

#### 使用方法

1. **手动触发构建**
   - 前往 GitHub 仓库的 Actions 页面
   - 选择 "Build and Push Docker Image to GHCR" 工作流
   - 点击 "Run workflow" 按钮
   - 可选择是否跳过缓存（首次构建建议使用缓存）

2. **拉取镜像**

   构建完成后，镜像会推送到 GHCR：

   ```bash
   # 登录 GHCR（首次使用需要）
   echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

   # 拉取镜像
   docker pull ghcr.io/YOUR_USERNAME/ai-ding-stock:latest

   # 运行容器
   docker run -d \
     --name stock-analyzer \
     -p 53290:9090 \
     -v $(pwd)/config_stock.json:/app/config_stock.json \
     -v $(pwd)/stock_analysis_logs:/app/stock_analysis_logs \
     -v $(pwd)/web:/app/web \
     ghcr.io/YOUR_USERNAME/ai-ding-stock:latest
   ```

   **注意**：请将 `YOUR_USERNAME` 替换为你的 GitHub 用户名或组织名。

3. **使用 Docker Compose**

   更新 `docker-compose.yml` 中的镜像地址：

   ```yaml
   services:
     stock-analyzer:
       image: ghcr.io/YOUR_USERNAME/ai-ding-stock:latest
       # ... 其他配置
   ```

#### 镜像信息

- **Registry**: `ghcr.io`
- **Tag**: `latest`（每次构建都会更新）
- **平台支持**: `linux/amd64`, `linux/arm64`

---

**Made with ❤️ for Stock Analysis**
