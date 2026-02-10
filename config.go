package main

import "os"

// Config 应用配置
type Config struct {
	DatabaseURL      string
	JWTSecret        string
	Port             string
	SupabaseURL      string // Supabase 项目 URL，如 https://xxx.supabase.co
	SupabaseKey      string // Supabase service_role key（用于服务端上传）
	SupabaseBucket   string // Storage bucket 名称
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/patpet?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Port:           getEnv("PORT", "8080"),
		SupabaseURL:    getEnv("SUPABASE_URL", ""),
		SupabaseKey:    getEnv("SUPABASE_SERVICE_KEY", ""),
		SupabaseBucket: getEnv("SUPABASE_BUCKET", "avatars"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
