package service

import (
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/repository"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type AdminPublicChatAgentService struct {
	userRepo *repository.UserRepository
}

var (
	ErrPublicChatAgentUserRequired    = errors.New("public chat agent user_id is required")
	ErrPublicChatAgentUserNotFound    = errors.New("public chat agent user not found")
	ErrPublicChatAgentUserInvalid     = errors.New("public chat agent user must be active admin, manager or support")
	ErrPublicChatAgentIDInvalid       = errors.New("public chat agent_id must be 50 characters or fewer")
	ErrPublicChatAgentIDTaken         = errors.New("public chat agent_id is already used")
	ErrPublicChatAgentStatusInvalid   = errors.New("public chat agent status must be active or inactive")
	ErrPublicChatAgentOnlineInvalid   = errors.New("public chat agent online_status must be online, busy, away or offline")
	ErrPublicChatAgentContactRequired = errors.New("public chat agent email or whatsapp is required for active profiles")
	ErrPublicChatAgentGroupInvalid    = errors.New("public chat agent group is invalid")
	ErrPublicChatGroupNameRequired    = errors.New("public chat group name is required")
	ErrPublicChatGroupCodeInvalid     = errors.New("public chat group code is invalid")
	ErrPublicChatGroupCodeTaken       = errors.New("public chat group code is already used")
	ErrPublicChatGroupStatusInvalid   = errors.New("public chat group status must be active or inactive")
	ErrPublicChatGroupNotFound        = errors.New("public chat group not found")
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
	ID             uint                        `json:"id"`
	AgentID        string                      `json:"agent_id"`
	UserID         *uint                       `json:"user_id"`
	Username       string                      `json:"username"`
	Email          string                      `json:"email"`
	DisplayName    string                      `json:"display_name"`
	RawRole        string                      `json:"raw_role"`
	NormalizedRole auth.Role                   `json:"normalized_role"`
	UserStatus     string                      `json:"user_status"`
	ProfileStatus  string                      `json:"profile_status"`
	OnlineStatus   string                      `json:"online_status"`
	Avatar         string                      `json:"avatar"`
	WhatsApp       string                      `json:"whatsapp"`
	GroupIDs       []uint                      `json:"group_ids"`
	Groups         []AdminPublicChatAgentGroup `json:"groups"`
	Exposed        bool                        `json:"exposed"`
}

type AdminPublicChatAgentGroup struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	SortOrder   int    `json:"sort_order"`
}

type AdminPublicChatAgentCandidate struct {
	UserID              uint                        `json:"user_id"`
	Username            string                      `json:"username"`
	Email               string                      `json:"email"`
	DisplayName         string                      `json:"display_name"`
	RawRole             string                      `json:"raw_role"`
	NormalizedRole      auth.Role                   `json:"normalized_role"`
	UserStatus          string                      `json:"user_status"`
	HasProfile          bool                        `json:"has_profile"`
	ProfileID           *uint                       `json:"profile_id"`
	AgentID             string                      `json:"agent_id"`
	ProfileName         string                      `json:"profile_name"`
	ProfileEmail        string                      `json:"profile_email"`
	ProfileAvatar       string                      `json:"profile_avatar"`
	ProfileWhatsApp     string                      `json:"profile_whatsapp"`
	ProfileStatus       string                      `json:"profile_status"`
	ProfileOnlineStatus string                      `json:"profile_online_status"`
	ProfileGroupIDs     []uint                      `json:"profile_group_ids"`
	ProfileGroups       []AdminPublicChatAgentGroup `json:"profile_groups"`
}

type AdminPublicChatAgentUpsertInput struct {
	UserID       uint   `json:"user_id"`
	AgentID      string `json:"agent_id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	WhatsApp     string `json:"whatsapp"`
	Status       string `json:"status"`
	OnlineStatus string `json:"online_status"`
	GroupIDs     []uint `json:"group_ids"`
}

type AdminPublicChatGroupUpsertInput struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	SortOrder   int    `json:"sort_order"`
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
	for _, item := range items {
		if !item.Exposed {
			continue
		}
		name := strings.TrimSpace(item.DisplayName)
		if name == "" {
			name = fmt.Sprintf("Profile #%d", item.ID)
		}
		if strings.TrimSpace(item.Email) == "" && strings.TrimSpace(item.WhatsApp) == "" {
			warnings = append(warnings, fmt.Sprintf("%s 未填写邮箱或 WhatsApp，前台将无法显示联系方式", name))
		}
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

func (s *AdminPublicChatAgentService) ListPublicChatGroups(limit int) ([]AdminPublicChatAgentGroup, error) {
	groups, err := s.userRepo.FindCustomerServiceAgentGroups(limit, true)
	if err != nil {
		return nil, err
	}
	items := make([]AdminPublicChatAgentGroup, 0, len(groups))
	for _, group := range groups {
		items = append(items, makeAdminPublicChatAgentGroup(group))
	}
	return items, nil
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
			item.ProfileOnlineStatus = profile.OnlineStatus
			item.ProfileGroupIDs = adminPublicChatAgentGroupIDs(profile.Groups)
			item.ProfileGroups = makeAdminPublicChatAgentGroups(profile.Groups)
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
	whatsApp := strings.TrimSpace(input.WhatsApp)
	if status == "active" {
		// 公开客服只要求能被联系到一次即可，展示标签与分组不参与这个判断。
		if !hasPublicChatAgentContact(email, whatsApp) {
			return nil, false, ErrPublicChatAgentContactRequired
		}
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
	profile.WhatsApp = whatsApp
	profile.Status = status
	profile.OnlineStatus = onlineStatus

	if created {
		if err := s.userRepo.CreateCustomerServiceAgentProfile(profile); err != nil {
			return nil, false, err
		}
	} else if err := s.userRepo.UpdateCustomerServiceAgentProfile(profile); err != nil {
		return nil, false, err
	}

	if err := s.userRepo.ReplaceCustomerServiceAgentProfileGroups(profile.ID, input.GroupIDs); err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, false, ErrPublicChatAgentGroupInvalid
		}
		return nil, false, err
	}

	savedProfile, err := s.userRepo.FindCustomerServiceAgentProfileByUserIDWithGroups(input.UserID)
	if err != nil {
		return nil, false, err
	}
	item := makeAdminPublicChatAgent(*savedProfile)
	return &item, created, nil
}

func (s *AdminPublicChatAgentService) UpsertPublicChatGroup(input AdminPublicChatGroupUpsertInput) (*AdminPublicChatAgentGroup, bool, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, false, ErrPublicChatGroupNameRequired
	}

	code := normalizePublicChatGroupCode(input.Code)
	if code == "" {
		code = normalizePublicChatGroupCode(name)
	}
	if !validPublicChatGroupCode(code) {
		return nil, false, ErrPublicChatGroupCodeInvalid
	}

	status, err := normalizePublicChatGroupStatus(input.Status)
	if err != nil {
		return nil, false, err
	}

	var group *user.AgentGroup
	created := input.ID == 0
	if created {
		group = &user.AgentGroup{}
	} else {
		group, err = s.userRepo.FindCustomerServiceAgentGroupByID(input.ID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return nil, false, ErrPublicChatGroupNotFound
			}
			return nil, false, err
		}
	}

	conflict, err := s.userRepo.FindCustomerServiceAgentGroupByCode(code)
	if err != nil && !repository.IsRecordNotFound(err) {
		return nil, false, err
	}
	if err == nil && conflict.ID != group.ID {
		return nil, false, ErrPublicChatGroupCodeTaken
	}

	group.Code = code
	group.Name = name
	group.Description = strings.TrimSpace(input.Description)
	group.Status = status
	group.SortOrder = input.SortOrder

	if created {
		if err := s.userRepo.CreateCustomerServiceAgentGroup(group); err != nil {
			return nil, false, err
		}
	} else if err := s.userRepo.UpdateCustomerServiceAgentGroup(group); err != nil {
		return nil, false, err
	}

	item := makeAdminPublicChatAgentGroup(*group)
	return &item, created, nil
}

func (s *AdminPublicChatAgentService) DeletePublicChatGroup(id uint) error {
	if id == 0 {
		return ErrPublicChatGroupNotFound
	}
	if _, err := s.userRepo.FindCustomerServiceAgentGroupByID(id); err != nil {
		if repository.IsRecordNotFound(err) {
			return ErrPublicChatGroupNotFound
		}
		return err
	}
	return s.userRepo.DeleteCustomerServiceAgentGroup(id)
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
		GroupIDs:       adminPublicChatAgentGroupIDs(agent.Groups),
		Groups:         makeAdminPublicChatAgentGroups(agent.Groups),
		Exposed:        isPublicChatAgentExposed(agent),
	}
}

func makeAdminPublicChatAgentGroups(groups []user.AgentGroup) []AdminPublicChatAgentGroup {
	items := make([]AdminPublicChatAgentGroup, 0, len(groups))
	for _, group := range groups {
		items = append(items, makeAdminPublicChatAgentGroup(group))
	}
	return items
}

func makeAdminPublicChatAgentGroup(group user.AgentGroup) AdminPublicChatAgentGroup {
	return AdminPublicChatAgentGroup{
		ID:          group.ID,
		Code:        group.Code,
		Name:        group.Name,
		Description: group.Description,
		Status:      group.Status,
		SortOrder:   group.SortOrder,
	}
}

func adminPublicChatAgentGroupIDs(groups []user.AgentGroup) []uint {
	ids := make([]uint, 0, len(groups))
	for _, group := range groups {
		if group.ID > 0 {
			ids = append(ids, group.ID)
		}
	}
	return ids
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

func normalizePublicChatGroupStatus(status string) (string, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = "active"
	}
	switch status {
	case "active", "inactive":
		return status, nil
	default:
		return "", ErrPublicChatGroupStatusInvalid
	}
}

func normalizePublicChatGroupCode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}

func hasPublicChatAgentContact(email, whatsApp string) bool {
	return strings.TrimSpace(email) != "" || strings.TrimSpace(whatsApp) != ""
}

func validPublicChatGroupCode(value string) bool {
	if value == "" || len([]rune(value)) > 50 {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}
	return true
}
