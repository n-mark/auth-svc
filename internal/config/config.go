package config

import (
	"fmt"
	"os"
)

// DB holds PostgreSQL connection parameters.
type DB struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type JWTConf struct {
	Secret string
}

type Config struct {
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	JWTSecret       string
	DBMaxRetries    string
	DBRetryInterval string
	AppPort         string
}

func Load() Config {
	return Config{
		DBHost:          mustEnv("DB_HOST"),
		DBPort:          mustEnv("DB_PORT"),
		DBUser:          mustEnv("DB_USER"),
		DBPassword:      mustEnv("DB_PASSWORD"),
		DBName:          mustEnv("DB_NAME"),
		JWTSecret:       mustEnv("JWT_SECRET"),
		DBMaxRetries:    mustEnv("DB_MAX_RETRIES"),
		DBRetryInterval: mustEnv("DB_RETRY_INTERVAL"),
		AppPort:         mustEnv("APP_PORT"),
	}
}

// DSN returns a PostgreSQL connection string.
func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required environment variable " + key + " is not set")
	}
	return v
}
