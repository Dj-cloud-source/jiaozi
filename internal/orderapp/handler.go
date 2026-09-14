package orderapp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"jiaozi/internal/auth"
	"jiaozi/internal/httpserver"
	"jiaozi/internal/train"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	value, exists := c.Get(auth.ContextUserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, httpserver.Error(41003, "unauthorized"))
		return
	}

	var req CreateOrderHTTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}
	if len(req.Passengers) != 1 {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}

	userID := value.(uint64)
	result, err := h.service.CreateOrder(c.Request.Context(), CreateOrderRequest{
		UserID:          userID,
		TrainID:         req.TrainID,
		SeatClass:       req.SeatClass,
		PassengerName:   req.Passengers[0].Name,
		PassengerIDCard: req.Passengers[0].IDCard,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCreateOrder):
			c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		case errors.Is(err, train.ErrTrainNotFound):
			c.JSON(http.StatusNotFound, httpserver.Error(42001, "train not found"))
		case errors.Is(err, ErrTrainNotOnSale):
			c.JSON(http.StatusConflict, httpserver.Error(43001, "train is not on sale"))
		case errors.Is(err, ErrPassengerAlreadyHasTicket):
			c.JSON(http.StatusConflict, httpserver.Error(43002, "passenger already has active ticket"))
		case errors.Is(err, ErrNoAvailableSeat), errors.Is(err, ErrSeatLockFailed):
			c.JSON(http.StatusConflict, httpserver.Error(43003, "no available seat"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	c.JSON(http.StatusCreated, httpserver.Success(NewCreateOrderResponse(result)))
}
