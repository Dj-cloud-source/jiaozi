package main

import (
	"log"

	"jiaozi/internal/config"
	"jiaozi/internal/httpserver"
)

func main() {
	cfg := config.Load()

	router := httpserver.NewRouter()

	log.Printf("jiaozi server listening on %s", cfg.ServerAddress())
	if err := router.Run(cfg.ServerAddress()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
