package config

import "os"

// Config 运行时配置（环境变量注入，见 .env.example）
type Config struct {
	Port      string
	DataDir   string
	JWTSecret string
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() Config {
	return Config{
		Port:      getenv("PORT", "8080"),
		DataDir:   getenv("DATA_DIR", "./data"),
		JWTSecret: getenv("JWT_SECRET", "dev-insecure-secret"),
	}
}
