package main

import (
	"log"

	"jiaozi/internal/auth"
	"jiaozi/internal/config"
	"jiaozi/internal/httpserver"
	"jiaozi/internal/platform/database"
)

func main() {
	cfg := config.Load()

	db, err := database.OpenMySQL(cfg.MySQL)
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	defer db.Close()

	authRepository := auth.NewRepository(db)
	authService := auth.NewService(authRepository)
	authHandler := auth.NewHandler(authService)

	router := httpserver.NewRouter(authHandler)

	log.Printf("jiaozi server listening on %s", cfg.ServerAddress())
	if err := router.Run(cfg.ServerAddress()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
