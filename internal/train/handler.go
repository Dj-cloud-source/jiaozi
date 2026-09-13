package train

import (
	"errors"
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

func (h *Handler) Create(c *gin.Context) {
	var req CreateTrainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}

	train, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTrain):
			c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		case errors.Is(err, ErrTrainAlreadyExists):
			c.JSON(http.StatusConflict, httpserver.Error(42005, "train already exists"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	c.JSON(http.StatusCreated, httpserver.Success(NewAdminTrainResponse(train)))
}

func (h *Handler) Publish(c *gin.Context) {
	trainID, err := strconv.ParseUint(c.Param("train_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}

	train, err := h.service.Publish(c.Request.Context(), trainID)
	if err != nil {
		switch {
		case errors.Is(err, ErrTrainNotFound):
			c.JSON(http.StatusNotFound, httpserver.Error(42001, "train not found"))
		case errors.Is(err, ErrTrainStatusInvalid):
			c.JSON(http.StatusConflict, httpserver.Error(47002, "admin operation not allowed"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(NewAdminTrainResponse(train)))
}

func (h *Handler) List(c *gin.Context) {
	req := ListTrainsRequest{
		TrainNo:            c.Query("train_no"),
		Date:               c.Query("date"),
		Sort:               c.Query("sort"),
		DepartureStationID: parseUintQuery(c.Query("departure_station_id")),
		ArrivalStationID:   parseUintQuery(c.Query("arrival_station_id")),
		Page:               parseIntQuery(c.Query("page")),
		PageSize:           parseIntQuery(c.Query("page_size")),
	}

	trains, err := h.service.List(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidTrain) {
			c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	response := make([]TrainResponse, 0, len(trains))
	for _, train := range trains {
		response = append(response, NewTrainResponse(train))
	}

	c.JSON(http.StatusOK, httpserver.Success(response))
}

func (h *Handler) Detail(c *gin.Context) {
	trainID, err := strconv.ParseUint(c.Param("train_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpserver.Error(40001, "invalid request"))
		return
	}

	train, err := h.service.Detail(c.Request.Context(), trainID)
	if err != nil {
		if errors.Is(err, ErrTrainNotFound) {
			c.JSON(http.StatusNotFound, httpserver.Error(42001, "train not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		return
	}

	c.JSON(http.StatusOK, httpserver.Success(NewTrainResponse(train)))
}

func parseUintQuery(value string) uint64 {
	parsed, _ := strconv.ParseUint(value, 10, 64)
	return parsed
}

func parseIntQuery(value string) int {
	parsed, _ := strconv.Atoi(value)
	return parsed
}
