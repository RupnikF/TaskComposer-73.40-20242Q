package repository

import (
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

func Initialize() *gorm.DB {
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
	err = connection.AutoMigrate(&Execution{}, &State{}, &Step{}, &KeyValueOutput{}, &KeyValueArgument{}, &KeyValueStep{}, &ExecutionParams{}, &Tags{})
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	err = connection.Use(otelgorm.NewPlugin())
	if err != nil {
		log.Printf("Failed to install instrumentation: %v", err)
	}
	return connection
}
