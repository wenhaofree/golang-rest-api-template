# 用户API使用指南

本文档介绍了更新后的用户API，支持第三方平台登录（如Apple ID）和完整的用户管理功能。

## 数据模型

### 用户模型字段

```go
type User struct {
    ID             uuid.UUID        // 用户唯一标识符
    Email          string           // 邮箱地址（唯一）
    IsActive       bool             // 是否激活
    IsSuperuser    bool             // 是否超级用户
    FullName       *string          // 全名（可选）
    HashedPassword *string          // 加密密码（仅邮箱注册用户）
    Platform       PlatformEnum     // 注册平台
    AuthProvider   AuthProviderEnum // 认证提供商
    ProviderUserID *string          // 第三方平台用户ID
    AvatarURL      *string          // 头像URL
    LastLogin      *time.Time       // 最后登录时间
    // 时间字段统一放在一起
    CreatedAt      time.Time        // 创建时间
    UpdatedAt      time.Time        // 更新时间
    DeletedAt      *time.Time       // 逻辑删除时间
}
```

### 枚举类型

#### 平台类型 (PlatformEnum)
- `web` - 网页端
- `mobile` - 移动端
- `app` - 应用端

#### 认证提供商 (AuthProviderEnum)
- `email` - 邮箱密码登录
- `apple_id` - Apple ID登录
- `google` - Google登录
- `facebook` - Facebook登录
- `github` - GitHub登录
- `wechat` - 微信登录

## API端点

### 1. 邮箱密码注册

**POST** `/api/v1/register`

```json
{
  "email": "user@example.com",
  "password": "password123",
  "full_name": "张三",
  "platform": "web",
  "auth_provider": "email"
}
```

**响应:**
```json
{
  "code": 0,
  "data": {
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "user@example.com",
      "is_active": true,
      "is_superuser": false,
      "full_name": "张三",
      "platform": "web",
      "auth_provider": "email",
      "avatar_url": null,
      "last_login": null,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  },
  "message": "Registration successful"
}
```

### 2. 邮箱密码登录

**POST** `/api/v1/login`

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**响应:**
```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "user@example.com",
      "is_active": true,
      "is_superuser": false,
      "full_name": "张三",
      "platform": "web",
      "auth_provider": "email",
      "avatar_url": null,
      "last_login": "2024-01-01T12:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  },
  "message": "success"
}
```

### 3. 第三方平台登录

**POST** `/api/v1/auth/third-party`

```json
{
  "auth_provider": "apple_id",
  "provider_user_id": "001234.567890abcdef.1234",
  "email": "user@privaterelay.appleid.com",
  "full_name": "张三",
  "avatar_url": "https://example.com/avatar.jpg",
  "platform": "mobile",
  "provider_token": "apple_id_token_here"
}
```

**响应:**
```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "email": "user@privaterelay.appleid.com",
      "is_active": true,
      "is_superuser": false,
      "full_name": "张三",
      "platform": "mobile",
      "created_at": "2024-01-01T00:00:00Z",
      "last_login": "2024-01-01T12:00:00Z",
      "auth_provider": "apple_id",
      "avatar_url": "https://example.com/avatar.jpg"
    }
  },
  "message": "success"
}
```

### 4. 刷新访问令牌

**POST** `/api/v1/auth/refresh-token`

**Headers:**
- `X-API-Key: <API_KEY>`

**请求体:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应:**
```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  },
  "message": "Tokens refreshed successfully"
}
```

**说明:**
- 使用有效的 refresh token 获取新的访问令牌和刷新令牌
- 旧的 refresh token 在成功刷新后会被立即撤销
- 新的 refresh token 有效期为7天（可配置）
- 访问令牌有效期为24小时（可配置）

### 5. 获取用户资料

**GET** `/api/v1/profile`

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`
- `X-API-Key: <API_KEY>`

**响应:**
```json
{
  "code": 0,
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "user@example.com",
    "is_active": true,
    "is_superuser": false,
    "full_name": "张三",
    "platform": "web",
    "created_at": "2024-01-01T00:00:00Z",
    "last_login": "2024-01-01T12:00:00Z",
    "auth_provider": "email",
    "avatar_url": null
  },
  "message": "success"
}
```

### 6. 更新用户资料

**PUT** `/api/v1/profile`

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`
- `X-API-Key: <API_KEY>`

```json
{
  "full_name": "李四",
  "avatar_url": "https://example.com/new-avatar.jpg"
}
```

### 7. 逻辑删除用户

**DELETE** `/api/v1/profile`

**Headers:**
- `Authorization: Bearer <JWT_TOKEN>`
- `X-API-Key: <API_KEY>`

**响应:**
```json
{
  "code": 0,
  "data": null,
  "message": "Account deleted successfully"
}
```

### 8. 获取用户列表

**GET** `/api/v1/users?offset=0&limit=10`

**Headers:**
- `X-API-Key: <API_KEY>`

## 特性说明

### JWT 双令牌认证系统
本API采用JWT双令牌认证机制：

- **访问令牌 (Access Token)**: 用于API访问认证，短期有效（默认24小时）
- **刷新令牌 (Refresh Token)**: 用于刷新访问令牌，长期有效（默认7天）

#### 令牌使用流程
1. 用户登录后获得访问令牌和刷新令牌
2. 使用访问令牌调用需要认证的API
3. 当访问令牌过期时，使用刷新令牌获取新的令牌对
4. 客户端应自动处理令牌刷新，保持用户登录状态

#### 令牌刷新最佳实践
```javascript
// 示例：JavaScript中的自动令牌刷新
async function apiCall(url, options = {}) {
  let token = localStorage.getItem('access_token');
  
  // 添加认证头
  const headers = {
    'X-API-Key': 'your-api-key',
    'Authorization': `Bearer ${token}`,
    ...options.headers
  };
  
  let response = await fetch(url, { ...options, headers });
  
  // 如果令牌过期，尝试刷新
  if (response.status === 401) {
    const refreshToken = localStorage.getItem('refresh_token');
    if (refreshToken) {
      const refreshResponse = await fetch('/api/v1/auth/refresh-token', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': 'your-api-key'
        },
        body: JSON.stringify({ refresh_token: refreshToken })
      });
      
      if (refreshResponse.ok) {
        const data = await refreshResponse.json();
        localStorage.setItem('access_token', data.data.token);
        localStorage.setItem('refresh_token', data.data.refresh_token);
        
        // 重新发起原请求
        headers['Authorization'] = `Bearer ${data.data.token}`;
        response = await fetch(url, { ...options, headers });
      } else {
        // 刷新失败，跳转到登录页
        window.location.href = '/login';
        return;
      }
    }
  }
  
  return response;
}
```

### 逻辑删除
- 用户删除采用逻辑删除方式，设置 `deleted_at` 字段
- 被删除的用户不会出现在用户列表中
- 被删除的用户无法登录

### 第三方登录流程
1. 客户端通过第三方平台（如Apple ID）获取用户信息和token
2. 调用 `/auth/third-party` 接口，传入第三方平台信息
3. 系统验证第三方token（需要实现验证逻辑）
4. 如果用户不存在，自动创建新用户
5. 返回JWT token和用户信息

### 安全考虑
- **密码安全**: 使用bcrypt加密存储，支持可配置的成本参数
- **双令牌认证**: JWT访问令牌 + 刷新令牌机制，提供更好的安全性
- **令牌轮换**: 刷新令牌使用时会自动轮换，旧令牌立即失效
- **API访问控制**: 所有接口都需要API Key进行访问控制
- **令牌缓存管理**: 刷新令牌存储在Redis中，支持主动撤销
- **用户状态验证**: 令牌刷新时会验证用户仍然处于活跃状态
- **防时序攻击**: 登录验证时使用dummy hash防止时序攻击
- **第三方登录**: 需要验证provider token（待实现具体验证逻辑）

### 环境变量配置
```bash
# JWT配置
JWT_SECRET_KEY=your-jwt-secret-key
JWT_EXPIRY_DURATION=24h                    # 访问令牌有效期（默认24小时）
REFRESH_TOKEN_EXPIRY_DURATION=7d           # 刷新令牌有效期（默认7天）

# bcrypt配置
BCRYPT_COST=12                             # bcrypt加密强度（4-15，默认12）

# API安全
API_SECRET_KEY=your-api-secret-key
```

## 数据库迁移

新的用户表结构会在应用启动时自动迁移。如果需要手动迁移现有数据，请参考以下SQL：

```sql
-- 添加新字段（如果从旧版本升级）
ALTER TABLE users ADD COLUMN IF NOT EXISTS id uuid DEFAULT gen_random_uuid();
ALTER TABLE users ADD COLUMN IF NOT EXISTS email varchar(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_superuser boolean DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS full_name varchar(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS hashed_password varchar;
ALTER TABLE users ADD COLUMN IF NOT EXISTS platform varchar(20) DEFAULT 'web';
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login timestamp(6);
ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_provider varchar(20) DEFAULT 'email';
ALTER TABLE users ADD COLUMN IF NOT EXISTS provider_user_id varchar(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url varchar(500);
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at timestamp(6);
```
