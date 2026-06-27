package service

import (
	"errors"
	"fmt"
	"tanzanite/internal/domain/auth"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailExists      = errors.New("email already exists")
	ErrUsernameExists   = errors.New("username already exists")
	ErrSelfDelete       = errors.New("cannot delete yourself")
	ErrSelfStatusChange = errors.New("cannot modify your own status")
	ErrSelfRoleChange   = errors.New("cannot modify your own role")
	ErrRoleForbidden    = errors.New("insufficient privileges for requested role change")
)

type UserCreateInput struct {
	Email     string
	Username  string
	Password  string
	FirstName string
	LastName  string
	Role      string
	Locale    string
	Status    string
}

type UserUpdateInput struct {
	Email     string
	Username  string
	Password  string
	FirstName string
	LastName  string
	Role      string
	Locale    string
	Status    string
}

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) ListUsers(page, pageSize int, role, status, search string) ([]user.User, int64, error) {
	return s.userRepo.FindAllWithFilters(page, pageSize, role, status, search)
}

func (s *UserService) GetUser(id uint) (*user.User, error) {
	u, err := s.userRepo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return u, err
}

func (s *UserService) CreateUser(input UserCreateInput, actorRole string) (*user.User, error) {
	role := auth.NormalizeRole(input.Role)
	if err := ensureActorCanAssignRole(actorRole, role); err != nil {
		return nil, err
	}

	if err := s.ensureEmailAvailable(input.Email, 0); err != nil {
		return nil, err
	}
	if err := s.ensureUsernameAvailable(input.Username, 0); err != nil {
		return nil, err
	}

	newUser := &user.User{
		Email:     input.Email,
		Username:  input.Username,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Role:      string(role),
		Locale:    input.Locale,
		Status:    input.Status,
	}

	if err := newUser.HashPassword(input.Password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.Create(newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *UserService) UpdateUser(id, actorID uint, actorRole string, input UserUpdateInput) (*user.User, error) {
	existingUser, err := s.GetUser(id)
	if err != nil {
		return nil, err
	}

	if err := ensureActorCanManageTarget(actorRole, existingUser.Role); err != nil {
		return nil, err
	}

	if input.Role != "" {
		nextRole := auth.NormalizeRole(input.Role)
		if actorID == id && nextRole != auth.NormalizeRole(existingUser.Role) {
			return nil, ErrSelfRoleChange
		}
		if err := ensureActorCanAssignRole(actorRole, nextRole); err != nil {
			return nil, err
		}
		existingUser.Role = string(nextRole)
	}

	if input.Status != "" {
		if actorID == id && input.Status != existingUser.Status {
			return nil, ErrSelfStatusChange
		}
		existingUser.Status = input.Status
	}

	if input.Email != "" && input.Email != existingUser.Email {
		if err := s.ensureEmailAvailable(input.Email, existingUser.ID); err != nil {
			return nil, err
		}
		existingUser.Email = input.Email
	}

	if input.Username != "" && input.Username != existingUser.Username {
		if err := s.ensureUsernameAvailable(input.Username, existingUser.ID); err != nil {
			return nil, err
		}
		existingUser.Username = input.Username
	}

	if input.Password != "" {
		if err := existingUser.HashPassword(input.Password); err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
	}

	if input.FirstName != "" {
		existingUser.FirstName = input.FirstName
	}
	if input.LastName != "" {
		existingUser.LastName = input.LastName
	}
	if input.Locale != "" {
		existingUser.Locale = input.Locale
	}

	if err := s.userRepo.Update(existingUser); err != nil {
		return nil, err
	}

	return existingUser, nil
}

func (s *UserService) DeleteUser(id, actorID uint) error {
	if actorID == id {
		return ErrSelfDelete
	}
	return s.userRepo.Delete(id)
}

func (s *UserService) UpdateUserStatus(id, actorID uint, actorRole, status string) error {
	if actorID == id {
		return ErrSelfStatusChange
	}

	existingUser, err := s.GetUser(id)
	if err != nil {
		return err
	}
	if err := ensureActorCanManageTarget(actorRole, existingUser.Role); err != nil {
		return err
	}

	return s.userRepo.UpdateStatus(id, status)
}

func (s *UserService) GetUserStats() (map[string]interface{}, error) {
	return s.userRepo.GetStats()
}

func (s *UserService) BatchDeleteUsers(ids []uint, actorID uint) (int, error) {
	for _, id := range ids {
		if id == actorID {
			return 0, ErrSelfDelete
		}
	}

	deleted := 0
	for _, id := range ids {
		if err := s.userRepo.Delete(id); err == nil {
			deleted++
		}
	}

	return deleted, nil
}

func (s *UserService) ensureEmailAvailable(email string, currentUserID uint) error {
	existingUser, err := s.userRepo.FindByEmail(email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if existingUser != nil && existingUser.ID != currentUserID {
		return ErrEmailExists
	}
	return nil
}

func (s *UserService) ensureUsernameAvailable(username string, currentUserID uint) error {
	existingUser, err := s.userRepo.FindByUsername(username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if existingUser != nil && existingUser.ID != currentUserID {
		return ErrUsernameExists
	}
	return nil
}

func ensureActorCanAssignRole(actorRole string, targetRole auth.Role) error {
	if targetRole == auth.RoleAdmin && auth.NormalizeRole(actorRole) != auth.RoleAdmin {
		return ErrRoleForbidden
	}
	return nil
}

func ensureActorCanManageTarget(actorRole, targetRole string) error {
	if auth.NormalizeRole(targetRole) == auth.RoleAdmin && auth.NormalizeRole(actorRole) != auth.RoleAdmin {
		return ErrRoleForbidden
	}
	return nil
}
