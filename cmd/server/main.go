package main

import (
	"minimal-service/internal/app"
	"minimal-service/internal/config"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()
	app.RunServer(cfg)
}
