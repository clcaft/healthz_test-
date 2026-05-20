package main

import (
	"log"

	"gl.eda1.ru/go/go-service-template/config"
	"gl.eda1.ru/go/go-service-template/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
