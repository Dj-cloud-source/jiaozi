package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jiaozi/internal/auth"
	"jiaozi/internal/user"
)

func NewRouter(authHandler *auth.Handler, authMiddleware gin.HandlerFunc, userHandler *user.Handler) *gin.Engine {
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
