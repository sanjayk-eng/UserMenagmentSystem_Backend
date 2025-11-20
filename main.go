package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/pkd/config"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/pkd/database"
)

// Init initializes environment variables, database connection, and runs auto-migrations.
// Returns the loaded environment configuration.
func Init() *config.ENV {
	// Load environment variables from .env or system environment
	env := config.LoadENV()

	// Connect to PostgreSQL using GORM
	database.Connection(env)

	// Run AutoMigrate to ensure all tables exist
	//database.AutoMigration()

	return env
}

func main() {
	// Initialize configuration and database
	env := Init()

	// Create a new Gin router
	r := gin.Default()

	fmt.Printf("Starting server on port %s\n", env.APP_PORT)

	// Start the Gin server
	if err := r.Run(env.APP_PORT); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
