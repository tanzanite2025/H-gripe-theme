package service

import (
	"errors"
	"fmt"
	"strings"
	"tanzanite/internal/domain/auth"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/repository"

	"gorm.io/gorm"
)

type AdminPublicChatAgentService struct {
	userRepo *repository.UserRepository
}

var (
	ErrPublicChatAgentUserRequired  = errors.New("public chat agent user_id is required")
	ErrPublicChatAgentUserNotFound  = errors.New("public chat agent user not found")
	ErrPublicChatAgentUserInvalid   = errors.New("public chat agent user must be active admin, manager or support")
	ErrPublicChatAgentIDInvalid     = errors.New("public chat agent_id must be 50 characters or fewer")
	ErrPublicChatAgentIDTaken       = errors.New("public chat agent_id is already used")
	ErrPublicChatAgentStatusInvalid = errors.New("public chat agent status must be active or inactive")
	ErrPublicChatAgentOnlineInvalid = errors.New("public chat agent online_status must be online, busy, away or offline")
)

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

type AdminPublicChatAgentCandidate struct {
	UserID          uint      `json:"user_id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	DisplayName     string    `json:"display_name"`
	RawRole         string    `json:"raw_role"`
	NormalizedRole  auth.Role `json:"normalized_role"`
	UserStatus      string    `json:"user_status"`
	HasProfile      bool      `json:"has_profile"`
	ProfileID       *uint     `json:"profile_id"`
	AgentID         string    `json:"agent_id"`
	ProfileName     string    `json:"profile_name"`
	ProfileEmail    string    `json:"profile_email"`
	ProfileAvatar   string    `json:"profile_avatar"`
	ProfileWhatsApp string    `json:"profile_whatsapp"`
	ProfileStatus   string    `json:"profile_status"`
}

type AdminPublicChatAgentUpsertInput struct {
	UserID       uint   `json:"user_id"`
	AgentID      string `json:"agent_id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Avatar       string `json:"avatar"`
	WhatsApp     string `json:"whatsapp"`
	Status       string `json:"status"`
	OnlineStatus string `json:"online_status"`
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

		item := makeAdminPublicChatAgent(agent)
		item.Exposed = exposed
		items = append(items, item)
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

func (s *AdminPublicChatAgentService) ListPublicChatAgentCandidates(limit int) ([]AdminPublicChatAgentCandidate, error) {
	users, err := s.userRepo.FindCustomerServiceAgents(limit)
	if err != nil {
		return nil, err
	}

	profiles, err := s.userRepo.FindAllCustomerServiceAgentProfiles(500)
	if err != nil {
		return nil, err
	}

	profileByUserID := make(map[uint]user.AgentProfile, len(profiles))
	for _, profile := range profiles {
		if profile.UserID == nil {
			continue
		}
		if _, exists := profileByUserID[*profile.UserID]; !exists {
			profileByUserID[*profile.UserID] = profile
		}
	}

	candidates := make([]AdminPublicChatAgentCandidate, 0, len(users))
	for _, candidateUser := range users {
		profile, hasProfile := profileByUserID[candidateUser.ID]
		item := AdminPublicChatAgentCandidate{
			UserID:         candidateUser.ID,
			Username:       candidateUser.Username,
			Email:          strings.TrimSpace(candidateUser.Email),
			DisplayName:    displayNameFromAdminUser(candidateUser),
			RawRole:        candidateUser.Role,
			NormalizedRole: auth.NormalizeRole(candidateUser.Role),
			UserStatus:     candidateUser.Status,
			HasProfile:     hasProfile,
		}
		if hasProfile {
			item.ProfileID = uintPointer(profile.ID)
			item.AgentID = profile.AgentID
			item.ProfileName = profile.DisplayName()
			item.ProfileEmail = profile.PublicEmail()
			item.ProfileAvatar = profile.Avatar
			item.ProfileWhatsApp = profile.WhatsApp
			item.ProfileStatus = profile.Status
		}
		candidates = append(candidates, item)
	}

	return candidates, nil
}

func (s *AdminPublicChatAgentService) UpsertPublicChatAgentProfile(input AdminPublicChatAgentUpsertInput) (*AdminPublicChatAgent, bool, error) {
	if input.UserID == 0 {
		return nil, false, ErrPublicChatAgentUserRequired
	}

	agentUser, err := s.userRepo.FindByID(input.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, ErrPublicChatAgentUserNotFound
		}
		return nil, false, err
	}
	if strings.TrimSpace(agentUser.Status) != "active" || !auth.IsCustomerServiceAgentRole(agentUser.Role) {
		return nil, false, ErrPublicChatAgentUserInvalid
	}

	existingProfile, err := s.userRepo.FindCustomerServiceAgentProfileByUserID(input.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		existingProfile = nil
	}

	agentID := strings.TrimSpace(input.AgentID)
	if agentID == "" {
		agentID = fmt.Sprintf("user-%d", input.UserID)
	}
	if len([]rune(agentID)) > 50 {
		return nil, false, ErrPublicChatAgentIDInvalid
	}

	conflictingProfile, err := s.userRepo.FindCustomerServiceAgentProfileByAgentID(agentID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	if err == nil && (existingProfile == nil || conflictingProfile.ID != existingProfile.ID) {
		return nil, false, ErrPublicChatAgentIDTaken
	}

	statusInput := input.Status
	if strings.TrimSpace(statusInput) == "" && existingProfile != nil {
		statusInput = existingProfile.Status
	}
	status, err := normalizePublicChatAgentProfileStatus(statusInput)
	if err != nil {
		return nil, false, err
	}

	onlineInput := input.OnlineStatus
	if strings.TrimSpace(onlineInput) == "" && existingProfile != nil {
		onlineInput = existingProfile.OnlineStatus
	}
	onlineStatus, err := normalizePublicChatAgentOnlineStatus(onlineInput)
	if err != nil {
		return nil, false, err
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = displayNameFromAdminUser(*agentUser)
	}
	email := strings.TrimSpace(input.Email)
	if email == "" {
		email = strings.TrimSpace(agentUser.Email)
	}

	created := existingProfile == nil
	profile := existingProfile
	if profile == nil {
		userID := agentUser.ID
		profile = &user.AgentProfile{UserID: &userID}
	}

	profile.AgentID = agentID
	profile.UserID = &agentUser.ID
	profile.Name = name
	profile.Email = email
	profile.Avatar = strings.TrimSpace(input.Avatar)
	profile.WhatsApp = strings.TrimSpace(input.WhatsApp)
	profile.Status = status
	profile.OnlineStatus = onlineStatus

	if created {
		if err := s.userRepo.CreateCustomerServiceAgentProfile(profile); err != nil {
			return nil, false, err
		}
	} else if err := s.userRepo.UpdateCustomerServiceAgentProfile(profile); err != nil {
		return nil, false, err
	}

	profile.User = agentUser
	item := makeAdminPublicChatAgent(*profile)
	return &item, created, nil
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

func makeAdminPublicChatAgent(agent user.AgentProfile) AdminPublicChatAgent {
	return AdminPublicChatAgent{
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
		Exposed:        isPublicChatAgentExposed(agent),
	}
}

func displayNameFromAdminUser(item user.User) string {
	fullName := strings.TrimSpace(strings.TrimSpace(item.FirstName) + " " + strings.TrimSpace(item.LastName))
	if fullName != "" {
		return fullName
	}
	if strings.TrimSpace(item.Username) != "" {
		return strings.TrimSpace(item.Username)
	}
	return strings.TrimSpace(item.Email)
}

func uintPointer(value uint) *uint {
	return &value
}

func normalizePublicChatAgentProfileStatus(status string) (string, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = "active"
	}
	switch status {
	case "active", "inactive":
		return status, nil
	default:
		return "", ErrPublicChatAgentStatusInvalid
	}
}

func normalizePublicChatAgentOnlineStatus(status string) (string, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = "offline"
	}
	switch status {
	case "online", "busy", "away", "offline":
		return status, nil
	default:
		return "", ErrPublicChatAgentOnlineInvalid
	}
}
