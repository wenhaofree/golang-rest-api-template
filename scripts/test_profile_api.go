package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Token string      `json:"token"`
		User  interface{} `json:"user"`
	} `json:"data"`
	Message string `json:"message"`
}

type ProfileResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func main() {
	// 配置
	baseURL := getEnv("BASE_URL", "http://localhost:8001")
	apiKey := getEnv("API_KEY", "cJGZ8L1sDcPezjOy1zacPJZxzZxrPObm2Ggs1U0V+fE=INSECURE")
	testEmail := getEnv("TEST_EMAIL", "user@privaterelay.appleid.com")
	testPassword := getEnv("TEST_PASSWORD", "password123")

	fmt.Println("=== 个人资料接口测试 ===")
	fmt.Printf("测试地址: %s\n", baseURL)
	fmt.Printf("测试用户: %s\n", testEmail)
	fmt.Println()

	// 步骤1：登录获取JWT token
	fmt.Println("🔐 步骤1: 用户登录...")
	token, err := loginAndGetToken(baseURL, apiKey, testEmail, testPassword)
	if err != nil {
		fmt.Printf("❌ 登录失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 登录成功，获取到token: %s...\n", token[:50])
	fmt.Println()

	// 步骤2：调试JWT token
	fmt.Println("🔍 步骤2: 调试JWT token...")
	debugJWTToken(token)
	fmt.Println()

	// 步骤3：测试个人资料接口
	fmt.Println("👤 步骤3: 获取个人资料...")
	err = testProfileAPI(baseURL, apiKey, token)
	if err != nil {
		fmt.Printf("❌ 个人资料接口测试失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 个人资料接口测试成功\n")
}

func loginAndGetToken(baseURL, apiKey, email, password string) (string, error) {
	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	jsonData, _ := json.Marshal(loginReq)
	req, err := http.NewRequest("POST", baseURL+"/api/v1/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("登录失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("解析登录响应失败: %v", err)
	}

	if !loginResp.Success || loginResp.Data.Token == "" {
		return "", fmt.Errorf("登录响应无效: %s", string(body))
	}

	return loginResp.Data.Token, nil
}

func debugJWTToken(token string) {
	fmt.Printf("Token: %s\n", token)
	
	// 这里可以添加JWT token解析逻辑
	// 为了简化，我们只显示token的基本信息
	if len(token) > 100 {
		fmt.Printf("Token长度: %d (正常)\n", len(token))
	} else {
		fmt.Printf("⚠️ Token长度: %d (可能异常)\n", len(token))
	}
}

func testProfileAPI(baseURL, apiKey, token string) error {
	req, err := http.NewRequest("GET", baseURL+"/api/v1/profile", nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))

	if resp.StatusCode != 200 {
		return fmt.Errorf("个人资料接口返回错误状态码: %d", resp.StatusCode)
	}

	var profileResp ProfileResponse
	if err := json.Unmarshal(body, &profileResp); err != nil {
		return fmt.Errorf("解析个人资料响应失败: %v", err)
	}

	if !profileResp.Success {
		return fmt.Errorf("个人资料接口返回失败: %s", profileResp.Message)
	}

	fmt.Printf("✅ 用户信息: %+v\n", profileResp.Data)
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
