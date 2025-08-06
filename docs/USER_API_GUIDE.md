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

### 4. 获取用户资料

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

### 5. 更新用户资料

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

### 6. 逻辑删除用户

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

### 7. 获取用户列表

**GET** `/api/v1/users?offset=0&limit=10`

**Headers:**
- `X-API-Key: <API_KEY>`

## 特性说明

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
- 密码使用bcrypt加密存储
- JWT token用于身份验证
- API Key用于接口访问控制
- 第三方登录需要验证provider token（待实现）

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
