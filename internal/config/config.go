package config

import (
	"os"
)

type Config struct {
	Port        string
	FirebaseKey string
	Collection  string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "3000"),
		FirebaseKey: getEnv("FIREBASE_SERVICE_ACCOUNT", "serviceAccountKey.json"),
		Collection:  getEnv("FIRESTORE_COLLECTION", "messages"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
