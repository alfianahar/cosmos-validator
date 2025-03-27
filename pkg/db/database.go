package db

import (
	"context"
	"cosmos-tracker/internal/models"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// establishes a connection to the database with optimized settings
func ConnectDB() {
	// Close any existing connection first
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
		}
		DB = nil
	}

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

	// Add connection parameters to fix the cached plan issue
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s application_name=cosmos_validator_app",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("SSLMODE"),
	)

	// Apply performance optimizations through gorm config
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 newLogger,
		SkipDefaultTransaction: true,
		PrepareStmt:            false, // Disable prepared statements to avoid caching issues
	})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// Execute DISCARD ALL to clear any statement cache
	if err := database.Exec("DISCARD ALL").Error; err != nil {
		log.Printf("⚠️ Warning: Failed to discard cached plans: %v", err)
		// Continue anyway as this is just a precaution
	}

	// Configure connection pool settings
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatal("❌ Failed to get database connection:", err)
	}

	// Set connection pool parameters for optimal performance
	sqlDB.SetMaxIdleConns(10)               // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(50)               // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(1 * time.Hour) // Maximum lifetime of a connection

	DB = database

	// First migrate the migration history table itself
	if err := DB.AutoMigrate(&models.MigrationHistory{}); err != nil {
		log.Fatal("❌ Failed to migrate migration history table: ", err)
	}

	// Prepare to track models being migrated
	modelsToMigrate := []interface{}{
		&models.HourlyDelegation{},
		&models.DailyDelegation{},
		&models.Watchlist{},
	}

	// Get model names for logging
	modelNames := make([]string, len(modelsToMigrate))
	for i, model := range modelsToMigrate {
		modelNames[i] = fmt.Sprintf("%T", model)
	}

	// Perform the migration
	migrationRecord := models.MigrationHistory{
		Models: strings.Join(modelNames, ", "),
		Status: "in_progress",
	}

	// Save initial migration record
	if err := DB.Create(&migrationRecord).Error; err != nil {
		log.Printf("⚠️ Failed to create migration history record: %v", err)
	}

	// Auto-migrate tables with indices for better query performance
	err = DB.AutoMigrate(modelsToMigrate...)

	// Update migration record with result
	if err != nil {
		migrationRecord.Status = "error"
		migrationRecord.ErrorMessage = err.Error()
		DB.Save(&migrationRecord)
		log.Fatal("❌ Failed to migrate database:", err)
	} else {
		migrationRecord.Status = "success"
		DB.Save(&migrationRecord)
	}

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatal("❌ Database connection verification failed: ", err)
	}

	log.Println("✅ Database connected and migrated successfully!")
}

// runs a function within a transaction
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

// returns the migration history records
func GetMigrationHistory(limit int) ([]models.MigrationHistory, error) {
	var history []models.MigrationHistory

	result := DB.Order("migrated_at DESC")

	if limit > 0 {
		result = result.Limit(limit)
	}

	err := result.Find(&history).Error
	return history, err
}
