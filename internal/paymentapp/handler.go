package paymentapp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"jiaozi/internal/auth"
	"jiaozi/internal/httpserver"
	"jiaozi/internal/order"
	"jiaozi/internal/payment"
	"jiaozi/internal/seat"
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

	paymentModel, err := h.service.Detail(c.Request.Context(), c.Param("payment_id"), userID)
	if err != nil {
		writePaymentError(c, err)
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(NewPaymentResponse(paymentModel)))
}

func (h *Handler) Pay(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, httpserver.Error(41003, "unauthorized"))
		return
	}

	paymentModel, err := h.service.StartPay(c.Request.Context(), c.Param("payment_id"), userID)
	if err != nil {
		writePaymentError(c, err)
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(NewPaymentResponse(paymentModel)))
}

func (h *Handler) MockSuccess(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, httpserver.Error(41003, "unauthorized"))
		return
	}

	result, err := h.service.MockSuccess(c.Request.Context(), c.Param("payment_id"), userID)
	if err != nil {
		writePaymentError(c, err)
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(NewMockSuccessResponse(result)))
}

func (h *Handler) MockFail(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, httpserver.Error(41003, "unauthorized"))
		return
	}

	result, err := h.service.MockFail(c.Request.Context(), c.Param("payment_id"), userID)
	if err != nil {
		writePaymentError(c, err)
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(NewMockFailResponse(result)))
}

func currentUserID(c *gin.Context) (uint64, bool) {
	value, exists := c.Get(auth.ContextUserIDKey)
	if !exists {
		return 0, false
	}

	userID, ok := value.(uint64)
	return userID, ok
}

func writePaymentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, payment.ErrInvalidPayment):
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
	case errors.Is(err, payment.ErrPaymentNotFound):
		c.JSON(http.StatusNotFound, httpserver.Error(44001, "payment not found"))
	case errors.Is(err, payment.ErrPaymentStatusInvalid), errors.Is(err, order.ErrOrderStatusInvalid):
		c.JSON(http.StatusConflict, httpserver.Error(44002, "payment status invalid"))
	case errors.Is(err, seat.ErrInvalidSeatAction):
		c.JSON(http.StatusConflict, httpserver.Error(44003, "seat status invalid"))
	default:
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
	}
}
