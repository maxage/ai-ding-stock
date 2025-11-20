#!/bin/bash

# ========================================
# 🔧 后端服务管理脚本
# ========================================

# 获取脚本所在目录作为项目目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_DIR="$SCRIPT_DIR"
PID_FILE="/tmp/stock_analyzer.pid"
CONFIG_FILE="config_stock.json"

cd "$PROJECT_DIR" || exit 1

# 从配置文件读取日志目录（如果没有则使用默认值）
if [ -f "$CONFIG_FILE" ]; then
    # 尝试使用jq解析，如果没有jq则使用grep+sed
    if command -v jq &> /dev/null; then
        LOG_DIR=$(jq -r '.log_dir // "stock_analysis_logs"' "$CONFIG_FILE")
    else
        # 使用grep和sed提取log_dir，默认值为stock_analysis_logs
        LOG_DIR=$(grep -o '"log_dir"[[:space:]]*:[[:space:]]*"[^"]*"' "$CONFIG_FILE" | sed 's/.*"log_dir"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' || echo "stock_analysis_logs")
    fi
    # 确保日志目录不为空
    if [ -z "$LOG_DIR" ] || [ "$LOG_DIR" = "null" ]; then
        LOG_DIR="stock_analysis_logs"
    fi
else
    # 配置文件不存在，使用默认值
    LOG_DIR="stock_analysis_logs"
fi

# 日志文件路径：日志目录下的stock_analyzer.log
LOG_FILE="$PROJECT_DIR/$LOG_DIR/stock_analyzer.log"

# 确保日志目录存在
mkdir -p "$PROJECT_DIR/$LOG_DIR"

case "$1" in
    start)
        echo "🚀 启动后端服务..."
        # 检查是否已经在运行
        if ps aux | grep -v grep | grep stock_analyzer > /dev/null; then
            echo "⚠️  后端服务已经在运行中"
            ps aux | grep stock_analyzer | grep -v grep
            exit 1
        fi
        
        # 清理可能占用的端口
        lsof -ti:9090 | xargs kill -9 2>/dev/null
        
        # 启动服务
        nohup ./stock_analyzer "$CONFIG_FILE" > "$LOG_FILE" 2>&1 &
        echo $! > "$PID_FILE"
        sleep 3
        
        # 检查启动状态
        if ps -p $(cat "$PID_FILE") > /dev/null 2>&1; then
            echo "✅ 后端服务启动成功！"
            echo "📊 进程ID: $(cat $PID_FILE)"
            echo "🌐 Web界面: http://localhost:9090"
            echo "📋 日志文件: $LOG_FILE"
        else
            echo "❌ 后端服务启动失败，请查看日志: $LOG_FILE"
            exit 1
        fi
        ;;
        
    stop)
        echo "🛑 停止后端服务..."
        if [ -f "$PID_FILE" ]; then
            PID=$(cat "$PID_FILE")
            if ps -p "$PID" > /dev/null 2>&1; then
                kill "$PID" 2>/dev/null
                sleep 2
                # 如果还在运行，强制停止
                if ps -p "$PID" > /dev/null 2>&1; then
                    kill -9 "$PID" 2>/dev/null
                fi
                echo "✅ 后端服务已停止 (PID: $PID)"
            else
                echo "⚠️  进程不存在 (PID: $PID)"
            fi
            rm -f "$PID_FILE"
        fi
        
        # 清理所有相关进程
        pkill -f stock_analyzer 2>/dev/null
        lsof -ti:9090 | xargs kill -9 2>/dev/null
        echo "✅ 已清理所有相关进程"
        ;;
        
    restart)
        echo "🔄 重启后端服务..."
        $0 stop
        sleep 2
        $0 start
        ;;
        
    status)
        echo "📊 后端服务状态:"
        if [ -f "$PID_FILE" ]; then
            PID=$(cat "$PID_FILE")
            if ps -p "$PID" > /dev/null 2>&1; then
                echo "✅ 运行中 (PID: $PID)"
                ps aux | grep stock_analyzer | grep -v grep | head -1
                echo ""
                echo "🌐 健康检查:"
                curl -s http://localhost:9090/health 2>/dev/null | jq . 2>/dev/null || curl -s http://localhost:9090/health
            else
                echo "❌ 未运行 (PID文件存在但进程不存在)"
            fi
        else
            if ps aux | grep -v grep | grep stock_analyzer > /dev/null; then
                echo "⚠️  进程在运行但没有PID文件"
                ps aux | grep stock_analyzer | grep -v grep
            else
                echo "❌ 未运行"
            fi
        fi
        ;;
        
    logs)
        if [ -f "$LOG_FILE" ]; then
            echo "📋 查看日志 (最后50行，Ctrl+C退出):"
            echo "═══════════════════════════════════════════"
            tail -50 "$LOG_FILE"
            echo "═══════════════════════════════════════════"
            echo "实时日志: tail -f $LOG_FILE"
        else
            echo "❌ 日志文件不存在: $LOG_FILE"
        fi
        ;;
        
    tail)
        if [ -f "$LOG_FILE" ]; then
            echo "📋 实时查看日志 (Ctrl+C退出):"
            tail -f "$LOG_FILE"
        else
            echo "❌ 日志文件不存在: $LOG_FILE"
        fi
        ;;
        
    *)
        echo "用法: $0 {start|stop|restart|status|logs|tail}"
        echo ""
        echo "命令说明:"
        echo "  start   - 启动后端服务"
        echo "  stop    - 停止后端服务"
        echo "  restart - 重启后端服务"
        echo "  status  - 查看服务状态"
        echo "  logs    - 查看最近50行日志"
        echo "  tail    - 实时查看日志"
        exit 1
        ;;
esac

