package train

import (
	"context"
	"regexp"
	"strings"
	"time"

	"jiaozi/internal/model"
)

var trainNoPattern = regexp.MustCompile(`^[A-Z][0-9]{1,4}$`)
var pricePattern = regexp.MustCompile(`^[0-9]+(\.[0-9]{1,2})?$`)

type repository interface {
	Create(ctx context.Context, params CreateTrainParams) (model.Train, error)
	Publish(ctx context.Context, trainID uint64) (model.Train, error)
	List(ctx context.Context, query ListTrainQuery) ([]TrainView, error)
	FindViewByID(ctx context.Context, trainID uint64) (TrainView, error)
	OpenDueTrains(ctx context.Context, now time.Time) (int64, error)
	StopDueTrains(ctx context.Context, now time.Time) (int64, error)
	DepartDueTrains(ctx context.Context, now time.Time) (int64, error)
}

type Service struct {
	repository repository
}

type CreateTrainParams struct {
	TrainNo              string
	DepartureDate        time.Time
	DepartureStationID   uint64
	ArrivalStationID     uint64
	DepartureTime        time.Time
	ArrivalTime          time.Time
	SaleStartTime        time.Time
	FirstClassPrice      string
	SecondClassPrice     string
	FirstClassSeatCount  uint64
	SecondClassSeatCount uint64
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, req CreateTrainRequest) (model.Train, error) {
	params, err := buildCreateTrainParams(req)
	if err != nil {
		return model.Train{}, err
	}

	return s.repository.Create(ctx, params)
}

func (s *Service) Publish(ctx context.Context, trainID uint64) (model.Train, error) {
	if trainID == 0 {
		return model.Train{}, ErrTrainNotFound
	}

	return s.repository.Publish(ctx, trainID)
}

func (s *Service) List(ctx context.Context, req ListTrainsRequest) ([]TrainView, error) {
	query, err := buildListTrainQuery(req)
	if err != nil {
		return nil, err
	}

	return s.repository.List(ctx, query)
}

func (s *Service) Detail(ctx context.Context, trainID uint64) (TrainView, error) {
	if trainID == 0 {
		return TrainView{}, ErrTrainNotFound
	}

	return s.repository.FindViewByID(ctx, trainID)
}

func (s *Service) OpenDueTrains(ctx context.Context, now time.Time) (int64, error) {
	return s.repository.OpenDueTrains(ctx, now)
}

func (s *Service) StopDueTrains(ctx context.Context, now time.Time) (int64, error) {
	return s.repository.StopDueTrains(ctx, now)
}

func (s *Service) DepartDueTrains(ctx context.Context, now time.Time) (int64, error) {
	return s.repository.DepartDueTrains(ctx, now)
}

func buildCreateTrainParams(req CreateTrainRequest) (CreateTrainParams, error) {
	trainNo := strings.TrimSpace(req.TrainNo)
	if !trainNoPattern.MatchString(trainNo) {
		return CreateTrainParams{}, ErrInvalidTrain
	}
	if req.DepartureStationID == 0 || req.ArrivalStationID == 0 || req.DepartureStationID == req.ArrivalStationID {
		return CreateTrainParams{}, ErrInvalidTrain
	}

	departureDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.DepartureDate))
	if err != nil {
		return CreateTrainParams{}, ErrInvalidTrain
	}

	departureTime, err := time.Parse(time.RFC3339, strings.TrimSpace(req.DepartureTime))
	if err != nil {
		return CreateTrainParams{}, ErrInvalidTrain
	}

	arrivalTime, err := time.Parse(time.RFC3339, strings.TrimSpace(req.ArrivalTime))
	if err != nil {
		return CreateTrainParams{}, ErrInvalidTrain
	}

	saleStartTime, err := time.Parse(time.RFC3339, strings.TrimSpace(req.SaleStartTime))
	if err != nil {
		return CreateTrainParams{}, ErrInvalidTrain
	}

	if !sameDate(departureDate, departureTime) || !sameDate(departureDate, arrivalTime) {
		return CreateTrainParams{}, ErrInvalidTrain
	}
	if !arrivalTime.After(departureTime) {
		return CreateTrainParams{}, ErrInvalidTrain
	}
	if req.FirstClassSeatCount == 0 && req.SecondClassSeatCount == 0 {
		return CreateTrainParams{}, ErrInvalidTrain
	}

	firstClassPrice := strings.TrimSpace(req.FirstClassPrice)
	secondClassPrice := strings.TrimSpace(req.SecondClassPrice)
	if req.FirstClassSeatCount > 0 && firstClassPrice == "" {
		return CreateTrainParams{}, ErrInvalidTrain
	}
	if req.SecondClassSeatCount > 0 && secondClassPrice == "" {
		return CreateTrainParams{}, ErrInvalidTrain
	}
	if req.FirstClassSeatCount > 0 && !pricePattern.MatchString(firstClassPrice) {
		return CreateTrainParams{}, ErrInvalidTrain
	}
	if req.SecondClassSeatCount > 0 && !pricePattern.MatchString(secondClassPrice) {
		return CreateTrainParams{}, ErrInvalidTrain
	}
	if req.FirstClassSeatCount == 0 {
		firstClassPrice = ""
	}
	if req.SecondClassSeatCount == 0 {
		secondClassPrice = ""
	}

	return CreateTrainParams{
		TrainNo:              trainNo,
		DepartureDate:        departureDate,
		DepartureStationID:   req.DepartureStationID,
		ArrivalStationID:     req.ArrivalStationID,
		DepartureTime:        departureTime,
		ArrivalTime:          arrivalTime,
		SaleStartTime:        saleStartTime,
		FirstClassPrice:      firstClassPrice,
		SecondClassPrice:     secondClassPrice,
		FirstClassSeatCount:  req.FirstClassSeatCount,
		SecondClassSeatCount: req.SecondClassSeatCount,
	}, nil
}

func sameDate(expectedDate time.Time, value time.Time) bool {
	year, month, day := value.Date()
	expectedYear, expectedMonth, expectedDay := expectedDate.Date()
	return year == expectedYear && month == expectedMonth && day == expectedDay
}

func buildListTrainQuery(req ListTrainsRequest) (ListTrainQuery, error) {
	date, err := time.Parse("2006-01-02", strings.TrimSpace(req.Date))
	if err != nil {
		return ListTrainQuery{}, ErrInvalidTrain
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	sort := strings.TrimSpace(req.Sort)
	if sort == "" {
		sort = "time_asc"
	}
	if sort != "time_asc" && sort != "time_desc" && sort != "price_asc" && sort != "price_desc" {
		return ListTrainQuery{}, ErrInvalidTrain
	}

	trainNo := strings.TrimSpace(req.TrainNo)
	if trainNo != "" {
		if !trainNoPattern.MatchString(trainNo) {
			return ListTrainQuery{}, ErrInvalidTrain
		}
		return ListTrainQuery{
			TrainNo:  trainNo,
			Date:     date,
			Page:     page,
			PageSize: pageSize,
			Sort:     sort,
		}, nil
	}

	if req.DepartureStationID == 0 || req.ArrivalStationID == 0 || req.DepartureStationID == req.ArrivalStationID {
		return ListTrainQuery{}, ErrInvalidTrain
	}

	return ListTrainQuery{
		DepartureStationID: req.DepartureStationID,
		ArrivalStationID:   req.ArrivalStationID,
		Date:               date,
		Page:               page,
		PageSize:           pageSize,
		Sort:               sort,
	}, nil
}
