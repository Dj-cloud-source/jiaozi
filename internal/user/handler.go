package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jiaozi/internal/auth"
	"jiaozi/internal/httpserver"
	"jiaozi/internal/order"
)

type Handler struct {
	service      *Service
	orderService *order.Service
}

func NewHandler(service *Service, orderService *order.Service) *Handler {
	return &Handler{service: service, orderService: orderService}
}

func (h *Handler) Me(c *gin.Context) {
	value, exists := c.Get(auth.ContextUserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, httpserver.Error(41003, "unauthorized"))
		return
	}

	userID := value.(uint64)
	currentUser, err := h.service.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(CurrentUserResponse{
		ID:        currentUser.ID,
		Phone:     currentUser.Phone,
		Nickname:  currentUser.Nickname,
		CreatedAt: currentUser.CreatedAt,
	}))
}

func (h *Handler) AdminList(c *gin.Context) {
	users, err := h.service.AdminList(c.Request.Context(), AdminListUsersRequest{
		Page:     parseIntQuery(c.Query("page")),
		PageSize: parseIntQuery(c.Query("page_size")),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(NewAdminUserListResponse(users)))
}

func (h *Handler) AdminDetail(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}

	user, err := h.service.AdminDetail(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			c.JSON(http.StatusNotFound, httpserver.Error(41004, "user not found"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	orders, err := h.orderService.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(NewAdminUserDetailResponse(user, orders)))
}

func parseIntQuery(value string) int {
	parsed, _ := strconv.Atoi(value)
	return parsed
}
