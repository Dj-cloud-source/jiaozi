package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"jiaozi/internal/httpserver"
)

const ContextUserIDKey = "user_id"

func NewMiddleware(tokens *TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token == header || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httpserver.Error(41003, "unauthorized"))
			return
		}

		userID, err := tokens.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httpserver.Error(41003, "unauthorized"))
			return
		}

		c.Set(ContextUserIDKey, userID)
		c.Next()
	}
}
