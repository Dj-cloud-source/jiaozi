package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"jiaozi/internal/auth"
)

func NewRouter(authHandler *auth.Handler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", health)

	api := router.Group("/api/v1")
	api.GET("/health", health)
	api.POST("/auth/register", authHandler.Register)

	return router
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, Success(gin.H{
		"service": "jiaozi",
		"status":  "ok",
	}))
}
