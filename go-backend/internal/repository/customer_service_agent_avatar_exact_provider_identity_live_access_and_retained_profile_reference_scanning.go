package repository

import (
	"fmt"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/user"
)

// Retention includes disabled/deleted profiles; public access is separately
// limited to live profile rows. All comparisons use complete object identity.
func (r *MediaRepository) customerServiceAgentAvatarReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("customer_service_agent_profiles") {
		return nil, nil
	}
	var references []media.AssetReference
	for afterID := uint(0); ; {
		var profiles []user.AgentProfile
		scan := r.db.Select("id", "avatar", "deleted_at").Where("id > ?", afterID).Order("id ASC").Limit(256)
		if query.IncludeRetainedDeletedPlacements {
			scan = scan.Unscoped()
		}
		if err := scan.Find(&profiles).Error; err != nil {
			return nil, err
		}
		for _, profile := range profiles {
			if !r.matchesExactRegisteredOrTrustedMediaReference(query, profile.Avatar) {
				continue
			}
			reference := newMediaReference(media.ReferenceCategorySupport, "customer_service_agent_avatar", profile.ID, 0, fmt.Sprintf("客服头像 #%d", profile.ID), "avatar")
			reference.UsageState = "active"
			if profile.DeletedAt.Valid {
				reference.UsageState = "retained_deleted_parent"
			}
			references = append(references, reference)
		}
		if len(profiles) < 256 {
			return references, nil
		}
		afterID = profiles[len(profiles)-1].ID
	}
}

func (r *MediaRepository) HasLiveCustomerServiceAgentAvatarReferenceForExactObject(key, reference string) (bool, error) {
	references, err := r.customerServiceAgentAvatarReferences(mediaAssetReferenceQuery{ObjectKeys: []string{key}, URLs: []string{reference, "/uploads/" + key, "uploads/" + key}})
	return len(references) > 0, err
}

func (r *UserRepository) HasLiveCustomerServiceAgentAvatarReferenceForExactObject(key, reference string, resolver PublicBusinessMediaReferenceObjectKeyResolver) (bool, error) {
	repo := NewMediaRepository(r.db)
	repo.ConfigurePublicBusinessMediaReferenceIdentityResolver(resolver)
	return repo.HasLiveCustomerServiceAgentAvatarReferenceForExactObject(key, reference)
}
