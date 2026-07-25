package setting

import "time"

// UpdateSettingRequest 更新设置请求
type UpdateSettingRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Group       string `json:"group"`
	Locale      string `json:"locale"`
	IsPublic    bool   `json:"is_public"`
	Description string `json:"description"`
}

// BatchUpdateSettingsRequest 批量更新设置请求
type BatchUpdateSettingsRequest struct {
	Settings []UpdateSettingRequest `json:"settings" binding:"required"`
}

// SettingResponse 设置响应（用于管理后台）
type SettingResponse struct {
	ID          uint      `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Type        string    `json:"type"`
	Group       string    `json:"group"`
	Locale      string    `json:"locale"`
	IsPublic    bool      `json:"is_public"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
