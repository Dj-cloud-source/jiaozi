package orderapp

import (
	"context"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"jiaozi/internal/common"
	"jiaozi/internal/model"
	"jiaozi/internal/order"
	"jiaozi/internal/passenger"
	"jiaozi/internal/payment"
	"jiaozi/internal/seat"
	"jiaozi/internal/train"
)

const paymentWindow = 10 * time.Minute

type Service struct {
	db                  *sqlx.DB
	trainService        *train.Service
	passengerRepository *passenger.Repository
	seatRepository      *seat.Repository
	orderRepository     *order.Repository
	paymentRepository   *payment.Repository
}

type CreateOrderRequest struct {
	UserID          uint64
	TrainID         uint64
	SeatClass       string
	PassengerName   string
	PassengerIDCard string
}

type CreateOrderResult struct {
	Order     model.TicketOrder
	Payment   model.Payment
	Passenger model.Passenger
	Seat      model.Seat
}

func NewService(
	db *sqlx.DB,
	trainService *train.Service,
	passengerRepository *passenger.Repository,
	seatRepository *seat.Repository,
	orderRepository *order.Repository,
	paymentRepository *payment.Repository,
) *Service {
	return &Service{
		db:                  db,
		trainService:        trainService,
		passengerRepository: passengerRepository,
		seatRepository:      seatRepository,
		orderRepository:     orderRepository,
		paymentRepository:   paymentRepository,
	}
}

func (s *Service) CreateOrder(ctx context.Context, req CreateOrderRequest) (CreateOrderResult, error) {
	req.SeatClass = strings.TrimSpace(req.SeatClass)
	req.PassengerName = strings.TrimSpace(req.PassengerName)
	req.PassengerIDCard = strings.TrimSpace(req.PassengerIDCard)

	if req.UserID == 0 || req.TrainID == 0 || req.SeatClass == "" || req.PassengerName == "" || req.PassengerIDCard == "" {
		return CreateOrderResult{}, ErrInvalidCreateOrder
	}

	trainView, err := s.trainService.Detail(ctx, req.TrainID)
	if err != nil {
		return CreateOrderResult{}, err
	}
	if trainView.Status != "ON_SALE" {
		return CreateOrderResult{}, ErrTrainNotOnSale
	}

	ticketPrice, err := priceForSeatClass(trainView, req.SeatClass)
	if err != nil {
		return CreateOrderResult{}, err
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return CreateOrderResult{}, err
	}
	defer tx.Rollback()

	passengerService := passenger.NewService(s.passengerRepository.WithExecutor(tx))
	seatService := seat.NewService(s.seatRepository.WithExecutor(tx))
	orderService := order.NewService(s.orderRepository.WithExecutor(tx))
	paymentService := payment.NewService(s.paymentRepository.WithExecutor(tx))

	hasActiveTicket, err := passengerService.HasActiveTicket(ctx, req.PassengerIDCard, req.TrainID)
	if err != nil {
		return CreateOrderResult{}, err
	}
	if hasActiveTicket {
		return CreateOrderResult{}, ErrPassengerAlreadyHasTicket
	}

	passengerModel, err := passengerService.Create(ctx, passenger.CreatePassengerParams{
		UserID: req.UserID,
		Name:   req.PassengerName,
		IDCard: req.PassengerIDCard,
	})
	if err != nil {
		return CreateOrderResult{}, err
	}

	availableSeats, err := seatService.ListAvailableSeats(ctx, req.TrainID, req.SeatClass)
	if err != nil {
		return CreateOrderResult{}, err
	}

	selectedSeat, err := selectSeat(availableSeats)
	if err != nil {
		return CreateOrderResult{}, err
	}

	orderID, err := common.NewUUID()
	if err != nil {
		return CreateOrderResult{}, err
	}

	locked, err := seatService.LockSeats(ctx, seat.LockSeatsParams{
		TrainID:   req.TrainID,
		SeatClass: req.SeatClass,
		SeatNos:   []uint64{selectedSeat.SeatNo},
		OrderID:   orderID,
	})
	if err != nil {
		return CreateOrderResult{}, err
	}
	if locked != 1 {
		return CreateOrderResult{}, ErrSeatLockFailed
	}

	paymentDeadline := time.Now().Add(paymentWindow)
	orderModel, err := orderService.Create(ctx, order.CreateTicketOrderRequest{
		ID:              orderID,
		UserID:          req.UserID,
		PassengerID:     passengerModel.ID,
		TrainID:         req.TrainID,
		SeatID:          selectedSeat.ID,
		TicketPrice:     ticketPrice,
		PaymentDeadline: paymentDeadline,
	})
	if err != nil {
		return CreateOrderResult{}, err
	}

	paymentModel, err := paymentService.Create(ctx, payment.CreatePaymentRequest{
		UserID:          req.UserID,
		Amount:          ticketPrice,
		PaymentDeadline: paymentDeadline,
	})
	if err != nil {
		return CreateOrderResult{}, err
	}

	err = paymentService.CreateOrderLinks(ctx, []payment.PaymentOrderLink{
		{
			PaymentID: paymentModel.ID,
			OrderID:   orderModel.ID,
			Amount:    ticketPrice,
		},
	})
	if err != nil {
		return CreateOrderResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return CreateOrderResult{}, err
	}

	return CreateOrderResult{
		Order:     orderModel,
		Payment:   paymentModel,
		Passenger: passengerModel,
		Seat:      selectedSeat,
	}, nil
}

func (s *Service) ListOrders(ctx context.Context, userID uint64) ([]order.OrderView, error) {
	return order.NewService(s.orderRepository).ListByUser(ctx, userID)
}

func (s *Service) OrderDetail(ctx context.Context, orderID string, userID uint64) (order.OrderView, error) {
	return order.NewService(s.orderRepository).Detail(ctx, orderID, userID)
}

func priceForSeatClass(trainView train.TrainView, seatClass string) (string, error) {
	switch seatClass {
	case "FIRST_CLASS":
		if trainView.FirstClassAvailableCount == 0 {
			return "", ErrNoAvailableSeat
		}
		return trainView.FirstClassPrice, nil
	case "SECOND_CLASS":
		if trainView.SecondClassAvailableCount == 0 {
			return "", ErrNoAvailableSeat
		}
		return trainView.SecondClassPrice, nil
	default:
		return "", ErrInvalidCreateOrder
	}
}

func selectSeat(availableSeats []model.Seat) (model.Seat, error) {
	seatNos := make([]uint64, 0, len(availableSeats))
	seatsByNo := make(map[uint64]model.Seat, len(availableSeats))
	for _, availableSeat := range availableSeats {
		seatNos = append(seatNos, availableSeat.SeatNo)
		seatsByNo[availableSeat.SeatNo] = availableSeat
	}

	selectedSeatNos, err := seat.AllocateSeats(seatNos, 1)
	if err != nil {
		return model.Seat{}, ErrNoAvailableSeat
	}

	selectedSeat, ok := seatsByNo[selectedSeatNos[0]]
	if !ok {
		return model.Seat{}, ErrNoAvailableSeat
	}

	return selectedSeat, nil
}
