#!/bin/sh
set -e

echo "🚀 启动 NOFX Stock Analyzer..."

# 检查配置文件是否存在
if [ ! -f /app/config_stock.json ]; then
    echo "⚠️  配置文件不存在，使用示例配置..."
    if [ -f /app/config_stock.json.example ]; then
        cp /app/config_stock.json.example /app/config_stock.json
        echo "✅ 已创建默认配置文件"
        echo "📝 请通过 Web 界面 (http://localhost:9090) 修改配置"
    else
        echo "❌ 错误：示例配置文件也不存在！"
        exit 1
    fi
else
    echo "✅ 配置文件已存在"
fi

# 检查日志目录
if [ ! -d /app/stock_analysis_logs ]; then
    mkdir -p /app/stock_analysis_logs
    echo "✅ 已创建日志目录"
fi

# 如果设置了API_TOKEN环境变量，更新到配置文件
if [ -n "$API_TOKEN" ]; then
    echo "🔑 检测到API_TOKEN环境变量，正在更新配置..."
    # 使用jq更新配置（如果可用），否则使用sed
    if command -v jq &> /dev/null; then
        jq --arg token "$API_TOKEN" '.api_token = $token' /app/config_stock.json > /tmp/config_stock.json.tmp && mv /tmp/config_stock.json.tmp /app/config_stock.json
        echo "✅ API Token已从环境变量更新到配置文件"
    else
        # 如果没有jq，使用sed简单处理（如果配置文件有api_token字段）
        if grep -q '"api_token"' /app/config_stock.json; then
            sed -i "s/\"api_token\":[^,}]*/\"api_token\": \"$API_TOKEN\"/g" /app/config_stock.json
            echo "✅ API Token已从环境变量更新到配置文件（使用sed）"
        fi
    fi
fi

# 显示配置信息
echo "📊 当前配置："
echo "   - API端口: $(grep -o '"api_server_port":[0-9]*' /app/config_stock.json | cut -d':' -f2 || echo '9090')"
echo "   - AI提供商: $(grep -o '"provider":"[^"]*"' /app/config_stock.json | cut -d':' -f2 | tr -d '"' || echo 'unknown')"
echo "   - 日志目录: /app/stock_analysis_logs"
if [ -n "$API_TOKEN" ]; then
    echo "   - API Token: 已从环境变量设置"
elif grep -q '"api_token"' /app/config_stock.json; then
    echo "   - API Token: 已配置（从配置文件）"
else
    echo "   - API Token: 未配置（将自动生成）"
fi

echo ""
echo "🌐 Web配置页面: http://localhost:9090"
echo "📡 API文档: http://localhost:9090/api/stocks"
echo ""
echo "🎯 启动应用..."

# 启动应用
exec ./stock_analyzer

