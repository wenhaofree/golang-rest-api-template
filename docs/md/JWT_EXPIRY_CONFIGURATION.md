# JWT Token 有效期配置指南

## 🎯 概述

JWT Token 有效期现在可以通过环境变量进行灵活配置，默认设置为 **24小时**，相比之前的5分钟大大提升了用户体验。

## ⚙️ 配置方法

### 1. 环境变量配置

在 `.env` 文件中设置：
```bash
# JWT Token 有效期配置
# 支持的时间单位: s(秒), m(分钟), h(小时), d(天)
JWT_EXPIRY_DURATION=24h
```

### 2. 支持的时间格式

| 格式 | 说明 | 示例 | 实际时长 |
|------|------|------|----------|
| `5m` | 5分钟 | `JWT_EXPIRY_DURATION=5m` | 5分钟 |
| `1h` | 1小时 | `JWT_EXPIRY_DURATION=1h` | 1小时 |
| `24h` | 24小时 | `JWT_EXPIRY_DURATION=24h` | 24小时 |
| `1d` | 1天 | `JWT_EXPIRY_DURATION=1d` | 24小时 |
| `7d` | 7天 | `JWT_EXPIRY_DURATION=7d` | 168小时 |

### 3. 安全限制

系统自动应用以下安全限制：
- **最小值**: 5分钟（防止过短导致频繁重新登录）
- **最大值**: 7天（防止安全风险）
- **默认值**: 24小时（平衡安全性和用户体验）

## 🌍 不同环境配置

### 开发环境 (.env.development)
```bash
# 开发环境 - 24小时，方便开发调试
JWT_EXPIRY_DURATION=24h
```

### 测试环境 (.env.test)
```bash
# 测试环境 - 较短时间，便于测试过期场景
JWT_EXPIRY_DURATION=1h
```

### 生产环境 (.env.production)
```bash
# 生产环境 - 24小时，平衡安全性和用户体验
JWT_EXPIRY_DURATION=24h
```

## 🧪 测试和验证

### 1. 检查当前配置
```bash
make test-jwt-expiry
```

输出示例：
```
=== JWT Token 有效期测试 ===

⚙️ 当前JWT配置:
  JWT_EXPIRY_DURATION: 24h
  解析后的时长: 24h0m0s
  等于: 24.0 小时
  等于: 1440 分钟
  JWT_SECRET_KEY: ✅ 已设置 (长度: 41)
```

### 2. 测试不同配置
```bash
# 测试各种时间格式
make test-jwt-expiry
```

### 3. 解析现有Token
```bash
# 解析并查看token的过期时间
make test-jwt-expiry-with-token TOKEN=<your_jwt_token>
```

## 📊 性能对比

### 修改前 vs 修改后

| 场景 | 修改前 | 修改后 | 改善 |
|------|--------|--------|------|
| Token有效期 | 5分钟 | 24小时 | **288倍** ⬆️ |
| 用户体验 | 频繁重新登录 | 一天内免登录 | 显著提升 |
| API调用便利性 | 需要频繁刷新token | 长时间有效 | 大幅改善 |

### 实际使用场景

**开发调试：**
- ✅ 不再需要频繁重新获取token
- ✅ 可以专注于功能开发而非认证问题
- ✅ 提高开发效率

**API测试：**
- ✅ 一次登录可以测试多个接口
- ✅ 减少测试脚本的复杂性
- ✅ 更稳定的自动化测试

## 🔒 安全考虑

### 1. 有效期选择原则

**短期有效期 (1-6小时):**
- ✅ 更高的安全性
- ❌ 用户体验较差
- 适用场景：高安全要求的系统

**中期有效期 (12-24小时):**
- ✅ 平衡安全性和用户体验
- ✅ 推荐配置
- 适用场景：大多数Web应用

**长期有效期 (1-7天):**
- ✅ 最佳用户体验
- ⚠️ 安全风险较高
- 适用场景：内部系统或低风险应用

### 2. 安全最佳实践

1. **定期轮换JWT密钥**
   ```bash
   # 生成新的JWT密钥
   make generate-jwt-key
   ```

2. **监控Token使用**
   - 记录异常的token使用模式
   - 监控过期token的访问尝试

3. **实施Token刷新机制**（可选）
   - 在token即将过期时自动刷新
   - 提供refresh token机制

## 🛠️ 故障排除

### 常见问题

**1. Token仍然很快过期**
```bash
# 检查配置是否生效
make test-jwt-expiry

# 确认环境变量
echo $JWT_EXPIRY_DURATION
```

**2. 配置格式错误**
```bash
# 测试不同格式
JWT_EXPIRY_DURATION=invalid make test-jwt-expiry
```

**3. Token解析失败**
```bash
# 调试token内容
make test-jwt-expiry-with-token TOKEN=<your_token>
```

### 配置验证

**验证步骤：**
1. 设置新的有效期配置
2. 重启应用服务
3. 登录获取新token
4. 验证token的过期时间

```bash
# 完整验证流程
echo "JWT_EXPIRY_DURATION=24h" >> .env
make run-dev &
sleep 5
make test-profile
make test-jwt-expiry-with-token TOKEN=<获取的token>
```

## 📋 配置检查清单

部署前确认：
- [ ] 设置了 `JWT_EXPIRY_DURATION` 环境变量
- [ ] 选择了合适的有效期（推荐24h）
- [ ] 确认JWT_SECRET_KEY已正确设置
- [ ] 测试了token生成和验证
- [ ] 验证了不同环境的配置

部署后验证：
- [ ] 新生成的token有效期正确
- [ ] 用户可以在有效期内正常使用
- [ ] 过期token被正确拒绝
- [ ] 日志显示正确的过期时间

## 💡 进一步优化

### 1. 动态有效期
根据用户角色或操作类型设置不同的有效期：
```go
// 管理员用户：较短有效期
// 普通用户：标准有效期
// 只读用户：较长有效期
```

### 2. Token刷新机制
实现自动token刷新，提升用户体验：
```go
// 在token即将过期时自动刷新
// 提供refresh token端点
```

### 3. 会话管理
结合Redis实现更复杂的会话管理：
```go
// 支持主动撤销token
// 实现单点登录(SSO)
// 设备管理功能
```
