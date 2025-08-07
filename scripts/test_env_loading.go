package main

import (
	"fmt"
	"golang-rest-api-template/pkg/config"
	"os"
)

func main() {
	fmt.Println("=== 环境变量配置加载测试 ===")
	fmt.Println()

	// 显示当前工作目录
	pwd, _ := os.Getwd()
	fmt.Printf("当前目录: %s\n", pwd)
	fmt.Println()

	// 检查.env文件是否存在
	envFiles := []string{".env", ".env.development", ".env.production", ".env.test", ".env.local"}
	fmt.Println("📁 .env文件检查:")
	for _, file := range envFiles {
		if _, err := os.Stat(file); err == nil {
			fmt.Printf("  ✅ %s 存在\n", file)
		} else {
			fmt.Printf("  ❌ %s 不存在\n", file)
		}
	}
	fmt.Println()

	// 加载配置
	fmt.Println("🔧 加载配置...")
	cfg := config.LoadConfig()
	fmt.Println()

	// 显示配置信息
	fmt.Println("📋 当前配置:")
	fmt.Printf("  数据库主机: %s\n", cfg.PostgresHost)
	fmt.Printf("  数据库名: %s\n", cfg.PostgresDB)
	fmt.Printf("  数据库用户: %s\n", cfg.PostgresUser)
	fmt.Printf("  数据库端口: %s\n", cfg.PostgresPort)
	fmt.Printf("  Redis主机: %s\n", cfg.RedisHost)
	fmt.Printf("  MongoDB启用: %t\n", cfg.MongoEnabled)
	if cfg.MongoEnabled {
		fmt.Printf("  MongoDB URI: %s\n", cfg.MongoURI)
		fmt.Printf("  MongoDB数据库: %s\n", cfg.MongoDB)
		fmt.Printf("  MongoDB集合: %s\n", cfg.MongoLogCollection)
	}
	fmt.Printf("  JWT密钥长度: %d\n", len(cfg.JWTSecretKey))
	fmt.Printf("  API密钥长度: %d\n", len(cfg.APISecretKey))
	fmt.Println()

	// 显示一些关键环境变量
	fmt.Println("🌍 关键环境变量:")
	envVars := []string{"GIN_MODE", "MONGO_ENABLED", "POSTGRES_DB", "PORT"}
	for _, envVar := range envVars {
		value := os.Getenv(envVar)
		if value != "" {
			fmt.Printf("  %s = %s\n", envVar, value)
		} else {
			fmt.Printf("  %s = (未设置)\n", envVar)
		}
	}
	fmt.Println()

	fmt.Println("✅ 配置加载测试完成")
}