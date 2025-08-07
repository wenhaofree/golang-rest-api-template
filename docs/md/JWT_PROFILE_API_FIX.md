# 个人资料接口JWT认证问题修复

## 🎯 问题诊断

### 错误现象
```
SELECT * FROM "users" WHERE email = '' AND deleted_at IS NULL ORDER BY "users"."id" LIMIT 1
```

**问题分析：**
- 个人资料接口查询时email字段为空字符串
- JWT token解析后获取的用户信息为空
- 导致数据库查询失败，返回404错误

## 🔍 根本原因

### JWT Token 结构不匹配

**JWT生成时：**
```go
// pkg/auth/auth.go
func GenerateToken(username string) (string, error) {
    claims := &jwt.StandardClaims{
        ExpiresAt: expirationTime,
        Issuer:    username,  // email存储在Issuer字段
    }
}
```

**JWT解析时：**
```go
// pkg/middleware/authenticateJWT.go (修复前)
claims := &auth.Claims{}  // 使用自定义Claims结构
c.Set("username", claims.Username)  // 读取Username字段
```

**问题：** 生成和解析使用了不同的Claims结构和字段！

## 🛠️ 解决方案

### 1. 修复JWT认证中间件

**修复前：**
```go
claims := &auth.Claims{}
c.Set("username", claims.Username)
```

**修复后：**
```go
// 使用StandardClaims来匹配token生成时的结构
claims := &jwt.StandardClaims{}

// 从Issuer字段获取email（与token生成时保持一致）
if claims.Issuer == "" {
    response.Unauthorized(c, "Invalid token: missing user information")
    c.Abort()
    return
}

c.Set("username", claims.Issuer)
```

### 2. 增强个人资料接口调试

**添加调试信息：**
```go
func (r *userRepository) GetUserProfile(c *gin.Context) {
    email, exists := c.Get("username")
    if !exists {
        response.Unauthorized(c, "User not authenticated")
        return
    }

    // 调试信息：检查获取到的email
    fmt.Printf("Debug - JWT extracted email: '%s'\n", email)
    
    // 类型断言确保email是字符串
    emailStr, ok := email.(string)
    if !ok || emailStr == "" {
        fmt.Printf("Debug - Invalid email type or empty: %T, value: %v\n", email, email)
        response.Unauthorized(c, "Invalid user information in token")
        return
    }
    
    // ... 其余代码
}
```

## 🧪 测试验证

### 1. 调试JWT Token

```bash
# 使用JWT调试工具
make debug-jwt TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 2. 测试个人资料接口

```bash
# 完整测试流程
make test-profile

# 或者完整API测试
make test-api-flow
```

### 3. 手动测试步骤

**步骤1：登录获取token**
```bash
curl -X POST 'http://127.0.0.1:8001/api/v1/login' \
  -H 'x-api-key: cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE' \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "user@privaterelay.appleid.com",
    "password": "password123"
  }'
```

**步骤2：使用token访问个人资料**
```bash
curl -X GET 'http://127.0.0.1:8001/api/v1/profile' \
  -H 'x-api-key: cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE' \
  -H 'Authorization: Bearer <YOUR_TOKEN>'
```

## 📊 修复效果

### 修复前
```
Debug - JWT extracted email: ''
SELECT * FROM "users" WHERE email = '' AND deleted_at IS NULL
❌ 404 User not found
```

### 修复后
```
Debug - JWT extracted email: 'user@privaterelay.appleid.com'
SELECT * FROM "users" WHERE email = 'user@privaterelay.appleid.com' AND deleted_at IS NULL
✅ 200 User profile returned
```

## 🔒 安全性考虑

### 1. JWT Token 验证增强
- 添加了Issuer字段空值检查
- 增强了token有效性验证
- 保持了原有的安全特性

### 2. 错误处理改进
- 更详细的错误信息
- 类型安全检查
- 调试信息输出

## 📋 部署检查清单

### 部署前确认
- [ ] JWT认证中间件已更新
- [ ] 个人资料接口已添加调试信息
- [ ] 测试脚本已创建
- [ ] Makefile已更新测试命令

### 部署后验证
- [ ] 登录接口正常工作
- [ ] JWT token包含正确的用户信息
- [ ] 个人资料接口返回正确数据
- [ ] 调试日志显示正确的email

## 🚀 快速修复步骤

1. **重启服务**
   ```bash
   make run-dev
   ```

2. **测试修复效果**
   ```bash
   make test-profile
   ```

3. **查看调试日志**
   ```bash
   # 观察服务日志中的调试信息
   # 应该看到正确的email而不是空字符串
   ```

## 💡 预防措施

### 1. 统一JWT Claims结构
建议创建统一的Claims结构，避免生成和解析不一致：

```go
type CustomClaims struct {
    Email string `json:"email"`
    jwt.StandardClaims
}
```

### 2. 添加单元测试
为JWT生成和解析添加单元测试，确保一致性。

### 3. API集成测试
定期运行完整的API测试流程，及早发现类似问题。

## 🔧 故障排除

### 常见问题

**1. Token仍然解析失败**
- 检查JWT_SECRET_KEY是否正确设置
- 确认token格式是否正确

**2. 用户仍然查询不到**
- 检查数据库中是否存在对应用户
- 确认email字段是否正确

**3. 第三方登录用户问题**
- 确认第三方登录时email字段正确保存
- 检查AuthProvider字段设置
