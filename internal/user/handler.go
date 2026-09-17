package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jiaozi/internal/adminaudit"
	"jiaozi/internal/auth"
	"jiaozi/internal/httpserver"
	"jiaozi/internal/order"
)

type Handler struct {
	service      *Service
	orderService *order.Service
	audit        *adminaudit.Service
}

func NewHandler(service *Service, orderService *order.Service, audit *adminaudit.Service) *Handler {
	return &Handler{service: service, orderService: orderService, audit: audit}
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

func (h *Handler) ResetPassword(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), userID, req); err != nil {
		switch {
		case errors.Is(err, ErrInvalidUser):
			c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		case errors.Is(err, ErrUserNotFound):
			c.JSON(http.StatusNotFound, httpserver.Error(41004, "user not found"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	if err := h.writeResetPasswordAudit(c, userID); err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(ResetPasswordResponse{
		UserID: userID,
		Status: "PASSWORD_RESET",
	}))
}

func parseIntQuery(value string) int {
	parsed, _ := strconv.Atoi(value)
	return parsed
}

func (h *Handler) writeResetPasswordAudit(c *gin.Context, userID uint64) error {
	adminID, ok := currentAdminID(c)
	if !ok {
		return auth.ErrUnauthorized
	}

	resourceID := strconv.FormatUint(userID, 10)
	afterData := map[string]interface{}{
		"password_reset": true,
	}
	afterDataBytes, err := json.Marshal(afterData)
	if err != nil {
		return err
	}
	afterDataJSON := string(afterDataBytes)

	_, err = h.audit.Create(c.Request.Context(), adminaudit.CreateRequest{
		AdminID:      adminID,
		Action:       "RESET_USER_PASSWORD",
		ResourceType: "USER",
		ResourceID:   &resourceID,
		AfterData:    &afterDataJSON,
	})
	return err
}

func currentAdminID(c *gin.Context) (uint64, bool) {
	value, exists := c.Get(auth.ContextAdminIDKey)
	if !exists {
		return 0, false
	}

	adminID, ok := value.(uint64)
	return adminID, ok
}
