#!/bin/bash

# 启动脚本 - 演示如何使用.env文件启动应用

set -e  # 遇到错误时退出

echo "=== Go REST API 启动脚本 ==="
echo

# 检查.env文件是否存在
if [ ! -f ".env" ]; then
    echo "⚠️  .env文件不存在，从示例文件创建..."
    if [ -f ".env.example" ]; then
        cp .env.example .env
        echo "✅ 已从.env.example创建.env文件"
        echo "📝 请根据你的环境修改.env文件中的配置"
    else
        echo "❌ .env.example文件也不存在，请手动创建.env文件"
        exit 1
    fi
fi

# 显示当前环境配置
echo "📋 当前环境配置："
echo "   GIN_MODE: $(grep '^GIN_MODE=' .env | cut -d'=' -f2 || echo 'debug')"
echo "   POSTGRES_DB: $(grep '^POSTGRES_DB=' .env | cut -d'=' -f2 || echo 'go_app_dev')"
echo "   MONGO_ENABLED: $(grep '^MONGO_ENABLED=' .env | cut -d'=' -f2 || echo 'true')"
echo "   PORT: $(grep '^PORT=' .env | cut -d'=' -f2 || echo '8001')"
echo

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装，请先安装Go"
    exit 1
fi

# 检查依赖
echo "🔍 检查Go模块依赖..."
go mod tidy
echo "✅ 依赖检查完成"
echo

# 启动应用
echo "🚀 启动应用..."
echo "   使用.env文件加载配置"
echo "   访问地址: http://localhost:$(grep '^PORT=' .env | cut -d'=' -f2 || echo '8001')"
echo "   API文档: http://localhost:$(grep '^PORT=' .env | cut -d'=' -f2 || echo '8001')/swagger/index.html"
echo "   健康检查: http://localhost:$(grep '^PORT=' .env | cut -d'=' -f2 || echo '8001')/api/v1/"
echo

# 运行应用
go run cmd/server/main.go