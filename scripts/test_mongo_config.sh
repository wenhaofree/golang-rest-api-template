#!/bin/bash

# MongoDB配置测试脚本
# 用于测试不同的MONGO_ENABLED配置值

echo "=== MongoDB Configuration Test ==="
echo

# 测试不同的MONGO_ENABLED值
test_values=("true" "false" "1" "0" "yes" "no" "on" "off" "enable" "disable" "enabled" "disabled")

for value in "${test_values[@]}"; do
    echo "Testing MONGO_ENABLED=$value"
    
    # 设置环境变量并运行配置测试
    MONGO_ENABLED=$value go run scripts/config_check.go
    
    echo "---"
done

echo "=== Test Complete ==="