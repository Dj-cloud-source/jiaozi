package ticket

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"jiaozi/internal/auth"
	"jiaozi/internal/httpserver"
	"jiaozi/internal/order"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Detail(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, httpserver.Error(41003, "unauthorized"))
		return
	}

	orderView, err := h.service.Detail(c.Request.Context(), c.Param("order_id"), userID)
	if err != nil {
		switch {
		case errors.Is(err, order.ErrInvalidTicketOrder):
			c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		case errors.Is(err, order.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, httpserver.Error(43004, "order not found"))
		case errors.Is(err, ErrTicketUnavailable):
			c.JSON(http.StatusConflict, httpserver.Error(45001, "ticket unavailable"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(NewResponse(orderView)))
}

func currentUserID(c *gin.Context) (uint64, bool) {
	value, exists := c.Get(auth.ContextUserIDKey)
	if !exists {
		return 0, false
	}

	userID, ok := value.(uint64)
	return userID, ok
}
