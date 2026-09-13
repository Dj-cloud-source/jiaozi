package main

import (
	"log"

	"jiaozi/internal/auth"
	"jiaozi/internal/config"
	"jiaozi/internal/httpserver"
	"jiaozi/internal/platform/database"
	"jiaozi/internal/user"
)

func main() {
	cfg := config.Load()

	db, err := database.OpenMySQL(cfg.MySQL)
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	defer db.Close()

	authRepository := auth.NewRepository(db)
	tokenManager := auth.NewTokenManager(cfg.Auth.TokenSecret, cfg.Auth.TokenExpireSeconds)
	authService := auth.NewService(authRepository, tokenManager)
	authHandler := auth.NewHandler(authService)
	authMiddleware := auth.NewMiddleware(tokenManager)

	userRepository := user.NewRepository(db)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)

	router := httpserver.NewRouter(authHandler, authMiddleware, userHandler)

	log.Printf("jiaozi server listening on %s", cfg.ServerAddress())
	if err := router.Run(cfg.ServerAddress()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
