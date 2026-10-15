package media

import "time"

const MediaUploadSessionCustomerServiceContext = "customer_service"

// A receipt lives for 24 hours, but its original/derivative writes share a
// bounded deadline. Cleanup must wait for both deadlines before retiring it.
const CustomerServiceAttachmentUploadMaximumWriteDuration = 5 * time.Minute

// MediaUploadSession tracks an incomplete upload before a message owns it.
// ObjectKey is allocated before writing bytes so failed uploads remain traceable.
type MediaUploadSession struct {
	ID                            string               `gorm:"type:varchar(36);primaryKey" json:"-"`
	BusinessContextType           string               `gorm:"type:varchar(40);not null;index:idx_media_upload_session_business_context,priority:1" json:"-"`
	BusinessContextID             uint                 `gorm:"not null;index:idx_media_upload_session_business_context,priority:2" json:"-"`
	UploadActorType               MediaUploadActorType `gorm:"type:varchar(32);not null" json:"-"`
	UploadActorAccountID          uint                 `gorm:"not null;default:0" json:"-"`
	UploadActorVisitorSessionHash string               `gorm:"type:varchar(64);not null;default:''" json:"-"`
	ObjectKey                     string               `gorm:"not null;uniqueIndex" json:"-"`
	MediaAssetID                  *uint                `gorm:"uniqueIndex" json:"-"`
	Status                        string               `gorm:"type:varchar(20);not null;index" json:"-"`
	ExpiresAt                     time.Time            `gorm:"not null;index" json:"-"`
	CreatedAt                     time.Time            `json:"-"`
	UpdatedAt                     time.Time            `json:"-"`
}

func (MediaUploadSession) TableName() string { return "media_upload_sessions" }

func (s MediaUploadSession) UploadActorIdentity() MediaUploadActorIdentity {
	return MediaUploadActorIdentity{ActorType: s.UploadActorType, AccountID: s.UploadActorAccountID, VisitorSessionHash: s.UploadActorVisitorSessionHash}
}

func (s MediaUploadSession) CustomerServiceAttachmentUploadWriteDeadline() time.Time {
	deadline := s.CreatedAt.Add(CustomerServiceAttachmentUploadMaximumWriteDuration)
	if s.ExpiresAt.Before(deadline) {
		return s.ExpiresAt
	}
	return deadline
}
