package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("=== 速率限制功能测试 ===")
	fmt.Println("注意：需要先启动服务器才能运行此测试")
	fmt.Println("启动命令：RATE_LIMIT_REQUESTS=3 RATE_LIMIT_WINDOW=10 go run cmd/server/main.go")
	fmt.Println()
	
	baseURL := "http://localhost:8000/api/v1/"
	
	fmt.Printf("测试URL: %s\n", baseURL)
	fmt.Println("配置：3 requests per 10 seconds")
	fmt.Println()
	
	// 发送多个请求测试速率限制
	for i := 1; i <= 6; i++ {
		fmt.Printf("发送第 %d 个请求... ", i)
		
		resp, err := http.Get(baseURL)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			continue
		}
		
		fmt.Printf("状态码: %d", resp.StatusCode)
		
		if resp.StatusCode == 429 {
			fmt.Printf(" (速率限制生效!) 🚫")
		} else if resp.StatusCode == 200 {
			fmt.Printf(" (请求成功) ✅")
		}
		
		resp.Body.Close()
		fmt.Println()
		
		// 短暂延迟
		time.Sleep(1 * time.Second)
	}
	
	fmt.Println()
	fmt.Println("=== 测试说明 ===")
	fmt.Println("- 前3个请求应该返回200状态码")
	fmt.Println("- 后续请求应该返回429状态码")
	fmt.Println("- 等待10秒后速率限制会重置")
	fmt.Println()
	fmt.Println("如果看到429错误，说明速率限制配置正常工作！")
}
