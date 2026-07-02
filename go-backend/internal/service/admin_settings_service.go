package service

import (
	"tanzanite/internal/domain/auth"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/repository"
)

type AdminSettingsService struct {
	settings *SettingService
	userRepo *repository.UserRepository
}

type AdminPublicChatAgentsOverview struct {
	Summary  AdminPublicChatAgentsSummary `json:"summary"`
	Agents   []AdminPublicChatAgent       `json:"agents"`
	Warnings []string                     `json:"warnings"`
}

type AdminPublicChatAgentsSummary struct {
	ProfileCount  int `json:"profile_count"`
	ExposedAgents int `json:"exposed_agents"`
}

type AdminPublicChatAgent struct {
	ID             uint      `json:"id"`
	AgentID        string    `json:"agent_id"`
	UserID         *uint     `json:"user_id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	DisplayName    string    `json:"display_name"`
	RawRole        string    `json:"raw_role"`
	NormalizedRole auth.Role `json:"normalized_role"`
	UserStatus     string    `json:"user_status"`
	ProfileStatus  string    `json:"profile_status"`
	OnlineStatus   string    `json:"online_status"`
	Avatar         string    `json:"avatar"`
	WhatsApp       string    `json:"whatsapp"`
	Exposed        bool      `json:"exposed"`
}

func NewAdminSettingsService(settings *SettingService, userRepo *repository.UserRepository) *AdminSettingsService {
	return &AdminSettingsService{
		settings: settings,
		userRepo: userRepo,
	}
}

func (s *AdminSettingsService) ListSettings(locale, group string) ([]setting.Setting, error) {
	if group != "" {
		return s.settings.GetByGroup(group, locale)
	}
	return s.settings.GetAll(locale)
}

func (s *AdminSettingsService) GetSetting(key, locale string) (*setting.Setting, error) {
	return s.settings.Get(key, locale)
}

func (s *AdminSettingsService) UpdateSetting(req setting.UpdateSettingRequest) (*setting.Setting, error) {
	st := normalizeSettingRequest(req)
	if err := s.settings.BatchSet([]setting.Setting{st}); err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *AdminSettingsService) BatchUpdateSettings(req setting.BatchUpdateSettingsRequest) (int, error) {
	settings := make([]setting.Setting, 0, len(req.Settings))
	for _, item := range req.Settings {
		settings = append(settings, normalizeSettingRequest(item))
	}

	if err := s.settings.BatchSet(settings); err != nil {
		return 0, err
	}

	return len(settings), nil
}

func (s *AdminSettingsService) DeleteSetting(key, locale string) error {
	return s.settings.Delete(key, locale)
}

func (s *AdminSettingsService) GetGroups() ([]string, error) {
	return s.settings.GetGroups()
}

func (s *AdminSettingsService) GetByGroup(group, locale string) ([]setting.Setting, error) {
	return s.settings.GetByGroup(group, locale)
}

func (s *AdminSettingsService) ListPublicChatAgents(limit int) (*AdminPublicChatAgentsOverview, error) {
	agents, err := s.userRepo.FindAllCustomerServiceAgentProfiles(limit)
	if err != nil {
		return nil, err
	}

	items := make([]AdminPublicChatAgent, 0, len(agents))
	exposedAgents := 0
	for _, agent := range agents {
		exposed := isPublicChatAgentExposed(agent)
		if exposed {
			exposedAgents++
		}

		items = append(items, AdminPublicChatAgent{
			ID:             agent.ID,
			AgentID:        agent.AgentID,
			UserID:         copyUserID(agent.UserID),
			Username:       usernameFromAgentProfile(agent),
			Email:          agent.PublicEmail(),
			DisplayName:    agent.DisplayName(),
			RawRole:        rawRoleFromAgentProfile(agent),
			NormalizedRole: normalizedRoleFromAgentProfile(agent),
			UserStatus:     userStatusFromAgentProfile(agent),
			ProfileStatus:  agent.Status,
			OnlineStatus:   agent.OnlineStatus,
			Avatar:         agent.Avatar,
			WhatsApp:       agent.WhatsApp,
			Exposed:        exposed,
		})
	}

	warnings := []string{}
	if len(agents) == 0 {
		warnings = append(warnings, "当前未配置 public chat 客服 profile")
	}

	return &AdminPublicChatAgentsOverview{
		Summary: AdminPublicChatAgentsSummary{
			ProfileCount:  len(agents),
			ExposedAgents: exposedAgents,
		},
		Agents:   items,
		Warnings: warnings,
	}, nil
}

func normalizeSettingRequest(req setting.UpdateSettingRequest) setting.Setting {
	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	settingType := req.Type
	if settingType == "" {
		settingType = "string"
	}

	return setting.Setting{
		Key:         req.Key,
		Value:       req.Value,
		Type:        settingType,
		Group:       req.Group,
		Locale:      locale,
		IsPublic:    req.IsPublic,
		Description: req.Description,
	}
}

func isPublicChatAgentExposed(agent user.AgentProfile) bool {
	return agent.UserID != nil &&
		agent.Status == "active" &&
		agent.User != nil &&
		agent.User.Status == "active" &&
		auth.IsCustomerServiceAgentRole(agent.User.Role)
}

func copyUserID(userID *uint) *uint {
	if userID == nil {
		return nil
	}
	value := *userID
	return &value
}

func usernameFromAgentProfile(agent user.AgentProfile) string {
	if agent.User == nil {
		return ""
	}
	return agent.User.Username
}

func rawRoleFromAgentProfile(agent user.AgentProfile) string {
	if agent.User == nil {
		return ""
	}
	return agent.User.Role
}

func normalizedRoleFromAgentProfile(agent user.AgentProfile) auth.Role {
	if agent.User == nil {
		return auth.RoleUser
	}
	return auth.NormalizeRole(agent.User.Role)
}

func userStatusFromAgentProfile(agent user.AgentProfile) string {
	if agent.User == nil {
		return ""
	}
	return agent.User.Status
}
