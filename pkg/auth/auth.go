package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// Claims struct to be encoded to JWT
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// RefreshClaims struct for refresh tokens
type RefreshClaims struct {
	Username string `json:"username"`
	TokenID  string `json:"token_id"`
	jwt.StandardClaims
}

var JwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

// getBcryptCost 获取bcrypt cost配置，默认为12
func getBcryptCost() int {
	costStr := os.Getenv("BCRYPT_COST")
	if costStr == "" {
		return 12 // 默认cost，平衡安全性和性能
	}

	cost, err := strconv.Atoi(costStr)
	if err != nil || cost < 4 || cost > 15 {
		return 12 // 无效值时使用默认值
	}

	return cost
}

func HashPassword(password string) (string, error) {
	cost := getBcryptCost()
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// GetDummyHash 获取用于防时序攻击的dummy hash
func GetDummyHash() string {
	cost := getBcryptCost()
	// 生成一个与当前cost匹配的dummy hash
	switch cost {
	case 10:
		return "$2a$10$dummy.hash.to.prevent.timing.attacks.abcdefghijklmnopqrstuvwxyz"
	case 11:
		return "$2a$11$dummy.hash.to.prevent.timing.attacks.abcdefghijklmnopqrstuvwxyz"
	case 12:
		return "$2a$12$dummy.hash.to.prevent.timing.attacks.abcdefghijklmnopqrstuvwxyz"
	case 13:
		return "$2a$13$dummy.hash.to.prevent.timing.attacks.abcdefghijklmnopqrstuvwxyz"
	default:
		return "$2a$12$dummy.hash.to.prevent.timing.attacks.abcdefghijklmnopqrstuvwxyz"
	}
}

// getJWTExpiryDuration 获取JWT token有效期配置
func getJWTExpiryDuration() time.Duration {
	durationStr := os.Getenv("JWT_EXPIRY_DURATION")
	if durationStr == "" {
		return 24 * time.Hour // 默认24小时
	}

	duration, err := parseDuration(durationStr)
	if err != nil {
		// 如果解析失败，使用默认值24小时
		return 24 * time.Hour
	}

	// 限制最小值为5分钟，最大值为7天
	if duration < 5*time.Minute {
		return 5 * time.Minute
	}
	if duration > 7*24*time.Hour {
		return 7 * 24 * time.Hour
	}

	return duration
}

// parseDuration 解析时间字符串，支持更多格式
func parseDuration(s string) (time.Duration, error) {
	// 先尝试标准的time.ParseDuration
	if duration, err := time.ParseDuration(s); err == nil {
		return duration, nil
	}

	// 支持天数格式 (如 "1d", "7d")
	if len(s) >= 2 && s[len(s)-1] == 'd' {
		if days, err := strconv.Atoi(s[:len(s)-1]); err == nil {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}

	return 0, fmt.Errorf("invalid duration format: %s", s)
}

func GenerateToken(username string) (string, error) {
	// 获取可配置的过期时间
	expiryDuration := getJWTExpiryDuration()
	expirationTime := time.Now().Add(expiryDuration).Unix()

	// Create the JWT claims, which includes the username and expiration time
	claims := &jwt.StandardClaims{
		// In JWT, the expiry time is expressed as unix milliseconds
		ExpiresAt: expirationTime,
		Issuer:    username,
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create the JWT string
	tokenString, err := token.SignedString(JwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateRandomKey() string {
	key := make([]byte, 32) // generate a 256 bit key
	_, err := rand.Read(key)
	if err != nil {
		panic("Failed to generate random key: " + err.Error())
	}

	return base64.StdEncoding.EncodeToString(key)
}

// getRefreshTokenExpiryDuration 获取refresh token有效期配置
func getRefreshTokenExpiryDuration() time.Duration {
	durationStr := os.Getenv("REFRESH_TOKEN_EXPIRY_DURATION")
	if durationStr == "" {
		return 7 * 24 * time.Hour // 默认7天
	}

	duration, err := parseDuration(durationStr)
	if err != nil {
		// 如果解析失败，使用默认值7天
		return 7 * 24 * time.Hour
	}

	// 限制最小值为1天，最大值为30天
	if duration < 24*time.Hour {
		return 24 * time.Hour
	}
	if duration > 30*24*time.Hour {
		return 30 * 24 * time.Hour
	}

	return duration
}

// GenerateRefreshToken 生成refresh token
func GenerateRefreshToken(username string) (string, string, error) {
	// 生成唯一的token ID
	tokenID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s_%d", username, time.Now().UnixNano())))
	
	// 获取可配置的过期时间
	expiryDuration := getRefreshTokenExpiryDuration()
	expirationTime := time.Now().Add(expiryDuration).Unix()

	// Create the refresh token claims
	claims := &RefreshClaims{
		Username: username,
		TokenID:  tokenID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime,
			Issuer:    "refresh",
			Subject:   username,
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create the refresh token string
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		return "", "", err
	}

	return tokenString, tokenID, nil
}

// ValidateRefreshToken 验证refresh token并返回用户信息
func ValidateRefreshToken(tokenString string) (string, string, error) {
	claims := &RefreshClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JwtKey, nil
	})

	if err != nil {
		return "", "", err
	}

	if !token.Valid {
		return "", "", fmt.Errorf("invalid refresh token")
	}

	// 检查是否是refresh token类型
	if claims.Issuer != "refresh" {
		return "", "", fmt.Errorf("not a refresh token")
	}

	return claims.Username, claims.TokenID, nil
}
