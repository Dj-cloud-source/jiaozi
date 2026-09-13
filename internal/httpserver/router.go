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

func NewRouter(authHandler AuthHandler, authMiddleware gin.HandlerFunc, userHandler UserHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", health)

	api := router.Group("/api/v1")
	api.GET("/health", health)
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	users := api.Group("/users")
	users.Use(authMiddleware)
	users.GET("/me", userHandler.Me)

	return router
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, Success(gin.H{
		"service": "jiaozi",
		"status":  "ok",
	}))
}
