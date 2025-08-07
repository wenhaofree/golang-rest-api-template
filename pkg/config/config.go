package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config 应用配置结构
type Config struct {
	// MongoDB配置
	MongoEnabled bool   `json:"mongo_enabled"`
	MongoURI     string `json:"mongo_uri"`
	MongoDB      string `json:"mongo_db"`
	MongoLogCollection string `json:"mongo_log_collection"`
	
	// 数据库配置
	PostgresHost     string `json:"postgres_host"`
	PostgresDB       string `json:"postgres_db"`
	PostgresUser     string `json:"postgres_user"`
	PostgresPassword string `json:"postgres_password"`
	PostgresPort     string `json:"postgres_port"`
	
	// Redis配置
	RedisHost string `json:"redis_host"`
	
	// JWT配置
	JWTSecretKey string `json:"jwt_secret_key"`
	APISecretKey string `json:"api_secret_key"`
}

// LoadConfig 从.env文件和环境变量加载配置
func LoadConfig() *Config {
	// 尝试加载.env文件，如果文件不存在也不会报错
	loadEnvFiles()
	
	return &Config{
		// MongoDB配置 - 默认启用，可通过环境变量禁用
		MongoEnabled:       getBoolEnv("MONGO_ENABLED", true),
		MongoURI:          getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:           getEnv("MONGO_DB", "logging"),
		MongoLogCollection: getEnv("MONGO_LOG_COLLECTION", "logs"),
		
		// 数据库配置
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresDB:       getEnv("POSTGRES_DB", "go_app_dev"),
		PostgresUser:     getEnv("POSTGRES_USER", "docker"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "password"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		
		// Redis配置
		RedisHost: getEnv("REDIS_HOST", "localhost"),
		
		// JWT配置
		JWTSecretKey: getEnv("JWT_SECRET_KEY", ""),
		APISecretKey: getEnv("API_SECRET_KEY", ""),
	}
}

// loadEnvFiles 按优先级加载.env文件
// 优先级：.env.local > .env.{environment} > .env
func loadEnvFiles() {
	envFiles := []string{
		".env",                    // 基础配置
		".env." + getGinMode(),    // 环境特定配置 (.env.development, .env.production)
		".env.local",              // 本地覆盖配置（通常在.gitignore中）
	}
	
	for _, file := range envFiles {
		if err := godotenv.Load(file); err != nil {
			// 只有.env文件不存在时才记录日志，其他文件可选
			if file == ".env" {
				log.Printf("Warning: .env file not found, using environment variables and defaults")
			}
		} else {
			log.Printf("Loaded configuration from %s", file)
		}
	}
}

// getGinMode 获取Gin运行模式，用于确定环境特定的.env文件
func getGinMode() string {
	mode := getEnv("GIN_MODE", "debug")
	switch mode {
	case "release":
		return "production"
	case "test":
		return "test"
	default:
		return "development"
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getBoolEnv 获取布尔类型环境变量，如果不存在则返回默认值
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		// 支持多种布尔值表示方式
		lowerValue := strings.ToLower(value)
		switch lowerValue {
		case "true", "1", "yes", "on", "enable", "enabled":
			return true
		case "false", "0", "no", "off", "disable", "disabled":
			return false
		default:
			// 如果无法解析，尝试使用strconv.ParseBool
			if parsed, err := strconv.ParseBool(value); err == nil {
				return parsed
			}
		}
	}
	return defaultValue
}