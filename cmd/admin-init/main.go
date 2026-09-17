package main

import (
	"context"
	"errors"
	"flag"
	"log"

	"jiaozi/internal/admin"
	"jiaozi/internal/auth"
	"jiaozi/internal/config"
	"jiaozi/internal/platform/database"
)

func main() {
	username := flag.String("username", "", "admin username")
	password := flag.String("password", "", "admin password")
	flag.Parse()

	cfg := config.Load()
	db, err := database.OpenMySQL(cfg.MySQL)
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	defer db.Close()

	repository := admin.NewRepository(db)
	service := admin.NewService(repository, auth.NewTokenManager(cfg.Auth.TokenSecret, cfg.Auth.TokenExpireSeconds))

	adminModel, err := service.Create(context.Background(), admin.CreateAdminRequest{
		Username: *username,
		Password: *password,
	})
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrInvalidCreateAdminRequest):
			log.Fatal("invalid admin init request: username is required and password must be at least 6 characters")
		case errors.Is(err, admin.ErrAdminAlreadyExists):
			log.Fatal("admin username already exists")
		default:
			log.Fatalf("create admin: %v", err)
		}
	}

	log.Printf("created admin id=%d username=%s", adminModel.ID, adminModel.Username)
}
