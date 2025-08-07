package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// LoginRequest 登录请求结构
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TestResult 测试结果
type TestResult struct {
	TotalRequests   int
	SuccessRequests int
	FailedRequests  int
	TotalTime       time.Duration
	AverageTime     time.Duration
	MinTime         time.Duration
	MaxTime         time.Duration
	RequestTimes    []time.Duration
}

// PerformanceTest 性能测试结构
type PerformanceTest struct {
	BaseURL    string
	APIKey     string
	Email      string
	Password   string
	Concurrent int
	Requests   int
}

func main() {
	// 配置测试参数
	test := PerformanceTest{
		BaseURL:    "http://localhost:8001",
		APIKey:     "cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE", // 从.env文件获取
		Email:      "test@example.com",
		Password:   "testpassword",
		Concurrent: 10,  // 并发数
		Requests:   100, // 总请求数
	}

	fmt.Println("=== 登录接口性能测试 ===")
	fmt.Printf("目标URL: %s/api/v1/login\n", test.BaseURL)
	fmt.Printf("并发数: %d\n", test.Concurrent)
	fmt.Printf("总请求数: %d\n", test.Requests)
	fmt.Println()

	// 执行性能测试
	result := test.RunTest()

	// 输出结果
	test.PrintResults(result)
}

// RunTest 执行性能测试
func (pt *PerformanceTest) RunTest() TestResult {
	var wg sync.WaitGroup
	var mu sync.Mutex

	result := TestResult{
		TotalRequests: pt.Requests,
		RequestTimes:  make([]time.Duration, 0, pt.Requests),
		MinTime:       time.Hour, // 初始化为一个大值
	}

	startTime := time.Now()

	// 创建信号量控制并发数
	semaphore := make(chan struct{}, pt.Concurrent)

	for i := 0; i < pt.Requests; i++ {
		wg.Add(1)
		go func(requestID int) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 执行单个请求
			duration, success := pt.makeLoginRequest()

			// 更新结果（需要加锁）
			mu.Lock()
			result.RequestTimes = append(result.RequestTimes, duration)
			if success {
				result.SuccessRequests++
			} else {
				result.FailedRequests++
			}

			// 更新最小和最大时间
			if duration < result.MinTime {
				result.MinTime = duration
			}
			if duration > result.MaxTime {
				result.MaxTime = duration
			}
			mu.Unlock()

		}(i)
	}

	wg.Wait()
	result.TotalTime = time.Since(startTime)

	// 计算平均时间
	if len(result.RequestTimes) > 0 {
		var total time.Duration
		for _, t := range result.RequestTimes {
			total += t
		}
		result.AverageTime = total / time.Duration(len(result.RequestTimes))
	}

	return result
}

// makeLoginRequest 执行单个登录请求
func (pt *PerformanceTest) makeLoginRequest() (time.Duration, bool) {
	loginReq := LoginRequest{
		Email:    pt.Email,
		Password: pt.Password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return 0, false
	}

	req, err := http.NewRequest("POST", pt.BaseURL+"/api/v1/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", pt.APIKey)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return duration, false
	}
	defer resp.Body.Close()

	// 读取响应体（避免连接泄漏）
	io.Copy(io.Discard, resp.Body)

	// 检查状态码
	success := resp.StatusCode == http.StatusOK

	return duration, success
}

// PrintResults 打印测试结果
func (pt *PerformanceTest) PrintResults(result TestResult) {
	fmt.Println("=== 测试结果 ===")
	fmt.Printf("总请求数: %d\n", result.TotalRequests)
	fmt.Printf("成功请求: %d\n", result.SuccessRequests)
	fmt.Printf("失败请求: %d\n", result.FailedRequests)
	fmt.Printf("成功率: %.2f%%\n", float64(result.SuccessRequests)/float64(result.TotalRequests)*100)
	fmt.Println()

	fmt.Println("=== 时间统计 ===")
	fmt.Printf("总耗时: %v\n", result.TotalTime)
	fmt.Printf("平均响应时间: %v\n", result.AverageTime)
	fmt.Printf("最快响应时间: %v\n", result.MinTime)
	fmt.Printf("最慢响应时间: %v\n", result.MaxTime)
	fmt.Printf("QPS (每秒请求数): %.2f\n", float64(result.TotalRequests)/result.TotalTime.Seconds())
	fmt.Println()

	// 计算百分位数
	if len(result.RequestTimes) > 0 {
		pt.printPercentiles(result.RequestTimes)
	}

	// 性能评估
	pt.evaluatePerformance(result)
}

// printPercentiles 打印百分位数统计
func (pt *PerformanceTest) printPercentiles(times []time.Duration) {
	// 简单排序（冒泡排序，适用于小数据集）
	for i := 0; i < len(times)-1; i++ {
		for j := 0; j < len(times)-i-1; j++ {
			if times[j] > times[j+1] {
				times[j], times[j+1] = times[j+1], times[j]
			}
		}
	}

	fmt.Println("=== 百分位数统计 ===")
	fmt.Printf("P50 (中位数): %v\n", times[len(times)*50/100])
	fmt.Printf("P90: %v\n", times[len(times)*90/100])
	fmt.Printf("P95: %v\n", times[len(times)*95/100])
	fmt.Printf("P99: %v\n", times[len(times)*99/100])
	fmt.Println()
}

// evaluatePerformance 性能评估
func (pt *PerformanceTest) evaluatePerformance(result TestResult) {
	fmt.Println("=== 性能评估 ===")

	successRate := float64(result.SuccessRequests) / float64(result.TotalRequests) * 100
	avgTimeMs := float64(result.AverageTime.Nanoseconds()) / 1000000

	// 成功率评估
	if successRate >= 99 {
		fmt.Println("✅ 成功率: 优秀")
	} else if successRate >= 95 {
		fmt.Println("⚠️  成功率: 良好")
	} else {
		fmt.Println("❌ 成功率: 需要改进")
	}

	// 响应时间评估
	if avgTimeMs <= 50 {
		fmt.Println("✅ 响应时间: 优秀")
	} else if avgTimeMs <= 100 {
		fmt.Println("⚠️  响应时间: 良好")
	} else if avgTimeMs <= 200 {
		fmt.Println("⚠️  响应时间: 一般")
	} else {
		fmt.Println("❌ 响应时间: 需要优化")
	}

	// QPS评估
	qps := float64(result.TotalRequests) / result.TotalTime.Seconds()
	if qps >= 100 {
		fmt.Println("✅ QPS: 优秀")
	} else if qps >= 50 {
		fmt.Println("⚠️  QPS: 良好")
	} else {
		fmt.Println("❌ QPS: 需要优化")
	}

	fmt.Println()
	fmt.Println("=== 优化建议 ===")
	if avgTimeMs > 100 {
		fmt.Println("- 考虑添加缓存机制")
		fmt.Println("- 优化数据库查询")
		fmt.Println("- 检查网络延迟")
	}
	if successRate < 99 {
		fmt.Println("- 检查错误日志")
		fmt.Println("- 增加错误处理")
		fmt.Println("- 检查系统资源")
	}
	if qps < 50 {
		fmt.Println("- 考虑增加并发处理能力")
		fmt.Println("- 优化代码性能")
		fmt.Println("- 检查系统瓶颈")
	}
}
