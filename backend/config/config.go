package config

import "os"

type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	JWTSecret     string
	ServerPort    string
	ReviewEnabled bool // true=审核模式，false=直通模式
	RedisHost     string
	RedisPort     string
	RedisPassword string
	TrustedProxies string // 信任的代理IP列表（逗号分隔），用于正确解析 X-Forwarded-For
}

var App *Config

func Init() {
	App = &Config{
		DBHost:        getEnv("DB_HOST", "127.0.0.1"),
		DBPort:        getEnv("DB_PORT", "3306"),
		DBUser:        getEnv("DB_USER", "root"),
		DBPassword:    getEnv("DB_PASSWORD", "123@123qwe"),
		DBName:        getEnv("DB_NAME", "go_yzs"),
		JWTSecret:     getEnv("JWT_SECRET", "go-yzs-secret-key-2026"),
		ServerPort:    getEnv("SERVER_PORT", "18881"),
		ReviewEnabled: getEnv("REVIEW_ENABLED", "true") == "true",
		RedisHost:     getEnv("REDIS_HOST", "127.0.0.1"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		TrustedProxies: getEnv("TRUSTED_PROXIES", ""), // 空字符串=不信任任何代理(直连模式)；"*"=信任所有；具体如 "127.0.0.1,10.0.0.0/8"
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
