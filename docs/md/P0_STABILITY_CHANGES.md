## 建议保存为文档文件
建议将下述内容保存为 docs/md/P0_STABILITY_CHANGES.md，作为本轮 P0 安全/稳定性优化的变更与使用指南。

## 概览
本次改动聚焦“稳定性 + 超时 + 连接池 + 配置集中化”：
- 配置集中：DB/Redis/请求超时等统一由 config 管理
- 数据库：连接池参数、指数退避重试、SQL 级 statement_timeout
- Redis：连接池与超时全面可配
- 上下文与超时：HTTP 请求级超时中间件 + service 层 DB 查询统一 WithContext
- 低侵入：路由、入口的调用签名小幅调整；业务行为不变

## 变更清单（按模块）

### 配置 config
- 扩展配置（PostgreSQL/Redis/请求超时/连接重试/查询超时等）
- 统一通过 config.LoadConfig() 注入下游，避免模块内直接读取 env

新增关键字段（节选）：
- PostgreSQL：POSTGRES_SSLMODE、POSTGRES_MAX_OPEN_CONNS、POSTGRES_MAX_IDLE_CONNS、POSTGRES_CONN_MAX_IDLE_TIME_SEC、POSTGRES_CONN_MAX_LIFETIME_SEC、DB_CONNECT_MAX_RETRIES、DB_CONNECT_INITIAL_BACKOFF_MS、DB_QUERY_TIMEOUT_MS
- Redis：REDIS_PORT、REDIS_PASSWORD、REDIS_DB、REDIS_POOL_SIZE、REDIS_MIN_IDLE_CONNS、REDIS_IDLE_TIMEOUT_SEC、REDIS_DIAL_TIMEOUT_MS、REDIS_READ_TIMEOUT_MS、REDIS_WRITE_TIMEOUT_MS
- 请求：REQUEST_TIMEOUT_MS

### 数据库 database
- NewDatabase(cfg) 接口变更：从 env 读取改为使用 cfg
- 使用 DSN 注入 statement_timeout，确保单条 SQL 超时
- 连接池参数完整可配；连接失败指数退避重试
- 保留 AutoMigrate（生产建议迁移工具）

接口签名位置参考：
````go path=pkg/database/db.go mode=EXCERPT
func NewDatabase(cfg *config.Config) *gorm.DB
````

### Redis cache
- NewRedisClient(cfg) 接口变更：注入 cfg 配置，设置 Pool/Timeout

接口签名位置参考：
````go path=pkg/cache/cache.go mode=EXCERPT
func NewRedisClient(cfg *config.Config) *redis.Client
````

### Router 与中间件
- 新增请求级超时中间件 RequestTimeout
- NewRouter 增加 requestTimeoutMs 参数，支持按配置注入

签名与使用参考：
````go path=pkg/api/router.go mode=EXCERPT
func NewRouter(..., requestTimeoutMs int) *gin.Engine
// r.Use(middleware.RequestTimeout(requestTimeoutMs))
````

### Service 层
- 所有 DB 调用统一使用 WithContext(ctx)，确保取消/超时可控

示例（user_service）：
````go path=pkg/services/user_service.go mode=EXCERPT
if err := s.db.WithContext(ctx).
  Where("email = ?", email).
  First(&u).Error(); err != nil { ... }
````

## 使用方法

### 1) 配置环境变量（新增/关键）
- 请求/查询超时：
  - REQUEST_TIMEOUT_MS=3000
  - DB_QUERY_TIMEOUT_MS=3000
- DB 重试与退避：
  - DB_CONNECT_MAX_RETRIES=5
  - DB_CONNECT_INITIAL_BACKOFF_MS=500
- DB 连接池：
  - POSTGRES_MAX_OPEN_CONNS=25
  - POSTGRES_MAX_IDLE_CONNS=25
  - POSTGRES_CONN_MAX_IDLE_TIME_SEC=300
  - POSTGRES_CONN_MAX_LIFETIME_SEC=1800
  - POSTGRES_SSLMODE=disable（生产按需改为 require/verify-full）
- Redis 连接与池化：
  - REDIS_PORT=6379
  - REDIS_PASSWORD=""
  - REDIS_DB=0
  - REDIS_POOL_SIZE=10
  - REDIS_MIN_IDLE_CONNS=2
  - REDIS_IDLE_TIMEOUT_SEC=300
  - REDIS_DIAL_TIMEOUT_MS=500
  - REDIS_READ_TIMEOUT_MS=1000
  - REDIS_WRITE_TIMEOUT_MS=1000

保留现有变量：
- POSTGRES_HOST/DB/USER/PASSWORD/PORT
- REDIS_HOST
- JWT_SECRET_KEY / API_SECRET_KEY
- MONGO_ENABLED / MONGO_URI / MONGO_DB / MONGO_LOG_COLLECTION
- LOG_LEVEL

建议：在 .env、.env.development、.env.production 中分别配置合理默认值。

### 2) 升级调用方式（若你有自定义入口）
- 初始化改为使用 cfg：
  - db := database.NewDatabase(cfg)
  - redis := cache.NewRedisClient(cfg)
  - r := api.NewRouter(..., cfg.RequestTimeoutMs)

main 中参考：
````go path=cmd/server/main.go mode=EXCERPT
redisClient := cache.NewRedisClient(cfg)
db := database.NewDatabase(cfg)
r := api.NewRouter(logger, mongoCol, dbWrapper, redisClient, &ctx, cfg.RequestTimeoutMs)
````

### 3) 运行与验证
- 编译/运行
  - go mod tidy
  - go build ./...
  - go run cmd/server/main.go
- 验证请求超时
  - 将 REQUEST_TIMEOUT_MS 调小（如 100ms），调用慢查询/慢接口观察是否 504/超时返回（业务函数需配合或造压）
- 验证 DB 超时
  - 将 DB_QUERY_TIMEOUT_MS 调小（如 100ms），触发慢 SQL，观察 DB 层超时行为（日志/错误码）
- 验证 Redis 连接
  - 修改 REDIS_* 参数，确认可连接并无阻塞

### 4) Docker/Compose
- 若使用 docker-compose，请在对应服务的 environment 中加入新增变量
- 生产建议：
  - POSTGRES_SSLMODE=require/verify-full（与证书配合）
  - 合理放大连接池与超时，结合流量与数据库资源评估

## 兼容性与注意事项
- API 行为不变；入口调用签名与 Router 签名有小幅变更（需传入 cfg 与 requestTimeoutMs）
- WithContext 会在请求取消或超时时中止 DB 操作，请确保外部调用耐心处理取消错误（context deadline exceeded）
- AutoMigrate 仍启用，生产应使用迁移工具替代
- 生产建议关闭 Swagger 或加认证保护（后续 P0 列表项）

## 回滚指引
- 若需回滚：
  - 将 main 中 NewDatabase/ NewRedisClient/ NewRouter 调用恢复为旧签名
  - 移除 RequestTimeout 中间件注册
  - service 层可不回退 WithContext（不影响功能）

## 后续推荐（下一步可执行）
- JWT 升级到 v4 + RegisteredClaims
- 合法 dummy bcrypt hash（启动时生成一次并缓存）
- 生产环境 Swagger 关闭或加认证
- CI 增加 golangci-lint / govulncheck
- 移除生产 AutoMigrate，改为迁移工具

如果你希望，我可以把以上内容落成 docs/md/P0_STABILITY_CHANGES.md 文件，并创建 .env.example 里补充所有新增配置项，是否需要我直接提交该文档与示例配置？
