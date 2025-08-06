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
1. 统一返回标准
2. 日志出入参
3. db时间统一
4. 