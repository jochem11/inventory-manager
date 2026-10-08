// Package database connects the services to MySQL.
package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds the MySQL connection settings.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// ConfigFromEnv reads DB_HOST, DB_PORT, DB_USER, DB_PASSWORD and DB_NAME,
// falling back to local development defaults and to defaultName, the
// service's own database. The password has no default: it comes from .env
// (see .env.example).
func ConfigFromEnv(defaultName string) Config {
	return Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "3306"),
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", ""),
		Name:     getEnv("DB_NAME", defaultName),
	}
}

// DSN builds the go-sql-driver/mysql connection string. parseTime is required
// so DATETIME columns scan into time.Time.
func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		c.User, c.Password, c.Host, c.Port, c.Name)
}

// Connect opens a GORM connection and verifies it with a ping.
func Connect(cfg Config) (*gorm.DB, error) {
	// TranslateError turns driver errors such as duplicate keys into GORM's
	// portable errors (gorm.ErrDuplicatedKey).
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		TranslateError: true,
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      logger.Warn,
			// Lookups that find nothing are expected (e.g. checking an email is
			// free), not errors worth logging.
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}

// CreateIfMissing creates cfg's database when it doesn't exist yet, so a new
// service needs no manual setup in development.
func CreateIfMissing(cfg Config) error {
	server := cfg
	server.Name = ""
	db, err := Connect(server)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	defer sqlDB.Close()
	// A case-insensitive collation (the "ci"), so LIKE filters and unique
	// indexes ignore case whatever the server's default is. Tables and columns
	// inherit it unless they set their own, like the ascii_bin ID columns.
	if err := db.Exec("CREATE DATABASE IF NOT EXISTS `" + cfg.Name + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci").Error; err != nil {
		return fmt.Errorf("create database %s: %w", cfg.Name, err)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
