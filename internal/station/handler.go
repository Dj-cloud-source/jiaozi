package station

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"jiaozi/internal/adminaudit"
	"jiaozi/internal/auth"
	"jiaozi/internal/httpserver"
	"jiaozi/internal/model"
)

type Handler struct {
	service *Service
	audit   *adminaudit.Service
}

func NewHandler(service *Service, audit *adminaudit.Service) *Handler {
	return &Handler{service: service, audit: audit}
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

func (h *Handler) AdminList(c *gin.Context) {
	stations, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	response := make([]AdminStationResponse, 0, len(stations))
	for _, station := range stations {
		response = append(response, AdminStationResponse{
			ID:     station.ID,
			Name:   station.Name,
			Status: station.Status,
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

	adminID, ok := currentAdminID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, httpserver.Error(41003, "unauthorized"))
		return
	}

	resourceID := strconv.FormatUint(station.ID, 10)
	afterData := map[string]interface{}{
		"id":     station.ID,
		"name":   station.Name,
		"status": station.Status,
	}
	afterDataBytes, err := json.Marshal(afterData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}
	afterDataJSON := string(afterDataBytes)

	if _, err := h.audit.Create(c.Request.Context(), adminaudit.CreateRequest{
		AdminID:      adminID,
		Action:       "CREATE_STATION",
		ResourceType: "STATION",
		ResourceID:   &resourceID,
		AfterData:    &afterDataJSON,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	c.JSON(http.StatusCreated, httpserver.Success(AdminStationResponse{
		ID:     station.ID,
		Name:   station.Name,
		Status: station.Status,
	}))
}

func (h *Handler) Update(c *gin.Context) {
	stationID, err := strconv.ParseUint(c.Param("station_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}

	var req UpdateStationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}

	beforeStation, err := h.service.Detail(c.Request.Context(), stationID)
	if err != nil {
		switch {
		case errors.Is(err, ErrStationNotFound):
			c.JSON(http.StatusNotFound, httpserver.Error(42006, "station not found"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	station, err := h.service.UpdateName(c.Request.Context(), stationID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidStationName):
			c.JSON(http.StatusBadRequest, httpserver.Error(42004, "invalid station"))
		case errors.Is(err, ErrStationNotFound):
			c.JSON(http.StatusNotFound, httpserver.Error(42006, "station not found"))
		case errors.Is(err, ErrStationExists):
			c.JSON(http.StatusConflict, httpserver.Error(42004, "invalid station"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	adminID, ok := currentAdminID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, httpserver.Error(41003, "unauthorized"))
		return
	}

	resourceID := strconv.FormatUint(station.ID, 10)
	beforeDataJSON, err := marshalStationAuditData(beforeStation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}
	afterDataJSON, err := marshalStationAuditData(station)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	if _, err := h.audit.Create(c.Request.Context(), adminaudit.CreateRequest{
		AdminID:      adminID,
		Action:       "UPDATE_STATION",
		ResourceType: "STATION",
		ResourceID:   &resourceID,
		BeforeData:   &beforeDataJSON,
		AfterData:    &afterDataJSON,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(AdminStationResponse{
		ID:     station.ID,
		Name:   station.Name,
		Status: station.Status,
	}))
}

func (h *Handler) Disable(c *gin.Context) {
	stationID, err := strconv.ParseUint(c.Param("station_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}

	beforeStation, err := h.service.Detail(c.Request.Context(), stationID)
	if err != nil {
		switch {
		case errors.Is(err, ErrStationNotFound):
			c.JSON(http.StatusNotFound, httpserver.Error(42006, "station not found"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	station, err := h.service.Disable(c.Request.Context(), stationID)
	if err != nil {
		switch {
		case errors.Is(err, ErrStationNotFound):
			c.JSON(http.StatusNotFound, httpserver.Error(42006, "station not found"))
		case errors.Is(err, ErrStationStatusInvalid):
			c.JSON(http.StatusConflict, httpserver.Error(42007, "station status invalid"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	adminID, ok := currentAdminID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, httpserver.Error(41003, "unauthorized"))
		return
	}

	resourceID := strconv.FormatUint(station.ID, 10)
	beforeDataJSON, err := marshalStationAuditData(beforeStation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}
	afterDataJSON, err := marshalStationAuditData(station)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	if _, err := h.audit.Create(c.Request.Context(), adminaudit.CreateRequest{
		AdminID:      adminID,
		Action:       "DISABLE_STATION",
		ResourceType: "STATION",
		ResourceID:   &resourceID,
		BeforeData:   &beforeDataJSON,
		AfterData:    &afterDataJSON,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(AdminStationResponse{
		ID:     station.ID,
		Name:   station.Name,
		Status: station.Status,
	}))
}

func currentAdminID(c *gin.Context) (uint64, bool) {
	value, exists := c.Get(auth.ContextAdminIDKey)
	if !exists {
		return 0, false
	}

	adminID, ok := value.(uint64)
	return adminID, ok
}

func marshalStationAuditData(station model.Station) (string, error) {
	data := map[string]interface{}{
		"id":     station.ID,
		"name":   station.Name,
		"status": station.Status,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(dataBytes), nil
}
