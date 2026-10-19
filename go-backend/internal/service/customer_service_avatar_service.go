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

	"github.com/google/uuid"

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
	siteURL  string
}

func NewCustomerServiceAvatarService(
	userRepo *repository.UserRepository,
	storageSvc storage.StorageService,
	outboxRepo *repository.OutboxRepository,
	siteURLs ...string,
) *CustomerServiceAvatarService {
	service := &CustomerServiceAvatarService{
		userRepo: userRepo,
		storage:  storageSvc,
		outbox:   outboxRepo,
	}
	if len(siteURLs) > 0 {
		service.siteURL = strings.TrimSpace(siteURLs[0])
	}
	if userRepo != nil {
		userRepo.ConfigureTransactionalCustomerServiceAvatarBindingValidation(func(ctx context.Context, reference string) error {
			key, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(service.storage, CustomerServiceAvatarStoragePrefix, reference)
			if err != nil {
				return err
			}
			return (&MediaService{storage: service.storage}).ValidateMediaObjectExistsBeforeRegisteringOwnership(ctx, key)
		})
	}
	return service
}

func (s *CustomerServiceAvatarService) Upload(ctx context.Context, userID uint, file *multipart.FileHeader) (string, error) {
	if s == nil || s.storage == nil || s.outbox == nil {
		return "", ErrCustomerServiceAvatarStorageUnavailable
	}
	if err := upload.ValidateSpecFile(file, string(upload.SpecCustomerServiceAvatar)); err != nil {
		return "", err
	}
	profileID, err := s.requireProfile(ctx, userID)
	if err != nil {
		return "", err
	}

	return uploadDedicatedImmediateOwnerObjectWithWriteAheadCleanupReceiptAndTransactionalBinding(ctx, s.outbox, s.storage, CustomerServiceAvatarStoragePrefix, profileID, file, func(tx *gorm.DB, avatarURL, _ string) error {
		transactionalService := *s
		transactionalService.userRepo = s.userRepo.WithTx(tx)
		err := transactionalService.replaceAvatarReference(tx.Statement.Context, userID, avatarURL)
		return err
	})
}

func (s *CustomerServiceAvatarService) Remove(ctx context.Context, userID uint) error {
	if s == nil || s.storage == nil {
		return ErrCustomerServiceAvatarStorageUnavailable
	}

	err := s.replaceAvatarReference(ctx, userID, "")
	if err != nil {
		return err
	}
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
	normalized, valid := storage.NormalizeObjectKey(key)
	if !valid {
		return false, nil
	}
	return s.userRepo.WithContext(ctx).HasLiveCustomerServiceAgentAvatarReferenceForExactObject(normalized, s.storage.GetURL(normalized), &MediaService{storage: s.storage, siteURL: s.siteURL})
}

func (s *CustomerServiceAvatarService) replaceAvatarReference(ctx context.Context, userID uint, avatarURL string) error {
	if s == nil || s.userRepo == nil || s.outbox == nil {
		return ErrCustomerServiceAvatarStorageUnavailable
	}

	_, err := s.userRepo.WithContext(ctx).ReplaceCustomerServiceAgentAvatar(
		userID,
		avatarURL,
		func(tx *gorm.DB, profileID uint, previousAvatarURL string) error {
			if !s.isManagedAvatarURL(previousAvatarURL) {
				return nil
			}
			_, err := s.enqueueCleanup(s.outbox.WithTx(tx), profileID, previousAvatarURL)
			return err
		},
	)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrCustomerServiceAvatarProfileNotFound
	}
	return err
}

func (s *CustomerServiceAvatarService) requireProfile(ctx context.Context, userID uint) (uint, error) {
	if s == nil || s.userRepo == nil || userID == 0 {
		return 0, ErrCustomerServiceAvatarProfileNotFound
	}
	profile, err := s.userRepo.WithContext(ctx).FindCustomerServiceAgentProfileByUserID(userID)
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
	_, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(s.storage, CustomerServiceAvatarStoragePrefix, avatarURL)
	return err == nil
}

func (s *CustomerServiceAvatarService) enqueueCleanup(repo *repository.OutboxRepository, profileID uint, avatarURL string) (*outbox.Event, error) {
	if repo == nil || profileID == 0 || !s.isManagedAvatarURL(avatarURL) {
		return nil, ErrObjectStorageCleanupUnavailable
	}

	payload, err := json.Marshal(outbox.CustomerServiceAvatarCleanupPayload{URL: strings.TrimSpace(avatarURL)})
	if err != nil {
		return nil, fmt.Errorf("encode customer-service avatar cleanup payload: %w", err)
	}
	now := time.Now().UTC()
	event := &outbox.Event{
		EventKey:      customerServiceAvatarCleanupEventKey(profileID, avatarURL),
		EventType:     outbox.EventTypeCustomerServiceAvatarCleanup,
		AggregateType: outbox.AggregateTypeCustomerServiceAgentProfile,
		AggregateID:   strconv.FormatUint(uint64(profileID), 10),
		Payload:       datatypes.JSON(payload),
		MaxAttempts:   customerServiceAvatarCleanupMaxAttempts,
		AvailableAt:   now,
	}
	return event, repo.CreateEvent(event)
}

func (s *CustomerServiceAvatarService) persistFailedAvatarUploadCompensationBeforeOwnerCheckedDeletion(ctx context.Context, profileID uint, reference string) error {
	if _, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(s.storage, CustomerServiceAvatarStoragePrefix, reference); err != nil {
		return fmt.Errorf("%w: %w", ErrObjectStorageCleanupUnavailable, err)
	}
	if s.outbox == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	cleanupContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	event, err := s.enqueueCleanup(s.outbox.WithContext(cleanupContext), profileID, reference)
	if err != nil {
		return fmt.Errorf("%w: persist customer-service avatar compensation: %w", ErrObjectStorageCleanupUnavailable, err)
	}
	_ = NewCustomerServiceAvatarCleanupHandler(s.userRepo, s.storage, s.siteURL).Handle(cleanupContext, *event)
	return nil
}

func IsCustomerServiceAvatarStorageKey(value string) bool {
	key, ok := storage.NormalizeObjectKey(value)
	if !ok {
		return false
	}
	return key == CustomerServiceAvatarStoragePrefix || strings.HasPrefix(key, CustomerServiceAvatarStoragePrefix+"/")
}

func customerServiceAvatarCleanupEventKey(profileID uint, avatarURL string) string {
	// A completed event for an earlier same-key detachment must never swallow
	// a later failed upload or detachment. This task's retries keep its identity.
	digest := sha256.Sum256([]byte(strings.TrimSpace(avatarURL) + "\x00" + uuid.NewString()))
	return fmt.Sprintf("%s:%d:%x", outbox.EventTypeCustomerServiceAvatarCleanup, profileID, digest[:])
}
