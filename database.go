package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

// DB is the global database connection
var DB *sql.DB

// RedisClient is the global Redis client
var RedisClient *redis.Client

// InitDatabases initializes the MySQL database connection
func InitDatabases() {
	// Get database connection details from environment variables or use defaults
	dbUser := getEnv("MYSQL_USER", "root")
	dbPassword := getEnv("MYSQL_PASSWORD", "")
	dbHost := getEnv("MYSQL_HOST", "localhost")
	dbPort := getEnv("MYSQL_PORT", "3306")

	// Build the DSN (Data Source Name) — database is always brainbolt
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/brainbolt?parseTime=true&charset=utf8mb4",
		dbUser, dbPassword, dbHost, dbPort)

	// Open the database connection
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	// Test the connection
	if err := DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	log.Println("Successfully connected to MySQL database")

	// Initialize Redis
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Successfully connected to Redis")
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
