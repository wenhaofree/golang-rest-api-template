# 用户模型更新日志

## 概述

本次更新完全重构了用户模型，以支持第三方平台登录（如Apple ID）和现代化的用户管理功能。更新基于提供的数据库表结构，实现了完整的用户认证和管理系统。

## 主要变更

### 1. 用户模型重构 (`pkg/models/user.go`)

#### 新增字段
- `ID`: 使用UUID作为主键，替代原来的自增ID
- `Email`: 邮箱地址，作为唯一标识符
- `IsActive`: 用户激活状态
- `IsSuperuser`: 超级用户标识
- `FullName`: 用户全名（可选）
- `HashedPassword`: 加密密码（仅邮箱注册用户）
- `Platform`: 注册平台（web/mobile/app）
- `LastLogin`: 最后登录时间
- `AuthProvider`: 认证提供商（email/apple_id/google等）
- `ProviderUserID`: 第三方平台用户ID
- `AvatarURL`: 头像URL
- `DeletedAt`: 逻辑删除时间戳

#### 移除字段
- `Username`: 改用Email作为唯一标识
- `Password`: 改为HashedPassword，并且可选（第三方登录用户无密码）

#### 新增枚举类型
```go
// 平台类型
type PlatformEnum string
const (
    PlatformWeb    PlatformEnum = "web"
    PlatformMobile PlatformEnum = "mobile"
    PlatformApp    PlatformEnum = "app"
)

// 认证提供商
type AuthProviderEnum string
const (
    AuthProviderEmail    AuthProviderEnum = "email"
    AuthProviderAppleID  AuthProviderEnum = "apple_id"
    AuthProviderGoogle   AuthProviderEnum = "google"
    AuthProviderFacebook AuthProviderEnum = "facebook"
    AuthProviderGithub   AuthProviderEnum = "github"
    AuthProviderWechat   AuthProviderEnum = "wechat"
)
```

#### 新增结构体
- `ThirdPartyLoginUser`: 第三方登录请求
- `RegisterUser`: 注册请求
- `UserResponse`: API响应格式
- `UpdateUser`: 用户信息更新

#### 新增方法
- `ToResponse()`: 转换为安全的响应格式
- `IsDeleted()`: 检查是否被逻辑删除
- `SoftDelete()`: 执行逻辑删除
- `BeforeCreate()`: GORM钩子，自动生成UUID

### 2. API接口更新 (`pkg/api/user.go`)

#### 更新现有接口
- **登录接口**: 改为使用邮箱+密码，支持账户状态检查
- **注册接口**: 支持完整的用户信息注册
- **用户列表**: 排除已删除用户，返回安全格式

#### 新增接口
- `POST /auth/third-party`: 第三方平台登录
- `GET /profile`: 获取当前用户资料
- `PUT /profile`: 更新用户资料
- `DELETE /profile`: 逻辑删除用户账户

#### 安全改进
- 密码字段不在API响应中返回
- 增加账户激活状态检查
- 支持逻辑删除，保护数据完整性
- 更新最后登录时间

### 3. 路由配置更新 (`pkg/api/router.go`)

新增路由：
```go
// 认证路由
v1.POST("/auth/third-party", middleware.APIKeyAuth(), userRepository.ThirdPartyLoginHandler)

// 用户个人资料路由（需要JWT认证）
v1.GET("/profile", middleware.APIKeyAuth(), middleware.JWTAuth(), userRepository.GetUserProfile)
v1.PUT("/profile", middleware.APIKeyAuth(), middleware.JWTAuth(), userRepository.UpdateUserProfile)
v1.DELETE("/profile", middleware.APIKeyAuth(), middleware.JWTAuth(), userRepository.SoftDeleteUser)
```

### 4. 依赖管理

新增依赖：
- `github.com/google/uuid v1.6.0`: UUID支持

### 5. 测试覆盖

新增测试文件 `pkg/models/user_test.go`：
- 用户模型方法测试
- 枚举类型验证
- 结构体功能测试
- 逻辑删除功能测试

## 数据库兼容性

### 自动迁移
GORM会自动处理表结构迁移，新字段会被自动添加。

### 手动迁移（如需要）
如果需要从旧版本迁移数据，可参考以下SQL：

```sql
-- 备份现有数据
CREATE TABLE users_backup AS SELECT * FROM users;

-- 添加新字段
ALTER TABLE users ADD COLUMN IF NOT EXISTS id uuid DEFAULT gen_random_uuid();
ALTER TABLE users ADD COLUMN IF NOT EXISTS email varchar(255);
-- ... 其他字段

-- 数据迁移示例
UPDATE users SET email = username || '@example.com' WHERE email IS NULL;
UPDATE users SET hashed_password = password WHERE hashed_password IS NULL;
```

## 向后兼容性

⚠️ **重要提醒**: 此次更新包含破坏性变更：

1. **API接口变更**: 登录接口从username改为email
2. **数据模型变更**: 主键从int改为UUID
3. **字段重命名**: password → hashed_password

建议在生产环境部署前：
1. 备份现有数据库
2. 测试数据迁移脚本
3. 更新客户端代码以适配新的API接口

## 使用指南

详细的API使用说明请参考：`docs/USER_API_GUIDE.md`

## 第三方登录集成

当前实现提供了第三方登录的框架，但需要根据具体的第三方平台实现token验证逻辑：

```go
// TODO: 实现第三方token验证
func validateThirdPartyToken(provider AuthProviderEnum, token string) bool {
    switch provider {
    case AuthProviderAppleID:
        // 验证Apple ID token
        return validateAppleIDToken(token)
    case AuthProviderGoogle:
        // 验证Google token
        return validateGoogleToken(token)
    // ... 其他平台
    }
    return false
}
```

## 测试结果

所有测试通过：
- ✅ 用户模型测试
- ✅ API接口测试  
- ✅ 现有功能回归测试
- ✅ 编译检查通过

## 下一步计划

1. 实现具体的第三方平台token验证
2. 添加用户权限管理
3. 实现邮箱验证功能
4. 添加密码重置功能
5. 实现用户活动日志
