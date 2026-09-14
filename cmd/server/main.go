package main

import (
	"context"
	"log"
	"time"

	"jiaozi/internal/auth"
	"jiaozi/internal/config"
	"jiaozi/internal/httpserver"
	"jiaozi/internal/order"
	"jiaozi/internal/orderapp"
	"jiaozi/internal/passenger"
	"jiaozi/internal/payment"
	"jiaozi/internal/paymentapp"
	"jiaozi/internal/platform/database"
	"jiaozi/internal/scheduler"
	"jiaozi/internal/seat"
	"jiaozi/internal/station"
	"jiaozi/internal/train"
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

	stationRepository := station.NewRepository(db)
	stationService := station.NewService(stationRepository)
	stationHandler := station.NewHandler(stationService)

	trainRepository := train.NewRepository(db)
	trainService := train.NewService(trainRepository)
	trainHandler := train.NewHandler(trainService)

	passengerRepository := passenger.NewRepository(db)
	seatRepository := seat.NewRepository(db)
	orderRepository := order.NewRepository(db)
	paymentRepository := payment.NewRepository(db)
	orderAppService := orderapp.NewService(
		db,
		trainService,
		passengerRepository,
		seatRepository,
		orderRepository,
		paymentRepository,
	)
	orderAppHandler := orderapp.NewHandler(orderAppService)

	paymentAppService := paymentapp.NewService(
		db,
		paymentRepository,
		orderRepository,
		seatRepository,
	)
	paymentAppHandler := paymentapp.NewHandler(paymentAppService)

	trainStatusScheduler := scheduler.NewTrainStatusScheduler(
		trainService,
		time.Duration(cfg.Scheduler.TrainStatusIntervalSeconds)*time.Second,
	)
	trainStatusScheduler.Start(context.Background())

	orderTimeoutScheduler := scheduler.NewOrderTimeoutScheduler(
		orderAppService,
		time.Duration(cfg.Scheduler.OrderTimeoutIntervalSeconds)*time.Second,
	)
	orderTimeoutScheduler.Start(context.Background())

	orderCompleteScheduler := scheduler.NewOrderCompleteScheduler(
		orderAppService,
		time.Duration(cfg.Scheduler.OrderCompleteIntervalSeconds)*time.Second,
	)
	orderCompleteScheduler.Start(context.Background())

	router := httpserver.NewRouter(authHandler, authMiddleware, userHandler, stationHandler, trainHandler, orderAppHandler, paymentAppHandler)

	log.Printf("jiaozi server listening on %s", cfg.ServerAddress())
	if err := router.Run(cfg.ServerAddress()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
