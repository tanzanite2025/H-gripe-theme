package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	CustomerServiceAvatarStoragePrefix = "customer-service/avatars"
	CustomerServiceAvatarCacheControl  = "public, max-age=31536000, immutable"

	// Cleanup is deliberately retried for a very long time. These objects are
	// single-owner, low-frequency resources and must not accumulate when a
	// transient object-store failure occurs.
	customerServiceAvatarCleanupMaxAttempts = 1_000_000
)

var (
	ErrCustomerServiceAvatarProfileNotFound    = errors.New("customer-service profile not found")
	ErrCustomerServiceAvatarStorageUnavailable = errors.New("customer-service avatar storage is not configured")
)

type CustomerServiceAvatarService struct {
	userRepo *repository.UserRepository
	storage  storage.StorageService
	outbox   *repository.OutboxRepository
}

func NewCustomerServiceAvatarService(
	userRepo *repository.UserRepository,
	storageSvc storage.StorageService,
	outboxRepo *repository.OutboxRepository,
) *CustomerServiceAvatarService {
	return &CustomerServiceAvatarService{
		userRepo: userRepo,
		storage:  storageSvc,
		outbox:   outboxRepo,
	}
}

func (s *CustomerServiceAvatarService) Upload(ctx context.Context, userID uint, file *multipart.FileHeader) (string, error) {
	if s == nil || s.storage == nil {
		return "", ErrCustomerServiceAvatarStorageUnavailable
	}
	if err := upload.ValidateSpecFile(file, string(upload.SpecCustomerServiceAvatar)); err != nil {
		return "", err
	}
	profileID, err := s.requireProfile(userID)
	if err != nil {
		return "", err
	}

	avatarURL, err := s.uploadAvatarObject(ctx, file)
	if err != nil {
		return "", fmt.Errorf("upload customer-service avatar: %w", err)
	}

	previousAvatarURL, err := s.replaceAvatarReference(userID, avatarURL)
	if err != nil {
		if cleanupErr := s.cleanupUnattachedAvatar(ctx, profileID, avatarURL); cleanupErr != nil {
			return "", errors.Join(err, cleanupErr)
		}
		return "", err
	}

	if previousAvatarURL != avatarURL {
		s.deleteManagedAvatarNow(ctx, previousAvatarURL)
	}
	return avatarURL, nil
}

func (s *CustomerServiceAvatarService) uploadAvatarObject(ctx context.Context, file *multipart.FileHeader) (string, error) {
	if cacheControlled, ok := s.storage.(storage.CacheControlledObjectUploader); ok {
		return cacheControlled.UploadWithPrefixAndCacheControl(
			ctx,
			file,
			CustomerServiceAvatarStoragePrefix,
			CustomerServiceAvatarCacheControl,
		)
	}
	return s.storage.UploadWithPrefix(ctx, file, CustomerServiceAvatarStoragePrefix)
}

func (s *CustomerServiceAvatarService) Remove(ctx context.Context, userID uint) error {
	if s == nil || s.storage == nil {
		return ErrCustomerServiceAvatarStorageUnavailable
	}

	previousAvatarURL, err := s.replaceAvatarReference(userID, "")
	if err != nil {
		return err
	}
	s.deleteManagedAvatarNow(ctx, previousAvatarURL)
	return nil
}

// CanServePublicAvatar only permits the current, dedicated avatar object.
// It is used by local upload serving; hosted storage still needs the same
// namespace policy at its bucket/CDN boundary.
func (s *CustomerServiceAvatarService) CanServePublicAvatar(ctx context.Context, key string) (bool, error) {
	if s == nil || s.userRepo == nil || s.storage == nil {
		return false, nil
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return false, err
		}
	}
	if !IsCustomerServiceAvatarStorageKey(key) {
		return false, nil
	}
	return s.userRepo.IsCurrentCustomerServiceAgentAvatarURL(s.storage.GetURL(key))
}

func (s *CustomerServiceAvatarService) replaceAvatarReference(userID uint, avatarURL string) (string, error) {
	if s == nil || s.userRepo == nil || s.outbox == nil {
		return "", ErrCustomerServiceAvatarStorageUnavailable
	}

	previousAvatarURL, err := s.userRepo.ReplaceCustomerServiceAgentAvatar(
		userID,
		avatarURL,
		func(tx *gorm.DB, profileID uint, previousAvatarURL string) error {
			if !s.isManagedAvatarURL(previousAvatarURL) {
				return nil
			}
			return s.enqueueCleanup(s.outbox.WithTx(tx), profileID, previousAvatarURL)
		},
	)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrCustomerServiceAvatarProfileNotFound
	}
	return previousAvatarURL, err
}

func (s *CustomerServiceAvatarService) cleanupUnattachedAvatar(ctx context.Context, profileID uint, avatarURL string) error {
	if !s.isManagedAvatarURL(avatarURL) {
		return nil
	}
	if err := s.storage.Delete(ctx, avatarURL); err == nil {
		return nil
	} else if s.outbox == nil {
		return fmt.Errorf("delete unattached customer-service avatar: %w", err)
	}

	if err := s.enqueueCleanup(s.outbox, profileID, avatarURL); err != nil {
		return fmt.Errorf("schedule unattached customer-service avatar cleanup: %w", err)
	}
	return nil
}

func (s *CustomerServiceAvatarService) deleteManagedAvatarNow(ctx context.Context, avatarURL string) {
	if !s.isManagedAvatarURL(avatarURL) {
		return
	}
	// The transactional outbox event already exists. A failed immediate delete
	// therefore remains durable work for the dispatcher without affecting the
	// successful avatar update response.
	_ = s.storage.Delete(ctx, avatarURL)
}

func (s *CustomerServiceAvatarService) requireProfile(userID uint) (uint, error) {
	if s == nil || s.userRepo == nil || userID == 0 {
		return 0, ErrCustomerServiceAvatarProfileNotFound
	}
	profile, err := s.userRepo.FindCustomerServiceAgentProfileByUserID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, ErrCustomerServiceAvatarProfileNotFound
	}
	if err != nil {
		return 0, err
	}
	return profile.ID, nil
}

func (s *CustomerServiceAvatarService) isManagedAvatarURL(avatarURL string) bool {
	if s == nil || s.storage == nil || strings.TrimSpace(avatarURL) == "" {
		return false
	}
	key, err := s.storage.ObjectKey(avatarURL)
	return err == nil && IsCustomerServiceAvatarStorageKey(key)
}

func (s *CustomerServiceAvatarService) enqueueCleanup(repo *repository.OutboxRepository, profileID uint, avatarURL string) error {
	if repo == nil || profileID == 0 || !s.isManagedAvatarURL(avatarURL) {
		return nil
	}

	payload, err := json.Marshal(outbox.CustomerServiceAvatarCleanupPayload{URL: strings.TrimSpace(avatarURL)})
	if err != nil {
		return fmt.Errorf("encode customer-service avatar cleanup payload: %w", err)
	}
	now := time.Now().UTC()
	return repo.CreateEvent(&outbox.Event{
		EventKey:      customerServiceAvatarCleanupEventKey(profileID, avatarURL),
		EventType:     outbox.EventTypeCustomerServiceAvatarCleanup,
		AggregateType: outbox.AggregateTypeCustomerServiceAgentProfile,
		AggregateID:   strconv.FormatUint(uint64(profileID), 10),
		Payload:       datatypes.JSON(payload),
		MaxAttempts:   customerServiceAvatarCleanupMaxAttempts,
		AvailableAt:   now,
	})
}

func IsCustomerServiceAvatarStorageKey(value string) bool {
	key, ok := storage.NormalizeObjectKey(value)
	if !ok {
		return false
	}
	return key == CustomerServiceAvatarStoragePrefix || strings.HasPrefix(key, CustomerServiceAvatarStoragePrefix+"/")
}

func customerServiceAvatarCleanupEventKey(profileID uint, avatarURL string) string {
	digest := sha256.Sum256([]byte(strings.TrimSpace(avatarURL)))
	return fmt.Sprintf("%s:%d:%x", outbox.EventTypeCustomerServiceAvatarCleanup, profileID, digest[:])
}
