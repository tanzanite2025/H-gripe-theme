package tracking

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// TrackingService 物流追踪服务接口
type TrackingService interface {
	Track(ctx context.Context, trackingNumber, carrier string) (*TrackingInfo, error)
	BatchTrack(ctx context.Context, trackings []TrackingRequest) ([]*TrackingInfo, error)
	DetectCarrier(ctx context.Context, trackingNumber string) ([]string, error)
}

// TrackingRequest 追踪请求
type TrackingRequest struct {
	TrackingNumber string `json:"tracking_number"`
	Carrier        string `json:"carrier,omitempty"`
}

// TrackingInfo 物流追踪信息
type TrackingInfo struct {
	TrackingNumber string           `json:"tracking_number"`
	Carrier        string           `json:"carrier"`
	Status         string           `json:"status"`
	StatusCode     int              `json:"status_code"`
	Events         []TrackingEvent  `json:"events"`
	Origin         *Location        `json:"origin,omitempty"`
	Destination    *Location        `json:"destination,omitempty"`
	EstimatedTime  *time.Time       `json:"estimated_time,omitempty"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// TrackingEvent 物流事件
type TrackingEvent struct {
	Time        time.Time `json:"time"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Location    *Location `json:"location,omitempty"`
}

// Location 位置信息
type Location struct {
	City    string `json:"city,omitempty"`
	State   string `json:"state,omitempty"`
	Country string `json:"country,omitempty"`
	ZipCode string `json:"zip_code,omitempty"`
}

// Config 追踪服务配置
type Config struct {
	Provider string // 17track, aftership, etc.
	APIKey   string
	BaseURL  string
	Timeout  time.Duration
}

// track17Service 17TRACK 服务实现
type track17Service struct {
	config     *Config
	httpClient *http.Client
}

// NewTrackingService 创建物流追踪服务
func NewTrackingService(config *Config) TrackingService {
	return &track17Service{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Track 追踪单个物流
func (s *track17Service) Track(ctx context.Context, trackingNumber, carrier string) (*TrackingInfo, error) {
	// 验证输入
	if err := validateTrackingNumber(trackingNumber); err != nil {
		return nil, err
	}

	// 构建请求
	reqBody := map[string]interface{}{
		"number": trackingNumber,
	}
	if carrier != "" {
		reqBody["carrier"] = carrier
	}

	// 发送请求
	resp, err := s.sendRequest(ctx, "/track", reqBody)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result struct {
		Code int `json:"code"`
		Data struct {
			Accepted []struct {
				Number  string `json:"number"`
				Carrier string `json:"carrier"`
				Track   struct {
					Status     int    `json:"status"`
					StatusText string `json:"status_text"`
					Events     []struct {
						Time        string `json:"time"`
						Status      string `json:"status"`
						Description string `json:"description"`
						Location    string `json:"location"`
					} `json:"events"`
					Origin      string `json:"origin"`
					Destination string `json:"destination"`
					UpdatedAt   string `json:"updated_at"`
				} `json:"track"`
			} `json:"accepted"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API error: code %d", result.Code)
	}

	if len(result.Data.Accepted) == 0 {
		return nil, fmt.Errorf("tracking number not found")
	}

	// 转换为标准格式
	accepted := result.Data.Accepted[0]
	trackInfo := &TrackingInfo{
		TrackingNumber: accepted.Number,
		Carrier:        accepted.Carrier,
		Status:         accepted.Track.StatusText,
		StatusCode:     accepted.Track.Status,
		Events:         make([]TrackingEvent, 0),
	}

	// 解析事件
	for _, event := range accepted.Track.Events {
		eventTime, _ := time.Parse(time.RFC3339, event.Time)
		trackInfo.Events = append(trackInfo.Events, TrackingEvent{
			Time:        eventTime,
			Status:      event.Status,
			Description: event.Description,
			Location:    parseLocation(event.Location),
		})
	}

	// 解析更新时间
	if accepted.Track.UpdatedAt != "" {
		updatedAt, _ := time.Parse(time.RFC3339, accepted.Track.UpdatedAt)
		trackInfo.UpdatedAt = updatedAt
	}

	return trackInfo, nil
}

// BatchTrack 批量追踪
func (s *track17Service) BatchTrack(ctx context.Context, trackings []TrackingRequest) ([]*TrackingInfo, error) {
	// 构建请求
	numbers := make([]map[string]string, len(trackings))
	for i, t := range trackings {
		numbers[i] = map[string]string{
			"number": t.TrackingNumber,
		}
		if t.Carrier != "" {
			numbers[i]["carrier"] = t.Carrier
		}
	}

	reqBody := map[string]interface{}{
		"numbers": numbers,
	}

	// 发送请求
	resp, err := s.sendRequest(ctx, "/track/batch", reqBody)
	if err != nil {
		return nil, err
	}

	// 解析响应 (简化版)
	var result struct {
		Code int `json:"code"`
		Data struct {
			Accepted []json.RawMessage `json:"accepted"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 转换结果
	results := make([]*TrackingInfo, 0)
	for _, raw := range result.Data.Accepted {
		// 这里需要根据实际 API 响应格式解析
		// 简化处理
		var info TrackingInfo
		if err := json.Unmarshal(raw, &info); err == nil {
			results = append(results, &info)
		}
	}

	return results, nil
}

// DetectCarrier 自动识别物流公司
func (s *track17Service) DetectCarrier(ctx context.Context, trackingNumber string) ([]string, error) {
	reqBody := map[string]interface{}{
		"number": trackingNumber,
	}

	resp, err := s.sendRequest(ctx, "/carrier/detect", reqBody)
	if err != nil {
		return nil, err
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			Carriers []struct {
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"carriers"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	carriers := make([]string, len(result.Data.Carriers))
	for i, c := range result.Data.Carriers {
		carriers[i] = c.Code
	}

	return carriers, nil
}

// sendRequest 发送 HTTP 请求
func (s *track17Service) sendRequest(ctx context.Context, endpoint string, body interface{}) ([]byte, error) {
	// 序列化请求体
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建请求
	url := s.config.BaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("17token", s.config.APIKey)

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// parseLocation 解析位置字符串
func parseLocation(locationStr string) *Location {
	if locationStr == "" {
		return nil
	}
	// 简化处理，实际需要根据格式解析
	return &Location{
		City: locationStr,
	}
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *Config {
	return &Config{
		Provider: getEnv("TRACKING_PROVIDER", "17track"),
		APIKey:   os.Getenv("TRACKING_API_KEY"),
		BaseURL:  getEnv("TRACKING_BASE_URL", "https://api.17track.net/track/v2"),
		Timeout:  30 * time.Second,
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// MockTrackingService 模拟追踪服务 (用于测试)
type MockTrackingService struct{}

func NewMockTrackingService() TrackingService {
	return &MockTrackingService{}
}

func (s *MockTrackingService) Track(ctx context.Context, trackingNumber, carrier string) (*TrackingInfo, error) {
	return &TrackingInfo{
		TrackingNumber: trackingNumber,
		Carrier:        carrier,
		Status:         "In Transit",
		StatusCode:     2,
		Events: []TrackingEvent{
			{
				Time:        time.Now().Add(-48 * time.Hour),
				Status:      "Picked Up",
				Description: "Package picked up by carrier",
				Location:    &Location{City: "Shanghai", Country: "China"},
			},
			{
				Time:        time.Now().Add(-24 * time.Hour),
				Status:      "In Transit",
				Description: "Package in transit",
				Location:    &Location{City: "Hong Kong", Country: "China"},
			},
		},
		UpdatedAt: time.Now(),
	}, nil
}

func (s *MockTrackingService) BatchTrack(ctx context.Context, trackings []TrackingRequest) ([]*TrackingInfo, error) {
	results := make([]*TrackingInfo, len(trackings))
	for i, t := range trackings {
		info, _ := s.Track(ctx, t.TrackingNumber, t.Carrier)
		results[i] = info
	}
	return results, nil
}

func (s *MockTrackingService) DetectCarrier(ctx context.Context, trackingNumber string) ([]string, error) {
	return []string{"ups", "fedex", "dhl"}, nil
}

// validateTrackingNumber 验证物流单号
func validateTrackingNumber(trackingNumber string) error {
	if trackingNumber == "" {
		return fmt.Errorf("tracking number cannot be empty")
	}

	// 移除空格
	trackingNumber = strings.TrimSpace(trackingNumber)

	// 检查长度 (一般物流单号在 8-40 个字符之间)
	if len(trackingNumber) < 8 || len(trackingNumber) > 40 {
		return fmt.Errorf("invalid tracking number length: must be between 8 and 40 characters")
	}

	// 检查是否只包含字母数字和连字符
	validChars := regexp.MustCompile(`^[A-Za-z0-9\-]+$`)
	if !validChars.MatchString(trackingNumber) {
		return fmt.Errorf("invalid tracking number format: only alphanumeric characters and hyphens allowed")
	}

	return nil
}

// validateBatchTrackingRequest 验证批量追踪请求
func validateBatchTrackingRequest(trackings []TrackingRequest) error {
	if len(trackings) == 0 {
		return fmt.Errorf("no tracking requests provided")
	}

	if len(trackings) > 100 {
		return fmt.Errorf("too many tracking requests: maximum 100 allowed")
	}

	for i, req := range trackings {
		if err := validateTrackingNumber(req.TrackingNumber); err != nil {
			return fmt.Errorf("invalid tracking request at index %d: %w", i, err)
		}
	}

	return nil
}
