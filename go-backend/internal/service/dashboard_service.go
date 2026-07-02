package service

import (
	"tanzanite/internal/domain/order"
	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/repository"
	"time"
)

type DashboardService struct {
	orderRepo        *repository.OrderRepository
	userRepo         *repository.UserRepository
	ticketRepo       *repository.TicketRepository
	subscriptionRepo *repository.SubscriptionRepository
}

func NewDashboardService(
	orderRepo *repository.OrderRepository,
	userRepo *repository.UserRepository,
	ticketRepo *repository.TicketRepository,
	subscriptionRepo *repository.SubscriptionRepository,
) *DashboardService {
	return &DashboardService{
		orderRepo:        orderRepo,
		userRepo:         userRepo,
		ticketRepo:       ticketRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

func (s *DashboardService) GetStats() map[string]interface{} {
	today := time.Now().Truncate(24 * time.Hour)

	orderStats, err := s.orderRepo.GetStats()
	if err != nil {
		orderStats = make(map[string]interface{})
	}

	totalUsers, err := s.userRepo.Count()
	if err != nil {
		totalUsers = 0
	}

	todayUsers, err := s.userRepo.CountByDateRange(today, time.Now())
	if err != nil {
		todayUsers = 0
	}

	ticketStats, err := s.ticketRepo.GetStats()
	if err != nil {
		ticketStats = make(map[string]interface{})
	}

	subscriptionStats, err := s.subscriptionRepo.GetStats()
	if err != nil {
		subscriptionStats = make(map[string]interface{})
	}

	return map[string]interface{}{
		"orders": map[string]interface{}{
			"total":         orderStats["total"],
			"today":         orderStats["today"],
			"pending":       orderStats["pending"],
			"processing":    orderStats["processing"],
			"completed":     orderStats["completed"],
			"revenue":       orderStats["total_revenue"],
			"today_revenue": orderStats["today_revenue"],
		},
		"users": map[string]interface{}{
			"total": totalUsers,
			"today": todayUsers,
		},
		"tickets": map[string]interface{}{
			"total":   ticketStats["total"],
			"open":    ticketStats["open"],
			"pending": ticketStats["pending"],
		},
		"subscriptions": map[string]interface{}{
			"total":  subscriptionStats["total"],
			"active": subscriptionStats["active"],
		},
		"timestamp": time.Now(),
	}
}

func (s *DashboardService) GetRecentOrders(limit int) ([]order.Order, error) {
	return s.orderRepo.FindRecent(limit)
}

func (s *DashboardService) GetRecentUsers(limit int) ([]*user.UserResponse, error) {
	users, err := s.userRepo.FindRecent(limit)
	if err != nil {
		return nil, err
	}

	responses := make([]*user.UserResponse, len(users))
	for i, item := range users {
		responses[i] = item.ToResponse()
	}

	return responses, nil
}

func (s *DashboardService) GetRecentTickets(limit int) ([]ticket.Ticket, error) {
	return s.ticketRepo.FindRecent(limit)
}

func (s *DashboardService) GetSalesChart(days int) (map[string]interface{}, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	salesData, err := s.orderRepo.GetSalesByDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data":       salesData,
		"start_date": startDate,
		"end_date":   endDate,
	}, nil
}
