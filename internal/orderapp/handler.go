package orderapp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"jiaozi/internal/auth"
	"jiaozi/internal/httpserver"
	"jiaozi/internal/order"
	"jiaozi/internal/train"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
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

func (h *Handler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, httpserver.Error(41003, "unauthorized"))
		return
	}

	orders, err := h.service.ListOrders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(NewOrderListResponse(orders)))
}

func (h *Handler) Detail(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, httpserver.Error(41003, "unauthorized"))
		return
	}

	orderView, err := h.service.OrderDetail(c.Request.Context(), c.Param("order_id"), userID)
	if err != nil {
		switch {
		case errors.Is(err, order.ErrInvalidTicketOrder):
			c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		case errors.Is(err, order.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, httpserver.Error(43004, "order not found"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(NewOrderDetailResponse(orderView)))
}

func (h *Handler) Cancel(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, httpserver.Error(41003, "unauthorized"))
		return
	}

	result, err := h.service.CancelOrder(c.Request.Context(), c.Param("order_id"), userID)
	if err != nil {
		switch {
		case errors.Is(err, order.ErrInvalidTicketOrder):
			c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		case errors.Is(err, order.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, httpserver.Error(43004, "order not found"))
		case errors.Is(err, order.ErrOrderStatusInvalid):
			c.JSON(http.StatusConflict, httpserver.Error(43005, "order cannot be cancelled"))
		case errors.Is(err, ErrSeatReleaseFailed):
			c.JSON(http.StatusConflict, httpserver.Error(43006, "seat release failed"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(NewCancelOrderResponse(result)))
}

func currentUserID(c *gin.Context) (uint64, bool) {
	value, exists := c.Get(auth.ContextUserIDKey)
	if !exists {
		return 0, false
	}

	userID, ok := value.(uint64)
	return userID, ok
}
