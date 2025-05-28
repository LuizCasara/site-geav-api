package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// DBConfig holds the configuration for the database connection
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDBConfigFromEnv creates a new DBConfig from environment variables
func NewDBConfigFromEnv() *DBConfig {
	return &DBConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "pgadmin"),
		DBName:   getEnv("DB_NAME", "geav"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}
}

// ConnectionString returns the connection string for the database
func (c *DBConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// NewDB creates a new database connection
func NewDB(config *DBConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return db, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// InitDB initializes the database connection
func InitDB() (*sql.DB, error) {
	config := NewDBConfigFromEnv()
	db, err := NewDB(config)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil, err
	}
	return db, nil
}
