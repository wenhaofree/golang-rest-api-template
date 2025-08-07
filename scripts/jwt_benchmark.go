package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

func main() {
	fmt.Println("=== JWT 性能测试 ===")
	fmt.Println()

	// 设置JWT密钥
	jwtKey := []byte("test-jwt-secret-key-for-performance-testing")

	// 测试JWT生成性能
	fmt.Println("📊 JWT生成性能测试:")

	testCount := 1000
	username := "test@example.com"

	start := time.Now()

	for i := 0; i < testCount; i++ {
		// 模拟JWT生成过程
		expirationTime := time.Now().Add(5 * time.Minute).Unix()

		claims := &jwt.StandardClaims{
			ExpiresAt: expirationTime,
			Issuer:    username,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		_, err := token.SignedString(jwtKey)

		if err != nil {
			fmt.Printf("❌ JWT生成失败: %v\n", err)
			return
		}
	}

	totalTime := time.Since(start)
	avgTime := totalTime / time.Duration(testCount)

	fmt.Printf("总测试次数: %d\n", testCount)
	fmt.Printf("总耗时: %v\n", totalTime)
	fmt.Printf("平均每次: %v\n", avgTime)
	fmt.Printf("每秒生成: %.0f tokens\n", float64(testCount)/totalTime.Seconds())

	// 性能评估
	if avgTime < 1*time.Millisecond {
		fmt.Println("✅ JWT生成性能: 优秀 (< 1ms)")
	} else if avgTime < 5*time.Millisecond {
		fmt.Println("⚠️  JWT生成性能: 良好 (< 5ms)")
	} else {
		fmt.Println("❌ JWT生成性能: 需要优化 (> 5ms)")
	}

	fmt.Println()

	// 测试单次JWT生成时间
	fmt.Println("🔍 单次JWT生成详细测试:")

	for i := 0; i < 10; i++ {
		start := time.Now()

		expirationTime := time.Now().Add(5 * time.Minute).Unix()
		claims := &jwt.StandardClaims{
			ExpiresAt: expirationTime,
			Issuer:    username,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)

		duration := time.Since(start)

		if err != nil {
			fmt.Printf("测试 %d: ❌ 失败 - %v\n", i+1, err)
		} else {
			fmt.Printf("测试 %d: ✅ %v (token长度: %d)\n", i+1, duration, len(tokenString))
		}
	}

	fmt.Println()
	fmt.Println("=== 结论 ===")
	fmt.Println("JWT生成通常不是性能瓶颈，每次生成时间通常 < 1ms")
	fmt.Println("如果JWT生成时间异常，请检查:")
	fmt.Println("- JWT密钥长度是否合适")
	fmt.Println("- 系统CPU负载情况")
	fmt.Println("- 内存使用情况")
}

func init() {
	// 设置环境变量，避免从空环境变量读取
	if os.Getenv("JWT_SECRET_KEY") == "" {
		os.Setenv("JWT_SECRET_KEY", "test-jwt-secret-key-for-performance-testing")
	}
}
