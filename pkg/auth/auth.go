package auth

import (
	"crypto/rand"
	"encoding/base64"
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

func GenerateToken(username string) (string, error) {
	// The expiration time after which the token will be invalid.
	expirationTime := time.Now().Add(5 * time.Minute).Unix()

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
