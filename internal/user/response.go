package user

import (
	"time"

	"jiaozi/internal/model"
	"jiaozi/internal/order"
)

type CurrentUserResponse struct {
	ID        uint64    `json:"id"`
	Phone     string    `json:"phone"`
	Nickname  string    `json:"nickname"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminUserResponse struct {
	ID        uint64 `json:"id"`
	Phone     string `json:"phone"`
	Nickname  string `json:"nickname"`
	CreatedAt string `json:"created_at"`
}

type AdminUserDetailResponse struct {
	ID        uint64                   `json:"id"`
	Phone     string                   `json:"phone"`
	Nickname  string                   `json:"nickname"`
	CreatedAt string                   `json:"created_at"`
	Orders    []AdminUserOrderResponse `json:"orders"`
}

type AdminUserOrderResponse struct {
	OrderID          string `json:"order_id"`
	TrainNo          string `json:"train_no"`
	DepartureStation string `json:"departure_station"`
	ArrivalStation   string `json:"arrival_station"`
	DepartureTime    string `json:"departure_time"`
	PassengerName    string `json:"passenger_name"`
	SeatClass        string `json:"seat_class"`
	SeatNo           uint64 `json:"seat_no"`
	TicketPrice      string `json:"ticket_price"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
}

func NewAdminUserResponse(user model.User) AdminUserResponse {
	return AdminUserResponse{
		ID:        user.ID,
		Phone:     user.Phone,
		Nickname:  user.Nickname,
		CreatedAt: formatTime(user.CreatedAt),
	}
}

func NewAdminUserListResponse(users []model.User) []AdminUserResponse {
	response := make([]AdminUserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, NewAdminUserResponse(user))
	}
	return response
}

func NewAdminUserDetailResponse(user model.User, orders []order.OrderView) AdminUserDetailResponse {
	response := AdminUserDetailResponse{
		ID:        user.ID,
		Phone:     user.Phone,
		Nickname:  user.Nickname,
		CreatedAt: formatTime(user.CreatedAt),
		Orders:    make([]AdminUserOrderResponse, 0, len(orders)),
	}
	for _, orderView := range orders {
		response.Orders = append(response.Orders, AdminUserOrderResponse{
			OrderID:          orderView.OrderID,
			TrainNo:          orderView.TrainNo,
			DepartureStation: orderView.DepartureStationName,
			ArrivalStation:   orderView.ArrivalStationName,
			DepartureTime:    formatTime(orderView.DepartureTime),
			PassengerName:    orderView.PassengerName,
			SeatClass:        orderView.SeatClass,
			SeatNo:           orderView.SeatNo,
			TicketPrice:      orderView.TicketPrice,
			Status:           orderView.Status,
			CreatedAt:        formatTime(orderView.CreatedAt),
		})
	}

	return response
}

func formatTime(value time.Time) string {
	return value.Format("2006-01-02T15:04:05-07:00")
}
