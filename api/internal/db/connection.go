package db

import (
	"fmt"
	"os"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/domain/auth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDatabaseConfig creates database config from environment variables
func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "hexabase"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// DSN returns the database connection string
func (cfg *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
}

// ConnectDatabase establishes database connection with retry logic
func ConnectDatabase(cfg *DatabaseConfig) (*gorm.DB, error) {
	// Configure logger
	logLevel := logger.Error
	if getEnv("DB_LOG_LEVEL", "error") == "info" {
		logLevel = logger.Info
	}

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	var db *gorm.DB
	var err error
	
	// Retry connection up to 10 times with exponential backoff
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(cfg.DSN()), config)
		if err == nil {
			// Test the connection
			sqlDB, testErr := db.DB()
			if testErr == nil && sqlDB.Ping() == nil {
				break
			}
			err = testErr
		}
		
		if i < maxRetries-1 {
			waitTime := time.Duration(i+1) * time.Second
			fmt.Printf("Database connection attempt %d failed, retrying in %v...\n", i+1, waitTime)
			time.Sleep(waitTime)
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// MigrateDatabase runs all migrations
func MigrateDatabase(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Organization{},
		&OrganizationUser{},
		&Plan{},
		&Workspace{},
		&Project{},
		&Group{},
		&GroupMembership{},
		&Role{},
		&RoleAssignment{},
		&VClusterProvisioningTask{},
		&StripeEvent{},
		// Node management models
		&NodePlan{},
		&WorkspaceNodeAllocation{},
		&DedicatedNode{},
		&NodeEvent{},
		// CI/CD models
		&Pipeline{},
		&PipelineRun{},
		&PipelineTemplate{},
		&WorkspaceProviderConfig{},
		&CICDCredential{},
		// Auth models
		&auth.AuthState{},
	)
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}