package database

import (
	"log"

	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/models"
	"github.com/sanjayk-eng/UserMenagmentSystem_Backend/pkd/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is the global GORM database connection instance.
// It is initialized in Connection() and used throughout the application.
var DB *gorm.DB

// Connection establishes a PostgreSQL connection using the provided ENV configuration.
// It sets the global DB variable. If connection fails, the function panics.
//
// Usage:
//
//	env := config.LoadENV()
//	database.Connection(env)
func Connection(env *config.ENV) {
	// Open connection to PostgreSQL using GORM
	db, err := gorm.Open(postgres.Open(env.DB_URL), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	// Assign the connection to the global variable
	DB = db

	log.Println("✅ Database connection established successfully")
}

// AutoMigration runs GORM auto-migration for all models.
// It creates or updates tables to match the defined models in the database.
// Must be called after Connection().
//
// Usage:
//
//	database.AutoMigration()
func AutoMigration() {
	// Ensure DB connection is initialized
	if DB == nil {
		panic("Database not initialized. Call Connection() first.")
	}

	// Run AutoMigrate for all models
	err := DB.AutoMigrate(
		&models.Employee{},
		&models.Role{},
		&models.LeaveType{},
		&models.Leave{},
		&models.LeaveBalance{},
		&models.LeaveAdjustment{},
		&models.PayrollRun{},
		&models.Payslip{},
		&models.Audit{},
	)

	if err != nil {
		panic("S Failed to create or migrate database schema")
	}

	log.Println("Database schema migrated successfully")
}
