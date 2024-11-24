package repository

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

var db *gorm.DB

func init() {
	// Load the .env file
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	// Construct the DSN
	dsn := "host=" + os.Getenv("POSTGRES_HOST") +
		" user=" + os.Getenv("POSTGRES_USER") +
		" password=" + os.Getenv("POSTGRES_PASSWORD") +
		" dbname=" + os.Getenv("POSTGRES_DB") +
		" port=5432" +
		" sslmode=disable" +
		" TimeZone=UTC"
	// Connect to the database
	connection, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Run migrations
	err = connection.AutoMigrate(&Execution{}, &State{}, &Step{}, &KeyValueOutput{}, &KeyValueArgument{}, &KeyValueStep{}, &ExecutionParams{})
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	db = connection
}
