package service

import (
	"math"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/domain/ticket"
	"commerce-platform/internal/repository"
)

type CustomerServiceAnalyticsService struct {
	ticketService  *TicketService
	contextService *CustomerServiceContextService
	orderRepo      *repository.OrderRepository
}

func NewCustomerServiceAnalyticsService(
	ticketService *TicketService,
	contextService *CustomerServiceContextService,
	orderRepo *repository.OrderRepository,
) *CustomerServiceAnalyticsService {
	return &CustomerServiceAnalyticsService{
		ticketService:  ticketService,
		contextService: contextService,
		orderRepo:      orderRepo,
	}
}

type CustomerServiceAnalyticsInput struct {
	Date                  string
	TimezoneOffsetMinutes int
	AgentUserID           uint
	CanViewAll            bool
}

type CustomerServiceAnalytics struct {
	Date                         string                           `json:"date"`
	TimezoneOffsetMinutes        int                              `json:"timezone_offset_minutes"`
	WindowStart                  time.Time                        `json:"window_start"`
	WindowEnd                    time.Time                        `json:"window_end"`
	TotalConversations           int                              `json:"total_conversations"`
	KnownRegionCount             int                              `json:"known_region_count"`
	UnknownRegionCount           int                              `json:"unknown_region_count"`
	MemberCustomerCount          int                              `json:"member_customer_count"`
	ConvertedMemberCustomerCount int                              `json:"converted_member_customer_count"`
	MemberConversionRate         float64                          `json:"member_conversion_rate"`
	AverageReplyIntervalSeconds  float64                          `json:"average_reply_interval_seconds"`
	ReplyIntervalCount           int                              `json:"reply_interval_count"`
	UnansweredCustomerTurns      int                              `json:"unanswered_customer_turns"`
	Regions                      []CustomerServiceAnalyticsRegion `json:"regions"`
}

type CustomerServiceAnalyticsRegion struct {
	RegionLabel  string  `json:"region_label"`
	RegionStatus string  `json:"region_status"`
	Count        int     `json:"count"`
	MemberCount  int     `json:"member_count"`
	VisitorCount int     `json:"visitor_count"`
	Percent      float64 `json:"percent"`
}

func (s *CustomerServiceAnalyticsService) ForAgent(input CustomerServiceAnalyticsInput) (*CustomerServiceAnalytics, error) {
	start, end, date := customerServiceAnalyticsWindow(input.Date, input.TimezoneOffsetMinutes)
	result := &CustomerServiceAnalytics{
		Date:                  date,
		TimezoneOffsetMinutes: input.TimezoneOffsetMinutes,
		WindowStart:           start,
		WindowEnd:             end,
		Regions:               []CustomerServiceAnalyticsRegion{},
	}
	if s == nil || s.ticketService == nil || s.contextService == nil {
		return result, nil
	}

	conversations, err := s.ticketService.ListCustomerServiceConversationsInWindowForAgent(
		start,
		end,
		input.AgentUserID,
		input.CanViewAll,
	)
	if err != nil {
		return nil, err
	}

	regionMap := map[string]*CustomerServiceAnalyticsRegion{}
	memberCustomerIDs := map[uint]struct{}{}
	replyIntervalTotalSeconds := 0.0
	replyIntervalCount := 0
	unansweredCustomerTurns := 0
	for _, conversation := range conversations {
		label := strings.TrimSpace(s.contextService.conversationCoarseRegionLabel(conversation))
		status := "unknown"
		if label != "" {
			status = "captured"
		}
		if label == "" {
			label = "未知区域"
		}
		if status == "" {
			status = "unknown"
		}

		item := regionMap[label]
		if item == nil {
			item = &CustomerServiceAnalyticsRegion{
				RegionLabel:  label,
				RegionStatus: status,
			}
			regionMap[label] = item
		}
		item.Count++
		if conversation.CustomerUserID != nil && *conversation.CustomerUserID > 0 {
			item.MemberCount++
		} else {
			item.VisitorCount++
		}

		if conversation.CustomerUserID != nil && *conversation.CustomerUserID > 0 {
			memberCustomerIDs[*conversation.CustomerUserID] = struct{}{}
		}

		intervalTotal, intervalCount, unansweredTurns := customerServiceReplyMetrics(conversation.Messages, start, end)
		replyIntervalTotalSeconds += intervalTotal
		replyIntervalCount += intervalCount
		unansweredCustomerTurns += unansweredTurns
	}

	result.TotalConversations = len(conversations)
	result.MemberCustomerCount = len(memberCustomerIDs)
	result.ReplyIntervalCount = replyIntervalCount
	result.UnansweredCustomerTurns = unansweredCustomerTurns
	if replyIntervalCount > 0 {
		result.AverageReplyIntervalSeconds = math.Round((replyIntervalTotalSeconds/float64(replyIntervalCount))*10) / 10
	}

	if s.orderRepo != nil && len(memberCustomerIDs) > 0 {
		memberUserIDs := make([]uint, 0, len(memberCustomerIDs))
		for userID := range memberCustomerIDs {
			memberUserIDs = append(memberUserIDs, userID)
		}
		convertedUserIDs, err := s.orderRepo.FindPaidUserIDsInWindow(memberUserIDs, start, end)
		if err != nil {
			return nil, err
		}
		result.ConvertedMemberCustomerCount = len(convertedUserIDs)
		if result.MemberCustomerCount > 0 {
			result.MemberConversionRate = math.Round(
				(float64(result.ConvertedMemberCustomerCount)/float64(result.MemberCustomerCount))*1000,
			) / 10
		}
	}

	for _, item := range regionMap {
		if item.RegionStatus == "captured" {
			result.KnownRegionCount += item.Count
		} else {
			result.UnknownRegionCount += item.Count
		}
		if result.TotalConversations > 0 {
			item.Percent = math.Round((float64(item.Count)/float64(result.TotalConversations))*1000) / 10
		}
		result.Regions = append(result.Regions, *item)
	}

	sort.SliceStable(result.Regions, func(i, j int) bool {
		if result.Regions[i].Count == result.Regions[j].Count {
			return result.Regions[i].RegionLabel < result.Regions[j].RegionLabel
		}
		return result.Regions[i].Count > result.Regions[j].Count
	})

	return result, nil
}

// customerServiceReplyMetrics pairs the latest customer turn with the next
// staff reply. Consecutive customer messages therefore count as one pending
// turn, which keeps a single staff reply from inflating the average.
func customerServiceReplyMetrics(messages []ticket.TicketMessage, start, end time.Time) (float64, int, int) {
	var pendingCustomerAt *time.Time
	totalSeconds := 0.0
	replyCount := 0
	unansweredTurns := 0

	for _, message := range messages {
		if message.IsInternal {
			continue
		}

		if message.CreatedAt.Before(start) {
			if message.IsStaff {
				pendingCustomerAt = nil
			} else {
				messageAt := message.CreatedAt
				pendingCustomerAt = &messageAt
			}
			continue
		}
		if !message.CreatedAt.Before(end) {
			continue
		}

		if message.IsStaff {
			if pendingCustomerAt == nil || message.CreatedAt.Before(*pendingCustomerAt) {
				continue
			}
			totalSeconds += message.CreatedAt.Sub(*pendingCustomerAt).Seconds()
			replyCount++
			pendingCustomerAt = nil
			continue
		}

		messageAt := message.CreatedAt
		pendingCustomerAt = &messageAt
	}

	if pendingCustomerAt != nil {
		unansweredTurns = 1
	}

	return totalSeconds, replyCount, unansweredTurns
}

func customerServiceAnalyticsWindow(date string, timezoneOffsetMinutes int) (time.Time, time.Time, string) {
	location := time.FixedZone("admin-local", timezoneOffsetMinutes*60)
	date = strings.TrimSpace(date)
	if date == "" {
		date = time.Now().In(location).Format("2006-01-02")
	}

	startLocal, err := time.ParseInLocation("2006-01-02", date, location)
	if err != nil {
		startLocal = time.Now().In(location)
		startLocal = time.Date(startLocal.Year(), startLocal.Month(), startLocal.Day(), 0, 0, 0, 0, location)
		date = startLocal.Format("2006-01-02")
	}

	endLocal := startLocal.AddDate(0, 0, 1)
	return startLocal.UTC(), endLocal.UTC(), date
}
