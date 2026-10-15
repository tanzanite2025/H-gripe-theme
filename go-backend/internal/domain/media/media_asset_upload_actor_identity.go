package media

import (
	"encoding/hex"
	"strings"
)

type MediaUploadActorType string

const (
	MediaUploadActorLegacyAccount MediaUploadActorType = "legacy_account"
	MediaUploadActorAdministrator MediaUploadActorType = "administrator"
	MediaUploadActorCustomer      MediaUploadActorType = "customer"
	MediaUploadActorVisitor       MediaUploadActorType = "visitor"
)

// MediaUploadActorIdentity is supplied by trusted application code, never by
// multipart form fields. AccountID uses the existing media uploader_id column.
// VisitorSessionHash is the verified session digest, never the cookie itself.
type MediaUploadActorIdentity struct {
	ActorType          MediaUploadActorType
	AccountID          uint
	VisitorSessionHash string
}

func (identity MediaUploadActorIdentity) IsValidMediaUploadActorIdentity() bool {
	switch identity.ActorType {
	case MediaUploadActorLegacyAccount, MediaUploadActorAdministrator, MediaUploadActorCustomer:
		return identity.AccountID > 0 && identity.VisitorSessionHash == ""
	case MediaUploadActorVisitor:
		if identity.AccountID != 0 || len(identity.VisitorSessionHash) != 64 ||
			strings.ToLower(identity.VisitorSessionHash) != identity.VisitorSessionHash {
			return false
		}
		_, err := hex.DecodeString(identity.VisitorSessionHash)
		return err == nil
	default:
		return false
	}
}
