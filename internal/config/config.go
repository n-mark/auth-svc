package config

import (
	"fmt"
	"net/url"
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
	BrokerType      string
	ConfirmBaseURL  string
}

type RabbitConfig struct {
	DSN          string
	AuthExchange string
	ProduceRK    string
}

func Load() Config {
	return Config{
		DBHost:          getEnv("DB_HOST", ""),
		DBPort:          getEnv("DB_PORT", ""),
		DBUser:          getEnv("DB_USER", ""),
		DBPassword:      getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		DBMaxRetries:    getEnv("DB_MAX_RETRIES", ""),
		DBRetryInterval: getEnv("DB_RETRY_INTERVAL", ""),
		AppPort:         getEnv("APP_PORT", ""),
		BrokerType:      getEnv("BROKER_TYPE", ""),
		ConfirmBaseURL:  getEnv("CONFIRM_BASE_URL", ""),
	}
}

// DSN returns a PostgreSQL connection string.
func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func GetRabbitConfig() RabbitConfig {
	user := os.Getenv("RABBIT_USERNAME")
	password := os.Getenv("RABBIT_PASSWORD")
	host := os.Getenv("RABBIT_HOST")
	port := os.Getenv("RABBIT_PORT")
	authExchange := getEnv("RABBIT_AUTH_PRODUCE_EXCHANGE", "auth")
	rk := getEnv("RABBIT_AUTH_PRODUCE_RK", "user.created")

	u := url.URL{Scheme: "amqp",
		User: url.UserPassword(user, password),
		Host: fmt.Sprintf("%s:%s", host, port)}

	return RabbitConfig{DSN: u.String(), AuthExchange: authExchange, ProduceRK: rk}
}
