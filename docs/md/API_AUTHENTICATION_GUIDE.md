# API 认证权限指南

## 🔐 认证机制说明

本API系统采用**双重认证机制**：
1. **API Key认证**：所有接口都需要
2. **JWT Token认证**：部分接口额外需要

## 📋 接口权限矩阵

### 🔓 仅需要API Key的接口

| 方法 | 路径 | 说明 | 认证要求 |
|------|------|------|----------|
| GET | `/api/v1/` | 健康检查 | 无 |
| GET | `/api/v1/books` | 获取书籍列表 | API Key |
| GET | `/api/v1/books/:id` | 获取单个书籍 | API Key |
| PUT | `/api/v1/books/:id` | 更新书籍 | API Key |
| DELETE | `/api/v1/books/:id` | 删除书籍 | API Key |
| GET | `/api/v1/users` | 获取用户列表 | API Key |
| POST | `/api/v1/login` | 用户登录 | API Key |
| POST | `/api/v1/register` | 用户注册 | API Key |
| POST | `/api/v1/auth/third-party` | 第三方登录 | API Key |

### 🔒 需要API Key + JWT Token的接口

| 方法 | 路径 | 说明 | 认证要求 |
|------|------|------|----------|
| POST | `/api/v1/books` | 创建书籍 | API Key + JWT |
| GET | `/api/v1/profile` | 获取个人资料 | API Key + JWT |
| PUT | `/api/v1/profile` | 更新个人资料 | API Key + JWT |
| DELETE | `/api/v1/profile` | 删除用户账户 | API Key + JWT |

## 🛠️ 使用方法

### 1. API Key认证

**所有请求都需要包含API Key：**
```bash
curl -H "X-API-Key: cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE" \
     http://127.0.0.1:8001/api/v1/books
```

### 2. JWT Token认证

**需要JWT的接口还要包含Authorization头：**

**步骤1：登录获取Token**
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/login' \
  -H 'X-API-Key: cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE' \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

**步骤2：使用Token访问受保护接口**
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/books' \
  -H 'X-API-Key: cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE' \
  -H 'Authorization: Bearer <YOUR_JWT_TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "书籍标题",
    "author": "作者姓名"
  }'
```

## ❌ 常见错误

### 1. 401 Unauthorized - 缺少API Key
```json
{
  "success": false,
  "message": "Unauthorized"
}
```
**解决方案：** 添加 `X-API-Key` 头

### 2. 401 Unauthorized - 缺少JWT Token
```json
{
  "success": false,
  "message": "Missing Authorization Header"
}
```
**解决方案：** 添加 `Authorization: Bearer <token>` 头

### 3. 401 Unauthorized - JWT Token无效
```json
{
  "success": false,
  "message": "Invalid token"
}
```
**解决方案：** 重新登录获取新token

### 4. 401 Unauthorized - JWT Token过期
```json
{
  "success": false,
  "message": "Invalid token"
}
```
**解决方案：** Token有效期5分钟，需要重新登录

## 🧪 测试工具

### 1. 快速测试单个接口

**测试书籍API：**
```bash
make test-books
```

**测试个人资料API：**
```bash
make test-profile
```

### 2. 完整API流程测试

```bash
make test-api-flow
```

### 3. JWT Token调试

```bash
make debug-jwt TOKEN=<your_jwt_token>
```

## 📝 请求示例

### 创建书籍（正确示例）

```bash
# 1. 先登录获取token
LOGIN_RESPONSE=$(curl -s -X POST 'http://127.0.0.1:8001/api/v1/login' \
  -H 'X-API-Key: cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE' \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "user@privaterelay.appleid.com",
    "password": "password123"
  }')

# 2. 提取token
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token')

# 3. 创建书籍
curl -X POST 'http://127.0.0.1:8001/api/v1/books' \
  -H 'X-API-Key: cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE' \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "我的书籍",
    "author": "我的作者"
  }'
```

### 获取书籍列表（只需API Key）

```bash
curl -X GET 'http://127.0.0.1:8001/api/v1/books?offset=0&limit=10' \
  -H 'X-API-Key: cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE'
```

## 🔧 故障排除

### 检查认证配置

1. **确认API Key正确**
   ```bash
   echo $API_SECRET_KEY
   ```

2. **确认JWT密钥正确**
   ```bash
   echo $JWT_SECRET_KEY
   ```

3. **检查token有效性**
   ```bash
   make debug-jwt TOKEN=<your_token>
   ```

### 常见问题解决

**问题1：所有接口都返回401**
- 检查API_SECRET_KEY环境变量
- 确认X-API-Key头格式正确

**问题2：创建书籍返回401，但获取书籍正常**
- 确认已添加Authorization头
- 检查JWT token是否有效
- 确认token未过期（5分钟有效期）

**问题3：个人资料接口正常，但创建书籍失败**
- 两个接口都需要JWT，检查请求头是否一致
- 确认使用的是同一个token

## 💡 最佳实践

1. **Token管理**
   - 及时刷新过期token
   - 安全存储token
   - 不要在URL中传递token

2. **错误处理**
   - 检查响应状态码
   - 根据错误信息调整请求
   - 实现自动重试机制

3. **安全考虑**
   - 使用HTTPS（生产环境）
   - 定期轮换API密钥
   - 监控异常访问
