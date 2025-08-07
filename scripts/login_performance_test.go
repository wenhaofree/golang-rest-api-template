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

type LoginResponse struct {
	Code    int                    `json:"code"`
	Data    map[string]interface{} `json:"data"`
	Message string                 `json:"message"`
}

func main() {
	baseURL := "http://localhost:8001"
	apiKey := "cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE"

	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "testpassword",
	}

	fmt.Println("=== 登录性能优化验证 ===")
	fmt.Printf("目标: %s/api/v1/login\n", baseURL)
	fmt.Println()

	// 预热请求
	fmt.Println("🔥 预热请求...")
	for i := 0; i < 3; i++ {
		testLogin(baseURL, apiKey, loginReq)
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\n📊 性能测试开始...")

	var times []time.Duration
	successCount := 0
	testCount := 10

	for i := 0; i < testCount; i++ {
		duration, success := testLogin(baseURL, apiKey, loginReq)
		times = append(times, duration)

		if success {
			successCount++
			fmt.Printf("测试 %2d: ✅ %8v", i+1, duration)
		} else {
			fmt.Printf("测试 %2d: ❌ %8v", i+1, duration)
		}

		// 性能评估
		if duration < 100*time.Millisecond {
			fmt.Printf(" (优秀)")
		} else if duration < 200*time.Millisecond {
			fmt.Printf(" (良好)")
		} else if duration < 500*time.Millisecond {
			fmt.Printf(" (一般)")
		} else {
			fmt.Printf(" (需要优化)")
		}
		fmt.Println()

		time.Sleep(50 * time.Millisecond)
	}

	// 计算统计数据
	var totalTime time.Duration
	minTime := times[0]
	maxTime := times[0]

	for _, t := range times {
		totalTime += t
		if t < minTime {
			minTime = t
		}
		if t > maxTime {
			maxTime = t
		}
	}

	avgTime := totalTime / time.Duration(len(times))

	// 计算P95
	sortedTimes := make([]time.Duration, len(times))
	copy(sortedTimes, times)
	for i := 0; i < len(sortedTimes)-1; i++ {
		for j := 0; j < len(sortedTimes)-i-1; j++ {
			if sortedTimes[j] > sortedTimes[j+1] {
				sortedTimes[j], sortedTimes[j+1] = sortedTimes[j+1], sortedTimes[j]
			}
		}
	}
	p95Time := sortedTimes[int(float64(len(sortedTimes))*0.95)]

	fmt.Println("\n=== 测试结果 ===")
	fmt.Printf("总测试次数: %d\n", testCount)
	fmt.Printf("成功次数: %d\n", successCount)
	fmt.Printf("成功率: %.1f%%\n", float64(successCount)/float64(testCount)*100)
	fmt.Printf("最快响应: %v\n", minTime)
	fmt.Printf("最慢响应: %v\n", maxTime)
	fmt.Printf("平均响应: %v\n", avgTime)
	fmt.Printf("P95响应: %v\n", p95Time)

	fmt.Println("\n=== 性能评估 ===")

	// 响应时间评估
	if avgTime < 100*time.Millisecond {
		fmt.Println("✅ 平均响应时间: 优秀 (< 100ms)")
	} else if avgTime < 200*time.Millisecond {
		fmt.Println("⚠️  平均响应时间: 良好 (< 200ms)")
	} else if avgTime < 500*time.Millisecond {
		fmt.Println("⚠️  平均响应时间: 一般 (< 500ms)")
	} else {
		fmt.Println("❌ 平均响应时间: 需要优化 (> 500ms)")
	}

	// P95评估
	if p95Time < 200*time.Millisecond {
		fmt.Println("✅ P95响应时间: 优秀 (< 200ms)")
	} else if p95Time < 400*time.Millisecond {
		fmt.Println("⚠️  P95响应时间: 良好 (< 400ms)")
	} else {
		fmt.Println("❌ P95响应时间: 需要优化 (> 400ms)")
	}

	// 成功率评估
	successRate := float64(successCount) / float64(testCount)
	if successRate >= 0.95 {
		fmt.Println("✅ 成功率: 优秀 (≥ 95%)")
	} else {
		fmt.Println("❌ 成功率: 需要改进 (< 95%)")
	}

	fmt.Println("\n=== 优化建议 ===")
	if avgTime > 200*time.Millisecond {
		fmt.Println("- 考虑降低 BCRYPT_COST 值")
		fmt.Println("- 检查数据库连接性能")
		fmt.Println("- 验证Redis缓存是否正常工作")
	}
	if maxTime > 1*time.Second {
		fmt.Println("- 存在异常慢的请求，需要检查系统资源")
	}
	if successRate < 0.95 {
		fmt.Println("- 检查应用错误日志")
		fmt.Println("- 验证测试数据是否正确")
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

	// 读取响应体
	body, _ := io.ReadAll(resp.Body)

	// 检查响应
	if resp.StatusCode == http.StatusOK {
		var loginResp LoginResponse
		if err := json.Unmarshal(body, &loginResp); err == nil && loginResp.Code == 0 {
			return duration, true
		}
	}

	return duration, false
}
