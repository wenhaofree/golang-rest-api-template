-- 性能优化索引脚本
-- 用于提升登录和用户查询性能

-- 1. 用户登录优化索引
-- 这个索引优化了基于email的登录查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email_active_deleted 
ON users (email, is_active) 
WHERE deleted_at IS NULL;

-- 2. 第三方登录优化索引  
-- 这个索引优化了第三方登录的查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_provider_auth_active 
ON users (provider_user_id, auth_provider, is_active) 
WHERE deleted_at IS NULL AND provider_user_id IS NOT NULL;

-- 3. 用户列表查询优化索引
-- 这个索引优化了用户列表的分页查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_created_at_active 
ON users (created_at DESC, is_active) 
WHERE deleted_at IS NULL;

-- 4. 最后登录时间索引（用于统计和排序）
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_last_login 
ON users (last_login DESC) 
WHERE deleted_at IS NULL AND last_login IS NOT NULL;

-- 5. 邮箱唯一性索引（如果不存在）
-- 注意：这个可能已经存在，根据实际情况调整
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email_unique_active 
ON users (email) 
WHERE deleted_at IS NULL;

-- 6. 复合索引用于用户状态查询
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_status_platform 
ON users (is_active, platform, auth_provider) 
WHERE deleted_at IS NULL;

-- 查看索引创建状态
-- 运行此脚本后，可以使用以下查询检查索引状态：

/*
-- 检查索引是否创建成功
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes 
WHERE tablename = 'users' 
ORDER BY indexname;

-- 检查索引大小
SELECT 
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes 
WHERE relname = 'users'
ORDER BY pg_relation_size(indexrelid) DESC;

-- 检查索引使用情况
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes 
WHERE relname = 'users'
ORDER BY idx_scan DESC;
*/

-- 性能测试查询
-- 可以使用 EXPLAIN ANALYZE 来测试这些查询的性能：

/*
-- 测试登录查询性能
EXPLAIN ANALYZE 
SELECT id, email, is_active, hashed_password, auth_provider, full_name, platform, avatar_url, last_login, created_at, updated_at
FROM users 
WHERE email = 'test@example.com' AND deleted_at IS NULL AND is_active = true;

-- 测试第三方登录查询性能
EXPLAIN ANALYZE 
SELECT * FROM users 
WHERE provider_user_id = 'google_123456' AND auth_provider = 'google' AND deleted_at IS NULL AND is_active = true;

-- 测试用户列表查询性能
EXPLAIN ANALYZE 
SELECT id, email, is_active, is_superuser, full_name, platform, auth_provider, avatar_url, last_login, created_at, updated_at
FROM users 
WHERE deleted_at IS NULL 
ORDER BY created_at DESC 
LIMIT 10 OFFSET 0;
*/