package db

import (
	"context"
	"cosmos-tracker/internal/models"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// establishes a connection to the database with optimized settings
func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found")
	}

	// Configure custom logger for better SQL debugging
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("SSLMODE"),
	)

	// Apply performance optimizations through gorm config
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 newLogger,
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// Configure connection pool settings
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatal("❌ Failed to get database connection:", err)
	}

	// Set connection pool parameters for optimal performance
	sqlDB.SetMaxIdleConns(2)                // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(10)               // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(1 * time.Hour) // Maximum lifetime of a connection

	DB = database

	// Auto-migrate tables with indices for better query performance
	err = DB.AutoMigrate(
		&models.HourlyDelegation{},
		&models.DailyDelegation{},
		&models.Watchlist{},
	)
	if err != nil {
		log.Fatal("❌ Failed to migrate database:", err)
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatal("❌ Database connection verification failed:", err)
	}

	log.Println("✅ Database connected and migrated successfully!")
}

// WithTransaction runs a function within a transaction
func WithTransaction(fn func(tx *gorm.DB) error) error {
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
