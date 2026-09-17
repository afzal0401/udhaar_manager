package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
	Port       string

	WhatsAppAccessToken     string
	WhatsAppPhoneNumberID   string
	WhatsAppOTPTemplate     string
	WhatsAppMessageTemplate string
	WhatsAppTemplateLang    string
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadDotEnv reads KEY=VALUE lines from a .env file (if present) into the process
// environment, without overriding vars already exported in the shell.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}

func Load() *Config {
	loadDotEnv(".env")

	return &Config{
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBName:     getEnv("DB_NAME", "udhaar_manager"),
		Port:       getEnv("PORT", "8080"),

		WhatsAppAccessToken:     getEnv("WHATSAPP_ACCESS_TOKEN", ""),
		WhatsAppPhoneNumberID:   getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
		WhatsAppOTPTemplate:     getEnv("WHATSAPP_OTP_TEMPLATE", ""),
		WhatsAppMessageTemplate: getEnv("WHATSAPP_MESSAGE_TEMPLATE", ""),
		WhatsAppTemplateLang:    getEnv("WHATSAPP_TEMPLATE_LANGUAGE", "en_US"),
	}
}
