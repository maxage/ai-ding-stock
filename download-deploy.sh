#!/bin/bash

# ================================
# AI股票分析系统 - 一键下载部署文件
# ================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 仓库信息
REPO_OWNER="maxage"
REPO_NAME="ai-ding-stock"
BRANCH="main"
BASE_URL="https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/${BRANCH}"

# 部署目录
DEPLOY_DIR="Ai-Ding-Stock"

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   AI股票分析系统 - 一键下载部署文件                        ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# 1. 创建部署目录
echo -e "${YELLOW}📁 创建部署目录: ${DEPLOY_DIR}${NC}"
if [ -d "${DEPLOY_DIR}" ]; then
    echo -e "${YELLOW}⚠️  目录已存在，将覆盖现有文件${NC}"
    read -p "是否继续？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${RED}❌ 已取消${NC}"
        exit 1
    fi
    rm -rf "${DEPLOY_DIR}"
fi

mkdir -p "${DEPLOY_DIR}"
cd "${DEPLOY_DIR}"
echo -e "${GREEN}✅ 目录创建成功${NC}"
echo ""

# 2. 创建 web 目录
echo -e "${YELLOW}📁 创建 web 目录${NC}"
mkdir -p web
echo -e "${GREEN}✅ web 目录创建成功${NC}"
echo ""

# 3. 下载文件
echo -e "${YELLOW}📥 开始下载必要文件...${NC}"
echo ""

# 下载 docker-compose.yml
echo -e "${BLUE}[1/4] 下载 docker-compose.yml${NC}"
if wget -q --show-progress "${BASE_URL}/docker-compose.yml" -O docker-compose.yml; then
    echo -e "${GREEN}✅ docker-compose.yml 下载成功${NC}"
else
    echo -e "${RED}❌ docker-compose.yml 下载失败${NC}"
    exit 1
fi

# 下载 nginx.conf
echo -e "${BLUE}[2/4] 下载 nginx.conf${NC}"
if wget -q --show-progress "${BASE_URL}/nginx.conf" -O nginx.conf; then
    echo -e "${GREEN}✅ nginx.conf 下载成功${NC}"
else
    echo -e "${RED}❌ nginx.conf 下载失败${NC}"
    exit 1
fi

# 下载 config_stock.json.example 并重命名为 config_stock.json
echo -e "${BLUE}[3/4] 下载 config_stock.json.example${NC}"
if wget -q --show-progress "${BASE_URL}/config_stock.json.example" -O config_stock.json; then
    echo -e "${GREEN}✅ config_stock.json 下载成功${NC}"
    echo -e "${YELLOW}⚠️  请编辑 config_stock.json 填写您的配置（TDX API URL、AI密钥等）${NC}"
else
    echo -e "${RED}❌ config_stock.json.example 下载失败${NC}"
    exit 1
fi

# 下载 web/config.html
echo -e "${BLUE}[4/4] 下载 web/config.html${NC}"
if wget -q --show-progress "${BASE_URL}/web/config.html" -O web/config.html; then
    echo -e "${GREEN}✅ web/config.html 下载成功${NC}"
else
    echo -e "${RED}❌ web/config.html 下载失败${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║    ✅ 所有文件下载完成！                                   ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# 4. 显示目录结构
echo -e "${BLUE}📂 当前目录结构：${NC}"
if command -v tree &> /dev/null; then
    tree -L 2 -a . 2>/dev/null || echo "├── docker-compose.yml"
    echo "├── config_stock.json"
    echo "├── nginx.conf"
    echo "└── web/"
    echo "    └── config.html"
else
    echo "├── docker-compose.yml"
    echo "├── config_stock.json"
    echo "├── nginx.conf"
    echo "└── web/"
    echo "    └── config.html"
fi
echo ""

# 5. 显示下一步操作
echo -e "${YELLOW}📋 下一步操作：${NC}"
echo ""
echo -e "${BLUE}1. 编辑配置文件：${NC}"
echo -e "   ${GREEN}cd ${DEPLOY_DIR}${NC}"
echo -e "   ${GREEN}vim config_stock.json${NC}  或  ${GREEN}nano config_stock.json${NC}"
echo ""
echo -e "${BLUE}2. 至少需要配置以下内容：${NC}"
echo -e "   - ${YELLOW}tdx_api_url${NC}: TDX API 服务地址（如: http://192.168.1.222:8181）"
echo -e "   - ${YELLOW}ai_config${NC}: AI 配置（DeepSeek/Qwen密钥等）"
echo -e "   - ${YELLOW}stocks${NC}: 要监控的股票代码列表"
echo -e "   - ${YELLOW}notification${NC}: 通知配置（钉钉/飞书，可选）"
echo ""
echo -e "${BLUE}3. 启动服务：${NC}"
echo -e "   ${GREEN}docker-compose pull${NC}        # 拉取远程镜像"
echo -e "   ${GREEN}docker-compose up -d${NC}       # 后台启动"
echo -e "   ${GREEN}docker-compose logs -f${NC}     # 查看日志"
echo ""
echo -e "${BLUE}4. 访问系统：${NC}"
echo -e "   ${GREEN}Web界面: http://localhost:53280${NC}"
echo -e "   ${GREEN}API接口: http://localhost:53290${NC}"
echo ""
echo -e "${YELLOW}💡 提示：配置文件也可以通过 Web 界面在线编辑${NC}"
echo ""
echo -e "${BLUE}📚 详细部署文档：${NC}"
echo -e "   https://github.com/${REPO_OWNER}/${REPO_NAME}/blob/main/doc/Docker-Compose部署指南.md"
echo ""

