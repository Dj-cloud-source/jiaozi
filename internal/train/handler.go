package train

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

func (h *Handler) Create(c *gin.Context) {
	var req CreateTrainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}

	train, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTrain):
			c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		case errors.Is(err, ErrTrainAlreadyExists):
			c.JSON(http.StatusConflict, httpserver.Error(42005, "train already exists"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	c.JSON(http.StatusCreated, httpserver.Success(NewAdminTrainResponse(train)))
}
