package service

import (
	"tanzanite/internal/domain/auth"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/repository"
)

type AdminPublicChatAgentService struct {
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

func NewAdminPublicChatAgentService(userRepo *repository.UserRepository) *AdminPublicChatAgentService {
	return &AdminPublicChatAgentService{
		userRepo: userRepo,
	}
}

func (s *AdminPublicChatAgentService) ListPublicChatAgents(limit int) (*AdminPublicChatAgentsOverview, error) {
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
