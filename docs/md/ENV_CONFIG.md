# 环境变量配置指南

## 概述

本项目支持通过 `.env` 文件和环境变量进行配置管理。配置系统具有以下特性：

- 🔧 **多层配置**: 支持基础配置、环境特定配置和本地覆盖
- 🔄 **自动加载**: 应用启动时自动加载相应的 `.env` 文件
- 🛡️ **安全管理**: 敏感配置文件不会被提交到版本控制
- 🌍 **环境隔离**: 不同环境使用不同的配置文件

## 配置文件层次

### 加载优先级（从低到高）

1. **`.env`** - 基础配置文件
2. **`.env.{environment}`** - 环境特定配置
   - `.env.development` (GIN_MODE=debug)
   - `.env.production` (GIN_MODE=release)  
   - `.env.test` (GIN_MODE=test)
3. **`.env.local`** - 本地覆盖配置
4. **系统环境变量** - 最高优先级

### 文件说明

| 文件名 | 用途 | 是否提交到Git |
|--------|------|---------------|
| `.env.example` | 配置模板和文档 | ✅ 是 |
| `.env` | 基础配置 | ❌ 否 |
| `.env.development` | 开发环境配置 | ✅ 是 |
| `.env.production` | 生产环境配置模板 | ✅ 是 |
| `.env.test` | 测试环境配置 | ✅ 是 |
| `.env.local` | 本地覆盖配置 | ❌ 否 |

## 配置项说明

### 数据库配置
```bash
POSTGRES_HOST=localhost          # PostgreSQL主机地址
POSTGRES_DB=goapi_db            # 数据库名
POSTGRES_USER=root              # 数据库用户名
POSTGRES_PASSWORD=password      # 数据库密码
POSTGRES_PORT=5432              # 数据库端口
```

### Redis配置
```bash
REDIS_HOST=localhost            # Redis主机地址
```

### JWT认证配置
```bash
JWT_SECRET_KEY=your-secret-key  # JWT签名密钥
API_SECRET_KEY=your-api-key     # API密钥
```

### MongoDB配置（日志存储）
```bash
MONGO_ENABLED=true              # 是否启用MongoDB
MONGO_URI=mongodb://localhost:27017  # MongoDB连接URI
MONGO_DB=logging                # MongoDB数据库名
MONGO_LOG_COLLECTION=logs       # 日志集合名
```

### 应用配置
```bash
GIN_MODE=debug                  # Gin运行模式: debug/release/test
LOG_LEVEL=info                  # 日志级别: debug/info/warn/error
PORT=8001                       # 应用端口
```

## 使用方法

### 1. 初始化配置

```bash
# 从示例文件创建配置
cp .env.example .env

# 编辑配置文件
vim .env
```

### 2. 启动应用

#### 方式一：使用启动脚本（推荐）
```bash
# 自动检查和加载.env文件
make run-env
# 或直接运行
./scripts/start.sh
```

#### 方式二：指定环境启动
```bash
# 开发环境
make run-dev

# 生产环境  
make run-prod

# 测试环境
make run-test
```

#### 方式三：直接运行
```bash
# Go程序会自动加载.env文件
go run cmd/server/main.go
```

### 3. 环境变量覆盖

```bash
# 临时覆盖配置
MONGO_ENABLED=false go run cmd/server/main.go

# 或者export后运行
export MONGO_ENABLED=false
go run cmd/server/main.go
```

## 不同环境的配置策略

### 开发环境 (`.env.development`)
```bash
GIN_MODE=debug
MONGO_ENABLED=false              # 简化开发环境
POSTGRES_DB=goapi_dev           # 开发数据库
PORT=8001
```

### 测试环境 (`.env.test`)
```bash
GIN_MODE=test
MONGO_ENABLED=false              # 加快测试速度
POSTGRES_DB=goapi_test          # 测试数据库
PORT=8002                       # 避免端口冲突
```

### 生产环境 (`.env.production`)
```bash
GIN_MODE=release
MONGO_ENABLED=true               # 启用完整日志
# 生产环境敏感信息通过环境变量注入
# POSTGRES_PASSWORD=${DB_PASSWORD}
```

## 安全最佳实践

### 1. 敏感信息管理
```bash
# ❌ 错误：在.env文件中存储生产密码
POSTGRES_PASSWORD=prod_secret_password

# ✅ 正确：使用环境变量占位符
POSTGRES_PASSWORD=${DB_PASSWORD}

# ✅ 正确：在部署时注入环境变量
export DB_PASSWORD="actual_secret_password"
```

### 2. 文件权限控制
```bash
# 设置.env文件权限，只有所有者可读写
chmod 600 .env
chmod 600 .env.local
```

### 3. Git忽略规则
```gitignore
# .gitignore
.env
.env.local
.env.*.local
```

## Docker集成

### Docker Compose
```yaml
# docker-compose.yml
services:
  app:
    build: .
    env_file:
      - .env                    # 基础配置
      - .env.production         # 生产环境配置
    environment:
      - DB_PASSWORD=${DB_PASSWORD}  # 从宿主机注入敏感信息
```

### Dockerfile
```dockerfile
# 复制配置文件（不包含敏感信息）
COPY .env.example .env.production ./
```

## 配置验证

### 检查配置加载
```bash
# 测试配置解析
go run scripts/config_check.go

# 测试不同的MONGO_ENABLED值
MONGO_ENABLED=false go run scripts/config_check.go
```

### 健康检查
```bash
# 访问健康检查端点查看配置状态
curl http://localhost:8001/api/v1/
```

## 故障排除

### 常见问题

1. **配置文件格式错误**
```bash
# ❌ 错误格式
export KEY=value
KEY = value
KEY=value with spaces

# ✅ 正确格式  
KEY=value
KEY="value with spaces"
KEY='value with spaces'
```

2. **配置不生效**
```bash
# 检查文件是否存在
ls -la .env*

# 检查文件内容
cat .env

# 检查环境变量
env | grep MONGO_ENABLED
```

3. **权限问题**
```bash
# 检查文件权限
ls -la .env

# 修复权限
chmod 600 .env
```

## 迁移指南

### 从export方式迁移到.env文件

#### 旧方式（export）
```bash
export POSTGRES_HOST=localhost
export POSTGRES_DB=mydb
go run cmd/server/main.go
```

#### 新方式（.env文件）
```bash
# .env
POSTGRES_HOST=localhost
POSTGRES_DB=mydb

# 直接运行，自动加载配置
go run cmd/server/main.go
```

### 团队协作建议

1. **提交配置模板**：提交 `.env.example` 和环境特定的 `.env.{env}` 文件
2. **忽略敏感文件**：确保 `.env` 和 `.env.local` 在 `.gitignore` 中
3. **文档化配置**：在README中说明必需的配置项
4. **使用默认值**：为所有配置项提供合理的默认值

## 总结

通过使用 `.env` 文件系统，项目获得了：

- ✅ **配置集中管理**：所有配置在文件中统一管理
- ✅ **环境隔离**：不同环境使用不同配置
- ✅ **团队协作友好**：配置模板可以共享
- ✅ **安全性**：敏感信息不会意外提交
- ✅ **灵活性**：支持多层配置覆盖
- ✅ **易于部署**：与Docker和CI/CD系统良好集成