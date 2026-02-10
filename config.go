package main

import "os"

// Config 应用配置
type Config struct {
	DatabaseURL    string
	JWTSecret      string
	Port           string
	CloudinaryURL  string // 格式: cloudinary://API_KEY:API_SECRET@CLOUD_NAME
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/patpet?sslmode=disable"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Port:          getEnv("PORT", "8080"),
		CloudinaryURL: getEnv("CLOUDINARY_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
