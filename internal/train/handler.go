package train

import (
	"errors"
	"net/http"
	"strconv"

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

func (h *Handler) Publish(c *gin.Context) {
	trainID, err := strconv.ParseUint(c.Param("train_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}

	train, err := h.service.Publish(c.Request.Context(), trainID)
	if err != nil {
		switch {
		case errors.Is(err, ErrTrainNotFound):
			c.JSON(http.StatusNotFound, httpserver.Error(42001, "train not found"))
		case errors.Is(err, ErrTrainStatusInvalid):
			c.JSON(http.StatusConflict, httpserver.Error(47002, "admin operation not allowed"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(NewAdminTrainResponse(train)))
}
