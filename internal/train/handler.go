package train

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

	if err := h.writeCreateTrainAudit(c, train); err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
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

	beforeTrain, err := h.service.Detail(c.Request.Context(), trainID)
	if err != nil {
		switch {
		case errors.Is(err, ErrTrainNotFound):
			c.JSON(http.StatusNotFound, httpserver.Error(42001, "train not found"))
		default:
			c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
		}
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

	if err := h.writePublishTrainAudit(c, beforeTrain, train); err != nil {
		c.JSON(http.StatusInternalServerError, httpserver.Error(50000, "internal server error"))
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

func (h *Handler) writeCreateTrainAudit(c *gin.Context, train model.Train) error {
	adminID, ok := currentAdminID(c)
	if !ok {
		return auth.ErrUnauthorized
	}

	resourceID := strconv.FormatUint(train.ID, 10)
	afterData := map[string]interface{}{
		"id":                      train.ID,
		"train_no":                train.TrainNo,
		"departure_date":          train.DepartureDate.Format("2006-01-02"),
		"departure_station_id":    train.DepartureStationID,
		"arrival_station_id":      train.ArrivalStationID,
		"departure_time":          train.DepartureTime.Format("2006-01-02T15:04:05-07:00"),
		"arrival_time":            train.ArrivalTime.Format("2006-01-02T15:04:05-07:00"),
		"sale_start_time":         train.SaleStartTime.Format("2006-01-02T15:04:05-07:00"),
		"first_class_price":       train.FirstClassPrice,
		"second_class_price":      train.SecondClassPrice,
		"first_class_seat_count":  train.FirstClassSeatCount,
		"second_class_seat_count": train.SecondClassSeatCount,
		"status":                  train.Status,
	}
	afterDataBytes, err := json.Marshal(afterData)
	if err != nil {
		return err
	}
	afterDataJSON := string(afterDataBytes)

	_, err = h.audit.Create(c.Request.Context(), adminaudit.CreateRequest{
		AdminID:      adminID,
		Action:       "CREATE_TRAIN",
		ResourceType: "TRAIN",
		ResourceID:   &resourceID,
		AfterData:    &afterDataJSON,
	})
	return err
}

func (h *Handler) writePublishTrainAudit(c *gin.Context, beforeTrain TrainView, afterTrain model.Train) error {
	adminID, ok := currentAdminID(c)
	if !ok {
		return auth.ErrUnauthorized
	}

	resourceID := strconv.FormatUint(afterTrain.ID, 10)
	beforeData := map[string]interface{}{
		"id":     beforeTrain.ID,
		"status": beforeTrain.Status,
	}
	beforeDataBytes, err := json.Marshal(beforeData)
	if err != nil {
		return err
	}
	beforeDataJSON := string(beforeDataBytes)

	afterData := map[string]interface{}{
		"id":     afterTrain.ID,
		"status": afterTrain.Status,
	}
	afterDataBytes, err := json.Marshal(afterData)
	if err != nil {
		return err
	}
	afterDataJSON := string(afterDataBytes)

	_, err = h.audit.Create(c.Request.Context(), adminaudit.CreateRequest{
		AdminID:      adminID,
		Action:       "PUBLISH_TRAIN",
		ResourceType: "TRAIN",
		ResourceID:   &resourceID,
		BeforeData:   &beforeDataJSON,
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
