# 使用说明：
## 本地终端开发运行：
1. 搭建postgresql，MongoDB，Redis

2. export 执行配置：
   export REDIS_HOST=localhost
   export POSTGRES_HOST=localhost
   export POSTGRES_DB=goapi_db
   export POSTGRES_USER=root
   export POSTGRES_PASSWORD=fuwenhao
   export POSTGRES_PORT=5432
   export JWT_SECRET_KEY=ObL89O3nOSSEj6tbdHako0cXtPErzBUfq8l8o/3KD9g=INSECURE
   export API_SECRET_KEY=cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE

3. 运行：
   go run cmd/server/main.go
4. 文档查看：
   http://127.0.0.1:8001/swagger/index.html#/

## 开发说明：

# TODO：
1. 统一返回标准-✅
2. 日志路由出入参
3. db时间统一
4. 用户表支持第三方登录信息字段+时间字段统一；✅
5. 定位APP后端支持接口实现；


# 其他说明：
我来帮你解释一下 `vendor/` 目录的作用和管理方式。首先让我查看一下这个目录的内容和项目的配置。

基于我对你的项目的分析，我来详细解释一下 `vendor/` 目录：

## vendor/ 目录是什么？

`vendor/` 目录是 Go 语言的**依赖管理机制**，它包含了你项目所有第三方依赖包的完整源代码副本。从你的项目中可以看到：

1. **包含的依赖**：所有在 `go.mod` 文件中声明的依赖包，包括：
   - 直接依赖（如 `gin-gonic/gin`、`go-redis/redis` 等）
   - 间接依赖（标记为 `// indirect` 的包）

2. **结构组织**：按照 Go 的导入路径组织，如：
   - `github.com/gin-gonic/gin`
   - `go.mongodb.org/mongo-driver`
   - `gorm.io/gorm`

## vendor/ 目录的作用

1. **依赖隔离**：确保项目使用特定版本的依赖，不受全局 Go 模块缓存影响
2. **构建可重现性**：保证在不同环境中构建结果一致
3. **离线构建**：无需网络连接即可构建项目
4. **版本锁定**：防止依赖包版本意外变更

## 是否需要 Git 提交？

**通常情况下不建议提交 vendor/ 目录**，原因如下：

1. **你的项目已经正确配置**：`.gitignore` 文件第16行已经排除了 vendor/ 目录
2. **文件体积巨大**：vendor/ 包含大量第三方代码，会显著增加仓库大小
3. **维护困难**：每次依赖更新都会产生大量文件变更
4. **现代 Go 工具链**：Go 1.11+ 的模块系统已经很好地解决了依赖管理问题

## 可以删除吗？

**可以安全删除**，因为：

1. **可重新生成**：运行 `go mod vendor` 命令可以重新创建
2. **不影响构建**：Go 工具链会自动从模块缓存或网络获取依赖
3. **已被忽略**：你的 `.gitignore` 已经排除了这个目录

## 建议的管理方式

1. **保持 .gitignore 配置**：继续忽略 vendor/ 目录
2. **依赖 go.mod 和 go.sum**：这两个文件已经足够确保依赖的可重现性
3. **按需生成**：只在特殊情况下（如需要离线构建）才运行 `go mod vendor`
4. **清理命令**：可以运行以下命令清理：
   ```bash
   rm -rf vendor/
   go mod tidy  # 清理不需要的依赖
   ```

总结：vendor/ 目录是 Go 的依赖管理工具，你的项目配置是正确的（已在 .gitignore 中排除），可以安全删除，不需要提交到 Git。
