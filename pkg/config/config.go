package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

// ENV holds all application environment variables in a structured way
type ENV struct {
	DB_URL          string
	APP_PORT        string
	SERACT_KEY      string
	FRONTEND_SERVER string
	SMTP_PORT       string
	SMTP_HOST       string
}

var (
	cfg  *ENV      // Singleton instance of ENV
	once sync.Once // Ensures LoadENV is executed only once (thread-safe)
)

// LoadENV loads environment variables from .env file (if exists)
// and system environment variables. It returns a singleton ENV instance.
//
// Usage:
//
//	env := config.LoadENV()
//	dbURL := env.DB.DB_URL
//	port  := env.PORT.PORT
func LoadENV() *ENV {
	once.Do(func() {
		// Attempt to load .env file; ignore error if file does not exist
		err := godotenv.Load()
		if err != nil {
			log.Println("⚠ No .env file found, using system environment variables")
		}

		// Populate ENV struct with environment variables
		cfg = &ENV{
			DB_URL:          os.Getenv("DB_URL"),   // Required: PostgreSQL URL
			APP_PORT:        os.Getenv("APP_PORT"), // Optional: server port
			SERACT_KEY:      os.Getenv("SECRATE_KEY"),
			FRONTEND_SERVER: os.Getenv("F_SERVER"),
			SMTP_PORT:       os.Getenv("SMTP_PORT"),
			SMTP_HOST:       os.Getenv("SMTP_HOST"),
		}
	})
	log.Println("✅ Environment variables loaded successfully")
	return cfg
}
