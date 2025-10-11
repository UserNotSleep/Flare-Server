package config

import (
	"os"
)

type Config struct {
	Port        string
	FirebaseKey string
	Collection  string
	JWTSecret   string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "3000"),
		FirebaseKey: getEnv("FIREBASE_SERVICE_ACCOUNT", "serviceAccountKey.json"),
		Collection:  getEnv("FIRESTORE_COLLECTION", "messages"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
