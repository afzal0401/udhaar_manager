package config

import "os"

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
	Port       string

	MSG91AuthKey  string
	GupshupAPIKey string
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Load() *Config {
	return &Config{
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBName:     getEnv("DB_NAME", "udhaar_manager"),
		Port:       getEnv("PORT", "8080"),

		// Leave empty until you sign up with a provider; notify package falls back to console logging.
		MSG91AuthKey:  getEnv("MSG91_AUTH_KEY", ""),
		GupshupAPIKey: getEnv("GUPSHUP_API_KEY", ""),
	}
}
