package main

import (
	"auth-service/internal/app"
	"auth-service/internal/config"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()
	app.RunServer(cfg)
}
