package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	// 配置
	baseURL := "http://localhost:8001"
	apiKey := "cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE"

	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "testpassword",
	}

	fmt.Println("=== 简单登录性能测试 ===")
	fmt.Printf("目标: %s/api/v1/login\n", baseURL)
	fmt.Println()

	// 执行多次测试
	var totalTime time.Duration
	successCount := 0
	testCount := 10

	for i := 0; i < testCount; i++ {
		duration, success := testLogin(baseURL, apiKey, loginReq)
		totalTime += duration

		if success {
			successCount++
			fmt.Printf("测试 %d: ✅ %v\n", i+1, duration)
		} else {
			fmt.Printf("测试 %d: ❌ %v\n", i+1, duration)
		}

		// 间隔一下，避免过于频繁
		time.Sleep(100 * time.Millisecond)
	}

	// 统计结果
	fmt.Println()
	fmt.Println("=== 测试结果 ===")
	fmt.Printf("总测试次数: %d\n", testCount)
	fmt.Printf("成功次数: %d\n", successCount)
	fmt.Printf("成功率: %.1f%%\n", float64(successCount)/float64(testCount)*100)
	fmt.Printf("平均响应时间: %v\n", totalTime/time.Duration(testCount))
	fmt.Printf("总耗时: %v\n", totalTime)

	// 性能评估
	avgTime := totalTime / time.Duration(testCount)
	fmt.Println()
	fmt.Println("=== 性能评估 ===")
	if avgTime < 50*time.Millisecond {
		fmt.Println("✅ 响应时间: 优秀 (< 50ms)")
	} else if avgTime < 100*time.Millisecond {
		fmt.Println("⚠️  响应时间: 良好 (< 100ms)")
	} else if avgTime < 200*time.Millisecond {
		fmt.Println("⚠️  响应时间: 一般 (< 200ms)")
	} else {
		fmt.Println("❌ 响应时间: 需要优化 (> 200ms)")
	}

	if float64(successCount)/float64(testCount) >= 0.95 {
		fmt.Println("✅ 成功率: 优秀 (≥ 95%)")
	} else {
		fmt.Println("❌ 成功率: 需要改进 (< 95%)")
	}
}

func testLogin(baseURL, apiKey string, loginReq LoginRequest) (time.Duration, bool) {
	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return 0, false
	}

	req, err := http.NewRequest("POST", baseURL+"/api/v1/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return duration, false
	}
	defer resp.Body.Close()

	// 读取响应体避免连接泄漏
	io.Copy(io.Discard, resp.Body)

	return duration, resp.StatusCode == http.StatusOK
}
