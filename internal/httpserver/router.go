package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	Register(*gin.Context)
	Login(*gin.Context)
}

type UserHandler interface {
	Me(*gin.Context)
}

type StationHandler interface {
	ListActive(*gin.Context)
	Create(*gin.Context)
}

type TrainHandler interface {
	Create(*gin.Context)
	Publish(*gin.Context)
	List(*gin.Context)
	Detail(*gin.Context)
}

type OrderHandler interface {
	Create(*gin.Context)
	List(*gin.Context)
	Detail(*gin.Context)
	Cancel(*gin.Context)
	Return(*gin.Context)
}

type PaymentHandler interface {
	Detail(*gin.Context)
	Pay(*gin.Context)
	MockSuccess(*gin.Context)
	MockFail(*gin.Context)
}

func NewRouter(
	authHandler AuthHandler,
	authMiddleware gin.HandlerFunc,
	userHandler UserHandler,
	stationHandler StationHandler,
	trainHandler TrainHandler,
	orderHandler OrderHandler,
	paymentHandler PaymentHandler,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", health)

	api := router.Group("/api/v1")
	api.GET("/health", health)
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)
	api.GET("/stations", stationHandler.ListActive)
	api.GET("/trains", trainHandler.List)
	api.GET("/trains/:train_id", trainHandler.Detail)

	admin := api.Group("/admin")
	admin.POST("/stations", stationHandler.Create)
	admin.POST("/trains", trainHandler.Create)
	admin.POST("/trains/:train_id/publish", trainHandler.Publish)

	users := api.Group("/users")
	users.Use(authMiddleware)
	users.GET("/me", userHandler.Me)

	orders := api.Group("/orders")
	orders.Use(authMiddleware)
	orders.POST("", orderHandler.Create)
	orders.GET("", orderHandler.List)
	orders.GET("/:order_id", orderHandler.Detail)
	orders.POST("/:order_id/cancel", orderHandler.Cancel)
	orders.POST("/:order_id/return", orderHandler.Return)

	payments := api.Group("/payments")
	payments.Use(authMiddleware)
	payments.GET("/:payment_id", paymentHandler.Detail)
	payments.POST("/:payment_id/pay", paymentHandler.Pay)
	payments.POST("/:payment_id/mock/success", paymentHandler.MockSuccess)
	payments.POST("/:payment_id/mock/fail", paymentHandler.MockFail)

	return router
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, Success(gin.H{
		"service": "jiaozi",
		"status":  "ok",
	}))
}
