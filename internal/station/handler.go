package station

import (
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

func (h *Handler) ListActive(c *gin.Context) {
	stations, err := h.service.ListActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	response := make([]StationResponse, 0, len(stations))
	for _, station := range stations {
		response = append(response, StationResponse{
			ID:   station.ID,
			Name: station.Name,
		})
	}

	c.JSON(http.StatusOK, httpserver.Success(response))
}
