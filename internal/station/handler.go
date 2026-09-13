package station

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

func (h *Handler) Create(c *gin.Context) {
	var req CreateStationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}

	station, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidStationName):
			c.JSON(http.StatusBadRequest, httpserver.Error(42004, "invalid station"))
		case errors.Is(err, ErrStationExists):
			c.JSON(http.StatusConflict, httpserver.Error(42004, "invalid station"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	c.JSON(http.StatusCreated, httpserver.Success(AdminStationResponse{
		ID:     station.ID,
		Name:   station.Name,
		Status: station.Status,
	}))
}
