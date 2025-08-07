# MongoDB 配置说明

## 概述

本项目使用MongoDB来存储API请求日志。从v1.1版本开始，支持通过环境变量配置来启用或禁用MongoDB功能。

## MongoDB的作用

MongoDB在本项目中主要用于：
- 存储API请求日志（包括请求方法、路径、状态码、响应时间、客户端IP等）
- 提供结构化的日志查询能力
- 支持日志分析和监控

## 配置选项

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `MONGO_ENABLED` | `true` | 是否启用MongoDB日志功能 |
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB连接URI |
| `MONGO_DB` | `logging` | MongoDB数据库名 |
| `MONGO_LOG_COLLECTION` | `logs` | 日志集合名 |

### MONGO_ENABLED 支持的值

启用MongoDB（以下值都表示启用）：
- `true`, `1`, `yes`, `on`, `enable`, `enabled`

禁用MongoDB（以下值都表示禁用）：
- `false`, `0`, `no`, `off`, `disable`, `disabled`

## 使用方法

### 1. 启用MongoDB（默认）

```bash
# 不设置或设置为true
export MONGO_ENABLED=true

# 或者
export MONGO_ENABLED=1
export MONGO_ENABLED=yes
export MONGO_ENABLED=on
export MONGO_ENABLED=enable
export MONGO_ENABLED=enabled
```

### 2. 禁用MongoDB

```bash
# 设置为false
export MONGO_ENABLED=false

# 或者
export MONGO_ENABLED=0
export MONGO_ENABLED=no
export MONGO_ENABLED=off
export MONGO_ENABLED=disable
export MONGO_ENABLED=disabled
```

### 3. Docker Compose配置

在 `docker-compose.yml` 中修改环境变量：

```yaml
services:
  backend:
    environment:
      # 禁用MongoDB
      MONGO_ENABLED: "false"
      # 其他配置...
```

### 4. 本地开发配置

创建 `.env` 文件：

```bash
# 禁用MongoDB
MONGO_ENABLED=false

# 或启用MongoDB并自定义配置
MONGO_ENABLED=true
MONGO_URI=mongodb://localhost:27017
MONGO_DB=my_app_logs
MONGO_LOG_COLLECTION=api_logs
```

## 功能影响

### 启用MongoDB时
- API请求日志会同时写入到：
  - 结构化日志（Zap logger）
  - MongoDB集合
- 支持通过MongoDB查询和分析日志
- 需要MongoDB服务运行

### 禁用MongoDB时
- API请求日志只写入到结构化日志（Zap logger）
- 不需要MongoDB服务
- 应用启动更快，资源占用更少
- 日志仍然可以通过应用日志查看

## 故障处理

### MongoDB连接失败时
- 如果`MONGO_ENABLED=true`但MongoDB连接失败，应用不会崩溃
- 日志会记录连接失败信息
- 请求日志会继续写入结构化日志
- 建议检查MongoDB服务状态和连接配置

### 性能考虑
- MongoDB日志写入是异步的，不会阻塞API请求
- 每个日志写入操作有5秒超时限制
- 禁用MongoDB可以减少网络和存储开销

## 日志格式

### 结构化日志格式（Zap）
```json
{
  "level": "info",
  "ts": 1640995200.123,
  "msg": "Request",
  "method": "GET",
  "path": "/api/v1/books",
  "status": 200,
  "duration": "15.2ms",
  "ip": "192.168.1.100",
  "user-agent": "Mozilla/5.0...",
  "errors": ""
}
```

### MongoDB日志格式
```json
{
  "_id": ObjectId("..."),
  "timestamp": ISODate("2021-12-31T12:00:00.123Z"),
  "method": "GET",
  "path": "/api/v1/books",
  "status": 200,
  "duration": 15,
  "ip": "192.168.1.100",
  "user-agent": "Mozilla/5.0...",
  "errors": ""
}
```

## 迁移指南

### 从启用MongoDB迁移到禁用
1. 设置 `MONGO_ENABLED=false`
2. 重启应用
3. 可选：备份现有MongoDB日志数据
4. 可选：停止MongoDB服务

### 从禁用MongoDB迁移到启用
1. 确保MongoDB服务运行
2. 设置MongoDB相关环境变量
3. 设置 `MONGO_ENABLED=true`
4. 重启应用