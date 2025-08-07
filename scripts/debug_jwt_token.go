package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run debug_jwt_token.go <JWT_TOKEN>")
		fmt.Println("示例: go run debug_jwt_token.go eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...")
		return
	}

	token := os.Args[1]
	fmt.Println("=== JWT Token 调试工具 ===")
	fmt.Printf("Token: %s\n\n", token)

	// 分割JWT token
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		fmt.Println("❌ 无效的JWT token格式")
		return
	}

	// 解析Header
	fmt.Println("📋 Header:")
	header, err := decodeJWTPart(parts[0])
	if err != nil {
		fmt.Printf("❌ 解析Header失败: %v\n", err)
	} else {
		fmt.Printf("%s\n\n", header)
	}

	// 解析Payload
	fmt.Println("📋 Payload:")
	payload, err := decodeJWTPart(parts[1])
	if err != nil {
		fmt.Printf("❌ 解析Payload失败: %v\n", err)
	} else {
		fmt.Printf("%s\n\n", payload)
		
		// 解析payload中的具体字段
		var claims map[string]interface{}
		if err := json.Unmarshal([]byte(payload), &claims); err == nil {
			fmt.Println("📊 Claims 详细信息:")
			for key, value := range claims {
				fmt.Printf("  %s: %v\n", key, value)
			}
			
			// 检查关键字段
			fmt.Println("\n🔍 关键字段检查:")
			if iss, ok := claims["iss"]; ok {
				fmt.Printf("  ✅ Issuer (iss): %v\n", iss)
			} else {
				fmt.Printf("  ❌ 缺少 Issuer (iss) 字段\n")
			}
			
			if exp, ok := claims["exp"]; ok {
				fmt.Printf("  ✅ Expiration (exp): %v\n", exp)
			} else {
				fmt.Printf("  ❌ 缺少 Expiration (exp) 字段\n")
			}
			
			if username, ok := claims["username"]; ok {
				fmt.Printf("  ✅ Username: %v\n", username)
			} else {
				fmt.Printf("  ❌ 缺少 Username 字段\n")
			}
		}
	}

	fmt.Println("\n📋 Signature:")
	fmt.Printf("Base64编码的签名: %s\n", parts[2])
}

func decodeJWTPart(part string) (string, error) {
	// JWT使用base64url编码，需要添加padding
	switch len(part) % 4 {
	case 2:
		part += "=="
	case 3:
		part += "="
	}

	// 替换base64url字符
	part = strings.ReplaceAll(part, "-", "+")
	part = strings.ReplaceAll(part, "_", "/")

	decoded, err := base64.StdEncoding.DecodeString(part)
	if err != nil {
		return "", err
	}

	// 格式化JSON
	var jsonData interface{}
	if err := json.Unmarshal(decoded, &jsonData); err != nil {
		return string(decoded), nil // 如果不是JSON，直接返回原始字符串
	}

	formatted, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return string(decoded), nil
	}

	return string(formatted), nil
}
