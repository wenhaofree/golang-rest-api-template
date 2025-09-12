setup:
	go get -u github.com/swaggo/swag/cmd/swag
	go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g ./cmd/server/main.go -o ./docs
	go get -u github.com/swaggo/gin-swagger
	go get -u github.com/swaggo/files

build-docker:
	docker compose build --no-cache

run-local:
	docker start dockerPostgres
	docker start dockerRedis
	docker start dockerMongo
	export REDIS_HOST=localhost
	export POSTGRES_DB=go_app_dev
	export POSTGRES_USER=docker
	export POSTGRES_PASSWORD=password
	export POSTGRES_PORT=5435
	export JWT_SECRET_KEY=ObL89O3nOSSEj6tbdHako0cXtPErzBUfq8l8o/3KD9g=INSECURE
	export API_SECRET_KEY=cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE
	export POSTGRES_HOST=localhost
	export MONGO_ENABLED=true
	export MONGO_URI=mongodb://localhost:27017
	go run cmd/server/main.go

run-local-no-mongo:
	docker start dockerPostgres
	docker start dockerRedis
	export REDIS_HOST=localhost
	export POSTGRES_DB=go_app_dev
	export POSTGRES_USER=docker
	export POSTGRES_PASSWORD=password
	export POSTGRES_PORT=5435
	export JWT_SECRET_KEY=ObL89O3nOSSEj6tbdHako0cXtPErzBUfq8l8o/3KD9g=INSECURE
	export API_SECRET_KEY=cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE
	export POSTGRES_HOST=localhost
	export MONGO_ENABLED=false
	go run cmd/server/main.go

# 使用.env文件启动（推荐方式）
run-env:
	@echo "🚀 使用.env文件启动应用..."
	@./scripts/start.sh

# 使用特定环境的.env文件启动
run-dev:
	@echo "🚀 使用开发环境配置启动..."
	@if [ -f .env.development ]; then \
		echo "📋 加载.env.development配置"; \
	fi
	@GIN_MODE=debug go run cmd/server/main.go

run-prod:
	@echo "🚀 使用生产环境配置启动..."
	@if [ -f .env.production ]; then \
		echo "📋 加载.env.production配置"; \
	fi
	@GIN_MODE=release go run cmd/server/main.go

run-test:
	@echo "🚀 使用测试环境配置启动..."
	@if [ -f .env.test ]; then \
		echo "📋 加载.env.test配置"; \
	fi
	@GIN_MODE=test go run cmd/server/main.go

up:
	docker compose up

up-no-mongo:
	docker compose -f docker-compose.no-mongo.yml up

down:
	docker compose down

down-no-mongo:
	docker compose -f docker-compose.no-mongo.yml down

restart:
	docker compose restart

restart-no-mongo:
	docker compose -f docker-compose.no-mongo.yml restart

build:
	go build -v ./...

test:
	go test -v ./... -race -cover

test-config:
	@echo "Testing MongoDB configuration parsing..."
	@./scripts/test_mongo_config.sh

# 性能测试
test-performance:
	@echo "🚀 Running login API performance test..."
	@go run scripts/performance_test.go

# 简单登录测试
test-login:
	@echo "🔐 Running simple login test..."
	@go run scripts/simple_login_test.go

# 登录性能测试
test-login-performance:
	@echo "⚡ Running login performance test..."
	@go run scripts/login_performance_test.go

# bcrypt性能测试
test-bcrypt:
	@echo "🔐 Running bcrypt benchmark..."
	@go run scripts/bcrypt_benchmark.go

# JWT性能测试
test-jwt:
	@echo "🎫 Running JWT benchmark..."
	@go run scripts/jwt_benchmark.go

# 快速性能检查
check-performance:
	@echo "⚡ Quick performance check..."
	@go run scripts/quick_performance_check.go

# 完整性能测试套件
test-all-performance:
	@echo "🚀 Running complete performance test suite..."
	@echo "1. Checking bcrypt performance..."
	@go run scripts/bcrypt_benchmark.go
	@echo "\n2. Checking JWT performance..."
	@go run scripts/jwt_benchmark.go
	@echo "\n3. Running login performance test..."
	@go run scripts/login_performance_test.go
	@echo "\n4. Quick performance summary..."
	@go run scripts/quick_performance_check.go

# JWT调试工具
debug-jwt:
	@echo "🔍 JWT Token 调试工具"
	@echo "用法: make debug-jwt TOKEN=<your_jwt_token>"
	@if [ -z "$(TOKEN)" ]; then \
		echo "请提供JWT token: make debug-jwt TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."; \
	else \
		go run scripts/debug_jwt_token.go $(TOKEN); \
	fi

# JWT有效期测试
test-jwt-expiry:
	@echo "⏰ JWT Token 有效期测试..."
	@go run scripts/test_jwt_expiry.go

# JWT有效期测试（带token解析）
test-jwt-expiry-with-token:
	@echo "⏰ JWT Token 有效期测试（带token解析）..."
	@echo "用法: make test-jwt-expiry-with-token TOKEN=<your_jwt_token>"
	@if [ -z "$(TOKEN)" ]; then \
		echo "请提供JWT token: make test-jwt-expiry-with-token TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."; \
		go run scripts/test_jwt_expiry.go; \
	else \
		go run scripts/test_jwt_expiry.go $(TOKEN); \
	fi

# 测试个人资料接口
test-profile:
	@echo "👤 测试个人资料接口..."
	@go run scripts/test_profile_api.go

# 测试书籍API
test-books:
	@echo "📚 测试书籍API..."
	@go run scripts/test_books_api.go

# 完整的API测试流程
test-api-flow:
	@echo "🚀 完整API测试流程..."
	@echo "1. 测试登录接口..."
	@make test-login
	@echo "\n2. 测试个人资料接口..."
	@make test-profile
	@echo "\n3. 测试书籍API..."
	@make test-books
	@echo "\n✅ API测试流程完成"

# 优化后的登录性能测试
test-optimized-login:
	@echo "🚀 优化后的登录性能测试..."
	@go run scripts/optimized_login_test.go

# 数据库索引优化
db-optimize:
	@echo "📊 Adding performance indexes to database..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "Please set DB_URL environment variable"; \
		echo "Example: make db-optimize DB_URL='postgres://user:pass@localhost:5432/dbname'"; \
		exit 1; \
	fi
	@psql $(DB_URL) -f scripts/add_performance_indexes.sql

# 数据库性能分析
db-analyze:
	@echo "📈 Analyzing database performance..."
	@if [ -z "$(DB_URL)" ]; then \
		echo "Please set DB_URL environment variable"; \
		exit 1; \
	fi
	@psql $(DB_URL) -c "ANALYZE users;"
	@echo "✅ Database analysis complete"

clean:
	docker stop go-rest-api-template
	docker stop dockerPostgres
	docker rm go-rest-api-template
	docker rm dockerPostgres
	docker rm dockerRedis
	docker image rm golang-rest-api-template-backend
	rm -rf .dbdata
