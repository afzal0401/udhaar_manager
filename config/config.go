package config

import "os"

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

func Load() *Config {
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
