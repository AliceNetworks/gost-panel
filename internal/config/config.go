package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
)

// DefaultJWTSecret is the insecure default - used only for detection
const DefaultJWTSecret = "gost-panel-secret-change-me"

// 默认配置常量
const (
	DefaultGitHubRawURL = "https://raw.githubusercontent.com/AliceNetworks/gost-panel/main/scripts"
	DefaultGOSTVersion  = "3.0.0-rc10"
)

type Config struct {
	ListenAddr     string   // 面板监听地址
	DBPath         string   // 数据库路径
	JWTSecret      string   // JWT 密钥
	AgentGRPCAddr  string   // Agent gRPC 监听地址
	Debug          bool     // 调试模式
	AllowedOrigins []string // 允许的 CORS 来源
	GitHubRawURL   string   // GitHub Raw 文件 URL
	GOSTVersion    string   // GOST 版本号
}

func Load() *Config {
	jwtSecret := getEnv("JWT_SECRET", "")

	// 如果未设置 JWT_SECRET，生成随机密钥并警告
	if jwtSecret == "" || jwtSecret == DefaultJWTSecret {
		if jwtSecret == DefaultJWTSecret {
			log.Println("WARNING: Using default JWT_SECRET is insecure! Please set JWT_SECRET environment variable.")
		}
		// 生成随机密钥 (256 bit)
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			log.Fatal("FATAL: Failed to generate random JWT secret:", err)
		}
		jwtSecret = hex.EncodeToString(randomBytes)
		log.Println("WARNING: Generated random JWT_SECRET for this session. Tokens will be invalidated on restart.")
		log.Println("WARNING: Set JWT_SECRET environment variable for persistent tokens.")
	}

	// 解析允许的 CORS 来源
	allowedOrigins := parseAllowedOrigins(getEnv("ALLOWED_ORIGINS", ""))

	return &Config{
		ListenAddr:     getEnv("LISTEN_ADDR", ":8080"),
		DBPath:         getEnv("DB_PATH", "./data/panel.db"),
		JWTSecret:      jwtSecret,
		AgentGRPCAddr:  getEnv("AGENT_GRPC_ADDR", ":9090"),
		Debug:          getEnv("DEBUG", "false") == "true",
		AllowedOrigins: allowedOrigins,
		GitHubRawURL:   getEnv("GITHUB_RAW_URL", DefaultGitHubRawURL),
		GOSTVersion:    getEnv("GOST_VERSION", DefaultGOSTVersion),
	}
}

// parseAllowedOrigins 解析允许的 CORS 来源
func parseAllowedOrigins(origins string) []string {
	if origins == "" {
		return nil // nil 表示使用默认策略
	}
	var result []string
	for _, origin := range splitAndTrim(origins, ",") {
		if origin != "" {
			result = append(result, origin)
		}
	}
	return result
}

// splitAndTrim 分割字符串并去除空格
func splitAndTrim(s, sep string) []string {
	if s == "" {
		return nil
	}
	parts := make([]string, 0)
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// splitString 简单的字符串分割
func splitString(s, sep string) []string {
	if sep == "" {
		return []string{s}
	}
	var result []string
	for {
		idx := indexOf(s, sep)
		if idx < 0 {
			result = append(result, s)
			break
		}
		result = append(result, s[:idx])
		s = s[idx+len(sep):]
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
