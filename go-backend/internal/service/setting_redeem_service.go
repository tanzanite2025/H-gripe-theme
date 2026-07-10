package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"tanzanite/internal/domain/setting"
)

// GetRedeemSettings 获取积分兑换配置
func (s *SettingService) GetRedeemSettings(locale string) (*setting.RedeemSettings, error) {
	cacheKey := settingsStructuredCacheKey("redeem", locale)

	// 尝试从缓存获取
	var redeemSettings setting.RedeemSettings
	if err := s.cache.Get(cacheKey, &redeemSettings); err == nil {
		return &redeemSettings, nil
	}

	// 从数据库获取
	settings, err := s.settingRepo.GetByGroup("redeem", locale)
	if err != nil {
		return nil, err
	}

	// 转换为 map 便于查询与校验
	m := make(map[string]string)
	for _, st := range settings {
		m[st.Key] = st.Value
	}

	// 强制校验必要配置项，实现 Fail Loudly 原则
	requiredKeys := []string{
		"tz_redeem_enabled",
		"tz_redeem_exchange_rate",
		"tz_redeem_min_points",
		"tz_redeem_max_value_per_day",
		"tz_redeem_card_expiry_days",
		"tz_redeem_preset_values",
	}

	for _, k := range requiredKeys {
		if _, ok := m[k]; !ok {
			errMsg := fmt.Sprintf("[CRITICAL] Redeem setting '%s' is missing in settings table", k)
			return nil, errors.New(errMsg)
		}
	}

	// 解析数据并映射至结构体
	redeemSettings.Enabled = m["tz_redeem_enabled"] == "1" || m["tz_redeem_enabled"] == "true"

	exchangeRate, err := strconv.Atoi(m["tz_redeem_exchange_rate"])
	if err != nil {
		return nil, fmt.Errorf("[CRITICAL] Invalid format for tz_redeem_exchange_rate: %v", err)
	}
	redeemSettings.ExchangeRate = exchangeRate

	minPoints, err := strconv.Atoi(m["tz_redeem_min_points"])
	if err != nil {
		return nil, fmt.Errorf("[CRITICAL] Invalid format for tz_redeem_min_points: %v", err)
	}
	redeemSettings.MinPoints = minPoints

	maxValue, err := strconv.ParseFloat(m["tz_redeem_max_value_per_day"], 64)
	if err != nil {
		return nil, fmt.Errorf("[CRITICAL] Invalid format for tz_redeem_max_value_per_day: %v", err)
	}
	redeemSettings.MaxValuePerDay = maxValue

	expiryDays, err := strconv.Atoi(m["tz_redeem_card_expiry_days"])
	if err != nil {
		return nil, fmt.Errorf("[CRITICAL] Invalid format for tz_redeem_card_expiry_days: %v", err)
	}
	redeemSettings.CardExpiryDays = expiryDays

	// 解析面值预设列表
	presetStr := m["tz_redeem_preset_values"]
	parts := strings.Split(presetStr, ",")
	presets := make([]float64, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		val, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return nil, fmt.Errorf("[CRITICAL] Invalid format for tz_redeem_preset_values item '%s': %v", p, err)
		}
		presets = append(presets, val)
	}
	redeemSettings.PresetValues = presets

	// 写入缓存
	_ = s.cache.Set(cacheKey, &redeemSettings, s.cacheTTL)

	return &redeemSettings, nil
}
