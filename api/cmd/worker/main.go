package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hexabase/hexabase-kaas/api/internal/config"
	"github.com/hexabase/hexabase-kaas/api/internal/db"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Fatal("Configuration validation failed", zap.Error(err))
	}

	// Connect to database
	dbConfig := &db.DatabaseConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	_, err = db.ConnectDatabase(dbConfig)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	logger.Info("Worker started successfully")

	// TODO: Initialize NATS connection
	// TODO: Start worker goroutines for different task types
	// TODO: Implement task processing logic

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Worker shutting down...")

	// TODO: Gracefully shutdown worker processes

	logger.Info("Worker exited")
}