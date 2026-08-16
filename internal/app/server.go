package app

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"auth-service/internal/auth"
	"auth-service/internal/config"
	"auth-service/internal/handler"
	"auth-service/internal/messaging"
	"auth-service/internal/prometheus"
	"auth-service/internal/repository"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewServer wires together all dependencies and returns a configured http.Handler.
func NewServer(cfg config.Config) http.Handler {
	conn := initDB(cfg)
	userRepo := repository.NewUserRepository(conn)

	passwordHasher := auth.NewBcryptHasher()
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, "myproject", "myproject-api")

	broker, err := messaging.InitBroker(cfg)
	if err != nil {
		log.Fatalf("init broker: %v", err)
	}
	if broker != nil {
		go broker.Run()
	}

	userHandler := handler.NewUserHandler(userRepo, passwordHasher, jwtManager, broker, cfg.ConfirmBaseURL)

	mux := http.NewServeMux()

	// public endpoints
	mux.HandleFunc("/health", handler.HealthHandler)
	mux.HandleFunc("/register", userHandler.Register)
	mux.HandleFunc("/confirm", userHandler.Confirm)
	mux.HandleFunc("/login", userHandler.Login)
	mux.HandleFunc("/validate", userHandler.Validate)
	mux.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)

	return prometheus.MetricsMiddleware(mux)
}

// Run starts the HTTP server and blocks.
func RunServer(cfg config.Config) {
	srv := &http.Server{
		Addr:    cfg.AppPort,
		Handler: NewServer(cfg),
	}
	log.Printf("Server listening on %s", cfg.AppPort)
	log.Fatal(srv.ListenAndServe())
}

func initDB(cfg config.Config) *sql.DB {
	maxRetries, err := strconv.Atoi(cfg.DBMaxRetries)
	if err != nil {
		log.Println("cannot get maxRetries value. set to default of 30 times...")
		maxRetries = 30
	}

	dbRetryInterval, err := strconv.Atoi(cfg.DBRetryInterval)
	if err != nil {
		log.Println("cannot get maxDelay value. set to default of 3 sec...")
		dbRetryInterval = 3
	}

	conn, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}

	for i := 1; i <= maxRetries; i++ {
		if err = conn.Ping(); err == nil {
			break
		}
		log.Printf("db.Ping attempt %d/%d failed: %v — retrying in %ds", i, maxRetries, err, dbRetryInterval)
		if i == maxRetries {
			log.Fatalf("could not connect to PostgreSQL after %d attempts", maxRetries)
		}
		time.Sleep(time.Duration(dbRetryInterval) * time.Second)
	}
	log.Println("Connected to PostgreSQL")

	return conn
}
