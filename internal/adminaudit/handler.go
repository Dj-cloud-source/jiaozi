package adminaudit

import (
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

func (h *Handler) List(c *gin.Context) {
	logs, err := h.service.List(c.Request.Context(), Query{
		Page:     parseIntQuery(c.Query("page")),
		PageSize: parseIntQuery(c.Query("page_size")),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(NewListResponse(logs)))
}

func parseIntQuery(value string) int {
	parsed, _ := strconv.Atoi(value)
	return parsed
}
