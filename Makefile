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

clean:
	docker stop go-rest-api-template
	docker stop dockerPostgres
	docker rm go-rest-api-template
	docker rm dockerPostgres
	docker rm dockerRedis
	docker image rm golang-rest-api-template-backend
	rm -rf .dbdata
