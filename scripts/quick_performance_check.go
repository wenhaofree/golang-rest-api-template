package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	fmt.Println("=== 快速性能检查 ===")
	
	// 检查当前bcrypt配置
	checkBcryptPerformance()
	
	// 检查环境变量
	checkEnvironmentConfig()
	
	// 提供优化建议
	provideOptimizationSuggestions()
}

func checkBcryptPerformance() {
	fmt.Println("\n🔐 bcrypt 性能测试:")
	
	// 获取当前配置的cost
	currentCost := getBcryptCost()
	fmt.Printf("当前配置 BCRYPT_COST: %d\n", currentCost)
	
	// 测试当前cost的性能
	testPassword := "test_password_123"
	
	// 测试哈希生成
	start := time.Now()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testPassword), currentCost)
	hashDuration := time.Since(start)
	
	if err != nil {
		fmt.Printf("❌ 哈希生成失败: %v\n", err)
		return
	}
	
	fmt.Printf("哈希生成时间: %v\n", hashDuration)
	
	// 测试密码验证
	start = time.Now()
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(testPassword))
	compareDuration := time.Since(start)
	
	if err != nil {
		fmt.Printf("❌ 密码验证失败: %v\n", err)
		return
	}
	
	fmt.Printf("密码验证时间: %v\n", compareDuration)
	
	// 性能评估
	if compareDuration < 50*time.Millisecond {
		fmt.Printf("✅ 性能评估: 优秀 (< 50ms)\n")
	} else if compareDuration < 100*time.Millisecond {
		fmt.Printf("⚠️ 性能评估: 良好 (< 100ms)\n")
	} else if compareDuration < 300*time.Millisecond {
		fmt.Printf("⚠️ 性能评估: 一般 (< 300ms)\n")
	} else if compareDuration < 1000*time.Millisecond {
		fmt.Printf("❌ 性能评估: 慢 (< 1s)\n")
	} else {
		fmt.Printf("❌ 性能评估: 很慢 (> 1s)\n")
	}
}

func checkEnvironmentConfig() {
	fmt.Println("\n⚙️ 环境配置检查:")
	
	configs := map[string]string{
		"BCRYPT_COST":    getBcryptCostStr(),
		"GIN_MODE":       os.Getenv("GIN_MODE"),
		"MONGO_ENABLED":  os.Getenv("MONGO_ENABLED"),
		"REDIS_HOST":     os.Getenv("REDIS_HOST"),
		"POSTGRES_HOST":  os.Getenv("POSTGRES_HOST"),
	}
	
	for key, value := range configs {
		if value == "" {
			fmt.Printf("⚠️ %s: 未设置 (使用默认值)\n", key)
		} else {
			fmt.Printf("✅ %s: %s\n", key, value)
		}
	}
}

func provideOptimizationSuggestions() {
	fmt.Println("\n💡 优化建议:")
	
	currentCost := getBcryptCost()
	
	if currentCost >= 14 {
		fmt.Println("❌ 紧急优化: BCRYPT_COST过高!")
		fmt.Println("   建议立即调整为:")
		fmt.Println("   - 开发环境: BCRYPT_COST=10")
		fmt.Println("   - 生产环境: BCRYPT_COST=12")
	} else if currentCost >= 13 {
		fmt.Println("⚠️ 建议优化: BCRYPT_COST较高")
		fmt.Println("   开发环境建议: BCRYPT_COST=10")
	} else if currentCost <= 10 {
		fmt.Println("✅ BCRYPT_COST配置合理")
	}
	
	// MongoDB建议
	mongoEnabled := os.Getenv("MONGO_ENABLED")
	if mongoEnabled == "true" || mongoEnabled == "" {
		fmt.Println("💡 性能提示: 开发环境可考虑设置 MONGO_ENABLED=false")
	}
	
	// 缓存建议
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		fmt.Println("⚠️ 缓存提示: 确保Redis服务正常运行以获得最佳性能")
	}
	
	fmt.Println("\n📋 快速优化步骤:")
	fmt.Println("1. 复制 .env.example 为 .env")
	fmt.Println("2. 设置 BCRYPT_COST=10 (开发环境)")
	fmt.Println("3. 设置 MONGO_ENABLED=false (开发环境)")
	fmt.Println("4. 重启应用服务")
	fmt.Println("5. 运行性能测试验证效果")
}

func getBcryptCost() int {
	costStr := os.Getenv("BCRYPT_COST")
	if costStr == "" {
		return 12 // 默认值
	}
	
	cost, err := strconv.Atoi(costStr)
	if err != nil || cost < 4 || cost > 15 {
		return 12
	}
	
	return cost
}

func getBcryptCostStr() string {
	costStr := os.Getenv("BCRYPT_COST")
	if costStr == "" {
		return "12 (默认)"
	}
	return costStr
}
