package main

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "testpassword123"

	fmt.Println("=== bcrypt 性能测试 ===")
	fmt.Println()

	// 测试不同的cost值
	costs := []int{10, 11, 12, 13, 14, 15}

	for _, cost := range costs {
		fmt.Printf("测试 cost = %d:\n", cost)

		// 测试密码哈希生成时间
		start := time.Now()
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
		hashDuration := time.Since(start)

		if err != nil {
			fmt.Printf("  ❌ 哈希生成失败: %v\n", err)
			continue
		}

		fmt.Printf("  哈希生成时间: %v\n", hashDuration)

		// 测试密码验证时间
		start = time.Now()
		err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
		compareDuration := time.Since(start)

		if err != nil {
			fmt.Printf("  ❌ 密码验证失败: %v\n", err)
		} else {
			fmt.Printf("  密码验证时间: %v\n", compareDuration)
		}

		// 性能评估
		if compareDuration < 50*time.Millisecond {
			fmt.Printf("  评估: ✅ 优秀 (< 50ms)\n")
		} else if compareDuration < 100*time.Millisecond {
			fmt.Printf("  评估: ⚠️  良好 (< 100ms)\n")
		} else if compareDuration < 500*time.Millisecond {
			fmt.Printf("  评估: ⚠️  一般 (< 500ms)\n")
		} else {
			fmt.Printf("  评估: ❌ 慢 (> 500ms)\n")
		}

		fmt.Println()
	}

	fmt.Println("=== 推荐配置 ===")
	fmt.Println("开发环境: cost = 10 (快速开发)")
	fmt.Println("测试环境: cost = 11 (平衡性能)")
	fmt.Println("生产环境: cost = 12 (安全性优先)")
	fmt.Println("高安全环境: cost = 13 (最高安全性)")
	fmt.Println()
	fmt.Println("注意: cost = 14 对于Web应用来说过高，会严重影响用户体验")
}
