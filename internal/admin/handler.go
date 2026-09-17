package admin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"jiaozi/internal/httpserver"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}

	response, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidLoginRequest):
			c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		case errors.Is(err, ErrInvalidUsernameOrPassword):
			c.JSON(http.StatusUnauthorized, httpserver.Error(48001, "invalid username or password"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(response))
}
